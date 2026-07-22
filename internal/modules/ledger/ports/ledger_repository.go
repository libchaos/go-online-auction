package ports

import (
	"context"

	"auction/internal/modules/ledger/domain/model"
)

type TransferInput struct {
	FromAccountID  uint64
	ToAccountID    uint64
	Amount         uint64
	IdempotencyKey string
	Reference      string
	Description    string
}

type FreezeInput struct {
	AccountID      uint64
	Amount         uint64
	IdempotencyKey string
	Reference      string
	Description    string
}

type UnfreezeInput struct {
	AccountID      uint64
	Amount         uint64
	IdempotencyKey string
	Reference      string
	Description    string
}

type WithdrawFromFrozenInput struct {
	AccountID             uint64
	CounterpartyAccountID uint64
	Amount                uint64
	IdempotencyKey        string
	Reference             string
	Description           string
}

type LedgerRepository interface {
	CreateAccount(ctx context.Context, owner string, currency string) (model.AccountModel, error)
	GetOrCreateAccountByOwner(ctx context.Context, owner string, currency string) (model.AccountModel, error)
	GetAccountByID(ctx context.Context, id uint64) (model.AccountModel, error)
	GetAccountByOwner(ctx context.Context, owner string, currency string) (model.AccountModel, error)
	Transfer(ctx context.Context, input TransferInput) (model.TransferModel, error)
	Freeze(ctx context.Context, input FreezeInput) (model.OperationModel, error)
	Unfreeze(ctx context.Context, input UnfreezeInput) (model.OperationModel, error)
	WithdrawFromFrozen(ctx context.Context, input WithdrawFromFrozenInput) (model.OperationModel, error)
}
