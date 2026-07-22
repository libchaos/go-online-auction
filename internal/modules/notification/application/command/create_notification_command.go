package command

import (
	"context"

	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/domain/event"
	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/event/envelope"
	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/logger"
)

type CreateNotificationCommandInput struct {
	UserID         uint64
	Type           string
	Title          string
	Body           string
	Payload        []byte
	Channels       []string
	IdempotencyKey string
}

type CreateNotificationCommandOutput struct {
	NotificationID uint64
	Created        bool
}

type CreateNotificationCommand struct {
	uowFactory ports.NotificationUnitOfWorkFactory
	logger     logger.Logger
}

func NewCreateNotificationCommand(
	uowFactory ports.NotificationUnitOfWorkFactory,
	logger logger.Logger,
) *CreateNotificationCommand {
	return &CreateNotificationCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

// Execute persists an in-app notification and, only when the row is newly
// inserted, writes a notification-created event to the outbox in the same
// transaction. The unique idempotency key makes redelivered source events a
// no-op so a user never receives a duplicate notification or SSE push.
func (command *CreateNotificationCommand) Execute(
	ctx context.Context,
	input CreateNotificationCommandInput,
) (CreateNotificationCommandOutput, error) {
	notificationType, typeErr := enum.NewNotificationTypeEnum(input.Type)
	if typeErr != nil {
		return CreateNotificationCommandOutput{}, typeErr
	}

	channels := make([]enum.NotificationChannelEnum, 0, len(input.Channels))
	for _, channel := range input.Channels {
		parsedChannel, channelErr := enum.NewNotificationChannelEnum(channel)
		if channelErr != nil {
			return CreateNotificationCommandOutput{}, channelErr
		}

		channels = append(channels, parsedChannel)
	}

	notification, buildErr := model.NewNotification(
		input.UserID,
		notificationType,
		input.Title,
		input.Body,
		input.Payload,
		channels,
		input.IdempotencyKey,
	)
	if buildErr != nil {
		return CreateNotificationCommandOutput{}, buildErr
	}

	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return CreateNotificationCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	persisted, created, saveErr := unitOfWork.NotificationRepository().Save(ctx, notification)
	if saveErr != nil {
		return CreateNotificationCommandOutput{}, saveErr
	}

	if !created {
		return CreateNotificationCommandOutput{NotificationID: persisted.ID(), Created: false}, nil
	}

	category := persisted.Category()
	notificationTypeValue := persisted.Type()
	createdEvent := event.NewNotificationCreatedEvent(
		persisted.ID(),
		persisted.UserID(),
		category.String(),
		notificationTypeValue.String(),
		persisted.Title(),
		persisted.Body(),
		persisted.Payload(),
		persisted.CreatedAt(),
	)

	outboxEvent, envelopeErr := envelope.ToNotificationCreatedOutboxEvent(createdEvent)
	if envelopeErr != nil {
		return CreateNotificationCommandOutput{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return CreateNotificationCommandOutput{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return CreateNotificationCommandOutput{}, completeErr
	}

	command.logger.Info().
		Uint64("notification_id", persisted.ID()).
		Uint64("user_id", persisted.UserID()).
		Str("type", notificationTypeValue.String()).
		Msg("notification created")

	return CreateNotificationCommandOutput{NotificationID: persisted.ID(), Created: true}, nil
}
