package query

import (
	"context"
	"time"

	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

const topBidsLimit = 10

type GetAuctionByIDQueryInput struct {
	ID uint64
}

type GetAuctionByIDQueryOutput struct {
	Auction AuctionOutput
	Bids    []BidOutput
}

type AuctionOutput struct {
	ID                      uint64
	ListingID               uint64
	State                   string
	TradingMode             string
	StartTime               *time.Time
	EndTime                 time.Time
	StartingPrice           *uint64
	PriceStep               *uint64
	ReservePrice            *uint64
	CurrentPrice            *uint64
	HighestBidAmountInCents *uint64
	WinnerUserID            *uint64
	WinningBidID            *uint64
	WinningBidAmountInCents *uint64
	AntiSnipeEnabled        bool
	ExtensionWindowSec      int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type BidOutput struct {
	ID               uint64
	UserID           uint64
	AmountInCents    uint64
	MaxAmountInCents *uint64
	CreatedAt        time.Time
}

type GetAuctionByIDQuery struct {
	auctionRepository ports.AuctionRepository
	bidRepository     ports.BidRepository
	logger            logger.Logger
}

func NewGetAuctionByIDQuery(
	auctionRepository ports.AuctionRepository,
	bidRepository ports.BidRepository,
	logger logger.Logger,
) *GetAuctionByIDQuery {
	return &GetAuctionByIDQuery{
		auctionRepository: auctionRepository,
		bidRepository:     bidRepository,
		logger:            logger,
	}
}

func (q *GetAuctionByIDQuery) Execute(
	ctx context.Context,
	input GetAuctionByIDQueryInput,
) (GetAuctionByIDQueryOutput, error) {
	auction, err := q.auctionRepository.FindByID(ctx, input.ID)
	if err != nil {
		return GetAuctionByIDQueryOutput{}, err
	}

	bids, err := q.bidRepository.FindTopBidsByAuctionID(ctx, input.ID, topBidsLimit)
	if err != nil {
		q.logger.Error().Err(err).Uint64("auction_id", input.ID).Msg("failed to fetch top bids")
		return GetAuctionByIDQueryOutput{}, err
	}

	bidOutputs := make([]BidOutput, 0, len(bids))
	for _, bid := range bids {
		var maxAmountInCents *uint64
		if bid.MaxAmount() != nil {
			amount := bid.MaxAmount().AmountInCents()
			maxAmountInCents = &amount
		}
		bidOutputs = append(bidOutputs, BidOutput{
			ID:               bid.ID(),
			UserID:           bid.UserID(),
			AmountInCents:    bid.Amount().AmountInCents(),
			MaxAmountInCents: maxAmountInCents,
			CreatedAt:        bid.CreatedAt(),
		})
	}

	state := auction.State()
	tradingMode := auction.TradingMode()
	return GetAuctionByIDQueryOutput{
		Auction: AuctionOutput{
			ID:                      auction.ID(),
			ListingID:               auction.ListingID(),
			State:                   state.String(),
			TradingMode:             tradingMode.String(),
			StartTime:               auction.StartTime(),
			EndTime:                 auction.EndTime(),
			StartingPrice:           auction.StartingPrice(),
			PriceStep:               auction.PriceStep(),
			ReservePrice:            auction.ReservePrice(),
			CurrentPrice:            auction.CurrentPrice(),
			HighestBidAmountInCents: auction.HighestBidAmount(),
			WinnerUserID:            auction.WinnerUserID(),
			WinningBidID:            auction.WinningBidID(),
			WinningBidAmountInCents: auction.WinningBidAmount(),
			AntiSnipeEnabled:        auction.AntiSnipeEnabled(),
			ExtensionWindowSec:      auction.ExtensionWindowSec(),
			CreatedAt:               auction.CreatedAt(),
			UpdatedAt:               auction.UpdatedAt(),
		},
		Bids: bidOutputs,
	}, nil
}
