package mapper

import (
	"time"

	"auction/internal/modules/ledger/domain/enum"
	"auction/internal/modules/ledger/domain/model"
	"auction/internal/modules/ledger/infra/sqlcgen"
)

type LedgerMapper struct{}

func NewLedgerMapper() *LedgerMapper {
	return &LedgerMapper{}
}

func (mapper *LedgerMapper) ToAccountDomain(account sqlcgen.LedgerAccount) (model.AccountModel, error) {
	return model.RestoreAccountModel(
		uint64(account.ID),
		account.Owner,
		uint64(account.Balance),
		uint64(account.FrozenBalance),
		account.Currency,
		uint64(account.Version),
		account.CreatedAt,
		account.UpdatedAt,
	)
}

func (mapper *LedgerMapper) ToCreateAccountParams(owner string, currency string) sqlcgen.CreateLedgerAccountParams {
	now := time.Now().UTC()

	return sqlcgen.CreateLedgerAccountParams{
		Owner:         owner,
		Balance:       0,
		FrozenBalance: 0,
		Currency:      currency,
		Version:       1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (mapper *LedgerMapper) ToOperationDomain(operation sqlcgen.LedgerOperation) (model.OperationModel, error) {
	operationType, err := enum.NewOperationTypeEnum(operation.OperationType)
	if err != nil {
		return model.OperationModel{}, err
	}

	status, err := enum.NewOperationStatusEnum(operation.Status)
	if err != nil {
		return model.OperationModel{}, err
	}

	return model.RestoreOperationModel(
		uint64(operation.ID),
		uint64(operation.AccountID),
		toUint64Ptr(operation.CounterpartyAccountID),
		operationType,
		uint64(operation.Amount),
		operation.IdempotencyKey,
		status,
		toString(operation.Reference),
		toString(operation.Description),
		operation.CreatedAt,
		operation.UpdatedAt,
	), nil
}

func (mapper *LedgerMapper) ToEntryDomain(entry sqlcgen.LedgerEntry) (model.EntryModel, error) {
	entryType, err := enum.NewEntryTypeEnum(entry.EntryType)
	if err != nil {
		return model.EntryModel{}, err
	}

	return model.RestoreEntryModel(
		entry.ID,
		uint64(entry.AccountID),
		entry.Amount,
		entryType,
		toUint64Ptr(entry.OperationID),
		entry.CreatedAt,
	), nil
}

func (mapper *LedgerMapper) ToTransferDomain(transfer sqlcgen.LedgerTransfer) (model.TransferModel, error) {
	return model.RestoreTransferModel(
		uint64(transfer.ID),
		uint64(transfer.FromAccountID),
		uint64(transfer.ToAccountID),
		uint64(transfer.Amount),
		transfer.IdempotencyKey,
		transfer.CreatedAt,
	), nil
}

func toString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func toUint64Ptr(value *int64) *uint64 {
	if value == nil {
		return nil
	}

	converted := uint64(*value)

	return &converted
}
