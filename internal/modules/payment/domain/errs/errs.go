package errs

import "errors"

var (
	ErrPaymentNotFound            = errors.New("payment not found")
	ErrPaymentAlreadyExists       = errors.New("payment with this out trade no already exists")
	ErrPaymentConcurrencyConflict = errors.New("payment was modified concurrently")
	ErrPaymentUserRequired        = errors.New("payment user id is required")
	ErrPaymentAmountRequired      = errors.New("payment amount is required")
	ErrPaymentCurrencyRequired    = errors.New("payment currency is required")
	ErrPaymentOutTradeNoRequired  = errors.New("payment out trade no is required")
	ErrPaymentTradeNoRequired     = errors.New("payment alipay trade no is required")
	ErrInvalidPaymentTransition   = errors.New("invalid payment status transition")
	ErrInvalidPaymentStatus       = errors.New("invalid payment status")

	ErrWithdrawalNotFound               = errors.New("withdrawal not found")
	ErrWithdrawalAlreadyExists          = errors.New("withdrawal with this out biz no already exists")
	ErrWithdrawalConcurrencyConflict    = errors.New("withdrawal was modified concurrently")
	ErrWithdrawalUserRequired           = errors.New("withdrawal user id is required")
	ErrWithdrawalAccountRequired        = errors.New("withdrawal ledger account id is required")
	ErrWithdrawalAlipayAccountRequired  = errors.New("withdrawal alipay account is required")
	ErrWithdrawalAlipayRealNameRequired = errors.New("withdrawal alipay real name is required")
	ErrWithdrawalAmountRequired         = errors.New("withdrawal amount is required")
	ErrWithdrawalCurrencyRequired       = errors.New("withdrawal currency is required")
	ErrWithdrawalOutBizNoRequired       = errors.New("withdrawal out biz no is required")
	ErrWithdrawalFrozenOpRequired       = errors.New("withdrawal frozen operation id is required")
	ErrInvalidWithdrawalTransition      = errors.New("invalid withdrawal status transition")
	ErrInvalidWithdrawalStatus          = errors.New("invalid withdrawal status")

	ErrAlipayNotifyVerification = errors.New("alipay asynchronous notification verification failed")
	ErrAlipayTradeNotPaid       = errors.New("alipay trade has not been paid")
)
