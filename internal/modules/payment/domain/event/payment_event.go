package event

import (
	"time"

	"github.com/google/uuid"
)

const (
	PaymentSuccessEventType      = "payment_success"
	WithdrawalRequestedEventType = "withdrawal_requested"
	WithdrawalCompletedEventType = "withdrawal_completed"
	WithdrawalFailedEventType    = "withdrawal_failed"
)

// NATS JetStream subjects used by the payment module. The deposit-success
// subject is consumed by the ledger credit consumer; the withdrawal-requested
// subject is consumed by the Alipay payout Saga consumer.
const (
	SubjectDepositSuccess      = "payment.evt.deposit.success"
	SubjectWithdrawalRequested = "payment.cmd.withdrawal.requested"
	SubjectWithdrawalCompleted = "payment.evt.withdrawal.completed"
	SubjectWithdrawalFailed    = "payment.evt.withdrawal.failed"
)

// PaymentDomainEvent carries the identity and timestamp shared by all payment
// domain events. EventID is used as the NATS Message-Id for deduplication.
type PaymentDomainEvent struct {
	eventID   string
	timestamp time.Time
}

func newPaymentDomainEvent() PaymentDomainEvent {
	return PaymentDomainEvent{
		eventID:   uuid.New().String(),
		timestamp: time.Now().UTC(),
	}
}

func (domainEvent PaymentDomainEvent) EventID() string {
	return domainEvent.eventID
}

func (domainEvent PaymentDomainEvent) Timestamp() time.Time {
	return domainEvent.timestamp
}

// PaymentSuccessEvent is emitted after a recharge order is confirmed paid. The
// deposit-success consumer credits the user's ledger account from the platform
// account using OutTradeNo as the idempotency key.
type PaymentSuccessEvent struct {
	PaymentDomainEvent
	paymentID     uint64
	userID        uint64
	amountInCents uint64
	currency      string
	outTradeNo    string
}

func NewPaymentSuccessEvent(
	paymentID uint64,
	userID uint64,
	amountInCents uint64,
	currency string,
	outTradeNo string,
) PaymentSuccessEvent {
	return PaymentSuccessEvent{
		PaymentDomainEvent: newPaymentDomainEvent(),
		paymentID:          paymentID,
		userID:             userID,
		amountInCents:      amountInCents,
		currency:           currency,
		outTradeNo:         outTradeNo,
	}
}

func (domainEvent PaymentSuccessEvent) PaymentID() uint64 {
	return domainEvent.paymentID
}

func (domainEvent PaymentSuccessEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent PaymentSuccessEvent) AmountInCents() uint64 {
	return domainEvent.amountInCents
}

func (domainEvent PaymentSuccessEvent) Currency() string {
	return domainEvent.currency
}

func (domainEvent PaymentSuccessEvent) OutTradeNo() string {
	return domainEvent.outTradeNo
}

// WithdrawalRequestedEvent is emitted after a withdrawal order freezes the
// user's funds. The withdrawal consumer calls Alipay to pay out, then either
// withdraws the frozen balance (success) or unfreezes it (failure). OutBizNo
// is the idempotency key for both the Alipay transfer and the ledger action.
type WithdrawalRequestedEvent struct {
	PaymentDomainEvent
	withdrawalID    uint64
	userID          uint64
	ledgerAccountID uint64
	alipayAccount   string
	alipayRealName  string
	amountInCents   uint64
	currency        string
	outBizNo        string
	frozenOpID      string
}

func NewWithdrawalRequestedEvent(
	withdrawalID uint64,
	userID uint64,
	ledgerAccountID uint64,
	alipayAccount string,
	alipayRealName string,
	amountInCents uint64,
	currency string,
	outBizNo string,
	frozenOpID string,
) WithdrawalRequestedEvent {
	return WithdrawalRequestedEvent{
		PaymentDomainEvent: newPaymentDomainEvent(),
		withdrawalID:       withdrawalID,
		userID:             userID,
		ledgerAccountID:    ledgerAccountID,
		alipayAccount:      alipayAccount,
		alipayRealName:     alipayRealName,
		amountInCents:      amountInCents,
		currency:           currency,
		outBizNo:           outBizNo,
		frozenOpID:         frozenOpID,
	}
}

