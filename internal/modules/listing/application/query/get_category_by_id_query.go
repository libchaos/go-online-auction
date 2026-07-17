package query

import (
	"context"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type GetCategoryByIDQueryInput struct {
	ID uint64
}

type GetCategoryByIDQueryOutput struct {
	Category CategoryOutput
}

type GetCategoryByIDQuery struct {
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewGetCategoryByIDQuery(
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *GetCategoryByIDQuery {
	return &GetCategoryByIDQuery{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (q *GetCategoryByIDQuery) Execute(
	ctx context.Context,
	input GetCategoryByIDQueryInput,
) (GetCategoryByIDQueryOutput, error) {
	category, err := q.categoryRepository.FindByID(ctx, input.ID)
	if err != nil {
		q.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to find category")
		return GetCategoryByIDQueryOutput{}, err
	}

	return GetCategoryByIDQueryOutput{
		Category: CategoryOutput{
			ID:        category.ID(),
			Name:      category.Name(),
			ParentID:  category.ParentID(),
			SortOrder: category.SortOrder(),
			CreatedAt: category.CreatedAt(),
			UpdatedAt: category.UpdatedAt(),
		},
	}, nil
}
