package command

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
)

type CloseAuctionCommandInput struct {
	AuctionID uint64
}

type CloseAuctionCommandOutput struct {
	ID           uint64
	ListingID    uint64
	State        string
	HighestBidID *uint64
	StartTime    time.Time
	EndTime      time.Time
	UpdatedAt    time.Time
}

type CloseAuctionCommand struct {
	uowFactory                  ports.AuctionUnitOfWorkFactory
	auctionEndedEventDispatcher ports.AuctionEndedEventDispatcher
	logger                      logger.Logger
}

func NewCloseAuctionCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	auctionEndedEventDispatcher ports.AuctionEndedEventDispatcher,
	logger logger.Logger,
) *CloseAuctionCommand {
	return &CloseAuctionCommand{
		uowFactory:                  uowFactory,
		auctionEndedEventDispatcher: auctionEndedEventDispatcher,
		logger:                      logger,
	}
}

func (c *CloseAuctionCommand) Execute(
	ctx context.Context,
	input CloseAuctionCommandInput,
) (CloseAuctionCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return CloseAuctionCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	auction, err := uow.AuctionRepository().FindByIDForUpdate(ctx, input.AuctionID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to find auction for update")
		return CloseAuctionCommandOutput{}, err
	}

	err = auction.Close()
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to close auction")
		return CloseAuctionCommandOutput{}, err
	}

	err = uow.AuctionRepository().Update(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to update auction")
		return CloseAuctionCommandOutput{}, err
	}

	var winningBidID *uint64
	var finalAmount *model.MoneyModel
	if auction.HighestBidID() != nil {
		winningBidID = auction.HighestBidID()
		var winningBid model.BidModel
		winningBid, err = uow.BidRepository().FindByID(ctx, *winningBidID)
		if err != nil {
			c.logger.Error().Err(err).
				Uint64("auction_id", input.AuctionID).
				Uint64("winning_bid_id", *winningBidID).
				Msg("failed to find winning bid")
			return CloseAuctionCommandOutput{}, err
		}
		amount := winningBid.Amount()
		finalAmount = &amount
	}

	err = uow.Complete(ctx)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to complete unit of work")
		return CloseAuctionCommandOutput{}, err
	}

	auctionEndedEvent := event.NewAuctionEndedEvent(
		auction.ID(),
		winningBidID,
		finalAmount,
	)

	err = c.auctionEndedEventDispatcher.Dispatch(ctx, auctionEndedEvent)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Str("event_id", auctionEndedEvent.EventID()).
			Msg("failed to dispatch AuctionEndedEvent")
		return CloseAuctionCommandOutput{}, err
	}

	state := auction.State()
	return CloseAuctionCommandOutput{
		ID:           auction.ID(),
		ListingID:    auction.ListingID(),
		State:        state.String(),
		HighestBidID: auction.HighestBidID(),
		StartTime:    auction.StartTime(),
		EndTime:      auction.EndTime(),
		UpdatedAt:    auction.UpdatedAt(),
	}, nil
}
