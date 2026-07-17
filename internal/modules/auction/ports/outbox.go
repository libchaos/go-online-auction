package ports

import (
	"context"
	"time"
)

// OutboxEvent is a serialized domain event recorded in the transactional outbox.
// It is inserted in the same database transaction as the state change that
// produced it and later relayed to the event stream by the outbox relay,
// guaranteeing at-least-once delivery; the relay uses EventID as the JetStream
// message ID so redeliveries are deduplicated server-side.
type OutboxEvent struct {
	ID            uint64 // assigned by the database, zero on save
	EventID       string
	EventType     string
	SchemaVersion int
	Subject       string
	Payload       []byte
	OccurredAt    time.Time
}

// OutboxRepository persists and drains the transactional outbox.
type OutboxRepository interface {
	// Save records an event in the outbox within the ambient transaction
	Save(ctx context.Context, event OutboxEvent) error
	// ListUnpublished returns pending events in insertion order
	ListUnpublished(ctx context.Context, limit int) ([]OutboxEvent, error)
	// MarkPublished flags an event as relayed; returns false when it was already marked
	MarkPublished(ctx context.Context, id uint64) (bool, error)
}
