package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	ledgerports "auction/internal/modules/ledger/ports"
	paymentports "auction/internal/modules/payment/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ paymentports.PaymentUnitOfWork = (*PaymentUnitOfWork)(nil)

type PaymentUnitOfWork struct {
	tx                   pgx.Tx
	paymentRepository    paymentports.PaymentRepository
	withdrawalRepository paymentports.WithdrawalRepository
	ledgerRepository     ledgerports.LedgerRepository
	outboxRepository     paymentports.PaymentOutboxRepository
	completed            bool
}

func (unitOfWork *PaymentUnitOfWork) PaymentRepository() paymentports.PaymentRepository {
	return unitOfWork.paymentRepository
}

func (unitOfWork *PaymentUnitOfWork) WithdrawalRepository() paymentports.WithdrawalRepository {
	return unitOfWork.withdrawalRepository
}

func (unitOfWork *PaymentUnitOfWork) LedgerRepository() ledgerports.LedgerRepository {
	return unitOfWork.ledgerRepository
}

func (unitOfWork *PaymentUnitOfWork) OutboxRepository() paymentports.PaymentOutboxRepository {
	return unitOfWork.outboxRepository
}

func (unitOfWork *PaymentUnitOfWork) Complete(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	if err := unitOfWork.tx.Commit(ctx); err != nil {
		return shareduow.ErrTransactionFailed
	}

	return nil
}

func (unitOfWork *PaymentUnitOfWork) Rollback(ctx context.Context) error {
	if unitOfWork.completed {
		return nil
	}

	unitOfWork.completed = true

	return unitOfWork.tx.Rollback(ctx)
}
