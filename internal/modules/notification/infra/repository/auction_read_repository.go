package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.AuctionReadPort = (*PostgresAuctionReadRepository)(nil)

// PostgresAuctionReadRepository is a read-only adapter over the shared bids
// table. It lets the notification module resolve auction recipients that a
// source event does not carry, without reaching into the auction module.
type PostgresAuctionReadRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresAuctionReadRepository(db sqlcgen.DBTX) *PostgresAuctionReadRepository {
	return &PostgresAuctionReadRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresAuctionReadRepository) FindPreviousHighestBidderID(
	ctx context.Context,
	auctionID uint64,
	currentBidderID uint64,
) (uint64, bool, error) {
	userID, err := repository.q.GetPreviousHighestBidder(ctx, sqlcgen.GetPreviousHighestBidderParams{
		AuctionID: int64(auctionID),
		UserID:    int64(currentBidderID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, err
	}

	return uint64(userID), true, nil
}

func (repository *PostgresAuctionReadRepository) FindBidderIDByBidID(
	ctx context.Context,
	bidID uint64,
) (uint64, bool, error) {
	userID, err := repository.q.GetBidderIDByBidID(ctx, int64(bidID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, err
	}

	return uint64(userID), true, nil
}
