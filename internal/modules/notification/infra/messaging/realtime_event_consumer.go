package messaging

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	sharednats "auction/internal/shared/modules/nats"
)

// EventConsumer streams notification.evt.{userID} events out of JetStream so the
// realtime hub can fan each one out to the recipient's connected SSE clients.
type EventConsumer interface {
	Consume(ctx context.Context, handler func(subject string, data []byte)) error
}

type JetStreamRealtimeEventConsumer struct {
	js jetstream.JetStream
}

func NewJetStreamRealtimeEventConsumer(js jetstream.JetStream) *JetStreamRealtimeEventConsumer {
	return &JetStreamRealtimeEventConsumer{js: js}
}

var _ EventConsumer = (*JetStreamRealtimeEventConsumer)(nil)

// Consume delivers only events published after the client connects (DeliverNew)
// with no acknowledgement, since a missed live push is recoverable from the
// notification-center REST endpoints on reconnect.
func (consumer *JetStreamRealtimeEventConsumer) Consume(
	ctx context.Context,
	handler func(subject string, data []byte),
) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(
		ctx,
		sharednats.StreamNotificationEvents,
		jetstream.ConsumerConfig{
			FilterSubject: sharednats.SubjectNotificationUserEvents,
			DeliverPolicy: jetstream.DeliverNewPolicy,
			AckPolicy:     jetstream.AckNonePolicy,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create notification realtime consumer: %w", err)
	}

	consumeContext, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		handler(msg.Subject(), msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming notification realtime events: %w", err)
	}

	<-ctx.Done()
	consumeContext.Stop()

	return nil
}
