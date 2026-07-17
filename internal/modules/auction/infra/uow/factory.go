package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/repository"
	"auction/internal/modules/auction/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.AuctionUnitOfWorkFactory = (*AuctionUnitOfWorkFactory)(nil)

type AuctionUnitOfWorkFactory struct {
	pool          *pgxpool.Pool
	auctionMapper *mapper.AuctionMapper
	bidMapper     *mapper.BidMapper
}

func NewAuctionUnitOfWorkFactory(
	pool *pgxpool.Pool,
	auctionMapper *mapper.AuctionMapper,
	bidMapper *mapper.BidMapper,
) *AuctionUnitOfWorkFactory {
	return &AuctionUnitOfWorkFactory{
		pool:          pool,
		auctionMapper: auctionMapper,
		bidMapper:     bidMapper,
	}
}

// Begin starts a transaction and returns a unit of work bound to it.
// pgx.Tx satisfies sqlcgen.DBTX, so the repositories run inside the transaction.
func (f *AuctionUnitOfWorkFactory) Begin(ctx context.Context) (ports.AuctionUnitOfWork, error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &AuctionUnitOfWork{
		tx:                tx,
		auctionRepository: repository.NewPostgresAuctionRepository(tx, f.auctionMapper),
		bidRepository:     repository.NewPostgresBidRepository(tx, f.bidMapper),
		outboxRepository:  repository.NewPostgresOutboxRepository(tx),
		completed:         false,
	}, nil
}
