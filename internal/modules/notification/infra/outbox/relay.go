package outbox

import (
	"context"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
)

const (
	defaultInterval  = 500 * time.Millisecond
	defaultBatchSize = 200
)

// Config controls the outbox relay polling behavior.
type Config struct {
	Interval  time.Duration
	BatchSize int
}

// Relay drains the notification module's transactional outbox (notification_outbox)
// and publishes pending events to JetStream. Delivery is at-least-once: an event
// is only marked published after JetStream acknowledges it, and redeliveries are
// deduplicated by the stream's duplicate window keyed on Nats-Msg-Id (the domain
// event ID). Multiple relay instances are safe: MarkPublished is guarded by
// `published_at IS NULL`, so duplicates are bounded and deduplicated anyway.
type Relay struct {
	outboxRepository ports.NotificationOutboxRepository
	js               jetstream.JetStream
	logger           logger.Logger
	interval         time.Duration
	batchSize        int

	stopOnce sync.Once
	done     chan struct{}
	stopped  chan struct{}
}

func NewRelay(
	outboxRepository ports.NotificationOutboxRepository,
	js jetstream.JetStream,
	logger logger.Logger,
	cfg Config,
) *Relay {
	interval := cfg.Interval
	if interval <= 0 {
		interval = defaultInterval
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	return &Relay{
		outboxRepository: outboxRepository,
		js:               js,
		logger:           logger,
		interval:         interval,
		batchSize:        batchSize,
		done:             make(chan struct{}),
		stopped:          make(chan struct{}),
	}
}

// Start launches the polling loop in a background goroutine.
func (relay *Relay) Start(ctx context.Context) {
	go relay.run(ctx)
}

// Stop signals the polling loop to exit and waits for it to finish.
func (relay *Relay) Stop() {
	relay.stopOnce.Do(func() {
		close(relay.done)
	})
	<-relay.stopped
}

func (relay *Relay) run(ctx context.Context) {
	defer close(relay.stopped)

	ticker := time.NewTicker(relay.interval)
	defer ticker.Stop()

	relay.logger.Info().
		Dur("interval", relay.interval).
		Int("batch_size", relay.batchSize).
		Msg("notification outbox relay started")

	for {
		select {
		case <-ctx.Done():
			relay.logger.Info().Msg("notification outbox relay stopped: context cancelled")
			return
		case <-relay.done:
			relay.logger.Info().Msg("notification outbox relay stopped")
			return
		case <-ticker.C:
			relay.Drain(ctx)
		}
	}
}

// Drain publishes one batch of pending outbox events. It keeps draining until
// the outbox has fewer pending events than the batch size, so bursts are
// flushed without waiting for the next tick.
func (relay *Relay) Drain(ctx context.Context) {
	for {
		published, pending := relay.drainBatch(ctx)
		if pending < relay.batchSize || published == 0 {
			return
		}
	}
}

func (relay *Relay) drainBatch(ctx context.Context) (int, int) {
	events, err := relay.outboxRepository.ListUnpublished(ctx, relay.batchSize)
	if err != nil {
		relay.logger.Error().Err(err).Msg("notification outbox relay: failed to list unpublished events")
		return 0, 0
	}

	published := 0

	for _, evt := range events {
		if ctx.Err() != nil {
			return published, len(events)
		}

		_, pubErr := relay.js.Publish(ctx, evt.Subject, evt.Payload, jetstream.WithMsgID(evt.EventID))
		if pubErr != nil {
			relay.logger.Error().Err(pubErr).
				Str("event_id", evt.EventID).
				Str("subject", evt.Subject).
				Msg("notification outbox relay: failed to publish event; will retry")
			return published, len(events)
		}

		if _, markErr := relay.outboxRepository.MarkPublished(ctx, evt.ID); markErr != nil {
			relay.logger.Error().Err(markErr).
				Str("event_id", evt.EventID).
				Msg("notification outbox relay: failed to mark event as published; duplicate delivery possible")
			return published, len(events)
		}

		published++
		notificationEventsRelayedTotal.Inc()
	}

	return published, len(events)
}
