package errs

import (
	"errors"
	"net/http"

	domainerrs "auction/internal/modules/deposit/domain/errs"
	"auction/pkg/errs"
)

var (
	ErrDepositNotFound = errs.New("DEPOSIT_01", "Deposit not found", http.StatusNotFound, nil)
	ErrDepositConflict = errs.New("DEPOSIT_02", "Deposit conflict", http.StatusConflict, nil)
	ErrDepositRequired = errs.New(
		"DEPOSIT_03",
		"A deposit is required to bid on this auction",
		http.StatusBadRequest,
		nil,
	)
	ErrDepositInsufficient = errs.New(
		"DEPOSIT_04",
		"Deposit amount is insufficient for this auction",
		http.StatusBadRequest,
		nil,
	)
	ErrDepositNotHeld = errs.New(
		"DEPOSIT_05",
		"No held deposit found for this user and auction",
		http.StatusBadRequest,
		nil,
	)
	ErrDepositExternalFailure   = errs.New("DEPOSIT_06", "External payment provider failed", http.StatusBadGateway, nil)
	ErrDepositInvalidTransition = errs.New("DEPOSIT_07", "Invalid deposit status transition", http.StatusConflict, nil)
	ErrDepositInvalidRequest    = errs.New("DEPOSIT_08", "Invalid request body", http.StatusBadRequest, nil)
	ErrDepositCaptureTooLarge   = errs.New(
		"DEPOSIT_09",
		"Capture amount exceeds held deposit",
		http.StatusBadRequest,
		nil,
	)
)

func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domainerrs.ErrDepositNotFound),
		errors.Is(err, domainerrs.ErrAuctionConfigNotFound):
		return ErrDepositNotFound
	case errors.Is(err, domainerrs.ErrDepositConcurrencyConflict),
		errors.Is(err, domainerrs.ErrDepositAlreadyExists),
		errors.Is(err, domainerrs.ErrInvalidDepositTransition):
		return ErrDepositConflict
	case errors.Is(err, domainerrs.ErrDepositRequired):
		return ErrDepositRequired
	case errors.Is(err, domainerrs.ErrDepositInsufficient):
		return ErrDepositInsufficient
	case errors.Is(err, domainerrs.ErrDepositNotHeld):
		return ErrDepositNotHeld
	case errors.Is(err, domainerrs.ErrDepositExternalFailure):
		return ErrDepositExternalFailure
	case errors.Is(err, domainerrs.ErrDepositCaptureTooLarge):
		return ErrDepositCaptureTooLarge
	case errors.Is(err, domainerrs.ErrDepositUserRequired),
		errors.Is(err, domainerrs.ErrDepositAuctionRequired),
		errors.Is(err, domainerrs.ErrDepositAmountRequired),
		errors.Is(err, domainerrs.ErrDepositCurrencyRequired),
		errors.Is(err, domainerrs.ErrDepositReferenceRequired),
		errors.Is(err, domainerrs.ErrDepositCancellationBlocked):
		return ErrDepositInvalidRequest
	default:
		return err
	}
}
