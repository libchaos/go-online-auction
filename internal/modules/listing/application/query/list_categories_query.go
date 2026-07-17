package query

import (
	"context"
	"time"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type ListCategoriesQueryInput struct {
	ParentID *uint64
}

type ListCategoriesQueryOutput struct {
	Categories []CategoryOutput
}

type CategoryOutput struct {
	ID        uint64
	Name      string
	ParentID  *uint64
	Depth     int32
	Path      string
	SortOrder int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListCategoriesQuery struct {
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewListCategoriesQuery(
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *ListCategoriesQuery {
	return &ListCategoriesQuery{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (q *ListCategoriesQuery) Execute(
	ctx context.Context,
	input ListCategoriesQueryInput,
) (ListCategoriesQueryOutput, error) {
	categories, err := q.categoryRepository.List(ctx, input.ParentID)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to list categories")
		return ListCategoriesQueryOutput{}, err
	}

	outputs := make([]CategoryOutput, 0, len(categories))
	for _, category := range categories {
		outputs = append(outputs, CategoryOutput{
			ID:        category.ID(),
			Name:      category.Name(),
			ParentID:  category.ParentID(),
			Depth:     category.Depth(),
			Path:      category.Path(),
			SortOrder: category.SortOrder(),
			CreatedAt: category.CreatedAt(),
			UpdatedAt: category.UpdatedAt(),
		})
	}

	return ListCategoriesQueryOutput{Categories: outputs}, nil
}
