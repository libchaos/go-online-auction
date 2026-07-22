package ports

import (
	"context"

	"auction/internal/modules/payment/domain/model"
)

// PaymentRepository persists recharge (user -> platform) orders.
type PaymentRepository interface {
	Save(ctx context.Context, payment model.PaymentModel) (model.PaymentModel, error)
	FindByID(ctx context.Context, id uint64) (model.PaymentModel, error)
	FindByOutTradeNo(ctx context.Context, outTradeNo string) (model.PaymentModel, error)
	Update(ctx context.Context, payment model.PaymentModel) (model.PaymentModel, error)
}
