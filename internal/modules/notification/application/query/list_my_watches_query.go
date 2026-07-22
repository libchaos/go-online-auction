package query

import (
	"context"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/ports"
)

type ListMyWatchesQueryInput struct {
	UserID uint64
	Limit  int
	Offset int
}

type ListMyWatchesQueryOutput struct {
	Watches []model.Watchlist
	Limit   int
	Offset  int
}

type ListMyWatchesQuery struct {
	watchlists ports.WatchlistRepository
}

func NewListMyWatchesQuery(watchlists ports.WatchlistRepository) *ListMyWatchesQuery {
	return &ListMyWatchesQuery{watchlists: watchlists}
}

// Execute returns a page of the user's watched products, newest first. The page
// size is clamped to the shared sane range used by the notification list query.
func (query *ListMyWatchesQuery) Execute(
	ctx context.Context,
	input ListMyWatchesQueryInput,
) (ListMyWatchesQueryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	watches, err := query.watchlists.ListByUser(ctx, input.UserID, limit, offset)
	if err != nil {
		return ListMyWatchesQueryOutput{}, err
	}

	return ListMyWatchesQueryOutput{Watches: watches, Limit: limit, Offset: offset}, nil
}
