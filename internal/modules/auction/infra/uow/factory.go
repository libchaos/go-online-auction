package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	shareduow "github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/uow"
)

var _ ports.AuctionUnitOfWorkFactory = (*AuctionUnitOfWorkFactory)(nil)

// AuctionUnitOfWorkFactory creates new AuctionUnitOfWork instances
type AuctionUnitOfWorkFactory struct {
	pool                     *pgxpool.Pool
	auctionRepositoryFactory func(shareduow.DBExecutor) ports.AuctionRepository
	bidRepositoryFactory     func(shareduow.DBExecutor) ports.BidRepository
}

func NewAuctionUnitOfWorkFactory(
	pool *pgxpool.Pool,
	auctionRepoFactory func(shareduow.DBExecutor) ports.AuctionRepository,
	bidRepoFactory func(shareduow.DBExecutor) ports.BidRepository,
) *AuctionUnitOfWorkFactory {
	return &AuctionUnitOfWorkFactory{
		pool:                     pool,
		auctionRepositoryFactory: auctionRepoFactory,
		bidRepositoryFactory:     bidRepoFactory,
	}
}

// Begin starts a new unit of work with a fresh transaction
func (f *AuctionUnitOfWorkFactory) Begin(ctx context.Context) (ports.AuctionUnitOfWork, error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, shareduow.ErrTransactionFailed
	}

	return &AuctionUnitOfWork{
		tx:                tx,
		auctionRepository: f.auctionRepositoryFactory(tx),
		bidRepository:     f.bidRepositoryFactory(tx),
		completed:         false,
	}, nil
}
