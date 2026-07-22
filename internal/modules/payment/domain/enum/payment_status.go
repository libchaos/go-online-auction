package enum

import (
	"auction/internal/modules/payment/domain/errs"
)

// PaymentStatus is the lifecycle state of a recharge (user -> platform) order.
type PaymentStatus string

const (
	PaymentStatusCreated PaymentStatus = "created"
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusFailed  PaymentStatus = "failed"
)

// ValidatePaymentStatus maps a stored string value to a PaymentStatus,
// returning ErrInvalidPaymentStatus when the value is not recognised.
func ValidatePaymentStatus(value string) (PaymentStatus, error) {
	switch PaymentStatus(value) {
	case PaymentStatusCreated, PaymentStatusSuccess, PaymentStatusFailed:
		return PaymentStatus(value), nil
	default:
		return "", errs.ErrInvalidPaymentStatus
	}
}
