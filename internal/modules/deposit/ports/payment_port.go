package ports

import (
	"context"

	"auction/internal/modules/deposit/domain/model"
)

type PaymentPort interface {
	Hold(
		ctx context.Context,
		userID uint64,
		amount model.MoneyModel,
		currency string,
		reference string,
	) (externalReference string, err error)
	Release(ctx context.Context, externalReference string) error
	Capture(ctx context.Context, externalReference string, amount model.MoneyModel) error
	Forfeit(ctx context.Context, externalReference string) error
}
