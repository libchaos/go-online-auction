package ports

import (
	"context"

	"auction/internal/modules/deposit/domain/model"
)

type AuctionDepositConfig struct {
	Required bool
	Amount   model.MoneyModel
}

type AuctionConfigPort interface {
	GetRequiredDeposit(ctx context.Context, auctionID uint64) (AuctionDepositConfig, error)
}

type AuctionWinnerPort interface {
	GetWinnerUserID(ctx context.Context, auctionID uint64) (*uint64, error)
}
