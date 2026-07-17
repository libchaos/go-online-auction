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
	Depth     int32
	Path      string
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
		parent, err := c.categoryRepository.FindByID(ctx, *input.ParentID)
		if err != nil {
			if errors.Is(err, errs.ErrCategoryNotFound) {
				return CreateCategoryCommandOutput{}, errs.ErrCategoryParentNotFound
			}
			c.logger.Error().Err(err).Uint64("parent_id", *input.ParentID).Msg("failed to find parent category")
			return CreateCategoryCommandOutput{}, err
		}

		if parent.Depth()+1 > model.MaxCategoryDepth {
			c.logger.Error().
				Uint64("parent_id", *input.ParentID).
				Int32("parent_depth", parent.Depth()).
				Msg("category depth exceeded")
			return CreateCategoryCommandOutput{}, errs.ErrCategoryDepthExceeded
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
		Depth:     persisted.Depth(),
		Path:      persisted.Path(),
		SortOrder: persisted.SortOrder(),
		CreatedAt: persisted.CreatedAt(),
	}, nil
}
