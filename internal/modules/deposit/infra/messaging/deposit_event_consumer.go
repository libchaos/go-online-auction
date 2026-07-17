package messaging

import (
	"context"
	"fmt"

	sharednats "auction/internal/shared/modules/nats"

	"github.com/nats-io/nats.go/jetstream"
)

type EventConsumer interface {
	Consume(ctx context.Context, handler func(subject string, data []byte)) error
}

type JetStreamDepositEventConsumer struct {
	js jetstream.JetStream
}

func NewJetStreamDepositEventConsumer(js jetstream.JetStream) *JetStreamDepositEventConsumer {
	return &JetStreamDepositEventConsumer{js: js}
}

var _ EventConsumer = (*JetStreamDepositEventConsumer)(nil)

func (consumer *JetStreamDepositEventConsumer) Consume(
	ctx context.Context,
	handler func(subject string, data []byte),
) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(
		ctx,
		sharednats.StreamDepositEvents,
		jetstream.ConsumerConfig{
			FilterSubject: sharednats.SubjectDepositEvents,
			DeliverPolicy: jetstream.DeliverNewPolicy,
			AckPolicy:     jetstream.AckNonePolicy,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create deposit events consumer: %w", err)
	}

	consumeContext, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		handler(msg.Subject(), msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming deposit events: %w", err)
	}

	<-ctx.Done()
	consumeContext.Stop()

	return nil
}
