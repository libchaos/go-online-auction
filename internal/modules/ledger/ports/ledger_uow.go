package ports

import "context"

type LedgerUnitOfWork interface {
	LedgerRepository() LedgerRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type LedgerUnitOfWorkFactory interface {
	Begin(ctx context.Context) (LedgerUnitOfWork, error)
}
