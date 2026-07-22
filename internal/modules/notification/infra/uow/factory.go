package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	notificationmapper "auction/internal/modules/notification/infra/mapper"
	"auction/internal/modules/notification/infra/outbox"
	"auction/internal/modules/notification/infra/repository"
	"auction/internal/modules/notification/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.NotificationUnitOfWorkFactory = (*NotificationUnitOfWorkFactory)(nil)

type NotificationUnitOfWorkFactory struct {
	pool               *pgxpool.Pool
	notificationMapper *notificationmapper.NotificationMapper
}

func NewNotificationUnitOfWorkFactory(
	pool *pgxpool.Pool,
	notificationMapper *notificationmapper.NotificationMapper,
) *NotificationUnitOfWorkFactory {
	return &NotificationUnitOfWorkFactory{
		pool:               pool,
		notificationMapper: notificationMapper,
	}
}

func (factory *NotificationUnitOfWorkFactory) Begin(ctx context.Context) (ports.NotificationUnitOfWork, error) {
	tx, err := factory.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &NotificationUnitOfWork{
		tx:                     tx,
		notificationRepository: repository.NewPostgresNotificationRepository(tx, factory.notificationMapper),
		preferenceRepository:   repository.NewPostgresPreferenceRepository(tx, factory.notificationMapper),
		outboxRepository:       outbox.NewPostgresNotificationOutboxRepository(tx),
		completed:              false,
	}, nil
}
