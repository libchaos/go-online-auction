package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/event/envelope"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type OffShelfSpuCommandInput struct {
	ID uint64
}

type OffShelfSpuCommandOutput struct {
	ID               uint64
	Status           string
	OffShelfSkuCount int
	UpdatedAt        time.Time
}

type OffShelfSpuCommand struct {
	uowFactory ports.ListingUnitOfWorkFactory
	logger     logger.Logger
}

func NewOffShelfSpuCommand(
	uowFactory ports.ListingUnitOfWorkFactory,
	logger logger.Logger,
) *OffShelfSpuCommand {
	return &OffShelfSpuCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

// Execute takes the SPU off shelf and cascades to all its published SKUs in
// the same transaction, emitting one event per SKU plus the SPU event.
// Lock ordering is always SPU before SKUs to avoid deadlocks.
func (c *OffShelfSpuCommand) Execute(
	ctx context.Context,
	input OffShelfSpuCommandInput,
) (OffShelfSpuCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return OffShelfSpuCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	spu, err := uow.SpuRepository().FindByIDForUpdate(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to find spu for update")
		return OffShelfSpuCommandOutput{}, err
	}

	if err = spu.OffShelf(); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to take spu off shelf")
		return OffShelfSpuCommandOutput{}, err
	}

	if err = uow.SpuRepository().Update(ctx, spu); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to update spu")
		return OffShelfSpuCommandOutput{}, err
	}

	skus, err := uow.SkuRepository().FindPublishedBySpuIDForUpdate(ctx, spu.ID())
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to find published skus for update")
		return OffShelfSpuCommandOutput{}, err
	}

	if err = c.offShelfSkus(ctx, uow, spu.ID(), skus); err != nil {
		return OffShelfSpuCommandOutput{}, err
	}

	spuEvent := event.NewSpuOffShelfEvent(spu.ID())
	outboxEvent, err := envelope.FromSpuOffShelf(spuEvent)
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to build SpuOffShelfEvent envelope")
		return OffShelfSpuCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("spu_id", input.ID).
			Str("event_id", spuEvent.EventID()).
			Msg("failed to save SpuOffShelfEvent to outbox")
		return OffShelfSpuCommandOutput{}, err
	}

	if err = uow.Complete(ctx); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to complete unit of work")
		return OffShelfSpuCommandOutput{}, err
	}

	status := spu.Status()
	return OffShelfSpuCommandOutput{
		ID:               spu.ID(),
		Status:           status.String(),
		OffShelfSkuCount: len(skus),
		UpdatedAt:        spu.UpdatedAt(),
	}, nil
}

// offShelfSkus takes each published SKU off shelf and records one
// SkuOffShelfEvent per SKU in the outbox, all within the ambient transaction.
func (c *OffShelfSpuCommand) offShelfSkus(
	ctx context.Context,
	uow ports.ListingUnitOfWork,
	spuID uint64,
	skus []model.SkuModel,
) error {
	for i := range skus {
		sku := skus[i]
		if err := sku.OffShelf(); err != nil {
			c.logger.Error().Err(err).Uint64("sku_id", sku.ID()).Msg("failed to take sku off shelf")
			return err
		}
		if err := uow.SkuRepository().Update(ctx, sku); err != nil {
			c.logger.Error().Err(err).Uint64("sku_id", sku.ID()).Msg("failed to update sku")
			return err
		}

		skuEvent := event.NewSkuOffShelfEvent(sku.ID(), spuID)
		outboxEvent, err := envelope.FromSkuOffShelf(skuEvent)
		if err != nil {
			c.logger.Error().Err(err).Uint64("sku_id", sku.ID()).Msg("failed to build SkuOffShelfEvent envelope")
			return err
		}
		if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
			c.logger.Error().Err(err).
				Uint64("sku_id", sku.ID()).
				Str("event_id", skuEvent.EventID()).
				Msg("failed to save SkuOffShelfEvent to outbox")
			return err
		}
	}

	return nil
}
