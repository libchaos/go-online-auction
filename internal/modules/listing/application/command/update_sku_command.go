package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type UpdateSkuCommandInput struct {
	ID           uint64
	SpecValues   map[string]string
	PriceInCents uint64
	Quantity     uint64
}

type UpdateSkuCommandOutput struct {
	ID           uint64
	SpuID        uint64
	SpecValues   map[string]string
	PriceInCents uint64
	Quantity     uint64
	Status       string
	UpdatedAt    time.Time
}

type UpdateSkuCommand struct {
	skuRepository ports.SkuRepository
	logger        logger.Logger
}

func NewUpdateSkuCommand(
	skuRepository ports.SkuRepository,
	logger logger.Logger,
) *UpdateSkuCommand {
	return &UpdateSkuCommand{
		skuRepository: skuRepository,
		logger:        logger,
	}
}

func (c *UpdateSkuCommand) Execute(
	ctx context.Context,
	input UpdateSkuCommandInput,
) (UpdateSkuCommandOutput, error) {
	sku, err := c.skuRepository.FindByID(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to find sku")
		return UpdateSkuCommandOutput{}, err
	}

	if err = sku.Update(input.SpecValues, input.PriceInCents, input.Quantity); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to update sku domain model")
		return UpdateSkuCommandOutput{}, err
	}

	if err = c.skuRepository.Update(ctx, sku); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to persist sku")
		return UpdateSkuCommandOutput{}, err
	}

	status := sku.Status()
	return UpdateSkuCommandOutput{
		ID:           sku.ID(),
		SpuID:        sku.SpuID(),
		SpecValues:   sku.SpecValues(),
		PriceInCents: sku.PriceInCents(),
		Quantity:     sku.Quantity(),
		Status:       status.String(),
		UpdatedAt:    sku.UpdatedAt(),
	}, nil
}
