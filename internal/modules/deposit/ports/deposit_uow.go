package ports

import "context"

type DepositUnitOfWork interface {
	DepositRepository() DepositRepository
	OutboxRepository() OutboxRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DepositUnitOfWorkFactory interface {
	Begin(ctx context.Context) (DepositUnitOfWork, error)
}
