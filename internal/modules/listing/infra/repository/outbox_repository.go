package repository

import (
	"context"

	"auction/internal/modules/listing/infra/sqlcgen"
	"auction/internal/modules/listing/ports"
)

var _ ports.ListingOutboxRepository = (*PostgresListingOutboxRepository)(nil)

// PostgresListingOutboxRepository writes listing events into the shared
// event_outbox table; the outbox relay owned by the auction module drains
// all pending rows regardless of subject.
type PostgresListingOutboxRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresListingOutboxRepository(db sqlcgen.DBTX) *PostgresListingOutboxRepository {
	return &PostgresListingOutboxRepository{q: sqlcgen.New(db)}
}

func (r *PostgresListingOutboxRepository) Save(ctx context.Context, event ports.OutboxEvent) error {
	return r.q.InsertListingOutboxEvent(ctx, sqlcgen.InsertListingOutboxEventParams{
		EventID:       event.EventID,
		EventType:     event.EventType,
		SchemaVersion: int32(event.SchemaVersion),
		Subject:       event.Subject,
		Payload:       event.Payload,
		OccurredAt:    event.OccurredAt,
	})
}
