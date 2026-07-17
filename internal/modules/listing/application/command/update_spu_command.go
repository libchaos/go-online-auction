package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type UpdateSpuCommandInput struct {
	ID          uint64
	Title       string
	Description string
	CategoryID  uint64
	Brand       *string
	Images      []string
}

type UpdateSpuCommandOutput struct {
	ID          uint64
	Title       string
	Description string
	CategoryID  uint64
	Brand       *string
	Images      []string
	Status      string
	UpdatedAt   time.Time
}

type UpdateSpuCommand struct {
	spuRepository      ports.SpuRepository
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewUpdateSpuCommand(
	spuRepository ports.SpuRepository,
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *UpdateSpuCommand {
	return &UpdateSpuCommand{
		spuRepository:      spuRepository,
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (c *UpdateSpuCommand) Execute(
	ctx context.Context,
	input UpdateSpuCommandInput,
) (UpdateSpuCommandOutput, error) {
	spu, err := c.spuRepository.FindByID(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to find spu")
		return UpdateSpuCommandOutput{}, err
	}

	if input.CategoryID != spu.CategoryID() {
		if _, err = c.categoryRepository.FindByID(ctx, input.CategoryID); err != nil {
			c.logger.Error().Err(err).Uint64("category_id", input.CategoryID).Msg("failed to find category")
			return UpdateSpuCommandOutput{}, err
		}
	}

	if err = spu.Update(input.Title, input.Description, input.CategoryID, input.Brand, input.Images); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to update spu domain model")
		return UpdateSpuCommandOutput{}, err
	}

	if err = c.spuRepository.Update(ctx, spu); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to persist spu")
		return UpdateSpuCommandOutput{}, err
	}

	status := spu.Status()
	return UpdateSpuCommandOutput{
		ID:          spu.ID(),
		Title:       spu.Title(),
		Description: spu.Description(),
		CategoryID:  spu.CategoryID(),
		Brand:       spu.Brand(),
		Images:      spu.Images(),
		Status:      status.String(),
		UpdatedAt:   spu.UpdatedAt(),
	}, nil
}
