package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/modules/auction/infra/event/envelope"
	sharednats "auction/internal/shared/modules/nats"
)

const replayBatchTimeout = 2 * time.Second

// errEndOfStream signals that the replay consumed all messages for the subject.
var errEndOfStream = errors.New("end of stream")

// ReplayFilter narrows an event replay to a time window. Zero values mean unbounded.
type ReplayFilter struct {
	// From replays events with timestamps >= From (temporal query lower bound)
	From time.Time
	// Until stops the replay at events with timestamps > Until (upper bound)
	Until time.Time
	// Limit caps the number of returned events; 0 means no cap
	Limit int
}

// EventReplayer reads historical events back from the event store.
type EventReplayer interface {
	// ReplayAuction returns the persisted event history of one auction in
	// publication order, optionally narrowed to a time window.
	ReplayAuction(ctx context.Context, auctionID uint64, filter ReplayFilter) ([]envelope.Envelope, error)
}

// JetStreamEventReplayer replays events from the AUCTION_EVENTS stream, which
// retains every published event (the event store). It uses ephemeral ordered
// consumers so replays never interfere with live consumers.
type JetStreamEventReplayer struct {
	js jetstream.JetStream
}

func NewJetStreamEventReplayer(js jetstream.JetStream) *JetStreamEventReplayer {
	return &JetStreamEventReplayer{js: js}
}

var _ EventReplayer = (*JetStreamEventReplayer)(nil)

func (r *JetStreamEventReplayer) ReplayAuction(
	ctx context.Context,
	auctionID uint64,
	filter ReplayFilter,
) ([]envelope.Envelope, error) {
	cfg := jetstream.OrderedConsumerConfig{
		FilterSubjects: []string{envelope.BuildSubject(auctionID)},
		DeliverPolicy:  jetstream.DeliverAllPolicy,
	}
	// Temporal query: start the replay at the first event at or after From
	if !filter.From.IsZero() {
		startTime := filter.From
		cfg.DeliverPolicy = jetstream.DeliverByStartTimePolicy
		cfg.OptStartTime = &startTime
	}

	consumer, err := r.js.OrderedConsumer(ctx, sharednats.StreamEvents, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create replay consumer: %w", err)
	}

	var events []envelope.Envelope
	for {
		if filter.Limit > 0 && len(events) >= filter.Limit {
			return events, nil
		}

		msg, fetchErr := r.nextMessage(ctx, consumer)
		if errors.Is(fetchErr, errEndOfStream) {
			return events, nil
		}
		if fetchErr != nil {
			return nil, fetchErr
		}

		env, decodeErr := envelope.Decode(msg.Data())
		if decodeErr != nil {
			return nil, decodeErr
		}

		// Temporal query upper bound: events are appended in timestamp order per
		// auction, so the first event past Until ends the replay.
		if !filter.Until.IsZero() && env.Timestamp.After(filter.Until) {
			return events, nil
		}

		events = append(events, env)
	}
}

// nextMessage fetches a single message, returning errEndOfStream when the
// stream has no more messages for the filtered subject.
func (r *JetStreamEventReplayer) nextMessage(
	ctx context.Context,
	consumer jetstream.Consumer,
) (jetstream.Msg, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	batch, err := consumer.Fetch(1, jetstream.FetchMaxWait(replayBatchTimeout))
	if err != nil {
		if errors.Is(err, jetstream.ErrNoMessages) {
			return nil, errEndOfStream
		}
		return nil, fmt.Errorf("failed to fetch replay message: %w", err)
	}

	for msg := range batch.Messages() {
		return msg, nil
	}
	if batch.Error() != nil && !errors.Is(batch.Error(), jetstream.ErrNoMessages) {
		return nil, fmt.Errorf("replay fetch failed: %w", batch.Error())
	}

	return nil, errEndOfStream
}
