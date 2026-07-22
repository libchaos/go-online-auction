package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	depositports "auction/internal/modules/deposit/ports"
	ledgerports "auction/internal/modules/ledger/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ depositports.DepositUnitOfWork = (*DepositUnitOfWork)(nil)

type DepositUnitOfWork struct {
	tx                pgx.Tx
	depositRepository depositports.DepositRepository
	ledgerRepository  ledgerports.LedgerRepository
	outboxRepository  depositports.DepositOutboxRepository
	completed         bool
}

func (unitOfWork *DepositUnitOfWork) DepositRepository() depositports.DepositRepository {
	return unitOfWork.depositRepository
}

func (unitOfWork *DepositUnitOfWork) LedgerRepository() ledgerports.LedgerRepository {
	return unitOfWork.ledgerRepository
}

func (unitOfWork *DepositUnitOfWork) OutboxRepository() depositports.DepositOutboxRepository {
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
