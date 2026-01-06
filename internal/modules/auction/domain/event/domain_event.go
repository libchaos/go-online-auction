package event

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent struct {
	eventID   string
	timestamp time.Time
}

func newDomainEvent() DomainEvent {
	return DomainEvent{
		eventID:   uuid.New().String(),
		timestamp: time.Now().UTC(),
	}
}

func (e DomainEvent) EventID() string {
	return e.eventID
}

func (e DomainEvent) Timestamp() time.Time {
	return e.timestamp
}
