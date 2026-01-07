package command

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
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
	uowFactory                    ports.AuctionUnitOfWorkFactory
	auctionStartedEventDispatcher ports.AuctionStartedEventDispatcher
	logger                        logger.Logger
}

func NewStartAuctionCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	auctionStartedEventDispatcher ports.AuctionStartedEventDispatcher,
	logger logger.Logger,
) *StartAuctionCommand {
	return &StartAuctionCommand{
		uowFactory:                    uowFactory,
		auctionStartedEventDispatcher: auctionStartedEventDispatcher,
		logger:                        logger,
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

	err = uow.Complete(ctx)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to complete unit of work")
		return StartAuctionCommandOutput{}, err
	}

	auctionStartedEvent := event.NewAuctionStartedEvent(
		auction.ID(),
		auction.ListingID(),
		auction.StartTime(),
		auction.EndTime(),
	)

	err = c.auctionStartedEventDispatcher.Dispatch(ctx, auctionStartedEvent)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Str("event_id", auctionStartedEvent.EventID()).
			Msg("failed to dispatch AuctionStartedEvent")
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
