package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/deposit/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.DepositUnitOfWork = (*DepositUnitOfWork)(nil)

type DepositUnitOfWork struct {
	tx                pgx.Tx
	depositRepository ports.DepositRepository
	outboxRepository  ports.OutboxRepository
	completed         bool
}

func (unitOfWork *DepositUnitOfWork) DepositRepository() ports.DepositRepository {
	return unitOfWork.depositRepository
}

func (unitOfWork *DepositUnitOfWork) OutboxRepository() ports.OutboxRepository {
	return unitOfWork.outboxRepository
}

func (unitOfWork *DepositUnitOfWork) Complete(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	if err := unitOfWork.tx.Commit(ctx); err != nil {
		return shareduow.ErrTransactionFailed
	}

	return nil
}

func (unitOfWork *DepositUnitOfWork) Rollback(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	return unitOfWork.tx.Rollback(ctx)
}
