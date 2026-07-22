package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"auction/internal/modules/ledger/domain/enum"
	"auction/internal/modules/ledger/domain/errs"
	"auction/internal/modules/ledger/domain/model"
	"auction/internal/modules/ledger/infra/mapper"
	"auction/internal/modules/ledger/infra/sqlcgen"
	"auction/internal/modules/ledger/ports"
)

const (
	pgUniqueViolationCode      = "23505"
	accountOwnerCurrencyIndex  = "uq_ledger_accounts_owner_currency"
	transfersIdempotencyIndex  = "uq_ledger_transfers_idempotency"
	operationsIdempotencyIndex = "uq_ledger_operations_idempotency"
)

var _ ports.LedgerRepository = (*PostgresLedgerRepository)(nil)

type PostgresLedgerRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.LedgerMapper
}

func NewPostgresLedgerRepository(db sqlcgen.DBTX, ledgerMapper *mapper.LedgerMapper) *PostgresLedgerRepository {
	return &PostgresLedgerRepository{
		q:      sqlcgen.New(db),
		mapper: ledgerMapper,
	}
}

func (repository *PostgresLedgerRepository) CreateAccount(
	ctx context.Context,
	owner string,
	currency string,
) (model.AccountModel, error) {
	row, err := repository.q.CreateLedgerAccount(ctx, repository.mapper.ToCreateAccountParams(owner, currency))
	if err != nil {
		if isUniqueViolation(err, accountOwnerCurrencyIndex) {
			return model.AccountModel{}, errs.ErrAccountAlreadyExists
		}

		return model.AccountModel{}, err
	}

	return repository.mapper.ToAccountDomain(row)
}

func (repository *PostgresLedgerRepository) GetOrCreateAccountByOwner(
	ctx context.Context,
	owner string,
	currency string,
) (model.AccountModel, error) {
	row, err := repository.q.CreateLedgerAccount(ctx, repository.mapper.ToCreateAccountParams(owner, currency))
	if err != nil {
		if !isUniqueViolation(err, accountOwnerCurrencyIndex) {
			return model.AccountModel{}, err
		}

		existing, findErr := repository.q.GetLedgerAccountByOwner(ctx, sqlcgen.GetLedgerAccountByOwnerParams{
			Owner:    owner,
			Currency: currency,
		})
		if findErr != nil {
			return model.AccountModel{}, findErr
		}

		return repository.mapper.ToAccountDomain(existing)
	}

	return repository.mapper.ToAccountDomain(row)
}

func (repository *PostgresLedgerRepository) GetAccountByID(
	ctx context.Context,
	id uint64,
) (model.AccountModel, error) {
	row, err := repository.q.GetLedgerAccountByID(ctx, int64(id))
	if err != nil {
		return model.AccountModel{}, mapAccountLookupError(err)
	}

	return repository.mapper.ToAccountDomain(row)
}

func (repository *PostgresLedgerRepository) GetAccountByOwner(
	ctx context.Context,
	owner string,
	currency string,
) (model.AccountModel, error) {
	row, err := repository.q.GetLedgerAccountByOwner(ctx, sqlcgen.GetLedgerAccountByOwnerParams{
		Owner:    owner,
		Currency: currency,
	})
	if err != nil {
		return model.AccountModel{}, mapAccountLookupError(err)
	}

	return repository.mapper.ToAccountDomain(row)
}

