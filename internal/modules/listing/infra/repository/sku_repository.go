package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/mapper"
	"auction/internal/modules/listing/infra/sqlcgen"
	"auction/internal/modules/listing/ports"
)

var _ ports.SkuRepository = (*PostgresSkuRepository)(nil)

type PostgresSkuRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.SkuMapper
}

func NewPostgresSkuRepository(db sqlcgen.DBTX, mapper *mapper.SkuMapper) *PostgresSkuRepository {
	return &PostgresSkuRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresSkuRepository) Create(ctx context.Context, sku model.SkuModel) (model.SkuModel, error) {
	params, err := r.mapper.ToCreateParams(sku)
	if err != nil {
		return model.SkuModel{}, err
	}

	row, err := r.q.CreateSku(ctx, params)
	if err != nil {
		return model.SkuModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresSkuRepository) FindByID(ctx context.Context, id uint64) (model.SkuModel, error) {
	row, err := r.q.GetSkuByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SkuModel{}, errs.ErrSkuNotFound
		}
		return model.SkuModel{}, err
	}

	return r.mapper.ToDomain(row)
}

// FindByIDForUpdate retrieves a SKU with row-level lock for update.
// Uses NOWAIT to fail fast under contention.
func (r *PostgresSkuRepository) FindByIDForUpdate(ctx context.Context, id uint64) (model.SkuModel, error) {
	row, err := r.q.GetSkuByIDForUpdate(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SkuModel{}, errs.ErrSkuNotFound
		}
		if isPgLockError(err) {
			return model.SkuModel{}, errs.ErrConcurrencyConflict
		}
		return model.SkuModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresSkuRepository) Update(ctx context.Context, sku model.SkuModel) error {
	params, err := r.mapper.ToUpdateParams(sku)
	if err != nil {
		return err
	}
	params.PreviousVersion = params.Version - 1 // Domain increments version before calling Update

	rowsAffected, err := r.q.UpdateSku(ctx, params)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrConcurrencyConflict
	}

	return nil
}

func (r *PostgresSkuRepository) FindBySpuID(ctx context.Context, spuID uint64) ([]model.SkuModel, error) {
	rows, err := r.q.ListSkusBySpuID(ctx, int64(spuID))
	if err != nil {
		return nil, err
	}

	return r.toDomainList(rows)
}

// FindPublishedBySpuIDForUpdate locks the published SKUs of an SPU for the
// off-shelf cascade. Uses NOWAIT to fail fast under contention.
func (r *PostgresSkuRepository) FindPublishedBySpuIDForUpdate(
	ctx context.Context,
	spuID uint64,
) ([]model.SkuModel, error) {
	rows, err := r.q.ListPublishedSkusBySpuIDForUpdate(ctx, int64(spuID))
	if err != nil {
		if isPgLockError(err) {
			return nil, errs.ErrConcurrencyConflict
		}
		return nil, err
	}

	return r.toDomainList(rows)
}

func (r *PostgresSkuRepository) toDomainList(rows []sqlcgen.Sku) ([]model.SkuModel, error) {
	skus := make([]model.SkuModel, 0, len(rows))
	for _, row := range rows {
		sku, err := r.mapper.ToDomain(row)
		if err != nil {
			return nil, err
		}
		skus = append(skus, sku)
	}

	return skus, nil
}
