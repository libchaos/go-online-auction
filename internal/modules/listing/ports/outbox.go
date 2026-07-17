package ports

import (
	"context"
	"time"
)

// OutboxEvent is a serialized domain event recorded in the shared transactional
// outbox (event_outbox). The listing module only writes rows; the outbox relay
// owned by the auction module drains all pending rows regardless of subject.
type OutboxEvent struct {
	EventID       string
	EventType     string
	SchemaVersion int
	Subject       string
	Payload       []byte
	OccurredAt    time.Time
}

// ListingOutboxRepository persists listing events in the transactional outbox
// within the ambient transaction.
type ListingOutboxRepository interface {
	Save(ctx context.Context, event OutboxEvent) error
}
