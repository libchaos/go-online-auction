package messaging

import (
	"context"
	"fmt"

	sharednats "auction/internal/shared/modules/nats"

	"github.com/nats-io/nats.go/jetstream"
)

// EventConsumer subscribes to the auction events subject and invokes handler for
// each delivered message until the context is cancelled.
type EventConsumer interface {
	Consume(ctx context.Context, handler func(subject string, data []byte)) error
}

type JetStreamEventConsumer struct {
	js jetstream.JetStream
}

func NewJetStreamEventConsumer(js jetstream.JetStream) *JetStreamEventConsumer {
	return &JetStreamEventConsumer{js: js}
}

var _ EventConsumer = (*JetStreamEventConsumer)(nil)

func (c *JetStreamEventConsumer) Consume(ctx context.Context, handler func(subject string, data []byte)) error {
	// Each WebSocket node uses its own ephemeral consumer so it receives the full
	// event stream for local fan-out, mirroring the previous per-node subscription.
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, sharednats.StreamEvents, jetstream.ConsumerConfig{
		FilterSubject: sharednats.SubjectEvents,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckNonePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create events consumer: %w", err)
	}

	consumeContext, err := consumer.Consume(func(msg jetstream.Msg) {
		handler(msg.Subject(), msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming events: %w", err)
	}

	<-ctx.Done()
	consumeContext.Stop()

	return nil
}
