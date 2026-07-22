package service

import (
	"context"

	"auction/internal/modules/ledger/domain/errs"
	"auction/internal/modules/ledger/domain/model"
	"auction/internal/modules/ledger/ports"
)

var _ ports.LedgerAccountService = (*LedgerAccountService)(nil)

type LedgerAccountService struct {
	uowFactory ports.LedgerUnitOfWorkFactory
}

func NewLedgerAccountService(uowFactory ports.LedgerUnitOfWorkFactory) *LedgerAccountService {
	return &LedgerAccountService{
		uowFactory: uowFactory,
	}
}

func (service *LedgerAccountService) CreateAccount(
	ctx context.Context,
	owner string,
	currency string,
) (model.AccountModel, error) {
	if owner == "" {
		return model.AccountModel{}, errs.ErrAccountOwnerRequired
	}

	if currency == "" {
		return model.AccountModel{}, errs.ErrAccountCurrencyRequired
	}

	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.AccountModel, error) {
		return repository.CreateAccount(ctx, owner, currency)
	})
}

func (service *LedgerAccountService) GetAccountByID(
	ctx context.Context,
	id uint64,
) (model.AccountModel, error) {
	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.AccountModel, error) {
		return repository.GetAccountByID(ctx, id)
	})
}

func (service *LedgerAccountService) GetAccountByOwner(
	ctx context.Context,
	owner string,
	currency string,
) (model.AccountModel, error) {
	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.AccountModel, error) {
		return repository.GetAccountByOwner(ctx, owner, currency)
	})
}

func (service *LedgerAccountService) Transfer(
	ctx context.Context,
	input ports.TransferInput,
) (model.TransferModel, error) {
	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.TransferModel, error) {
		return repository.Transfer(ctx, input)
	})
}

func (service *LedgerAccountService) Freeze(
	ctx context.Context,
	input ports.FreezeInput,
) (model.OperationModel, error) {
	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.OperationModel, error) {
		return repository.Freeze(ctx, input)
	})
}

func (service *LedgerAccountService) Unfreeze(
	ctx context.Context,
	input ports.UnfreezeInput,
) (model.OperationModel, error) {
	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.OperationModel, error) {
		return repository.Unfreeze(ctx, input)
	})
}

func (service *LedgerAccountService) WithdrawFromFrozen(
	ctx context.Context,
	input ports.WithdrawFromFrozenInput,
) (model.OperationModel, error) {
	return runTx(ctx, service.uowFactory, func(repository ports.LedgerRepository) (model.OperationModel, error) {
		return repository.WithdrawFromFrozen(ctx, input)
	})
}

func runTx[T any](
	ctx context.Context,
	factory ports.LedgerUnitOfWorkFactory,
	fn func(ports.LedgerRepository) (T, error),
) (T, error) {
	unitOfWork, err := factory.Begin(ctx)
	if err != nil {
		var zero T

		return zero, err
	}

	defer func() { _ = unitOfWork.Rollback(ctx) }()

	result, fnErr := fn(unitOfWork.LedgerRepository())
	if fnErr != nil {
		var zero T

		return zero, fnErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		var zero T

		return zero, completeErr
	}

	return result, nil
}
