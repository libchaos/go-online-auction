package query

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
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
	EndTime                 time.Time
	CreatedAt               time.Time
	StartTime               *time.Time
	HighestBidAmountInCents *uint64
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
		auctionOutputs = append(auctionOutputs, AuctionSummaryOutput{
			ID:                      auction.ID(),
			ListingID:               auction.ListingID(),
			State:                   state.String(),
			StartTime:               auction.StartTime(),
			EndTime:                 auction.EndTime(),
			HighestBidAmountInCents: auction.HighestBidAmount(),
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
