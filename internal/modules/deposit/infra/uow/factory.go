package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/modules/deposit/infra/mapper"
	"auction/internal/modules/deposit/infra/outbox"
	"auction/internal/modules/deposit/infra/repository"
	"auction/internal/modules/deposit/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.DepositUnitOfWorkFactory = (*DepositUnitOfWorkFactory)(nil)

type DepositUnitOfWorkFactory struct {
	pool          *pgxpool.Pool
	depositMapper *mapper.DepositMapper
}

func NewDepositUnitOfWorkFactory(
	pool *pgxpool.Pool,
	depositMapper *mapper.DepositMapper,
) *DepositUnitOfWorkFactory {
	return &DepositUnitOfWorkFactory{
		pool:          pool,
		depositMapper: depositMapper,
	}
}

func (factory *DepositUnitOfWorkFactory) Begin(ctx context.Context) (ports.DepositUnitOfWork, error) {
	tx, err := factory.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &DepositUnitOfWork{
		tx:                tx,
		depositRepository: repository.NewPostgresDepositRepository(tx, factory.depositMapper),
		outboxRepository:  outbox.NewPostgresOutboxRepository(tx),
		completed:         false,
	}, nil
}
