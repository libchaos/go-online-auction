package envelope

import (
	"encoding/json"
	"time"

	"auction/internal/modules/payment/domain/event"
	"auction/internal/modules/payment/ports"
)

const schemaVersion = 1

// PaymentSuccessPayload is the JSON body published for a recharge-confirmed
// event. The deposit-success consumer credits the user's ledger account.
type PaymentSuccessPayload struct {
	EventID       string `json:"event_id"`
	PaymentID     uint64 `json:"payment_id"`
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	OutTradeNo    string `json:"out_trade_no"`
	OccurredAt    string `json:"occurred_at"`
}

// WithdrawalRequestedPayload is the JSON body published after a withdrawal
// order freezes the user's funds. The withdrawal consumer calls Alipay.
type WithdrawalRequestedPayload struct {
	EventID         string `json:"event_id"`
	WithdrawalID    uint64 `json:"withdrawal_id"`
	UserID          uint64 `json:"user_id"`
	LedgerAccountID uint64 `json:"ledger_account_id"`
	AlipayAccount   string `json:"alipay_account"`
	AlipayRealName  string `json:"alipay_real_name"`
	AmountInCents   uint64 `json:"amount_in_cents"`
	Currency        string `json:"currency"`
	OutBizNo        string `json:"out_biz_no"`
	FrozenOpID      string `json:"frozen_op_id"`
	OccurredAt      string `json:"occurred_at"`
}

func ToDepositSuccessOutboxEvent(domainEvent event.PaymentSuccessEvent) (ports.OutboxEvent, error) {
	payload := PaymentSuccessPayload{
		EventID:       domainEvent.EventID(),
		PaymentID:     domainEvent.PaymentID(),
		UserID:        domainEvent.UserID(),
		AmountInCents: domainEvent.AmountInCents(),
		Currency:      domainEvent.Currency(),
		OutTradeNo:    domainEvent.OutTradeNo(),
		OccurredAt:    domainEvent.Timestamp().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxEvent{}, err
	}

	return ports.OutboxEvent{
		EventID:       domainEvent.EventID(),
		EventType:     event.PaymentSuccessEventType,
		SchemaVersion: schemaVersion,
		Subject:       event.SubjectDepositSuccess,
		Payload:       body,
		OccurredAt:    domainEvent.Timestamp(),
	}, nil
}

// WithdrawalCompletedPayload is the JSON body published after a successful
// Alipay payout. The notification consumer turns it into an in-app message.
type WithdrawalCompletedPayload struct {
	EventID       string `json:"event_id"`
	WithdrawalID  uint64 `json:"withdrawal_id"`
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	OutBizNo      string `json:"out_biz_no"`
	AlipayOrderID string `json:"alipay_order_id"`
	OccurredAt    string `json:"occurred_at"`
}

// WithdrawalFailedPayload is the JSON body published after a failed Alipay
// payout whose funds were compensated back to the user's balance.
type WithdrawalFailedPayload struct {
	EventID       string `json:"event_id"`
	WithdrawalID  uint64 `json:"withdrawal_id"`
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	OutBizNo      string `json:"out_biz_no"`
	FailReason    string `json:"fail_reason"`
	OccurredAt    string `json:"occurred_at"`
}

func ToWithdrawalRequestedOutboxEvent(domainEvent event.WithdrawalRequestedEvent) (ports.OutboxEvent, error) {
	payload := WithdrawalRequestedPayload{
		EventID:         domainEvent.EventID(),
		WithdrawalID:    domainEvent.WithdrawalID(),
		UserID:          domainEvent.UserID(),
		LedgerAccountID: domainEvent.LedgerAccountID(),
		AlipayAccount:   domainEvent.AlipayAccount(),
		AlipayRealName:  domainEvent.AlipayRealName(),
		AmountInCents:   domainEvent.AmountInCents(),
		Currency:        domainEvent.Currency(),
		OutBizNo:        domainEvent.OutBizNo(),
		FrozenOpID:      domainEvent.FrozenOpID(),
		OccurredAt:      domainEvent.Timestamp().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxEvent{}, err
	}

	return ports.OutboxEvent{
		EventID:       domainEvent.EventID(),
		EventType:     event.WithdrawalRequestedEventType,
		SchemaVersion: schemaVersion,
		Subject:       event.SubjectWithdrawalRequested,
		Payload:       body,
		OccurredAt:    domainEvent.Timestamp(),
	}, nil
}

func ToWithdrawalCompletedOutboxEvent(domainEvent event.WithdrawalCompletedEvent) (ports.OutboxEvent, error) {
	payload := WithdrawalCompletedPayload{
		EventID:       domainEvent.EventID(),
		WithdrawalID:  domainEvent.WithdrawalID(),
		UserID:        domainEvent.UserID(),
		AmountInCents: domainEvent.AmountInCents(),
		Currency:      domainEvent.Currency(),
		OutBizNo:      domainEvent.OutBizNo(),
		AlipayOrderID: domainEvent.AlipayOrderID(),
		OccurredAt:    domainEvent.Timestamp().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxEvent{}, err
	}

	return ports.OutboxEvent{
		EventID:       domainEvent.EventID(),
		EventType:     event.WithdrawalCompletedEventType,
		SchemaVersion: schemaVersion,
		Subject:       event.SubjectWithdrawalCompleted,
		Payload:       body,
		OccurredAt:    domainEvent.Timestamp(),
	}, nil
}

func ToWithdrawalFailedOutboxEvent(domainEvent event.WithdrawalFailedEvent) (ports.OutboxEvent, error) {
	payload := WithdrawalFailedPayload{
		EventID:       domainEvent.EventID(),
		WithdrawalID:  domainEvent.WithdrawalID(),
		UserID:        domainEvent.UserID(),
		AmountInCents: domainEvent.AmountInCents(),
		Currency:      domainEvent.Currency(),
		OutBizNo:      domainEvent.OutBizNo(),
		FailReason:    domainEvent.FailReason(),
		OccurredAt:    domainEvent.Timestamp().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxEvent{}, err
	}

	return ports.OutboxEvent{
		EventID:       domainEvent.EventID(),
		EventType:     event.WithdrawalFailedEventType,
		SchemaVersion: schemaVersion,
		Subject:       event.SubjectWithdrawalFailed,
		Payload:       body,
		OccurredAt:    domainEvent.Timestamp(),
	}, nil
}
