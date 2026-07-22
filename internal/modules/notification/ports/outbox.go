package ports

import (
	"context"
	"time"
)

// OutboxEvent is a single row in the notification module's transactional outbox.
type OutboxEvent struct {
	ID            uint64
	EventID       string
	EventType     string
	SchemaVersion int
	Subject       string
	Payload       []byte
	OccurredAt    time.Time
}

// NotificationOutboxRepository drains pending notification events and publishes
// them to NATS JetStream via the module's own relay so the SSE realtime hub can
// fan each event out to the recipient user's connected clients.
type NotificationOutboxRepository interface {
	Save(ctx context.Context, event OutboxEvent) error
	SaveEmailRequest(ctx context.Context, event OutboxEvent) error
	ListUnpublished(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, id uint64) (bool, error)
}
