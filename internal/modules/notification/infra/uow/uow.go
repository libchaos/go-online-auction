package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	notificationports "auction/internal/modules/notification/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ notificationports.NotificationUnitOfWork = (*NotificationUnitOfWork)(nil)

type NotificationUnitOfWork struct {
	tx                     pgx.Tx
	notificationRepository notificationports.NotificationRepository
	preferenceRepository   notificationports.PreferenceRepository
	outboxRepository       notificationports.NotificationOutboxRepository
	completed              bool
}

func (unitOfWork *NotificationUnitOfWork) NotificationRepository() notificationports.NotificationRepository {
	return unitOfWork.notificationRepository
}

func (unitOfWork *NotificationUnitOfWork) PreferenceRepository() notificationports.PreferenceRepository {
	return unitOfWork.preferenceRepository
}

func (unitOfWork *NotificationUnitOfWork) OutboxRepository() notificationports.NotificationOutboxRepository {
	return unitOfWork.outboxRepository
}

func (unitOfWork *NotificationUnitOfWork) Complete(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	if err := unitOfWork.tx.Commit(ctx); err != nil {
		return shareduow.ErrTransactionFailed
	}

	return nil
}

func (unitOfWork *NotificationUnitOfWork) Rollback(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	return unitOfWork.tx.Rollback(ctx)
}
