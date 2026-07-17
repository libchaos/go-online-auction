package command

import (
	"context"
	"errors"
	"time"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type CreateCategoryCommandInput struct {
	Name      string
	ParentID  *uint64
	SortOrder int32
}

type CreateCategoryCommandOutput struct {
	ID        uint64
	Name      string
	ParentID  *uint64
	SortOrder int32
	CreatedAt time.Time
}

type CreateCategoryCommand struct {
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewCreateCategoryCommand(
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *CreateCategoryCommand {
	return &CreateCategoryCommand{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (c *CreateCategoryCommand) Execute(
	ctx context.Context,
	input CreateCategoryCommandInput,
) (CreateCategoryCommandOutput, error) {
	if input.ParentID != nil {
		_, err := c.categoryRepository.FindByID(ctx, *input.ParentID)
		if err != nil {
			if errors.Is(err, errs.ErrCategoryNotFound) {
				return CreateCategoryCommandOutput{}, errs.ErrCategoryParentNotFound
			}
			c.logger.Error().Err(err).Uint64("parent_id", *input.ParentID).Msg("failed to find parent category")
			return CreateCategoryCommandOutput{}, err
		}
	}

	category, err := model.NewCategoryModel(input.Name, input.ParentID, input.SortOrder)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create category domain model")
		return CreateCategoryCommandOutput{}, err
	}

	persisted, err := c.categoryRepository.Create(ctx, category)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist category")
		return CreateCategoryCommandOutput{}, err
	}

	return CreateCategoryCommandOutput{
		ID:        persisted.ID(),
		Name:      persisted.Name(),
		ParentID:  persisted.ParentID(),
		SortOrder: persisted.SortOrder(),
		CreatedAt: persisted.CreatedAt(),
	}, nil
}
