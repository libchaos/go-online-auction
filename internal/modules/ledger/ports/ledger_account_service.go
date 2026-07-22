package ports

import (
	"context"

	"auction/internal/modules/ledger/domain/model"
)

type LedgerAccountService interface {
	CreateAccount(ctx context.Context, owner string, currency string) (model.AccountModel, error)
	GetAccountByID(ctx context.Context, id uint64) (model.AccountModel, error)
	GetAccountByOwner(ctx context.Context, owner string, currency string) (model.AccountModel, error)
	Transfer(ctx context.Context, input TransferInput) (model.TransferModel, error)
	Freeze(ctx context.Context, input FreezeInput) (model.OperationModel, error)
	Unfreeze(ctx context.Context, input UnfreezeInput) (model.OperationModel, error)
	WithdrawFromFrozen(ctx context.Context, input WithdrawFromFrozenInput) (model.OperationModel, error)
}
