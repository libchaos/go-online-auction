package httperrs

import (
	"errors"
	"net/http"

	domainerrs "auction/internal/modules/deposit/domain/errs"
	ledgerdomainerrs "auction/internal/modules/ledger/domain/errs"
	notificationdomainerrs "auction/internal/modules/notification/domain/errs"
	paymentdomainerrs "auction/internal/modules/payment/domain/errs"
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

var (
	ErrPaymentNotFound       = errs.New("PAYMENT_01", "Payment or withdrawal not found", http.StatusNotFound, nil)
	ErrPaymentConflict       = errs.New("PAYMENT_02", "Payment or withdrawal conflict", http.StatusConflict, nil)
	ErrPaymentInvalidRequest = errs.New("PAYMENT_03", "Invalid payment request", http.StatusBadRequest, nil)
	ErrPaymentAlipayFailure  = errs.New("PAYMENT_04", "Alipay gateway failure", http.StatusBadGateway, nil)
)

var (
	ErrNotificationNotFound       = errs.New("NOTIFICATION_01", "Notification not found", http.StatusNotFound, nil)
	ErrNotificationConflict       = errs.New("NOTIFICATION_02", "Notification conflict", http.StatusConflict, nil)
	ErrNotificationInvalidRequest = errs.New(
		"NOTIFICATION_03",
		"Invalid notification request",
		http.StatusBadRequest,
		nil,
	)
)

var (
	ErrWatchNotFound       = errs.New("WATCH_01", "Watch not found", http.StatusNotFound, nil)
	ErrWatchInvalidRequest = errs.New("WATCH_02", "Invalid watch request", http.StatusBadRequest, nil)
)

func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	mappings := []struct {
		domainErr error
		httpErr   error
	}{
		{domainerrs.ErrDepositNotFound, ErrDepositNotFound},
		{domainerrs.ErrAuctionConfigNotFound, ErrDepositNotFound},
		{domainerrs.ErrDepositConcurrencyConflict, ErrDepositConflict},
		{domainerrs.ErrDepositAlreadyExists, ErrDepositConflict},
		{domainerrs.ErrInvalidDepositTransition, ErrDepositConflict},
		{domainerrs.ErrDepositRequired, ErrDepositRequired},
		{domainerrs.ErrDepositInsufficient, ErrDepositInsufficient},
		{domainerrs.ErrDepositNotHeld, ErrDepositNotHeld},
		{domainerrs.ErrDepositExternalFailure, ErrDepositExternalFailure},
		{domainerrs.ErrDepositCaptureTooLarge, ErrDepositCaptureTooLarge},
		{ledgerdomainerrs.ErrInsufficientBalance, ErrDepositInsufficient},
		{ledgerdomainerrs.ErrInsufficientFrozenBalance, ErrDepositInsufficient},
		{ledgerdomainerrs.ErrAccountConcurrencyConflict, ErrDepositConflict},
		{ledgerdomainerrs.ErrAccountAlreadyExists, ErrDepositConflict},
		{ledgerdomainerrs.ErrDuplicateIdempotencyKey, ErrDepositConflict},
		{ledgerdomainerrs.ErrAccountNotFound, ErrDepositNotFound},
		{paymentdomainerrs.ErrPaymentNotFound, ErrPaymentNotFound},
		{paymentdomainerrs.ErrPaymentConcurrencyConflict, ErrPaymentConflict},
		{paymentdomainerrs.ErrPaymentAlreadyExists, ErrPaymentConflict},
		{paymentdomainerrs.ErrWithdrawalNotFound, ErrPaymentNotFound},
		{paymentdomainerrs.ErrWithdrawalConcurrencyConflict, ErrPaymentConflict},
		{paymentdomainerrs.ErrWithdrawalAlreadyExists, ErrPaymentConflict},
		{paymentdomainerrs.ErrAlipayNotifyVerification, ErrPaymentAlipayFailure},
		{paymentdomainerrs.ErrAlipayTradeNotPaid, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrPaymentUserRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrPaymentAmountRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrPaymentCurrencyRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrPaymentOutTradeNoRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrPaymentTradeNoRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalUserRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalAccountRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalAlipayAccountRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalAlipayRealNameRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalAmountRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalCurrencyRequired, ErrPaymentInvalidRequest},
		{paymentdomainerrs.ErrWithdrawalOutBizNoRequired, ErrPaymentInvalidRequest},
		{domainerrs.ErrDepositUserRequired, ErrDepositInvalidRequest},
		{domainerrs.ErrDepositAuctionRequired, ErrDepositInvalidRequest},
		{domainerrs.ErrDepositAmountRequired, ErrDepositInvalidRequest},
		{domainerrs.ErrDepositCurrencyRequired, ErrDepositInvalidRequest},
		{domainerrs.ErrDepositReferenceRequired, ErrDepositInvalidRequest},
		{domainerrs.ErrDepositCancellationBlocked, ErrDepositInvalidRequest},
		{notificationdomainerrs.ErrNotificationNotFound, ErrNotificationNotFound},
		{notificationdomainerrs.ErrNotificationAlreadyRead, ErrNotificationConflict},
		{notificationdomainerrs.ErrNotificationUserRequired, ErrNotificationInvalidRequest},
		{notificationdomainerrs.ErrNotificationTitleRequired, ErrNotificationInvalidRequest},
		{notificationdomainerrs.ErrNotificationBodyRequired, ErrNotificationInvalidRequest},
		{notificationdomainerrs.ErrNotificationChannelsEmpty, ErrNotificationInvalidRequest},
		{notificationdomainerrs.ErrPreferencesUserRequired, ErrNotificationInvalidRequest},
		{notificationdomainerrs.ErrPreferencesInvalid, ErrNotificationInvalidRequest},
		{notificationdomainerrs.ErrWatchUserRequired, ErrWatchInvalidRequest},
		{notificationdomainerrs.ErrWatchSpuRequired, ErrWatchInvalidRequest},
		{notificationdomainerrs.ErrWatchNotFound, ErrWatchNotFound},
	}

	for _, mapping := range mappings {
		if errors.Is(err, mapping.domainErr) {
			return mapping.httpErr
		}
	}

	return err
}
