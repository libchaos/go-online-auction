package errs

import "errors"

var (
	ErrAccountNotFound             = errors.New("ledger account not found")
	ErrAccountAlreadyExists        = errors.New("ledger account already exists for this owner and currency")
	ErrAccountOwnerRequired        = errors.New("account owner must not be empty")
	ErrAccountCurrencyRequired     = errors.New("account currency must not be empty")
	ErrInsufficientBalance         = errors.New("account available balance is insufficient")
	ErrInsufficientFrozenBalance   = errors.New("account frozen balance is insufficient")
	ErrDuplicateIdempotencyKey     = errors.New("ledger operation with this idempotency key was already processed")
	ErrAccountConcurrencyConflict  = errors.New("account was modified by another transaction")
	ErrInvalidOperationType        = errors.New("invalid ledger operation type")
	ErrSameAccountTransfer         = errors.New("transfer source and destination accounts must differ")
	ErrTransferAmountRequired      = errors.New("transfer amount must be greater than zero")
	ErrCounterpartyAccountRequired = errors.New("withdraw from frozen requires a counterparty account")
)
