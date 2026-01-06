package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/entity"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/mapper"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/uow"
)

var _ ports.BidRepository = (*PostgresBidRepository)(nil)

type PostgresBidRepository struct {
	db     uow.DBExecutor
	mapper *mapper.BidMapper
}

func NewPostgresBidRepository(db uow.DBExecutor, mapper *mapper.BidMapper) *PostgresBidRepository {
	return &PostgresBidRepository{
		db:     db,
		mapper: mapper,
	}
}

func (r *PostgresBidRepository) Create(ctx context.Context, bid model.BidModel) error {
	e := r.mapper.ToEntity(bid)

	query := `
		INSERT INTO bids (auction_id, user_id, amount_in_cents, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		e.AuctionID,
		e.UserID,
		e.AmountInCents,
		e.Currency,
		e.Status,
		e.CreatedAt,
		e.UpdatedAt,
	).Scan(&e.ID)

	return err
}

func (r *PostgresBidRepository) FindByID(ctx context.Context, id uint64) (model.BidModel, error) {
	query := `
		SELECT id, auction_id, user_id, amount_in_cents, currency, status, created_at, updated_at
		FROM bids
		WHERE id = $1`

	var e entity.BidEntity
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.AuctionID,
		&e.UserID,
		&e.AmountInCents,
		&e.Currency,
		&e.Status,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BidModel{}, errs.ErrBidNotFound
		}
		return model.BidModel{}, err
	}

	return r.mapper.ToDomain(e)
}

func (r *PostgresBidRepository) FindByAuctionID(ctx context.Context, auctionID uint64) ([]model.BidModel, error) {
	query := `
		SELECT id, auction_id, user_id, amount_in_cents, currency, status, created_at, updated_at
		FROM bids
		WHERE auction_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, auctionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []model.BidModel
	for rows.Next() {
		var e entity.BidEntity
		if err := rows.Scan(
			&e.ID,
			&e.AuctionID,
			&e.UserID,
			&e.AmountInCents,
			&e.Currency,
			&e.Status,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}

		bid, err := r.mapper.ToDomain(e)
		if err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bids, nil
}

func (r *PostgresBidRepository) Update(ctx context.Context, bid model.BidModel) error {
	e := r.mapper.ToEntity(bid)

	query := `
		UPDATE bids
		SET status = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.Exec(ctx, query,
		e.Status,
		e.UpdatedAt,
		e.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errs.ErrBidNotFound
	}

	return nil
}
