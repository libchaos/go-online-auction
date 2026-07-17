package errs

import "errors"

var (
	ErrDepositNotFound            = errors.New("deposit not found")
	ErrDepositConcurrencyConflict = errors.New("deposit was modified by another transaction")
	ErrAuctionConfigNotFound      = errors.New("auction not found")
	ErrAuctionWinnerNotFound      = errors.New("auction winner not found")
	ErrDepositRequired            = errors.New("a deposit is required to bid on this auction")
	ErrDepositInsufficient        = errors.New("deposit amount is insufficient for this auction")
	ErrDepositNotHeld             = errors.New("no held deposit found for this user and auction")
	ErrDepositAlreadyHeld         = errors.New("a deposit is already held for this user and auction")
	ErrDepositAlreadyExists       = errors.New("a deposit already exists for this user and auction")
	ErrInvalidDepositTransition   = errors.New("invalid deposit status transition")
	ErrDepositExternalFailure     = errors.New("external payment provider failed to hold funds")
	ErrDepositCancellationBlocked = errors.New("only a pending deposit can be cancelled")
	ErrInsufficientBalance        = errors.New("monetary amount is insufficient for subtraction")
	ErrDepositUserRequired        = errors.New("user id must be greater than zero")
	ErrDepositAuctionRequired     = errors.New("auction id must be greater than zero")
	ErrDepositAmountRequired      = errors.New("deposit amount must be greater than zero")
	ErrDepositCurrencyRequired    = errors.New("deposit currency must not be empty")
	ErrDepositReferenceRequired   = errors.New("deposit reference must not be empty")
	ErrDepositCaptureTooLarge     = errors.New("capture amount exceeds held deposit")
)
