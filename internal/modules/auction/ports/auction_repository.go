package ports

import (
	"context"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
)

type AuctionRepository interface {
	Create(ctx context.Context, auction model.AuctionModel) error
	FindByID(ctx context.Context, id uint64) (model.AuctionModel, error)
	Update(ctx context.Context, auction model.AuctionModel) error
}
