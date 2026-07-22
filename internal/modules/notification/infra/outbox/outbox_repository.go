package outbox

import (
	"context"

	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.NotificationOutboxRepository = (*PostgresNotificationOutboxRepository)(nil)

type PostgresNotificationOutboxRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresNotificationOutboxRepository(db sqlcgen.DBTX) *PostgresNotificationOutboxRepository {
	return &PostgresNotificationOutboxRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresNotificationOutboxRepository) Save(ctx context.Context, event ports.OutboxEvent) error {
	return repository.q.InsertNotificationOutboxEvent(ctx, sqlcgen.InsertNotificationOutboxEventParams{
		EventID:       event.EventID,
		EventType:     event.EventType,
		SchemaVersion: int32(event.SchemaVersion),
		Subject:       event.Subject,
		Payload:       event.Payload,
		OccurredAt:    event.OccurredAt,
	})
}

func (repository *PostgresNotificationOutboxRepository) ListUnpublished(
	ctx context.Context,
	limit int,
) ([]ports.OutboxEvent, error) {
	rows, err := repository.q.ListUnpublishedNotificationOutboxEvents(ctx, int32(limit))
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

func (repository *PostgresNotificationOutboxRepository) MarkPublished(
	ctx context.Context,
	id uint64,
) (bool, error) {
	rowsAffected, err := repository.q.MarkNotificationOutboxEventPublished(ctx, int64(id))
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// SaveEmailRequest inserts an email-request row. The deterministic event id
// (source_event_id:user_id:email) plus the outbox's UNIQUE(event_id) and the
// ON CONFLICT DO NOTHING clause make a redelivered source event a no-op, so a
// user never receives a duplicate email.
func (repository *PostgresNotificationOutboxRepository) SaveEmailRequest(
	ctx context.Context,
	event ports.OutboxEvent,
) error {
	return repository.q.InsertEmailRequestOutboxEvent(ctx, sqlcgen.InsertEmailRequestOutboxEventParams{
		EventID:       event.EventID,
		EventType:     event.EventType,
		SchemaVersion: int32(event.SchemaVersion),
		Subject:       event.Subject,
		Payload:       event.Payload,
		OccurredAt:    event.OccurredAt,
	})
}
