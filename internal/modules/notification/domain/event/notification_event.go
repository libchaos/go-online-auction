package event

import (
	"time"

	"github.com/google/uuid"
)

const (
	NotificationCreatedEventType = "notification_created"
)

type NotificationDomainEvent struct {
	eventID   string
	timestamp time.Time
}

func newNotificationDomainEvent() NotificationDomainEvent {
	return NotificationDomainEvent{
		eventID:   uuid.New().String(),
		timestamp: time.Now().UTC(),
	}
}

func (domainEvent NotificationDomainEvent) EventID() string {
	return domainEvent.eventID
}

func (domainEvent NotificationDomainEvent) Timestamp() time.Time {
	return domainEvent.timestamp
}

type NotificationCreatedEvent struct {
	NotificationDomainEvent
	notificationID   uint64
	userID           uint64
	category         string
	notificationType string
	title            string
	body             string
	payload          []byte
	createdAt        time.Time
}

func NewNotificationCreatedEvent(
	notificationID uint64,
	userID uint64,
	category string,
	notificationType string,
	title string,
	body string,
	payload []byte,
	createdAt time.Time,
) NotificationCreatedEvent {
	return NotificationCreatedEvent{
		NotificationDomainEvent: newNotificationDomainEvent(),
		notificationID:          notificationID,
		userID:                  userID,
		category:                category,
		notificationType:        notificationType,
		title:                   title,
		body:                    body,
		payload:                 payload,
		createdAt:               createdAt,
	}
}

func (domainEvent NotificationCreatedEvent) NotificationID() uint64 {
	return domainEvent.notificationID
}

func (domainEvent NotificationCreatedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent NotificationCreatedEvent) Category() string {
	return domainEvent.category
}

func (domainEvent NotificationCreatedEvent) Type() string {
	return domainEvent.notificationType
}

func (domainEvent NotificationCreatedEvent) Title() string {
	return domainEvent.title
}

func (domainEvent NotificationCreatedEvent) Body() string {
	return domainEvent.body
}

func (domainEvent NotificationCreatedEvent) Payload() []byte {
	return domainEvent.payload
}

func (domainEvent NotificationCreatedEvent) CreatedAt() time.Time {
	return domainEvent.createdAt
}
