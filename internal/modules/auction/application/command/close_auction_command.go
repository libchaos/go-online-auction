package command

import (
	"context"
	"time"

	"auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/event/envelope"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

type CloseAuctionCommandInput struct {
	AuctionID uint64
}

type CloseAuctionCommandOutput struct {
	ID        uint64
	ListingID uint64
	State     string
	StartTime *time.Time
	EndTime   time.Time
	UpdatedAt time.Time
}

type CloseAuctionCommand struct {
	uowFactory ports.AuctionUnitOfWorkFactory
	logger     logger.Logger
}

func NewCloseAuctionCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	logger logger.Logger,
) *CloseAuctionCommand {
	return &CloseAuctionCommand{
		uowFactory: uowFactory,
		logger:     logger,
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

	bids, err := uow.BidRepository().FindByAuctionID(ctx, auction.ID())
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to find bids for auction")
		return CloseAuctionCommandOutput{}, err
	}

	err = auction.Close(bids)
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
	if auction.WinnerUserID() != nil && auction.WinningBidAmount() != nil {
		winningBidID = auction.WinningBidID()
		amount := model.NewMoneyModel(*auction.WinningBidAmount())
		finalAmount = &amount
	}

	// Record the event in the transactional outbox so it commits atomically
	// with the state change; the outbox relay delivers it to JetStream.
	auctionEndedEvent := event.NewAuctionEndedEvent(
		auction.ID(),
		winningBidID,
		finalAmount,
	)
	outboxEvent, err := envelope.FromAuctionEnded(auctionEndedEvent)
	if err != nil {
		c.logger.Error().
			Err(err).
			Uint64("auction_id", input.AuctionID).
			Msg("failed to build AuctionEndedEvent envelope")
		return CloseAuctionCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Str("event_id", auctionEndedEvent.EventID()).
			Msg("failed to save AuctionEndedEvent to outbox")
		return CloseAuctionCommandOutput{}, err
	}

	err = uow.Complete(ctx)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to complete unit of work")
		return CloseAuctionCommandOutput{}, err
	}

	state := auction.State()
	return CloseAuctionCommandOutput{
		ID:        auction.ID(),
		ListingID: auction.ListingID(),
		State:     state.String(),
		StartTime: auction.StartTime(),
		EndTime:   auction.EndTime(),
		UpdatedAt: auction.UpdatedAt(),
	}, nil
}
