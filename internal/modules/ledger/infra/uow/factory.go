package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/modules/ledger/infra/mapper"
	"auction/internal/modules/ledger/infra/repository"
	"auction/internal/modules/ledger/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.LedgerUnitOfWorkFactory = (*LedgerUnitOfWorkFactory)(nil)

type LedgerUnitOfWorkFactory struct {
	pool   *pgxpool.Pool
	mapper *mapper.LedgerMapper
}

func NewLedgerUnitOfWorkFactory(
	pool *pgxpool.Pool,
	ledgerMapper *mapper.LedgerMapper,
) *LedgerUnitOfWorkFactory {
	return &LedgerUnitOfWorkFactory{
		pool:   pool,
		mapper: ledgerMapper,
	}
}

func (factory *LedgerUnitOfWorkFactory) Begin(ctx context.Context) (ports.LedgerUnitOfWork, error) {
	tx, err := factory.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &LedgerUnitOfWork{
		tx:               tx,
		ledgerRepository: repository.NewPostgresLedgerRepository(tx, factory.mapper),
		completed:        false,
	}, nil
}
