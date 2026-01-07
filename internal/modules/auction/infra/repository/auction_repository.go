package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/entity"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/mapper"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/uow"
)

var _ ports.AuctionRepository = (*PostgresAuctionRepository)(nil)

type PostgresAuctionRepository struct {
	db     uow.DBExecutor
	mapper *mapper.AuctionMapper
}

func NewPostgresAuctionRepository(db uow.DBExecutor, mapper *mapper.AuctionMapper) *PostgresAuctionRepository {
	return &PostgresAuctionRepository{
		db:     db,
		mapper: mapper,
	}
}

func (r *PostgresAuctionRepository) Create(ctx context.Context, auction model.AuctionModel) error {
	e := r.mapper.ToEntity(auction)

	query := `
		INSERT INTO auctions (listing_id, end_time, state, highest_bid_id, highest_bid_amount_in_cents, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		e.ListingID,
		e.EndTime,
		e.State,
		e.HighestBidID,
		e.HighestBidAmountInCents,
		e.Version,
		e.CreatedAt,
		e.UpdatedAt,
	).Scan(&e.ID)

	return err
}

func (r *PostgresAuctionRepository) FindByID(ctx context.Context, id uint64) (model.AuctionModel, error) {
	query := `
		SELECT id, listing_id, start_time, end_time, state, highest_bid_id, highest_bid_amount_in_cents, version, created_at, updated_at
		FROM auctions
		WHERE id = $1`

	var e entity.AuctionEntity
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.ListingID,
		&e.StartTime,
		&e.EndTime,
		&e.State,
		&e.HighestBidID,
		&e.HighestBidAmountInCents,
		&e.Version,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AuctionModel{}, errs.ErrAuctionNotFound
		}
		return model.AuctionModel{}, err
	}

	return r.mapper.ToDomain(e)
}

// FindByIDForUpdate retrieves an auction with row-level lock for update
// Uses NOWAIT to fail fast under contention
func (r *PostgresAuctionRepository) FindByIDForUpdate(ctx context.Context, id uint64) (model.AuctionModel, error) {
	query := `
		SELECT id, listing_id, start_time, end_time, state, highest_bid_id, highest_bid_amount_in_cents, version, created_at, updated_at
		FROM auctions
		WHERE id = $1
		FOR UPDATE NOWAIT`

	var e entity.AuctionEntity
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.ListingID,
		&e.StartTime,
		&e.EndTime,
		&e.State,
		&e.HighestBidID,
		&e.HighestBidAmountInCents,
		&e.Version,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AuctionModel{}, errs.ErrAuctionNotFound
		}
		// Check for lock_not_available error (NOWAIT)
		if isPgLockError(err) {
			return model.AuctionModel{}, errs.ErrConcurrencyConflict
		}
		return model.AuctionModel{}, err
	}

	return r.mapper.ToDomain(e)
}

func (r *PostgresAuctionRepository) Update(ctx context.Context, auction model.AuctionModel) error {
	e := r.mapper.ToEntity(auction)
	previousVersion := e.Version - 1 // Domain increments version before calling Update

	query := `
		UPDATE auctions
		SET listing_id = $1, start_time = $2, end_time = $3, state = $4, 
			highest_bid_id = $5, highest_bid_amount_in_cents = $6, version = $7, updated_at = $8
		WHERE id = $9 AND version = $10`

	result, err := r.db.Exec(ctx, query,
		e.ListingID,
		e.StartTime,
		e.EndTime,
		e.State,
		e.HighestBidID,
		e.HighestBidAmountInCents,
		e.Version,
		e.UpdatedAt,
		e.ID,
		previousVersion,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errs.ErrConcurrencyConflict
	}

	return nil
}

func (r *PostgresAuctionRepository) FindAllPaginated(
	ctx context.Context,
	state *enum.AuctionStateEnum,
	limit, offset int,
) ([]model.AuctionModel, error) {
	var query string
	var args []any

	if state != nil {
		query = `
			SELECT id, listing_id, start_time, end_time, state, highest_bid_id, highest_bid_amount_in_cents, version, created_at, updated_at
			FROM auctions
			WHERE state = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []any{*state, limit, offset}
	} else {
		query = `
			SELECT id, listing_id, start_time, end_time, state, highest_bid_id, highest_bid_amount_in_cents, version, created_at, updated_at
			FROM auctions
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`
		args = []any{limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	auctions := []model.AuctionModel{}
	for rows.Next() {
		var e entity.AuctionEntity
		if scanErr := rows.Scan(
			&e.ID,
			&e.ListingID,
			&e.StartTime,
			&e.EndTime,
			&e.State,
			&e.HighestBidID,
			&e.HighestBidAmountInCents,
			&e.Version,
			&e.CreatedAt,
			&e.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}

		auction, mapErr := r.mapper.ToDomain(e)
		if mapErr != nil {
			return nil, mapErr
		}
		auctions = append(auctions, auction)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return auctions, nil
}

func (r *PostgresAuctionRepository) Count(ctx context.Context, state *enum.AuctionStateEnum) (uint64, error) {
	var query string
	var args []any

	if state != nil {
		query = `SELECT COUNT(*) FROM auctions WHERE state = $1`
		args = []any{*state}
	} else {
		query = `SELECT COUNT(*) FROM auctions`
		args = []any{}
	}

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func isPgLockError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 55P03 = lock_not_available
		return pgErr.Code == "55P03"
	}
	return false
}
