package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/modules/listing/infra/mapper"
	"auction/internal/modules/listing/infra/repository"
	"auction/internal/modules/listing/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.ListingUnitOfWorkFactory = (*ListingUnitOfWorkFactory)(nil)

type ListingUnitOfWorkFactory struct {
	pool      *pgxpool.Pool
	spuMapper *mapper.SpuMapper
	skuMapper *mapper.SkuMapper
}

func NewListingUnitOfWorkFactory(
	pool *pgxpool.Pool,
	spuMapper *mapper.SpuMapper,
	skuMapper *mapper.SkuMapper,
) *ListingUnitOfWorkFactory {
	return &ListingUnitOfWorkFactory{
		pool:      pool,
		spuMapper: spuMapper,
		skuMapper: skuMapper,
	}
}

// Begin starts a transaction and returns a unit of work bound to it.
// pgx.Tx satisfies sqlcgen.DBTX, so the repositories run inside the transaction.
func (f *ListingUnitOfWorkFactory) Begin(ctx context.Context) (ports.ListingUnitOfWork, error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &ListingUnitOfWork{
		tx:               tx,
		spuRepository:    repository.NewPostgresSpuRepository(tx, f.spuMapper),
		skuRepository:    repository.NewPostgresSkuRepository(tx, f.skuMapper),
		outboxRepository: repository.NewPostgresListingOutboxRepository(tx),
		completed:        false,
	}, nil
}
