package command

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
)

type PlaceBidCommandInput struct {
	AuctionID     uint64
	UserID        uint64
	AmountInCents uint64
}

type PlaceBidCommandOutput struct {
	ID            uint64
	AuctionID     uint64
	UserID        uint64
	AmountInCents uint64
	CreatedAt     time.Time
}

type PlaceBidCommand struct {
	uowFactory               ports.AuctionUnitOfWorkFactory
	bidPlacedEventDispatcher ports.BidPlacedEventDispatcher
	logger                   logger.Logger
}

func NewPlaceBidCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	bidPlacedEventDispatcher ports.BidPlacedEventDispatcher,
	logger logger.Logger,
) *PlaceBidCommand {
	return &PlaceBidCommand{
		uowFactory:               uowFactory,
		bidPlacedEventDispatcher: bidPlacedEventDispatcher,
		logger:                   logger,
	}
}

func (c *PlaceBidCommand) Execute(
	ctx context.Context,
	input PlaceBidCommandInput,
) (PlaceBidCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return PlaceBidCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	auction, err := uow.AuctionRepository().FindByIDForUpdate(ctx, input.AuctionID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to find auction for update")
		return PlaceBidCommandOutput{}, err
	}

	money := model.NewMoneyModel(input.AmountInCents)

	var currentHighestBid *model.BidModel
	if highest, ok := c.restoreCurrentHighestBid(auction, input.AuctionID); ok {
		currentHighestBid = &highest
	}

	bid, err := model.NewBidModel(input.AuctionID, input.UserID, money)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Uint64("user_id", input.UserID).
			Msg("failed to create bid model")
		return PlaceBidCommandOutput{}, err
	}

	err = auction.PlaceBid(bid.ID(), money, currentHighestBid)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Uint64("user_id", input.UserID).
			Msg("failed to place bid on auction")
		return PlaceBidCommandOutput{}, err
	}

	err = uow.BidRepository().Create(ctx, bid)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Uint64("user_id", input.UserID).
			Msg("failed to persist bid")
		return PlaceBidCommandOutput{}, err
	}

	err = uow.AuctionRepository().Update(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to update auction")
		return PlaceBidCommandOutput{}, err
	}

	err = uow.Complete(ctx)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to complete unit of work")
		return PlaceBidCommandOutput{}, err
	}

	bidPlacedEvent := event.NewBidPlacedEvent(
		bid.ID(),
		auction.ID(),
		input.UserID,
		money,
	)

	err = c.bidPlacedEventDispatcher.Dispatch(ctx, bidPlacedEvent)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Str("event_id", bidPlacedEvent.EventID()).
			Msg("failed to dispatch BidPlacedEvent")
		return PlaceBidCommandOutput{}, err
	}

	return PlaceBidCommandOutput{
		ID:            bid.ID(),
		AuctionID:     bid.AuctionID(),
		UserID:        bid.UserID(),
		AmountInCents: bid.Amount().AmountInCents(),
		CreatedAt:     bid.CreatedAt(),
	}, nil
}

func (c *PlaceBidCommand) restoreCurrentHighestBid(
	auction model.AuctionModel,
	auctionID uint64,
) (model.BidModel, bool) {
	if auction.HighestBidID() == nil || auction.HighestBidAmount() == nil {
		return model.BidModel{}, false
	}

	highestMoney := model.NewMoneyModel(*auction.HighestBidAmount())
	highest, err := model.RestoreBidModel(
		*auction.HighestBidID(),
		auction.ID(),
		0,
		highestMoney,
		time.Time{},
		time.Time{},
	)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", auctionID).
			Uint64("highest_bid_id", *auction.HighestBidID()).
			Msg("failed to restore current highest bid")
		return model.BidModel{}, false
	}

	return highest, true
}
