package ports

import (
	"context"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
)

type AuctionRepository interface {
	Create(ctx context.Context, auction model.AuctionModel) error
	FindByID(ctx context.Context, id uint64) (model.AuctionModel, error)
	FindByIDForUpdate(ctx context.Context, id uint64) (model.AuctionModel, error)
	Update(ctx context.Context, auction model.AuctionModel) error
	FindAllPaginated(ctx context.Context, state *enum.AuctionStateEnum, limit, offset int) ([]model.AuctionModel, error)
	Count(ctx context.Context, state *enum.AuctionStateEnum) (uint64, error)
}
