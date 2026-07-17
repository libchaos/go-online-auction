package query

import (
	"context"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type GetSkuByIDQueryInput struct {
	ID uint64
}

type GetSkuByIDQueryOutput struct {
	Sku SkuOutput
}

type GetSkuByIDQuery struct {
	skuRepository ports.SkuRepository
	logger        logger.Logger
}

func NewGetSkuByIDQuery(
	skuRepository ports.SkuRepository,
	logger logger.Logger,
) *GetSkuByIDQuery {
	return &GetSkuByIDQuery{
		skuRepository: skuRepository,
		logger:        logger,
	}
}

func (q *GetSkuByIDQuery) Execute(
	ctx context.Context,
	input GetSkuByIDQueryInput,
) (GetSkuByIDQueryOutput, error) {
	sku, err := q.skuRepository.FindByID(ctx, input.ID)
	if err != nil {
		q.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to find sku")
		return GetSkuByIDQueryOutput{}, err
	}

	status := sku.Status()
	return GetSkuByIDQueryOutput{
		Sku: SkuOutput{
			ID:           sku.ID(),
			SpuID:        sku.SpuID(),
			SpecValues:   sku.SpecValues(),
			PriceInCents: sku.PriceInCents(),
			Quantity:     sku.Quantity(),
			Status:       status.String(),
			Version:      sku.Version(),
			CreatedAt:    sku.CreatedAt(),
			UpdatedAt:    sku.UpdatedAt(),
		},
	}, nil
}
