package ports

import (
	"context"

	ledgerports "auction/internal/modules/ledger/ports"
)

type DepositUnitOfWork interface {
	DepositRepository() DepositRepository
	LedgerRepository() ledgerports.LedgerRepository
	OutboxRepository() DepositOutboxRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DepositUnitOfWorkFactory interface {
	Begin(ctx context.Context) (DepositUnitOfWork, error)
}
