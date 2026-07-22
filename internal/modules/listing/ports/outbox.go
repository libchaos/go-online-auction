package ports

import (
	"context"
	"time"
)

// OutboxEvent is a serialized domain event recorded in the listing transactional
// outbox (listing_outbox). The listing module owns both the outbox table and its
// own relay that drains it, so the two can be deployed independently of other
// bounded contexts.
type OutboxEvent struct {
	ID            uint64
	EventID       string
	EventType     string
	SchemaVersion int
	Subject       string
	Payload       []byte
	OccurredAt    time.Time
}

// ListingOutboxRepository persists listing events in the transactional outbox
// within the ambient transaction and drains them via the outbox relay.
type ListingOutboxRepository interface {
	Save(ctx context.Context, event OutboxEvent) error
	ListUnpublished(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, id uint64) (bool, error)
}
