package ports

import (
	"context"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
)

type BidRepository interface {
	Create(ctx context.Context, bid model.BidModel) error
	FindByID(ctx context.Context, id uint64) (model.BidModel, error)
	FindByAuctionID(ctx context.Context, auctionID uint64) ([]model.BidModel, error)
	FindTopBidsByAuctionID(ctx context.Context, auctionID uint64, limit int) ([]model.BidModel, error)
	Update(ctx context.Context, bid model.BidModel) error
}
