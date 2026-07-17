package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/infra/event/envelope"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type PublishSkuCommandInput struct {
	ID uint64
}

type PublishSkuCommandOutput struct {
	ID        uint64
	SpuID     uint64
	Status    string
	UpdatedAt time.Time
}

type PublishSkuCommand struct {
	uowFactory ports.ListingUnitOfWorkFactory
	logger     logger.Logger
}

func NewPublishSkuCommand(
	uowFactory ports.ListingUnitOfWorkFactory,
	logger logger.Logger,
) *PublishSkuCommand {
	return &PublishSkuCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

// Execute publishes the SKU. The parent SPU must be published; it is locked
// first (SPU before SKU) so a concurrent SPU off-shelf cascade serializes
// against this publish instead of deadlocking.
func (c *PublishSkuCommand) Execute(
	ctx context.Context,
	input PublishSkuCommandInput,
) (PublishSkuCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return PublishSkuCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	sku, err := uow.SkuRepository().FindByID(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to find sku")
		return PublishSkuCommandOutput{}, err
	}

	spu, err := uow.SpuRepository().FindByIDForUpdate(ctx, sku.SpuID())
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", sku.SpuID()).Msg("failed to find parent spu for update")
		return PublishSkuCommandOutput{}, err
	}

	spuStatus := spu.Status()
	if !spuStatus.IsPublished() {
		return PublishSkuCommandOutput{}, errs.ErrSpuMustBePublished
	}

	sku, err = uow.SkuRepository().FindByIDForUpdate(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to find sku for update")
		return PublishSkuCommandOutput{}, err
	}

	if err = sku.Publish(); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to publish sku")
		return PublishSkuCommandOutput{}, err
	}

	if err = uow.SkuRepository().Update(ctx, sku); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to update sku")
		return PublishSkuCommandOutput{}, err
	}

	// Record the event in the transactional outbox so it commits atomically
	// with the state change; the outbox relay delivers it to JetStream.
	publishedEvent := event.NewSkuPublishedEvent(sku.ID(), sku.SpuID(), sku.PriceInCents(), sku.Quantity())
	outboxEvent, err := envelope.FromSkuPublished(publishedEvent)
	if err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to build SkuPublishedEvent envelope")
		return PublishSkuCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("sku_id", input.ID).
			Str("event_id", publishedEvent.EventID()).
			Msg("failed to save SkuPublishedEvent to outbox")
		return PublishSkuCommandOutput{}, err
	}

	if err = uow.Complete(ctx); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to complete unit of work")
		return PublishSkuCommandOutput{}, err
	}

	status := sku.Status()
	return PublishSkuCommandOutput{
		ID:        sku.ID(),
		SpuID:     sku.SpuID(),
		Status:    status.String(),
		UpdatedAt: sku.UpdatedAt(),
	}, nil
}
