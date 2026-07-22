package ports

import (
	"context"
	"time"
)

// OutboxEvent is a single row in the payment module's transactional outbox.
type OutboxEvent struct {
	ID            uint64
	EventID       string
	EventType     string
	SchemaVersion int
	Subject       string
	Payload       []byte
	OccurredAt    time.Time
}

// PaymentOutboxRepository drains pending payment events and publishes them to
// NATS JetStream via the module's own relay.
type PaymentOutboxRepository interface {
	Save(ctx context.Context, event OutboxEvent) error
	ListUnpublished(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, id uint64) (bool, error)
}
