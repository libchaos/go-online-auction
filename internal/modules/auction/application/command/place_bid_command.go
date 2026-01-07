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

	// Domain validation provides fast-fail for invalid bids
	// DB trigger with FOR UPDATE ensures final integrity against race conditions
	auction, err := uow.AuctionRepository().FindByID(ctx, input.AuctionID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to find auction")
		return PlaceBidCommandOutput{}, err
	}

	money := model.NewMoneyModel(input.AmountInCents)

	bid, err := model.NewBidModel(input.AuctionID, input.UserID, money)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Uint64("user_id", input.UserID).
			Msg("failed to create bid model")
		return PlaceBidCommandOutput{}, err
	}

	err = auction.PlaceBid(money)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Uint64("user_id", input.UserID).
			Msg("failed to place bid on auction")
		return PlaceBidCommandOutput{}, err
	}

	persistedBid, err := uow.BidRepository().Create(ctx, bid)
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
		persistedBid.ID(),
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
		ID:            persistedBid.ID(),
		AuctionID:     persistedBid.AuctionID(),
		UserID:        persistedBid.UserID(),
		AmountInCents: persistedBid.Amount().AmountInCents(),
		CreatedAt:     persistedBid.CreatedAt(),
	}, nil
}
