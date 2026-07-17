package query

import (
	"context"
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type ListAuctionsQueryInput struct {
	State  *string
	Limit  int
	Offset int
}

type ListAuctionsQueryOutput struct {
	Auctions   []AuctionSummaryOutput
	TotalCount uint64
	Limit      int
	Offset     int
}

type AuctionSummaryOutput struct {
	ID                      uint64
	ListingID               uint64
	State                   string
	TradingMode             string
	EndTime                 time.Time
	CreatedAt               time.Time
	StartTime               *time.Time
	StartingPrice           *uint64
	ReservePrice            *uint64
	CurrentPrice            *uint64
	HighestBidAmountInCents *uint64
	WinnerUserID            *uint64
	WinningBidAmountInCents *uint64
	AntiSnipeEnabled        bool
	ExtensionWindowSec      int64
}

type ListAuctionsQuery struct {
	auctionRepository ports.AuctionRepository
	logger            logger.Logger
}

func NewListAuctionsQuery(
	auctionRepository ports.AuctionRepository,
	logger logger.Logger,
) *ListAuctionsQuery {
	return &ListAuctionsQuery{
		auctionRepository: auctionRepository,
		logger:            logger,
	}
}

func (q *ListAuctionsQuery) Execute(
	ctx context.Context,
	input ListAuctionsQueryInput,
) (ListAuctionsQueryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	var state *enum.AuctionStateEnum
	if input.State != nil {
		s, err := enum.NewAuctionStateEnum(*input.State)
		if err != nil {
			q.logger.Error().Err(err).Msg("invalid auction state")
			return ListAuctionsQueryOutput{}, err
		}
		state = &s
	}

	auctions, err := q.auctionRepository.FindAllPaginated(ctx, state, limit, offset)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to fetch auctions")
		return ListAuctionsQueryOutput{}, err
	}

	totalCount, err := q.auctionRepository.Count(ctx, state)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to count auctions")
		return ListAuctionsQueryOutput{}, err
	}

	auctionOutputs := make([]AuctionSummaryOutput, 0, len(auctions))
	for _, auction := range auctions {
		state := auction.State()
		tradingMode := auction.TradingMode()
		auctionOutputs = append(auctionOutputs, AuctionSummaryOutput{
			ID:                      auction.ID(),
			ListingID:               auction.ListingID(),
			State:                   state.String(),
			TradingMode:             tradingMode.String(),
			StartTime:               auction.StartTime(),
			EndTime:                 auction.EndTime(),
			StartingPrice:           auction.StartingPrice(),
			ReservePrice:            auction.ReservePrice(),
			CurrentPrice:            auction.CurrentPrice(),
			HighestBidAmountInCents: auction.HighestBidAmount(),
			WinnerUserID:            auction.WinnerUserID(),
			WinningBidAmountInCents: auction.WinningBidAmount(),
			AntiSnipeEnabled:        auction.AntiSnipeEnabled(),
			ExtensionWindowSec:      auction.ExtensionWindowSec(),
			CreatedAt:               auction.CreatedAt(),
		})
	}

	return ListAuctionsQueryOutput{
		Auctions:   auctionOutputs,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
