package command

import (
	"context"
	"time"

	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

type CancelAuctionCommandInput struct {
	AuctionID uint64
}

type CancelAuctionCommandOutput struct {
	ID        uint64
	ListingID uint64
	State     string
	StartTime *time.Time
	EndTime   time.Time
	UpdatedAt time.Time
}

type CancelAuctionCommand struct {
	uowFactory ports.AuctionUnitOfWorkFactory
	logger     logger.Logger
}

func NewCancelAuctionCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	logger logger.Logger,
) *CancelAuctionCommand {
	return &CancelAuctionCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (c *CancelAuctionCommand) Execute(
	ctx context.Context,
	input CancelAuctionCommandInput,
) (CancelAuctionCommandOutput, error) {
	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return CancelAuctionCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	auction, err := uow.AuctionRepository().FindByIDForUpdate(ctx, input.AuctionID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to find auction for update")
		return CancelAuctionCommandOutput{}, err
	}

	err = auction.Cancel()
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to cancel auction")
		return CancelAuctionCommandOutput{}, err
	}

	err = uow.AuctionRepository().Update(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to update auction")
		return CancelAuctionCommandOutput{}, err
	}

	err = uow.Complete(ctx)
	if err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", input.AuctionID).Msg("failed to complete unit of work")
		return CancelAuctionCommandOutput{}, err
	}

	state := auction.State()
	return CancelAuctionCommandOutput{
		ID:        auction.ID(),
		ListingID: auction.ListingID(),
		State:     state.String(),
		StartTime: auction.StartTime(),
		EndTime:   auction.EndTime(),
		UpdatedAt: auction.UpdatedAt(),
	}, nil
}
