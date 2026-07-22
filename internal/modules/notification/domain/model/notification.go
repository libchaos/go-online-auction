package model

import (
	"time"

	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/domain/errs"
)

type NotificationModel struct {
	id               uint64
	userID           uint64
	category         enum.NotificationCategoryEnum
	notificationType enum.NotificationTypeEnum
	title            string
	body             string
	payload          []byte
	channels         []enum.NotificationChannelEnum
	idempotencyKey   string
	readAt           *time.Time
	createdAt        time.Time
}

func NewNotification(
	userID uint64,
	notificationType enum.NotificationTypeEnum,
	title string,
	body string,
	payload []byte,
	channels []enum.NotificationChannelEnum,
	idempotencyKey string,
) (NotificationModel, error) {
	if userID == 0 {
		return NotificationModel{}, errs.ErrNotificationUserRequired
	}

	if title == "" {
		return NotificationModel{}, errs.ErrNotificationTitleRequired
	}

	if body == "" {
		return NotificationModel{}, errs.ErrNotificationBodyRequired
	}

	if len(channels) == 0 {
		return NotificationModel{}, errs.ErrNotificationChannelsEmpty
	}

	normalizedPayload := payload
	if normalizedPayload == nil {
		normalizedPayload = []byte("{}")
	}

	return NotificationModel{
		userID:           userID,
		category:         notificationType.Category(),
		notificationType: notificationType,
		title:            title,
		body:             body,
		payload:          normalizedPayload,
		channels:         channels,
		idempotencyKey:   idempotencyKey,
		createdAt:        time.Now().UTC(),
	}, nil
}

func RestoreNotificationModel(
	id uint64,
	userID uint64,
	category string,
	notificationType string,
	title string,
	body string,
	payload []byte,
	channels []string,
	idempotencyKey string,
	readAt *time.Time,
	createdAt time.Time,
) (NotificationModel, error) {
	parsedCategory, err := enum.NewNotificationCategoryEnum(category)
	if err != nil {
		return NotificationModel{}, err
	}

	parsedType, err := enum.NewNotificationTypeEnum(notificationType)
	if err != nil {
		return NotificationModel{}, err
	}

	parsedChannels := make([]enum.NotificationChannelEnum, 0, len(channels))
	for _, channel := range channels {
		parsedChannel, channelErr := enum.NewNotificationChannelEnum(channel)
		if channelErr != nil {
			return NotificationModel{}, channelErr
		}

		parsedChannels = append(parsedChannels, parsedChannel)
	}

	return NotificationModel{
		id:               id,
		userID:           userID,
		category:         parsedCategory,
		notificationType: parsedType,
		title:            title,
		body:             body,
		payload:          payload,
		channels:         parsedChannels,
		idempotencyKey:   idempotencyKey,
		readAt:           readAt,
		createdAt:        createdAt,
	}, nil
}

func (notification *NotificationModel) MarkRead() error {
	if notification.readAt != nil {
		return errs.ErrNotificationAlreadyRead
	}

	now := time.Now().UTC()
	notification.readAt = &now

	return nil
}

func (notification *NotificationModel) IsRead() bool {
	return notification.readAt != nil
}

func (notification *NotificationModel) ID() uint64 {
	return notification.id
}

func (notification *NotificationModel) UserID() uint64 {
	return notification.userID
}

func (notification *NotificationModel) Category() enum.NotificationCategoryEnum {
	return notification.category
}

func (notification *NotificationModel) Type() enum.NotificationTypeEnum {
	return notification.notificationType
}

func (notification *NotificationModel) Title() string {
	return notification.title
}

func (notification *NotificationModel) Body() string {
	return notification.body
}

func (notification *NotificationModel) Payload() []byte {
	return notification.payload
}

func (notification *NotificationModel) Channels() []enum.NotificationChannelEnum {
	return notification.channels
}

func (notification *NotificationModel) IdempotencyKey() string {
	return notification.idempotencyKey
}

func (notification *NotificationModel) ReadAt() *time.Time {
	return notification.readAt
}

func (notification *NotificationModel) CreatedAt() time.Time {
	return notification.createdAt
}
