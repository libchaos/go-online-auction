package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/ledger/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.LedgerUnitOfWork = (*LedgerUnitOfWork)(nil)

type LedgerUnitOfWork struct {
	tx               pgx.Tx
	ledgerRepository ports.LedgerRepository
	completed        bool
}

func (unitOfWork *LedgerUnitOfWork) LedgerRepository() ports.LedgerRepository {
	return unitOfWork.ledgerRepository
}

func (unitOfWork *LedgerUnitOfWork) Complete(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	if err := unitOfWork.tx.Commit(ctx); err != nil {
		return shareduow.ErrTransactionFailed
	}

	return nil
}

func (unitOfWork *LedgerUnitOfWork) Rollback(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	return unitOfWork.tx.Rollback(ctx)
}
