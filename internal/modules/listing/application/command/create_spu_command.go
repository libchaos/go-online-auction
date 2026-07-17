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

type CreateSpuCommandInput struct {
	Title       string
	Description string
	CategoryID  uint64
	Brand       *string
	Images      []string
}

type CreateSpuCommandOutput struct {
	ID          uint64
	Title       string
	Description string
	CategoryID  uint64
	Brand       *string
	Images      []string
	Status      string
	CreatedAt   time.Time
}

type CreateSpuCommand struct {
	spuRepository      ports.SpuRepository
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewCreateSpuCommand(
	spuRepository ports.SpuRepository,
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *CreateSpuCommand {
	return &CreateSpuCommand{
		spuRepository:      spuRepository,
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (c *CreateSpuCommand) Execute(
	ctx context.Context,
	input CreateSpuCommandInput,
) (CreateSpuCommandOutput, error) {
	if input.CategoryID == 0 {
		return CreateSpuCommandOutput{}, errs.ErrSpuCategoryRequired
	}

	if _, err := c.categoryRepository.FindByID(ctx, input.CategoryID); err != nil {
		if errors.Is(err, errs.ErrCategoryNotFound) {
			return CreateSpuCommandOutput{}, errs.ErrCategoryNotFound
		}
		c.logger.Error().Err(err).Uint64("category_id", input.CategoryID).Msg("failed to find category")
		return CreateSpuCommandOutput{}, err
	}

	spu, err := model.NewSpuModel(input.Title, input.Description, input.CategoryID, input.Brand, input.Images)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create spu domain model")
		return CreateSpuCommandOutput{}, err
	}

	persisted, err := c.spuRepository.Create(ctx, spu)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist spu")
		return CreateSpuCommandOutput{}, err
	}

	status := persisted.Status()
	return CreateSpuCommandOutput{
		ID:          persisted.ID(),
		Title:       persisted.Title(),
		Description: persisted.Description(),
		CategoryID:  persisted.CategoryID(),
		Brand:       persisted.Brand(),
		Images:      persisted.Images(),
		Status:      status.String(),
		CreatedAt:   persisted.CreatedAt(),
	}, nil
}