func (repository *PostgresLedgerRepository) Transfer(
	ctx context.Context,
	input ports.TransferInput,
) (model.TransferModel, error) {
	if input.Amount == 0 {
		return model.TransferModel{}, errs.ErrTransferAmountRequired
	}

	if input.FromAccountID == input.ToAccountID {
		return model.TransferModel{}, errs.ErrSameAccountTransfer
	}

	if existing, err := repository.q.GetLedgerTransferByIdempotency(ctx, input.IdempotencyKey); err == nil {
		return repository.mapper.ToTransferDomain(existing)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return model.TransferModel{}, err
	}

	fromRow, err := repository.q.GetLedgerAccountByIDForUpdate(ctx, int64(input.FromAccountID))
	if err != nil {
		return model.TransferModel{}, mapAccountLookupError(err)
	}

	toRow, err := repository.q.GetLedgerAccountByIDForUpdate(ctx, int64(input.ToAccountID))
	if err != nil {
		return model.TransferModel{}, mapAccountLookupError(err)
	}

	if uint64(fromRow.Balance) < input.Amount {
		return model.TransferModel{}, errs.ErrInsufficientBalance
	}

	now := time.Now().UTC()

	if err = repository.transferBalances(ctx, fromRow, toRow, int64(input.Amount), now); err != nil {
		return model.TransferModel{}, err
	}

	if err = repository.writeTransferEntries(ctx, fromRow, toRow, int64(input.Amount), now); err != nil {
		return model.TransferModel{}, err
	}

	transfer, err := repository.q.CreateLedgerTransfer(ctx, sqlcgen.CreateLedgerTransferParams{
		FromAccountID:  fromRow.ID,
		ToAccountID:    toRow.ID,
		Amount:         int64(input.Amount),
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
	})
	if err != nil {
		recovered, recoverErr := repository.recoverTransferOnConflict(ctx, err, input.IdempotencyKey)
		if recoverErr != nil {
			return model.TransferModel{}, recoverErr
		}

		return recovered, nil
	}

	return repository.mapper.ToTransferDomain(transfer)
}