func (domainEvent WithdrawalRequestedEvent) WithdrawalID() uint64 {
	return domainEvent.withdrawalID
}

func (domainEvent WithdrawalRequestedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent WithdrawalRequestedEvent) LedgerAccountID() uint64 {
	return domainEvent.ledgerAccountID
}

func (domainEvent WithdrawalRequestedEvent) AlipayAccount() string {
	return domainEvent.alipayAccount
}

func (domainEvent WithdrawalRequestedEvent) AlipayRealName() string {
	return domainEvent.alipayRealName
}

func (domainEvent WithdrawalRequestedEvent) AmountInCents() uint64 {
	return domainEvent.amountInCents
}

func (domainEvent WithdrawalRequestedEvent) Currency() string {
	return domainEvent.currency
}

func (domainEvent WithdrawalRequestedEvent) OutBizNo() string {
	return domainEvent.outBizNo
}

func (domainEvent WithdrawalRequestedEvent) FrozenOpID() string {
	return domainEvent.frozenOpID
}

// WithdrawalCompletedEvent is emitted after the payout Saga confirms a
// successful Alipay transfer and permanently withdraws the frozen balance. It
// notifies the user that the funds have reached their Alipay account.
type WithdrawalCompletedEvent struct {
	PaymentDomainEvent
	withdrawalID  uint64
	userID        uint64
	amountInCents uint64
	currency      string
	outBizNo      string
	alipayOrderID string
}

func NewWithdrawalCompletedEvent(
	withdrawalID uint64,
	userID uint64,
	amountInCents uint64,
	currency string,
	outBizNo string,
	alipayOrderID string,
) WithdrawalCompletedEvent {
	return WithdrawalCompletedEvent{
		PaymentDomainEvent: newPaymentDomainEvent(),
		withdrawalID:       withdrawalID,
		userID:             userID,
		amountInCents:      amountInCents,
		currency:           currency,
		outBizNo:           outBizNo,
		alipayOrderID:      alipayOrderID,
	}
}

func (domainEvent WithdrawalCompletedEvent) WithdrawalID() uint64 {
	return domainEvent.withdrawalID
}

func (domainEvent WithdrawalCompletedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent WithdrawalCompletedEvent) AmountInCents() uint64 {
	return domainEvent.amountInCents
}

func (domainEvent WithdrawalCompletedEvent) Currency() string {
	return domainEvent.currency
}

func (domainEvent WithdrawalCompletedEvent) OutBizNo() string {
	return domainEvent.outBizNo
}

func (domainEvent WithdrawalCompletedEvent) AlipayOrderID() string {
	return domainEvent.alipayOrderID
}

// WithdrawalFailedEvent is emitted after the payout Saga fails the Alipay
// transfer and compensates by unfreezing the reserved funds. It notifies the
// user that the withdrawal did not go through and the balance was restored.
type WithdrawalFailedEvent struct {
	PaymentDomainEvent
	withdrawalID  uint64
	userID        uint64
	amountInCents uint64
	currency      string
	outBizNo      string
	failReason    string
}

func NewWithdrawalFailedEvent(
	withdrawalID uint64,
	userID uint64,
	amountInCents uint64,
	currency string,
	outBizNo string,
	failReason string,
) WithdrawalFailedEvent {
	return WithdrawalFailedEvent{
		PaymentDomainEvent: newPaymentDomainEvent(),
		withdrawalID:       withdrawalID,
		userID:             userID,
		amountInCents:      amountInCents,
		currency:           currency,
		outBizNo:           outBizNo,
		failReason:         failReason,
	}
}

func (domainEvent WithdrawalFailedEvent) WithdrawalID() uint64 {
	return domainEvent.withdrawalID
}

func (domainEvent WithdrawalFailedEvent) UserID() uint64 {
	return domainEvent.userID
}

func (domainEvent WithdrawalFailedEvent) AmountInCents() uint64 {
	return domainEvent.amountInCents
}

func (domainEvent WithdrawalFailedEvent) Currency() string {
	return domainEvent.currency
}

func (domainEvent WithdrawalFailedEvent) OutBizNo() string {
	return domainEvent.outBizNo
}

func (domainEvent WithdrawalFailedEvent) FailReason() string {
	return domainEvent.failReason
}
