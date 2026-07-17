package command

import (
	"context"
	"time"

	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/infra/event/envelope"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type PublishSpuCommandInput struct {
	ID uint64
}

type PublishSpuCommandOutput struct {
	ID        uint64
	Status    string
	UpdatedAt time.Time
}

type PublishSpuCommand struct {
	uowFactory ports.ListingUnitOfWorkFactory
	logger     logger.Logger
}

func NewPublishSpuCommand(
	uowFactory ports.ListingUnitOfWorkFactory,
	logger logger.Logger,
) *PublishSpuCommand {
	return &PublishSpuCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (c *PublishSpuCommand) Execute(
	ctx context.Context,
	input PublishSpuCommandInput,
) (PublishSpuCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return PublishSpuCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	spu, err := uow.SpuRepository().FindByIDForUpdate(ctx, input.ID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to find spu for update")
		return PublishSpuCommandOutput{}, err
	}

	if err = spu.Publish(); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to publish spu")
		return PublishSpuCommandOutput{}, err
	}

	if err = uow.SpuRepository().Update(ctx, spu); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to update spu")
		return PublishSpuCommandOutput{}, err
	}

	// Record the event in the transactional outbox so it commits atomically
	// with the state change; the outbox relay delivers it to JetStream.
	publishedEvent := event.NewSpuPublishedEvent(spu.ID())
	outboxEvent, err := envelope.FromSpuPublished(publishedEvent)
	if err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to build SpuPublishedEvent envelope")
		return PublishSpuCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("spu_id", input.ID).
			Str("event_id", publishedEvent.EventID()).
			Msg("failed to save SpuPublishedEvent to outbox")
		return PublishSpuCommandOutput{}, err
	}

	if err = uow.Complete(ctx); err != nil {
		c.logger.Error().Err(err).Uint64("spu_id", input.ID).Msg("failed to complete unit of work")
		return PublishSpuCommandOutput{}, err
	}

	status := spu.Status()
	return PublishSpuCommandOutput{
		ID:        spu.ID(),
		Status:    status.String(),
		UpdatedAt: spu.UpdatedAt(),
	}, nil
}
