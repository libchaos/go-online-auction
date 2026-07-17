package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/sqlcgen"
	"auction/internal/modules/auction/ports"
)

const (
	// PostgreSQL error codes
	pgCheckViolationCode  = "23514"
	pgUniqueViolationCode = "23505"

	// Error message patterns
	bidMustBeHigherErrMsg = "must be higher than current highest bid"

	// Unique index enforcing (auction_id, idempotency_key)
	bidIdempotencyIndexName = "ux_bid_auction_idempotency"
)

var _ ports.BidRepository = (*PostgresBidRepository)(nil)

type PostgresBidRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.BidMapper
}

func NewPostgresBidRepository(db sqlcgen.DBTX, mapper *mapper.BidMapper) *PostgresBidRepository {
	return &PostgresBidRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresBidRepository) Create(
	ctx context.Context,
	bid model.BidModel,
	idempotencyKey string,
) (model.BidModel, error) {
	row, err := r.q.CreateBid(ctx, r.mapper.ToCreateParams(bid, idempotencyKey))
	if err != nil {
		if ok, mapped := mapPostgresCreateError(err); ok {
			return model.BidModel{}, mapped
		}
		return model.BidModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func mapPostgresCreateError(err error) (bool, error) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false, nil
	}
	if pgErr.Code == pgCheckViolationCode && strings.Contains(pgErr.Message, bidMustBeHigherErrMsg) {
		return true, errs.ErrBidMustExceedHighest
	}
	if pgErr.Code == pgUniqueViolationCode && strings.Contains(pgErr.ConstraintName, bidIdempotencyIndexName) {
		return true, errs.ErrBidDuplicateIdempotencyKey
	}
	return false, nil
}

func (r *PostgresBidRepository) FindByID(ctx context.Context, id uint64) (model.BidModel, error) {
	row, err := r.q.GetBidByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BidModel{}, errs.ErrBidNotFound
		}
		return model.BidModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresBidRepository) FindByAuctionID(ctx context.Context, auctionID uint64) ([]model.BidModel, error) {
	rows, err := r.q.ListBidsByAuctionID(ctx, int64(auctionID))
	if err != nil {
		return nil, err
	}

	return r.mapBidRows(rows, nil)
}

func (r *PostgresBidRepository) Update(ctx context.Context, bid model.BidModel) error {
	rowsAffected, err := r.q.UpdateBid(ctx, sqlcgen.UpdateBidParams{
		UpdatedAt: bid.UpdatedAt(),
		ID:        int64(bid.ID()),
	})
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrBidNotFound
	}

	return nil
}

func (r *PostgresBidRepository) FindTopBidsByAuctionID(
	ctx context.Context,
	auctionID uint64,
	limit int,
) ([]model.BidModel, error) {
	rows, err := r.q.ListTopBidsByAuctionID(ctx, sqlcgen.ListTopBidsByAuctionIDParams{
		AuctionID: int64(auctionID),
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, err
	}

	return r.mapBidRows(rows, []model.BidModel{})
}

func (r *PostgresBidRepository) mapBidRows(rows []sqlcgen.Bid, bids []model.BidModel) ([]model.BidModel, error) {
	for _, row := range rows {
		bid, err := r.mapper.ToDomain(row)
		if err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}
