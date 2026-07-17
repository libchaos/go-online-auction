package outbox

import (
	"context"

	"auction/internal/modules/deposit/infra/sqlcgen"
	"auction/internal/modules/deposit/ports"
)

var _ ports.OutboxRepository = (*PostgresOutboxRepository)(nil)

type PostgresOutboxRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresOutboxRepository(db sqlcgen.DBTX) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresOutboxRepository) Save(ctx context.Context, event ports.OutboxEvent) error {
	err := repository.q.InsertOutboxEvent(ctx, sqlcgen.InsertOutboxEventParams{
		EventID:       event.EventID,
		EventType:     event.EventType,
		SchemaVersion: int32(event.SchemaVersion),
		Subject:       event.Subject,
		Payload:       event.Payload,
		OccurredAt:    event.OccurredAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func (repository *PostgresOutboxRepository) ListUnpublished(
	ctx context.Context,
	limit int,
) ([]ports.OutboxEvent, error) {
	rows, err := repository.q.ListUnpublishedOutboxEvents(ctx, int32(limit))
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

func (repository *PostgresOutboxRepository) MarkPublished(
	ctx context.Context,
	id uint64,
) (bool, error) {
	rowsAffected, err := repository.q.MarkOutboxEventPublished(ctx, int64(id))
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
