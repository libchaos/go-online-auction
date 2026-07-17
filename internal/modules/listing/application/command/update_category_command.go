package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type UpdateCategoryCommandInput struct {
	ID        uint64
	Name      string
	SortOrder int32
}

type UpdateCategoryCommandOutput struct {
	ID        uint64
	Name      string
	ParentID  *uint64
	SortOrder int32
	UpdatedAt time.Time
}

type UpdateCategoryCommand struct {
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewUpdateCategoryCommand(
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *UpdateCategoryCommand {
	return &UpdateCategoryCommand{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (c *UpdateCategoryCommand) Execute(
	ctx context.Context,
	input UpdateCategoryCommandInput,
) (UpdateCategoryCommandOutput, error) {
	category, err := c.categoryRepository.FindByID(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to find category")
		return UpdateCategoryCommandOutput{}, err
	}

	if err = category.Update(input.Name, input.SortOrder); err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to update category domain model")
		return UpdateCategoryCommandOutput{}, err
	}

	if err = c.categoryRepository.Update(ctx, category); err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to persist category")
		return UpdateCategoryCommandOutput{}, err
	}

	return UpdateCategoryCommandOutput{
		ID:        category.ID(),
		Name:      category.Name(),
		ParentID:  category.ParentID(),
		SortOrder: category.SortOrder(),
		UpdatedAt: category.UpdatedAt(),
	}, nil
}
