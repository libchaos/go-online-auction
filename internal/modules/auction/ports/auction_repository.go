package ports

import (
	"context"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/model"
)

type AuctionRepository interface {
	Create(ctx context.Context, auction model.AuctionModel) (model.AuctionModel, error)
	FindByID(ctx context.Context, id uint64) (model.AuctionModel, error)
	FindByIDForUpdate(ctx context.Context, id uint64) (model.AuctionModel, error)
	Update(ctx context.Context, auction model.AuctionModel) error
	FindAllPaginated(ctx context.Context, state *enum.AuctionStateEnum, limit, offset int) ([]model.AuctionModel, error)
	Count(ctx context.Context, state *enum.AuctionStateEnum) (uint64, error)
	// FindIDsDueToStart returns IDs of draft auctions whose scheduled start time has passed
	FindIDsDueToStart(ctx context.Context, limit int) ([]uint64, error)
	// FindIDsDueToClose returns IDs of active auctions whose end time has passed
	FindIDsDueToClose(ctx context.Context, limit int) ([]uint64, error)
}
