package model

import (
	"time"

	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/errs"
)

// WithdrawalModel is the withdrawal aggregate (platform -> user Alipay). When
// created the user's ledger balance is frozen; a successful Alipay payout
// moves it to SUCCESS (permanent debit via WithdrawFromFrozen), while a failed
// payout moves it to FAILED and the balance is compensated back via Unfreeze.
type WithdrawalModel struct {
	id              uint64
	userID          uint64
	ledgerAccountID uint64
	alipayAccount   string
	alipayRealName  string
	amountInCents   uint64
	currency        string
	status          enum.WithdrawalStatus
	outBizNo        string
	frozenOpID      string
	alipayOrderID   string
	failReason      string
	version         uint64
	createdAt       time.Time
	updatedAt       time.Time
}

func NewWithdrawal(
	userID uint64,
	ledgerAccountID uint64,
	alipayAccount string,
	alipayRealName string,
	amountInCents uint64,
	currency string,
	outBizNo string,
) (WithdrawalModel, error) {
	if userID == 0 {
		return WithdrawalModel{}, errs.ErrWithdrawalUserRequired
	}
	if ledgerAccountID == 0 {
		return WithdrawalModel{}, errs.ErrWithdrawalAccountRequired
	}
	if alipayAccount == "" {
		return WithdrawalModel{}, errs.ErrWithdrawalAlipayAccountRequired
	}
	if alipayRealName == "" {
		return WithdrawalModel{}, errs.ErrWithdrawalAlipayRealNameRequired
	}
	if amountInCents == 0 {
		return WithdrawalModel{}, errs.ErrWithdrawalAmountRequired
	}
	if currency == "" {
		return WithdrawalModel{}, errs.ErrWithdrawalCurrencyRequired
	}
	if outBizNo == "" {
		return WithdrawalModel{}, errs.ErrWithdrawalOutBizNoRequired
	}

	now := time.Now().UTC()

	return WithdrawalModel{
		userID:          userID,
		ledgerAccountID: ledgerAccountID,
		alipayAccount:   alipayAccount,
		alipayRealName:  alipayRealName,
		amountInCents:   amountInCents,
		currency:        currency,
		status:          enum.WithdrawalStatusCreated,
		outBizNo:        outBizNo,
		version:         1,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func RestoreWithdrawalModel(
	id uint64,
	userID uint64,
	ledgerAccountID uint64,
	alipayAccount string,
	alipayRealName string,
	amountInCents uint64,
	currency string,
	status enum.WithdrawalStatus,
	outBizNo string,
	frozenOpID string,
	alipayOrderID string,
	failReason string,
	version uint64,
	createdAt time.Time,
	updatedAt time.Time,
) (WithdrawalModel, error) {
	return WithdrawalModel{
		id:              id,
		userID:          userID,
		ledgerAccountID: ledgerAccountID,
		alipayAccount:   alipayAccount,
		alipayRealName:  alipayRealName,
		amountInCents:   amountInCents,
		currency:        currency,
		status:          status,
		outBizNo:        outBizNo,
		frozenOpID:      frozenOpID,
		alipayOrderID:   alipayOrderID,
		failReason:      failReason,
		version:         version,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

// MarkFrozen records the ledger freeze operation that reserved the funds. It
// fails if the withdrawal is no longer in the CREATED state.
func (withdrawal *WithdrawalModel) MarkFrozen(frozenOpID string) error {
	if withdrawal.status != enum.WithdrawalStatusCreated {
		return errs.ErrInvalidWithdrawalTransition
	}
	if frozenOpID == "" {
		return errs.ErrWithdrawalFrozenOpRequired
	}

	withdrawal.status = enum.WithdrawalStatusFrozen
	withdrawal.frozenOpID = frozenOpID
	withdrawal.version++
	withdrawal.updatedAt = time.Now().UTC()

	return nil
}

// MarkSuccess moves the withdrawal to SUCCESS after a confirmed Alipay payout.
// It fails if the withdrawal is not in the FROZEN state.
func (withdrawal *WithdrawalModel) MarkSuccess(alipayOrderID string) error {
	if withdrawal.status != enum.WithdrawalStatusFrozen {
		return errs.ErrInvalidWithdrawalTransition
	}

	withdrawal.status = enum.WithdrawalStatusSuccess
	withdrawal.alipayOrderID = alipayOrderID
	withdrawal.version++
	withdrawal.updatedAt = time.Now().UTC()

	return nil
}

// MarkFailed moves the withdrawal to FAILED and records the failure reason for
// the compensation (unfreeze) step. It fails if the withdrawal is not FROZEN.
func (withdrawal *WithdrawalModel) MarkFailed(failReason string) error {
	if withdrawal.status != enum.WithdrawalStatusFrozen {
		return errs.ErrInvalidWithdrawalTransition
	}

	withdrawal.status = enum.WithdrawalStatusFailed
	withdrawal.failReason = failReason
	withdrawal.version++
	withdrawal.updatedAt = time.Now().UTC()

	return nil
}

func (withdrawal *WithdrawalModel) ID() uint64 {
	return withdrawal.id
}

func (withdrawal *WithdrawalModel) UserID() uint64 {
	return withdrawal.userID
}

func (withdrawal *WithdrawalModel) LedgerAccountID() uint64 {
	return withdrawal.ledgerAccountID
}

func (withdrawal *WithdrawalModel) AlipayAccount() string {
	return withdrawal.alipayAccount
}

func (withdrawal *WithdrawalModel) AlipayRealName() string {
	return withdrawal.alipayRealName
}

func (withdrawal *WithdrawalModel) AmountInCents() uint64 {
	return withdrawal.amountInCents
}

func (withdrawal *WithdrawalModel) Currency() string {
	return withdrawal.currency
}

func (withdrawal *WithdrawalModel) Status() enum.WithdrawalStatus {
	return withdrawal.status
}

func (withdrawal *WithdrawalModel) OutBizNo() string {
	return withdrawal.outBizNo
}

func (withdrawal *WithdrawalModel) FrozenOpID() string {
	return withdrawal.frozenOpID
}

func (withdrawal *WithdrawalModel) AlipayOrderID() string {
	return withdrawal.alipayOrderID
}

func (withdrawal *WithdrawalModel) FailReason() string {
	return withdrawal.failReason
}

func (withdrawal *WithdrawalModel) Version() uint64 {
	return withdrawal.version
}

func (withdrawal *WithdrawalModel) CreatedAt() time.Time {
	return withdrawal.createdAt
}

func (withdrawal *WithdrawalModel) UpdatedAt() time.Time {
	return withdrawal.updatedAt
}
