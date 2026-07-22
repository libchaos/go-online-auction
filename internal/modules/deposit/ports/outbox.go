package ports

import (
	"context"
	"time"
)

type OutboxEvent struct {
	ID            uint64
	EventID       string
	EventType     string
	SchemaVersion int
	Subject       string
	Payload       []byte
	OccurredAt    time.Time
}

type DepositOutboxRepository interface {
	Save(ctx context.Context, event OutboxEvent) error
	ListUnpublished(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, id uint64) (bool, error)
}
