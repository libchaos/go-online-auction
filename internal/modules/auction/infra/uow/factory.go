package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/mapper"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/repository"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	shareduow "github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/uow"
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
		completed:         false,
	}, nil
}
