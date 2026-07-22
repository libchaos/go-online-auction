package repository

import (
	"context"

	"auction/internal/modules/listing/infra/sqlcgen"
	"auction/internal/modules/listing/ports"
)

var _ ports.ListingOutboxRepository = (*PostgresListingOutboxRepository)(nil)

// PostgresListingOutboxRepository writes listing events into the listing's own
// listing_outbox table. The outbox relay owned by the listing module drains the
// pending rows, so the listing bounded context is independently deployable.
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

func (r *PostgresListingOutboxRepository) ListUnpublished(
	ctx context.Context,
	limit int,
) ([]ports.OutboxEvent, error) {
	rows, err := r.q.ListUnpublishedOutboxEvents(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	events := make([]ports.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, ports.OutboxEvent{
			ID:            uint64(row.ID),
			EventID:       row.EventID,
			EventType:     row.EventType,
			SchemaVersion: int(row.SchemaVersion),
			Subject:       row.Subject,
			Payload:       row.Payload,
			OccurredAt:    row.OccurredAt,
		})
	}

	return events, nil
}

func (r *PostgresListingOutboxRepository) MarkPublished(
	ctx context.Context,
	id uint64,
) (bool, error) {
	rowsAffected, err := r.q.MarkOutboxEventPublished(ctx, int64(id))
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
