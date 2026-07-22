package outbox

import (
	"context"

	"auction/internal/modules/payment/infra/sqlcgen"
	"auction/internal/modules/payment/ports"
)

var _ ports.PaymentOutboxRepository = (*PostgresPaymentOutboxRepository)(nil)

type PostgresPaymentOutboxRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresPaymentOutboxRepository(db sqlcgen.DBTX) *PostgresPaymentOutboxRepository {
	return &PostgresPaymentOutboxRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresPaymentOutboxRepository) Save(ctx context.Context, event ports.OutboxEvent) error {
	return repository.q.InsertPaymentOutboxEvent(ctx, sqlcgen.InsertPaymentOutboxEventParams{
		EventID:       event.EventID,
		EventType:     event.EventType,
		SchemaVersion: int32(event.SchemaVersion),
		Subject:       event.Subject,
		Payload:       event.Payload,
		OccurredAt:    event.OccurredAt,
	})
}

func (repository *PostgresPaymentOutboxRepository) ListUnpublished(
	ctx context.Context,
	limit int,
) ([]ports.OutboxEvent, error) {
	rows, err := repository.q.ListUnpublishedPaymentOutboxEvents(ctx, int32(limit))
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

func (repository *PostgresPaymentOutboxRepository) MarkPublished(
	ctx context.Context,
	id uint64,
) (bool, error) {
	rowsAffected, err := repository.q.MarkPaymentOutboxEventPublished(ctx, int64(id))
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
