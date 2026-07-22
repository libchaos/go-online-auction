package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
	sharednats "auction/internal/shared/modules/nats"
)

const (
	emailConsumerDurableName = "notification-email-dispatch"
	emailAckWait             = 30 * time.Second
	emailMaxDeliver          = 5
)

type emailRequestedPayload struct {
	EventID string `json:"event_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

// EmailDispatchConsumer subscribes to notification.evt.email.requested on the
// NOTIFICATION_EVENTS stream and turns each request into a real email via the
// EmailPort. It uses an explicit-ack durable consumer so a message is only
// acknowledged after the transport succeeds; a send failure is negatively
// acknowledged for redelivery (bounded by MaxDeliver).
type EmailDispatchConsumer struct {
	js              jetstream.JetStream
	email           ports.EmailPort
	logger          logger.Logger
	consumeContexts []jetstream.ConsumeContext
}

func NewEmailDispatchConsumer(
	js jetstream.JetStream,
	email ports.EmailPort,
	logger logger.Logger,
) *EmailDispatchConsumer {
	return &EmailDispatchConsumer{js: js, email: email, logger: logger}
}

func (consumer *EmailDispatchConsumer) Start(ctx context.Context) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(
		ctx,
		sharednats.StreamNotificationEvents,
		jetstream.ConsumerConfig{
			Durable:       emailConsumerDurableName,
			FilterSubject: sharednats.SubjectNotificationEmailRequested,
			DeliverPolicy: jetstream.DeliverAllPolicy,
			AckPolicy:     jetstream.AckExplicitPolicy,
			AckWait:       emailAckWait,
			MaxDeliver:    emailMaxDeliver,
			ReplayPolicy:  jetstream.ReplayInstantPolicy,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create email dispatch consumer: %w", err)
	}

	consumeContext, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		consumer.handle(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("failed to start email dispatch consumer: %w", err)
	}

	consumer.consumeContexts = append(consumer.consumeContexts, consumeContext)

	return nil
}

func (consumer *EmailDispatchConsumer) Stop() {
	for _, consumeContext := range consumer.consumeContexts {
		if consumeContext != nil {
			consumeContext.Drain()
		}
	}
}

func (consumer *EmailDispatchConsumer) handle(ctx context.Context, msg jetstream.Msg) {
	if sendErr := consumer.processEmailRequest(ctx, msg.Data()); sendErr != nil {
		consumer.logger.Error().Err(sendErr).
			Str("subject", msg.Subject()).
			Msg("failed to send email; redelivering")

		if nackErr := msg.Nak(); nackErr != nil {
			consumer.logger.Error().Err(nackErr).Msg("failed to nack email request")
		}

		return
	}

	if ackErr := msg.Ack(); ackErr != nil {
		consumer.logger.Error().Err(ackErr).Msg("failed to ack email after successful send")
	}
}

// processEmailRequest decodes an email request and sends it. It is the unit of
// work behind each consumed message, exposed separately so it can be tested
// without a live JetStream connection.
func (consumer *EmailDispatchConsumer) processEmailRequest(ctx context.Context, data []byte) error {
	var payload emailRequestedPayload
	if unmarshalErr := json.Unmarshal(data, &payload); unmarshalErr != nil {
		return fmt.Errorf("failed to decode email request: %w", unmarshalErr)
	}

	message := ports.EmailMessage{
		To:       payload.To,
		Subject:  payload.Subject,
		HTMLBody: renderEmailHTML(payload.Title, payload.Body),
		TextBody: payload.Body,
	}

	return consumer.email.Send(ctx, message)
}

func renderEmailHTML(title, body string) string {
	return fmt.Sprintf(
		"<!DOCTYPE html><html><head><meta charset=\"UTF-8\"></head><body>"+
			"<h2>%s</h2><p>%s</p></body></html>",
		html.EscapeString(title),
		html.EscapeString(body),
	)
}
