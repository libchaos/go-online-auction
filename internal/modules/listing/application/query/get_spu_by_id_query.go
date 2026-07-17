package query

import (
	"context"
	"time"

	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type GetSpuByIDQueryInput struct {
	ID uint64
}

type GetSpuByIDQueryOutput struct {
	ID          uint64
	Title       string
	Description string
	CategoryID  uint64
	Brand       *string
	Images      []string
	Status      string
	Version     uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Skus        []SkuOutput
}

type SkuOutput struct {
	ID           uint64
	SpuID        uint64
	SpecValues   map[string]string
	PriceInCents uint64
	Quantity     uint64
	Status       string
	Version      uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GetSpuByIDQuery struct {
	spuRepository ports.SpuRepository
	skuRepository ports.SkuRepository
	logger        logger.Logger
}

func NewGetSpuByIDQuery(
	spuRepository ports.SpuRepository,
	skuRepository ports.SkuRepository,
	logger logger.Logger,
) *GetSpuByIDQuery {
	return &GetSpuByIDQuery{
		spuRepository: spuRepository,
		skuRepository: skuRepository,
		logger:        logger,
	}
}

func (q *GetSpuByIDQuery) Execute(
	ctx context.Context,
	input GetSpuByIDQueryInput,
) (GetSpuByIDQueryOutput, error) {
	spu, err := q.spuRepository.FindByID(ctx, input.ID)
	if err != nil {
		q.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to find spu")
		return GetSpuByIDQueryOutput{}, err
	}

	skus, err := q.skuRepository.FindBySpuID(ctx, spu.ID())
	if err != nil {
		q.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to find skus by spu")
		return GetSpuByIDQueryOutput{}, err
	}

	skuOutputs := make([]SkuOutput, 0, len(skus))
	for _, sku := range skus {
		status := sku.Status()
		skuOutputs = append(skuOutputs, SkuOutput{
			ID:           sku.ID(),
			SpuID:        sku.SpuID(),
			SpecValues:   sku.SpecValues(),
			PriceInCents: sku.PriceInCents(),
			Quantity:     sku.Quantity(),
			Status:       status.String(),
			Version:      sku.Version(),
			CreatedAt:    sku.CreatedAt(),
			UpdatedAt:    sku.UpdatedAt(),
		})
	}

	status := spu.Status()
	return GetSpuByIDQueryOutput{
		ID:          spu.ID(),
		Title:       spu.Title(),
		Description: spu.Description(),
		CategoryID:  spu.CategoryID(),
		Brand:       spu.Brand(),
		Images:      spu.Images(),
		Status:      status.String(),
		Version:     spu.Version(),
		CreatedAt:   spu.CreatedAt(),
		UpdatedAt:   spu.UpdatedAt(),
		Skus:        skuOutputs,
	}, nil
}
