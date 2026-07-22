package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	ledgermapper "auction/internal/modules/ledger/infra/mapper"
	ledgerrepository "auction/internal/modules/ledger/infra/repository"
	paymentmapper "auction/internal/modules/payment/infra/mapper"
	"auction/internal/modules/payment/infra/outbox"
	"auction/internal/modules/payment/infra/repository"
	"auction/internal/modules/payment/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.PaymentUnitOfWorkFactory = (*PaymentUnitOfWorkFactory)(nil)

type PaymentUnitOfWorkFactory struct {
	pool          *pgxpool.Pool
	paymentMapper *paymentmapper.PaymentMapper
	ledgerMapper  *ledgermapper.LedgerMapper
}

func NewPaymentUnitOfWorkFactory(
	pool *pgxpool.Pool,
	paymentMapper *paymentmapper.PaymentMapper,
	ledgerMapper *ledgermapper.LedgerMapper,
) *PaymentUnitOfWorkFactory {
	return &PaymentUnitOfWorkFactory{
		pool:          pool,
		paymentMapper: paymentMapper,
		ledgerMapper:  ledgerMapper,
	}
}

func (factory *PaymentUnitOfWorkFactory) Begin(ctx context.Context) (ports.PaymentUnitOfWork, error) {
	tx, err := factory.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &PaymentUnitOfWork{
		tx:                   tx,
		paymentRepository:    repository.NewPostgresPaymentRepository(tx, factory.paymentMapper),
		withdrawalRepository: repository.NewPostgresWithdrawalRepository(tx, factory.paymentMapper),
		ledgerRepository:     ledgerrepository.NewPostgresLedgerRepository(tx, factory.ledgerMapper),
		outboxRepository:     outbox.NewPostgresPaymentOutboxRepository(tx),
		completed:            false,
	}, nil
}
