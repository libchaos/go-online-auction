package ports

import (
	"context"

	"auction/internal/modules/notification/domain/model"
)

type WatchlistRepository interface {
	Save(ctx context.Context, watchlist model.Watchlist) (model.Watchlist, error)
	Remove(ctx context.Context, userID, spuID uint64) error
	ListByUser(ctx context.Context, userID uint64, limit, offset int) ([]model.Watchlist, error)
	FindWatcherIDsBySpuID(ctx context.Context, spuID uint64) ([]uint64, error)
}
