package query

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
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
	StartTime               *time.Time
	EndTime                 time.Time
	HighestBidID            *uint64
	HighestBidAmountInCents *uint64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type BidOutput struct {
	ID            uint64
	UserID        uint64
	AmountInCents uint64
	CreatedAt     time.Time
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
		bidOutputs = append(bidOutputs, BidOutput{
			ID:            bid.ID(),
			UserID:        bid.UserID(),
			AmountInCents: bid.Amount().AmountInCents(),
			CreatedAt:     bid.CreatedAt(),
		})
	}

	state := auction.State()
	return GetAuctionByIDQueryOutput{
		Auction: AuctionOutput{
			ID:                      auction.ID(),
			ListingID:               auction.ListingID(),
			State:                   state.String(),
			StartTime:               auction.StartTime(),
			EndTime:                 auction.EndTime(),
			HighestBidID:            auction.HighestBidID(),
			HighestBidAmountInCents: auction.HighestBidAmount(),
			CreatedAt:               auction.CreatedAt(),
			UpdatedAt:               auction.UpdatedAt(),
		},
		Bids: bidOutputs,
	}, nil
}
