package ports

import (
	"context"

	"auction/internal/modules/payment/domain/model"
)

// WithdrawalRepository persists withdrawal (platform -> user Alipay) orders.
type WithdrawalRepository interface {
	Save(ctx context.Context, withdrawal model.WithdrawalModel) (model.WithdrawalModel, error)
	FindByID(ctx context.Context, id uint64) (model.WithdrawalModel, error)
	FindByOutBizNo(ctx context.Context, outBizNo string) (model.WithdrawalModel, error)
	Update(ctx context.Context, withdrawal model.WithdrawalModel) (model.WithdrawalModel, error)
}