func (repository *PostgresLedgerRepository) Freeze(
	ctx context.Context,
	input ports.FreezeInput,
) (model.OperationModel, error) {
	if existing, err := repository.q.GetLedgerOperationByIdempotency(ctx, input.IdempotencyKey); err == nil {
		return repository.mapper.ToOperationDomain(existing)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return model.OperationModel{}, err
	}

	acct, err := repository.q.GetLedgerAccountByIDForUpdate(ctx, int64(input.AccountID))
	if err != nil {
		return model.OperationModel{}, mapAccountLookupError(err)
	}

	if uint64(acct.Balance) < input.Amount {
		return model.OperationModel{}, errs.ErrInsufficientBalance
	}

	now := time.Now().UTC()

	if err = repository.applyBalanceDelta(ctx, acct, -int64(input.Amount), int64(input.Amount), now); err != nil {
		return model.OperationModel{}, err
	}

	op, err := repository.q.CreateLedgerOperation(ctx, sqlcgen.CreateLedgerOperationParams{
		AccountID:      acct.ID,
		OperationType:  enum.EnumOperationTypeFreeze,
		Amount:         int64(input.Amount),
		IdempotencyKey: input.IdempotencyKey,
		Status:         enum.EnumOperationStatusCommitted,
		Reference:      toNullableString(input.Reference),
		Description:    toNullableString(input.Description),
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		recovered, recoverErr := repository.recoverOperationOnConflict(ctx, err, input.IdempotencyKey)
		if recoverErr != nil {
			return model.OperationModel{}, recoverErr
		}

		return recovered, nil
	}

	if err = repository.writeEntry(ctx, acct.ID, -int64(input.Amount), enum.EnumEntryTypeFreeze, &op.ID, now); err != nil {
		return model.OperationModel{}, err
	}

	return repository.mapper.ToOperationDomain(op)
}

func (repository *PostgresLedgerRepository) Unfreeze(
	ctx context.Context,
	input ports.UnfreezeInput,
) (model.OperationModel, error) {
	if existing, err := repository.q.GetLedgerOperationByIdempotency(ctx, input.IdempotencyKey); err == nil {
		return repository.mapper.ToOperationDomain(existing)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return model.OperationModel{}, err
	}

	acct, err := repository.q.GetLedgerAccountByIDForUpdate(ctx, int64(input.AccountID))
	if err != nil {
		return model.OperationModel{}, mapAccountLookupError(err)
	}

	if uint64(acct.FrozenBalance) < input.Amount {
		return model.OperationModel{}, errs.ErrInsufficientFrozenBalance
	}

	now := time.Now().UTC()

	if err = repository.applyBalanceDelta(ctx, acct, int64(input.Amount), -int64(input.Amount), now); err != nil {
		return model.OperationModel{}, err
	}

	op, err := repository.q.CreateLedgerOperation(ctx, sqlcgen.CreateLedgerOperationParams{
		AccountID:      acct.ID,
		OperationType:  enum.EnumOperationTypeUnfreeze,
		Amount:         int64(input.Amount),
		IdempotencyKey: input.IdempotencyKey,
		Status:         enum.EnumOperationStatusCommitted,
		Reference:      toNullableString(input.Reference),
		Description:    toNullableString(input.Description),
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		recovered, recoverErr := repository.recoverOperationOnConflict(ctx, err, input.IdempotencyKey)
		if recoverErr != nil {
			return model.OperationModel{}, recoverErr
		}

		return recovered, nil
	}

	if err = repository.writeEntry(ctx, acct.ID, int64(input.Amount), enum.EnumEntryTypeUnfreeze, &op.ID, now); err != nil {
		return model.OperationModel{}, err
	}

	return repository.mapper.ToOperationDomain(op)
}

func (repository *PostgresLedgerRepository) WithdrawFromFrozen(
	ctx context.Context,
	input ports.WithdrawFromFrozenInput,
) (model.OperationModel, error) {
	if input.CounterpartyAccountID == 0 {
		return model.OperationModel{}, errs.ErrCounterpartyAccountRequired
	}

	if existing, err := repository.q.GetLedgerOperationByIdempotency(ctx, input.IdempotencyKey); err == nil {
		return repository.mapper.ToOperationDomain(existing)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return model.OperationModel{}, err
	}

	acct, err := repository.q.GetLedgerAccountByIDForUpdate(ctx, int64(input.AccountID))
	if err != nil {
		return model.OperationModel{}, mapAccountLookupError(err)
	}

	counterparty, err := repository.q.GetLedgerAccountByIDForUpdate(ctx, int64(input.CounterpartyAccountID))
	if err != nil {
		return model.OperationModel{}, mapAccountLookupError(err)
	}

	if uint64(acct.FrozenBalance) < input.Amount {
		return model.OperationModel{}, errs.ErrInsufficientFrozenBalance
	}

	now := time.Now().UTC()

	if err = repository.applyWithdrawBalances(ctx, acct, counterparty, int64(input.Amount), now); err != nil {
		return model.OperationModel{}, err
	}

	counterpartyID := input.CounterpartyAccountID

	op, err := repository.q.CreateLedgerOperation(ctx, sqlcgen.CreateLedgerOperationParams{
		AccountID:             acct.ID,
		CounterpartyAccountID: toInt64Ptr(&counterpartyID),
		OperationType:         enum.EnumOperationTypeWithdrawFromFrozen,
		Amount:                int64(input.Amount),
		IdempotencyKey:        input.IdempotencyKey,
		Status:                enum.EnumOperationStatusCommitted,
		Reference:             toNullableString(input.Reference),
		Description:           toNullableString(input.Description),
		CreatedAt:             now,
		UpdatedAt:             now,
	})
	if err != nil {
		recovered, recoverErr := repository.recoverOperationOnConflict(ctx, err, input.IdempotencyKey)
		if recoverErr != nil {
			return model.OperationModel{}, recoverErr
		}

		return recovered, nil
	}

	if err = repository.writeWithdrawEntries(ctx, acct, counterparty, op.ID, int64(input.Amount), now); err != nil {
		return model.OperationModel{}, err
	}

	return repository.mapper.ToOperationDomain(op)
}

func (repository *PostgresLedgerRepository) applyBalanceDelta(
	ctx context.Context,
	account sqlcgen.LedgerAccount,
	balanceDelta int64,
	frozenDelta int64,
	now time.Time,
) error {
	_, err := repository.q.UpdateLedgerAccountBalance(ctx, sqlcgen.UpdateLedgerAccountBalanceParams{
		ID:            account.ID,
		Balance:       account.Balance + balanceDelta,
		FrozenBalance: account.FrozenBalance + frozenDelta,
		Version:       account.Version + 1,
		UpdatedAt:     now,
		Version_2:     account.Version,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrAccountConcurrencyConflict
		}

		return err
	}

	return nil
}

func (repository *PostgresLedgerRepository) writeEntry(
	ctx context.Context,
	accountID int64,
	amount int64,
	entryType string,
	operationID *int64,
	now time.Time,
) error {
	_, err := repository.q.CreateLedgerEntry(ctx, sqlcgen.CreateLedgerEntryParams{
		AccountID:   accountID,
		Amount:      amount,
		EntryType:   entryType,
		OperationID: operationID,
		CreatedAt:   now,
	})

	return err
}

func (repository *PostgresLedgerRepository) transferBalances(
	ctx context.Context,
	fromRow sqlcgen.LedgerAccount,
	toRow sqlcgen.LedgerAccount,
	amount int64,
	now time.Time,
) error {
	if err := repository.applyBalanceDelta(ctx, fromRow, -amount, 0, now); err != nil {
		return err
	}

	return repository.applyBalanceDelta(ctx, toRow, amount, 0, now)
}

func (repository *PostgresLedgerRepository) writeTransferEntries(
	ctx context.Context,
	fromRow sqlcgen.LedgerAccount,
	toRow sqlcgen.LedgerAccount,
	amount int64,
	now time.Time,
) error {
	if err := repository.writeEntry(ctx, fromRow.ID, -amount, enum.EnumEntryTypeTransferOut, nil, now); err != nil {
		return err
	}

	return repository.writeEntry(ctx, toRow.ID, amount, enum.EnumEntryTypeTransferIn, nil, now)
}

func (repository *PostgresLedgerRepository) applyWithdrawBalances(
	ctx context.Context,
	acct sqlcgen.LedgerAccount,
	counterparty sqlcgen.LedgerAccount,
	amount int64,
	now time.Time,
) error {
	if err := repository.applyBalanceDelta(ctx, acct, 0, -amount, now); err != nil {
		return err
	}

	return repository.applyBalanceDelta(ctx, counterparty, amount, 0, now)
}

func (repository *PostgresLedgerRepository) writeWithdrawEntries(
	ctx context.Context,
	acct sqlcgen.LedgerAccount,
	counterparty sqlcgen.LedgerAccount,
	operationID int64,
	amount int64,
	now time.Time,
) error {
	if err := repository.writeEntry(ctx, acct.ID, -amount, enum.EnumEntryTypeWithdrawFromFrozen, &operationID, now); err != nil {
		return err
	}

	return repository.writeEntry(
		ctx,
		counterparty.ID,
		amount,
		enum.EnumEntryTypeWithdrawFromFrozenCredit,
		&operationID,
		now,
	)
}

func (repository *PostgresLedgerRepository) recoverTransferOnConflict(
	ctx context.Context,
	createErr error,
	key string,
) (model.TransferModel, error) {
	if !isUniqueViolation(createErr, transfersIdempotencyIndex) {
		return model.TransferModel{}, createErr
	}

	row, err := repository.q.GetLedgerTransferByIdempotency(ctx, key)
	if err != nil {
		return model.TransferModel{}, createErr
	}

	return repository.mapper.ToTransferDomain(row)
}

func (repository *PostgresLedgerRepository) recoverOperationOnConflict(
	ctx context.Context,
	createErr error,
	key string,
) (model.OperationModel, error) {
	if !isUniqueViolation(createErr, operationsIdempotencyIndex) {
		return model.OperationModel{}, createErr
	}

	row, err := repository.q.GetLedgerOperationByIdempotency(ctx, key)
	if err != nil {
		return model.OperationModel{}, createErr
	}

	return repository.mapper.ToOperationDomain(row)
}

func mapAccountLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrAccountNotFound
	}

	return err
}

func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolationCode && pgErr.ConstraintName == constraintName
	}

	return false
}

func toNullableString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func toInt64Ptr(value *uint64) *int64 {
	if value == nil {
		return nil
	}

	converted := int64(*value)

	return &converted
}
