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

var _ ports.SpuRepository = (*PostgresSpuRepository)(nil)

type PostgresSpuRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.SpuMapper
}

func NewPostgresSpuRepository(db sqlcgen.DBTX, mapper *mapper.SpuMapper) *PostgresSpuRepository {
	return &PostgresSpuRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresSpuRepository) Create(ctx context.Context, spu model.SpuModel) (model.SpuModel, error) {
	params, err := r.mapper.ToCreateParams(spu)
	if err != nil {
		return model.SpuModel{}, err
	}

	row, err := r.q.CreateSpu(ctx, params)
	if err != nil {
		return model.SpuModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresSpuRepository) FindByID(ctx context.Context, id uint64) (model.SpuModel, error) {
	row, err := r.q.GetSpuByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SpuModel{}, errs.ErrSpuNotFound
		}
		return model.SpuModel{}, err
	}

	return r.mapper.ToDomain(row)
}

// FindByIDForUpdate retrieves an SPU with row-level lock for update.
// Uses NOWAIT to fail fast under contention.
func (r *PostgresSpuRepository) FindByIDForUpdate(ctx context.Context, id uint64) (model.SpuModel, error) {
	row, err := r.q.GetSpuByIDForUpdate(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SpuModel{}, errs.ErrSpuNotFound
		}
		if isPgLockError(err) {
			return model.SpuModel{}, errs.ErrConcurrencyConflict
		}
		return model.SpuModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresSpuRepository) Update(ctx context.Context, spu model.SpuModel) error {
	params, err := r.mapper.ToUpdateParams(spu)
	if err != nil {
		return err
	}
	params.PreviousVersion = params.Version - 1 // Domain increments version before calling Update

	rowsAffected, err := r.q.UpdateSpu(ctx, params)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrConcurrencyConflict
	}

	return nil
}

func (r *PostgresSpuRepository) List(ctx context.Context, filter ports.ListSpusFilter) ([]model.SpuModel, error) {
	params := sqlcgen.ListSpusParams{
		Status:     toNullListingStatus(filter.Status),
		CategoryID: toNullableInt64Filter(filter.CategoryID),
		Limit:      int32(filter.Limit),
		Offset:     int32(filter.Offset),
	}

	rows, err := r.q.ListSpus(ctx, params)
	if err != nil {
		return nil, err
	}

	spus := make([]model.SpuModel, 0, len(rows))
	for _, row := range rows {
		spu, mapErr := r.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}
		spus = append(spus, spu)
	}

	return spus, nil
}

func (r *PostgresSpuRepository) Count(ctx context.Context, filter ports.ListSpusFilter) (uint64, error) {
	count, err := r.q.CountSpus(ctx, sqlcgen.CountSpusParams{
		Status:     toNullListingStatus(filter.Status),
		CategoryID: toNullableInt64Filter(filter.CategoryID),
	})
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}
