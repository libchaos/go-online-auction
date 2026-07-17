package messaging

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

// EventPublisher publishes a serialized domain event payload to a JetStream subject.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

type JetStreamEventPublisher struct {
	js jetstream.JetStream
}

func NewJetStreamEventPublisher(js jetstream.JetStream) *JetStreamEventPublisher {
	return &JetStreamEventPublisher{js: js}
}

var _ EventPublisher = (*JetStreamEventPublisher)(nil)

func (p *JetStreamEventPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := p.js.Publish(ctx, subject, data)
	if err != nil {
		return err
	}
	eventsPublishedTotal.Inc()
	return nil
}
