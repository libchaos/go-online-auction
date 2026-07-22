package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/notification/domain/errs"
	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/mapper"
	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.NotificationRepository = (*PostgresNotificationRepository)(nil)

type PostgresNotificationRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.NotificationMapper
}

func NewPostgresNotificationRepository(
	db sqlcgen.DBTX,
	notificationMapper *mapper.NotificationMapper,
) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{
		q:      sqlcgen.New(db),
		mapper: notificationMapper,
	}
}

// Save inserts a notification. The insert is idempotent on the idempotency key:
// a redelivered source event hits ON CONFLICT DO NOTHING, which returns no row,
// so the second return value reports whether the row was newly created.
func (repository *PostgresNotificationRepository) Save(
	ctx context.Context,
	notification model.NotificationModel,
) (model.NotificationModel, bool, error) {
	row, err := repository.q.InsertNotification(ctx, repository.mapper.ToInsertNotificationParams(notification))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NotificationModel{}, false, nil
		}

		return model.NotificationModel{}, false, err
	}

	persisted, mapErr := repository.mapper.ToNotificationDomain(row)
	if mapErr != nil {
		return model.NotificationModel{}, false, mapErr
	}

	return persisted, true, nil
}

func (repository *PostgresNotificationRepository) FindByID(
	ctx context.Context,
	id uint64,
) (model.NotificationModel, error) {
	row, err := repository.q.GetNotificationByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NotificationModel{}, errs.ErrNotificationNotFound
		}

		return model.NotificationModel{}, err
	}

	return repository.mapper.ToNotificationDomain(row)
}

func (repository *PostgresNotificationRepository) ListByUser(
	ctx context.Context,
	userID uint64,
	limit int,
	offset int,
) ([]model.NotificationModel, error) {
	rows, err := repository.q.ListNotificationsByUser(ctx, sqlcgen.ListNotificationsByUserParams{
		UserID: int64(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	return repository.toDomainSlice(rows)
}

func (repository *PostgresNotificationRepository) ListUnreadByUser(
	ctx context.Context,
	userID uint64,
	limit int,
	offset int,
) ([]model.NotificationModel, error) {
	rows, err := repository.q.ListUnreadNotificationsByUser(ctx, sqlcgen.ListUnreadNotificationsByUserParams{
		UserID: int64(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	return repository.toDomainSlice(rows)
}

func (repository *PostgresNotificationRepository) CountUnread(
	ctx context.Context,
	userID uint64,
) (uint64, error) {
	count, err := repository.q.GetUnreadNotificationCount(ctx, int64(userID))
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func (repository *PostgresNotificationRepository) MarkRead(
	ctx context.Context,
	id uint64,
	userID uint64,
) (bool, error) {
	rowsAffected, err := repository.q.MarkNotificationRead(ctx, sqlcgen.MarkNotificationReadParams{
		ID:     int64(id),
		UserID: int64(userID),
	})
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository *PostgresNotificationRepository) MarkAllRead(
	ctx context.Context,
	userID uint64,
) (uint64, error) {
	rowsAffected, err := repository.q.MarkAllNotificationsRead(ctx, int64(userID))
	if err != nil {
		return 0, err
	}

	return uint64(rowsAffected), nil
}

func (repository *PostgresNotificationRepository) Delete(
	ctx context.Context,
	id uint64,
	userID uint64,
) (bool, error) {
	rowsAffected, err := repository.q.DeleteNotification(ctx, sqlcgen.DeleteNotificationParams{
		ID:     int64(id),
		UserID: int64(userID),
	})
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository *PostgresNotificationRepository) toDomainSlice(
	rows []sqlcgen.Notification,
) ([]model.NotificationModel, error) {
	notifications := make([]model.NotificationModel, 0, len(rows))
	for index := range rows {
		notification, err := repository.mapper.ToNotificationDomain(rows[index])
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
