package command

import (
	"context"
	"time"

	"auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/infra/event/envelope"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

type StartAuctionCommandInput struct {
	AuctionID uint64
}

type StartAuctionCommandOutput struct {
	ID        uint64
	ListingID uint64
	State     string
	StartTime *time.Time
	EndTime   time.Time
	UpdatedAt time.Time
}

type StartAuctionCommand struct {
	uowFactory ports.AuctionUnitOfWorkFactory
	logger     logger.Logger
}

func NewStartAuctionCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	logger logger.Logger,
) *StartAuctionCommand {
	return &StartAuctionCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (c *StartAuctionCommand) Execute(
	ctx context.Context,
	input StartAuctionCommandInput,
) (StartAuctionCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return StartAuctionCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	auction, err := uow.AuctionRepository().FindByIDForUpdate(ctx, input.AuctionID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to find auction for update")
		return StartAuctionCommandOutput{}, err
	}

	err = auction.Start()
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to start auction")
		return StartAuctionCommandOutput{}, err
	}

	err = uow.AuctionRepository().Update(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to update auction")
		return StartAuctionCommandOutput{}, err
	}

	// Record the event in the transactional outbox so it commits atomically
	// with the state change; the outbox relay delivers it to JetStream.
	auctionStartedEvent := event.NewAuctionStartedEvent(
		auction.ID(),
		auction.ListingID(),
		auction.StartTime(),
		auction.EndTime(),
	)
	outboxEvent, err := envelope.FromAuctionStarted(auctionStartedEvent)
	if err != nil {
		c.logger.Error().
			Err(err).
			Uint64("auction_id", input.AuctionID).
			Msg("failed to build AuctionStartedEvent envelope")
		return StartAuctionCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Str("event_id", auctionStartedEvent.EventID()).
			Msg("failed to save AuctionStartedEvent to outbox")
		return StartAuctionCommandOutput{}, err
	}

	err = uow.Complete(ctx)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to complete unit of work")
		return StartAuctionCommandOutput{}, err
	}

	state := auction.State()
	return StartAuctionCommandOutput{
		ID:        auction.ID(),
		ListingID: auction.ListingID(),
		State:     state.String(),
		StartTime: auction.StartTime(),
		EndTime:   auction.EndTime(),
		UpdatedAt: auction.UpdatedAt(),
	}, nil
}
