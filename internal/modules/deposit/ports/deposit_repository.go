package ports

import (
	"context"

	"auction/internal/modules/deposit/domain/model"
)

type DepositRepository interface {
	Save(ctx context.Context, deposit model.DepositModel) (model.DepositModel, error)
	FindByID(ctx context.Context, id uint64) (model.DepositModel, error)
	FindByUserAndAuction(ctx context.Context, userID uint64, auctionID uint64) (model.DepositModel, error)
	ListByUser(ctx context.Context, userID uint64) ([]model.DepositModel, error)
	ListHeldByAuction(ctx context.Context, auctionID uint64) ([]model.DepositModel, error)
	Update(ctx context.Context, deposit model.DepositModel) (model.DepositModel, error)
}
