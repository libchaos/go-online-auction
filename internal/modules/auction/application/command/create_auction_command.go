package command

import (
	"context"
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

type CreateAuctionCommandInput struct {
	ListingID          uint64
	EndTime            time.Time
	TradingMode        string
	StartingPrice      *uint64
	PriceStep          *uint64
	ReservePrice       *uint64
	AntiSnipeEnabled   bool
	ExtensionWindowSec int64
	// StartTime optionally schedules the auction to be started automatically
	// by the auction scheduler; nil means the auction is started manually.
	StartTime *time.Time
}

type CreateAuctionCommandOutput struct {
	ID                 uint64
	ListingID          uint64
	State              string
	TradingMode        string
	StartingPrice      *uint64
	PriceStep          *uint64
	ReservePrice       *uint64
	AntiSnipeEnabled   bool
	ExtensionWindowSec int64
	StartTime          *time.Time
	EndTime            time.Time
	CreatedAt          time.Time
}

type CreateAuctionCommand struct {
	auctionRepository ports.AuctionRepository
	listingValidator  ports.ListingValidator
	logger            logger.Logger
}

func NewCreateAuctionCommand(
	auctionRepository ports.AuctionRepository,
	listingValidator ports.ListingValidator,
	logger logger.Logger,
) *CreateAuctionCommand {
	return &CreateAuctionCommand{
		auctionRepository: auctionRepository,
		listingValidator:  listingValidator,
		logger:            logger,
	}
}

func (c *CreateAuctionCommand) Execute(
	ctx context.Context,
	input CreateAuctionCommandInput,
) (CreateAuctionCommandOutput, error) {
	auctionable, err := c.listingValidator.IsAuctionable(ctx, input.ListingID)
	if err != nil {
		c.logger.Error().Err(err).Uint64("listing_id", input.ListingID).Msg("failed to validate listing")
		return CreateAuctionCommandOutput{}, err
	}
	if !auctionable {
		return CreateAuctionCommandOutput{}, errs.ErrListingNotAvailable
	}

	tradingMode, err := enum.NewTradingModeEnum(input.TradingMode)
	if err != nil {
		c.logger.Error().Err(err).Msg("invalid trading mode")
		return CreateAuctionCommandOutput{}, err
	}

	auction, err := model.NewAuctionModelWithMode(
		input.ListingID,
		input.EndTime,
		tradingMode,
		input.StartingPrice,
		input.PriceStep,
		input.ReservePrice,
		input.AntiSnipeEnabled,
		input.ExtensionWindowSec,
		input.StartTime,
	)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create auction domain model")
		return CreateAuctionCommandOutput{}, err
	}

	persistedAuction, err := c.auctionRepository.Create(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist auction")
		return CreateAuctionCommandOutput{}, err
	}

	state := persistedAuction.State()
	tradingMode = persistedAuction.TradingMode()
	return CreateAuctionCommandOutput{
		ID:                 persistedAuction.ID(),
		ListingID:          persistedAuction.ListingID(),
		State:              state.String(),
		TradingMode:        tradingMode.String(),
		StartingPrice:      persistedAuction.StartingPrice(),
		PriceStep:          persistedAuction.PriceStep(),
		ReservePrice:       persistedAuction.ReservePrice(),
		AntiSnipeEnabled:   persistedAuction.AntiSnipeEnabled(),
		ExtensionWindowSec: persistedAuction.ExtensionWindowSec(),
		StartTime:          persistedAuction.StartTime(),
		EndTime:            persistedAuction.EndTime(),
		CreatedAt:          persistedAuction.CreatedAt(),
	}, nil
}
