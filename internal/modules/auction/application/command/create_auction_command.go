package command

import (
	"context"
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/event/envelope"
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
	uowFactory       ports.AuctionUnitOfWorkFactory
	listingValidator ports.ListingValidator
	resolver         strategy.Resolver
	logger           logger.Logger
}

func NewCreateAuctionCommand(
	uowFactory ports.AuctionUnitOfWorkFactory,
	listingValidator ports.ListingValidator,
	resolver strategy.Resolver,
	logger logger.Logger,
) *CreateAuctionCommand {
	return &CreateAuctionCommand{
		uowFactory:       uowFactory,
		listingValidator: listingValidator,
		resolver:         resolver,
		logger:           logger,
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
		c.resolver,
	)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create auction domain model")
		return CreateAuctionCommandOutput{}, err
	}

	uow, err := c.uowFactory.Begin(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to begin unit of work")
		return CreateAuctionCommandOutput{}, err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	persistedAuction, err := uow.AuctionRepository().Create(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist auction")
		return CreateAuctionCommandOutput{}, err
	}

	// Record the creation event in the transactional outbox so it commits
	// atomically with the row insert; the outbox relay delivers it to JetStream.
	createdTradingMode := persistedAuction.TradingMode()
	auctionCreatedEvent := event.NewAuctionCreatedEvent(
		persistedAuction.ID(),
		persistedAuction.ListingID(),
		createdTradingMode.String(),
		persistedAuction.StartingPrice(),
		persistedAuction.PriceStep(),
		persistedAuction.ReservePrice(),
		persistedAuction.AntiSnipeEnabled(),
		persistedAuction.ExtensionWindowSec(),
		persistedAuction.StartTime(),
		persistedAuction.EndTime(),
	)
	outboxEvent, err := envelope.FromAuctionCreated(auctionCreatedEvent)
	if err != nil {
		c.logger.Error().
			Err(err).
			Uint64("auction_id", persistedAuction.ID()).
			Msg("failed to build AuctionCreatedEvent envelope")
		return CreateAuctionCommandOutput{}, err
	}
	if err = uow.OutboxRepository().Save(ctx, outboxEvent); err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", persistedAuction.ID()).
			Str("event_id", auctionCreatedEvent.EventID()).
			Msg("failed to save AuctionCreatedEvent to outbox")
		return CreateAuctionCommandOutput{}, err
	}

	if err = uow.Complete(ctx); err != nil {
		c.logger.Error().Err(err).Uint64("auction_id", persistedAuction.ID()).Msg("failed to complete unit of work")
		return CreateAuctionCommandOutput{}, err
	}

	state := persistedAuction.State()
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
