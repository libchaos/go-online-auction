package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type CreateSkuCommandInput struct {
	SpuID        uint64
	SpecValues   map[string]string
	PriceInCents uint64
	Quantity     uint64
}

type CreateSkuCommandOutput struct {
	ID           uint64
	SpuID        uint64
	SpecValues   map[string]string
	PriceInCents uint64
	Quantity     uint64
	Status       string
	CreatedAt    time.Time
}

type CreateSkuCommand struct {
	skuRepository ports.SkuRepository
	spuRepository ports.SpuRepository
	logger        logger.Logger
}

func NewCreateSkuCommand(
	skuRepository ports.SkuRepository,
	spuRepository ports.SpuRepository,
	logger logger.Logger,
) *CreateSkuCommand {
	return &CreateSkuCommand{
		skuRepository: skuRepository,
		spuRepository: spuRepository,
		logger:        logger,
	}
}

func (c *CreateSkuCommand) Execute(
	ctx context.Context,
	input CreateSkuCommandInput,
) (CreateSkuCommandOutput, error) {
	if _, err := c.spuRepository.FindByID(ctx, input.SpuID); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.SpuID).Msg("failed to find spu")
		return CreateSkuCommandOutput{}, err
	}

	sku, err := model.NewSkuModel(input.SpuID, input.SpecValues, input.PriceInCents, input.Quantity)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create sku domain model")
		return CreateSkuCommandOutput{}, err
	}

	persisted, err := c.skuRepository.Create(ctx, sku)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist sku")
		return CreateSkuCommandOutput{}, err
	}

	status := persisted.Status()
	return CreateSkuCommandOutput{
		ID:           persisted.ID(),
		SpuID:        persisted.SpuID(),
		SpecValues:   persisted.SpecValues(),
		PriceInCents: persisted.PriceInCents(),
		Quantity:     persisted.Quantity(),
		Status:       status.String(),
		CreatedAt:    persisted.CreatedAt(),
	}, nil
}
