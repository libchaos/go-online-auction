package command

import (
	"context"
	"fmt"
	"time"

	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/infra/event/envelope"
	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
)

type CreateEmailRequestCommandInput struct {
	SourceEventID string
	UserID        uint64
	To            string
	Subject       string
	Title         string
	Body          string
}

type CreateEmailRequestCommand struct {
	outbox ports.NotificationOutboxRepository
	logger logger.Logger
}

func NewCreateEmailRequestCommand(
	outbox ports.NotificationOutboxRepository,
	logger logger.Logger,
) *CreateEmailRequestCommand {
	return &CreateEmailRequestCommand{outbox: outbox, logger: logger}
}

// Execute enqueues an email by writing an idempotent request row to the
// transactional outbox. The deterministic event id (source_event_id:user_id:email)
// combined with the outbox's UNIQUE(event_id) + ON CONFLICT DO NOTHING clause
// guarantees a redelivered source event never enqueues a duplicate email. The
// shared relay later publishes it to notification.evt.email.requested, where the
// EmailDispatchConsumer picks it up and calls the EmailPort.
func (command *CreateEmailRequestCommand) Execute(
	ctx context.Context,
	input CreateEmailRequestCommandInput,
) error {
	eventID := fmt.Sprintf("%s:%d:%s", input.SourceEventID, input.UserID, enum.EnumNotificationChannelEmail)

	outboxEvent, buildErr := envelope.BuildEmailRequestedOutboxEvent(
		eventID,
		input.To,
		input.Subject,
		input.Title,
		input.Body,
		time.Now().UTC(),
	)
	if buildErr != nil {
		return buildErr
	}

	if saveErr := command.outbox.SaveEmailRequest(ctx, outboxEvent); saveErr != nil {
		return saveErr
	}

	command.logger.Info().
		Uint64("user_id", input.UserID).
		Str("to", input.To).
		Str("subject", input.Subject).
		Msg("email request enqueued")

	return nil
}
