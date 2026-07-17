package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/sqlcgen"
	"auction/internal/modules/deposit/ports"
)

var _ ports.AuctionConfigPort = (*PostgresAuctionConfigRepository)(nil)
var _ ports.AuctionWinnerPort = (*PostgresAuctionConfigRepository)(nil)

type PostgresAuctionConfigRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresAuctionConfigRepository(db sqlcgen.DBTX) *PostgresAuctionConfigRepository {
	return &PostgresAuctionConfigRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresAuctionConfigRepository) GetRequiredDeposit(
	ctx context.Context,
	auctionID uint64,
) (ports.AuctionDepositConfig, error) {
	row, err := repository.q.GetAuctionDepositConfig(ctx, int64(auctionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.AuctionDepositConfig{}, errs.ErrAuctionConfigNotFound
		}

		return ports.AuctionDepositConfig{}, err
	}

	return ports.AuctionDepositConfig{
		Required: row.DepositRequired,
		Amount:   model.NewMoneyModel(uint64(row.DepositAmountInCents)),
	}, nil
}

func (repository *PostgresAuctionConfigRepository) GetWinnerUserID(
	ctx context.Context,
	auctionID uint64,
) (*uint64, error) {
	row, err := repository.q.GetAuctionWinner(ctx, int64(auctionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrAuctionConfigNotFound
		}

		return nil, err
	}

	if row == nil {
		return nil, errs.ErrAuctionWinnerNotFound
	}

	winner := uint64(*row)

	return &winner, nil
}
