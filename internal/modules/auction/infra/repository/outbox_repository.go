package repository

import (
	"context"

	"auction/internal/modules/auction/infra/sqlcgen"
	"auction/internal/modules/auction/ports"
)

var _ ports.OutboxRepository = (*PostgresOutboxRepository)(nil)

type PostgresOutboxRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresOutboxRepository(db sqlcgen.DBTX) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{q: sqlcgen.New(db)}
}

func (r *PostgresOutboxRepository) Save(ctx context.Context, event ports.OutboxEvent) error {
	return r.q.InsertOutboxEvent(ctx, sqlcgen.InsertOutboxEventParams{
		EventID:       event.EventID,
		EventType:     event.EventType,
		SchemaVersion: int32(event.SchemaVersion),
		Subject:       event.Subject,
		Payload:       event.Payload,
		OccurredAt:    event.OccurredAt,
	})
}

func (r *PostgresOutboxRepository) ListUnpublished(ctx context.Context, limit int) ([]ports.OutboxEvent, error) {
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

func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, id uint64) (bool, error) {
	rowsAffected, err := r.q.MarkOutboxEventPublished(ctx, int64(id))
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
