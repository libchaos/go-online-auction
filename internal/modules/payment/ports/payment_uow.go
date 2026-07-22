package ports

import (
	"context"

	ledgerports "auction/internal/modules/ledger/ports"
)

// PaymentUnitOfWork bundles the repositories that must commit atomically
// inside a single PostgreSQL transaction: the payment/withdrawal stores, the
// ledger store (freeze / transfer / unfreeze), and the outbox store.
type PaymentUnitOfWork interface {
	PaymentRepository() PaymentRepository
	WithdrawalRepository() WithdrawalRepository
	LedgerRepository() ledgerports.LedgerRepository
	OutboxRepository() PaymentOutboxRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// PaymentUnitOfWorkFactory opens a new unit of work on a pooled connection.
type PaymentUnitOfWorkFactory interface {
	Begin(ctx context.Context) (PaymentUnitOfWork, error)
}
