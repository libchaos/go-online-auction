package command

import (
	"context"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type DeleteCategoryCommandInput struct {
	ID uint64
}

type DeleteCategoryCommand struct {
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewDeleteCategoryCommand(
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *DeleteCategoryCommand {
	return &DeleteCategoryCommand{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (c *DeleteCategoryCommand) Execute(ctx context.Context, input DeleteCategoryCommandInput) error {
	if _, err := c.categoryRepository.FindByID(ctx, input.ID); err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to find category")
		return err
	}

	children, err := c.categoryRepository.CountChildren(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to count category children")
		return err
	}
	if children > 0 {
		return errs.ErrCategoryHasChildren
	}

	spus, err := c.categoryRepository.CountSpusByCategory(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to count spus by category")
		return err
	}
	if spus > 0 {
		return errs.ErrCategoryInUse
	}

	if err = c.categoryRepository.Delete(ctx, input.ID); err != nil {
		c.logger.Error().Err(err).Uint64("category_id", input.ID).Msg("failed to delete category")
		return err
	}

	return nil
}
