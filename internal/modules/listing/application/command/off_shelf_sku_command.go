package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/infra/event/envelope"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type OffShelfSkuCommandInput struct {
	ID uint64
}

type OffShelfSkuCommandOutput struct {
	ID        uint64
	SpuID     uint64
	Status    string
	UpdatedAt time.Time
}

type OffShelfSkuCommand struct {
	uowFactory ports.ListingUnitOfWorkFactory
	logger     logger.Logger
}

func NewOffShelfSkuCommand(
	uowFactory ports.ListingUnitOfWorkFactory,
	logger logger.Logger,
) *OffShelfSkuCommand {
	return &OffShelfSkuCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (c *OffShelfSkuCommand) Execute(
	ctx context.Context,
	input OffShelfSkuCommandInput,
) (OffShelfSkuCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return OffShelfSkuCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	sku, err := uow.SkuRepository().FindByIDForUpdate(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to find sku for update")
		return OffShelfSkuCommandOutput{}, err
	}

	if err = sku.OffShelf(); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to take sku off shelf")
		return OffShelfSkuCommandOutput{}, err
	}

	if err = uow.SkuRepository().Update(ctx, sku); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to update sku")
		return OffShelfSkuCommandOutput{}, err
	}

	// Record the event in the transactional outbox so it commits atomically
	// with the state change; the outbox relay delivers it to JetStream.
	offShelfEvent := event.NewSkuOffShelfEvent(sku.ID(), sku.SpuID())
	outboxEvent, err := envelope.FromSkuOffShelf(offShelfEvent)
	if err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to build SkuOffShelfEvent envelope")
		return OffShelfSkuCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("sku_id", input.ID).
			Str("event_id", offShelfEvent.EventID()).
			Msg("failed to save SkuOffShelfEvent to outbox")
		return OffShelfSkuCommandOutput{}, err
	}

	if err = uow.Complete(ctx); err != nil {
		c.logger.Error().Err(err).Uint64("sku_id", input.ID).Msg("failed to complete unit of work")
		return OffShelfSkuCommandOutput{}, err
	}

	status := sku.Status()
	return OffShelfSkuCommandOutput{
		ID:        sku.ID(),
		SpuID:     sku.SpuID(),
		Status:    status.String(),
		UpdatedAt: sku.UpdatedAt(),
	}, nil
}
