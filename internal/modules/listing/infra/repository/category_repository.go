package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/mapper"
	"auction/internal/modules/listing/infra/sqlcgen"
	"auction/internal/modules/listing/ports"
)

var _ ports.CategoryRepository = (*PostgresCategoryRepository)(nil)

type PostgresCategoryRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.CategoryMapper
}

func NewPostgresCategoryRepository(db sqlcgen.DBTX, mapper *mapper.CategoryMapper) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresCategoryRepository) Create(
	ctx context.Context,
	category model.CategoryModel,
) (model.CategoryModel, error) {
	inserted, err := r.q.CreateCategory(ctx, r.mapper.ToCreateParams(category))
	if err != nil {
		return model.CategoryModel{}, err
	}

	finalized, err := r.q.FinalizeCategoryHierarchy(ctx, inserted.ID)
	if err != nil {
		return model.CategoryModel{}, err
	}

	return r.mapper.ToDomain(finalized)
}

func (r *PostgresCategoryRepository) FindByID(ctx context.Context, id uint64) (model.CategoryModel, error) {
	row, err := r.q.GetCategoryByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CategoryModel{}, errs.ErrCategoryNotFound
		}
		return model.CategoryModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, category model.CategoryModel) error {
	params := r.mapper.ToUpdateParams(category)
	params.PreviousVersion = params.Version - 1 // Domain increments version before calling Update

	rowsAffected, err := r.q.UpdateCategory(ctx, params)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrConcurrencyConflict
	}

	return nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id uint64) error {
	rowsAffected, err := r.q.DeleteCategory(ctx, int64(id))
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrCategoryNotFound
	}

	return nil
}

func (r *PostgresCategoryRepository) List(ctx context.Context, parentID *uint64) ([]model.CategoryModel, error) {
	var rows []sqlcgen.Category
	var err error

	if parentID != nil {
		rows, err = r.q.ListCategoriesByParent(ctx, int64(*parentID))
	} else {
		rows, err = r.q.ListRootCategories(ctx)
	}
	if err != nil {
		return nil, err
	}

	categories := make([]model.CategoryModel, 0, len(rows))
	for _, row := range rows {
		category, mapErr := r.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *PostgresCategoryRepository) ListAll(ctx context.Context) ([]model.CategoryModel, error) {
	rows, err := r.q.ListAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]model.CategoryModel, 0, len(rows))
	for _, row := range rows {
		category, mapErr := r.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *PostgresCategoryRepository) ListDescendants(ctx context.Context, id uint64) ([]model.CategoryModel, error) {
	rows, err := r.q.ListCategoryDescendants(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	categories := make([]model.CategoryModel, 0, len(rows))
	for _, row := range rows {
		category, mapErr := r.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *PostgresCategoryRepository) CountChildren(ctx context.Context, id uint64) (uint64, error) {
	count, err := r.q.CountCategoryChildren(ctx, int64(id))
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func (r *PostgresCategoryRepository) CountSpusByCategory(ctx context.Context, id uint64) (uint64, error) {
	count, err := r.q.CountSpusByCategory(ctx, int64(id))
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func isPgLockError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 55P03 = lock_not_available
		return pgErr.Code == "55P03"
	}
	return false
}
