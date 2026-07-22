package model

import (
	"time"

	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/errs"
)

// PaymentModel is the recharge aggregate (user -> platform). It is created
// when a face-to-face prepay order is opened and moves to SUCCESS or FAILED
// when Alipay's asynchronous notification is processed.
type PaymentModel struct {
	id            uint64
	userID        uint64
	amountInCents uint64
	currency      string
	status        enum.PaymentStatus
	outTradeNo    string
	qrCodeURL     string
	alipayTradeNo string
	version       uint64
	createdAt     time.Time
	updatedAt     time.Time
}

func NewPayment(
	userID uint64,
	amountInCents uint64,
	currency string,
	outTradeNo string,
	qrCodeURL string,
) (PaymentModel, error) {
	if userID == 0 {
		return PaymentModel{}, errs.ErrPaymentUserRequired
	}
	if amountInCents == 0 {
		return PaymentModel{}, errs.ErrPaymentAmountRequired
	}
	if currency == "" {
		return PaymentModel{}, errs.ErrPaymentCurrencyRequired
	}
	if outTradeNo == "" {
		return PaymentModel{}, errs.ErrPaymentOutTradeNoRequired
	}

	now := time.Now().UTC()

	return PaymentModel{
		userID:        userID,
		amountInCents: amountInCents,
		currency:      currency,
		status:        enum.PaymentStatusCreated,
		outTradeNo:    outTradeNo,
		qrCodeURL:     qrCodeURL,
		version:       1,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func RestorePaymentModel(
	id uint64,
	userID uint64,
	amountInCents uint64,
	currency string,
	status enum.PaymentStatus,
	outTradeNo string,
	qrCodeURL string,
	alipayTradeNo string,
	version uint64,
	createdAt time.Time,
	updatedAt time.Time,
) (PaymentModel, error) {
	return PaymentModel{
		id:            id,
		userID:        userID,
		amountInCents: amountInCents,
		currency:      currency,
		status:        status,
		outTradeNo:    outTradeNo,
		qrCodeURL:     qrCodeURL,
		alipayTradeNo: alipayTradeNo,
		version:       version,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

// MarkSuccess moves the payment to SUCCESS once Alipay confirms the trade was
// paid. It fails if the payment is no longer in the CREATED state.
func (payment *PaymentModel) MarkSuccess(tradeNo string) error {
	if payment.status != enum.PaymentStatusCreated {
		return errs.ErrInvalidPaymentTransition
	}
	if tradeNo == "" {
		return errs.ErrPaymentTradeNoRequired
	}

	payment.status = enum.PaymentStatusSuccess
	payment.alipayTradeNo = tradeNo
	payment.version++
	payment.updatedAt = time.Now().UTC()

	return nil
}

// MarkFailed moves the payment to FAILED. It fails if the payment is no longer
// in the CREATED state.
func (payment *PaymentModel) MarkFailed() error {
	if payment.status != enum.PaymentStatusCreated {
		return errs.ErrInvalidPaymentTransition
	}

	payment.status = enum.PaymentStatusFailed
	payment.version++
	payment.updatedAt = time.Now().UTC()

	return nil
}

func (payment *PaymentModel) ID() uint64 {
	return payment.id
}

func (payment *PaymentModel) UserID() uint64 {
	return payment.userID
}

func (payment *PaymentModel) AmountInCents() uint64 {
	return payment.amountInCents
}

func (payment *PaymentModel) Currency() string {
	return payment.currency
}

func (payment *PaymentModel) Status() enum.PaymentStatus {
	return payment.status
}

func (payment *PaymentModel) OutTradeNo() string {
	return payment.outTradeNo
}

func (payment *PaymentModel) QRCodeURL() string {
	return payment.qrCodeURL
}

func (payment *PaymentModel) AlipayTradeNo() string {
	return payment.alipayTradeNo
}

func (payment *PaymentModel) Version() uint64 {
	return payment.version
}

func (payment *PaymentModel) CreatedAt() time.Time {
	return payment.createdAt
}

func (payment *PaymentModel) UpdatedAt() time.Time {
	return payment.updatedAt
}
