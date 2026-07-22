package enum

import (
	"auction/internal/modules/payment/domain/errs"
)

// WithdrawalStatus is the lifecycle state of a withdrawal (platform -> user
// Alipay) order. Funds are frozen when the order is created, then either
// permanently withdrawn on a successful Alipay payout or unfrozen on failure.
type WithdrawalStatus string

const (
	WithdrawalStatusCreated WithdrawalStatus = "created"
	WithdrawalStatusFrozen  WithdrawalStatus = "frozen"
	WithdrawalStatusSuccess WithdrawalStatus = "success"
	WithdrawalStatusFailed  WithdrawalStatus = "failed"
)

// ValidateWithdrawalStatus maps a stored string value to a WithdrawalStatus,
// returning ErrInvalidWithdrawalStatus when the value is not recognised.
func ValidateWithdrawalStatus(value string) (WithdrawalStatus, error) {
	switch WithdrawalStatus(value) {
	case WithdrawalStatusCreated, WithdrawalStatusFrozen, WithdrawalStatusSuccess, WithdrawalStatusFailed:
		return WithdrawalStatus(value), nil
	default:
		return "", errs.ErrInvalidWithdrawalStatus
	}
}
