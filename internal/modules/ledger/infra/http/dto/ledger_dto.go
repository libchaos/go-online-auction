package dto

type CreateAccountRequest struct {
	Owner    string `json:"owner"`
	Currency string `json:"currency"`
}

type AccountResponse struct {
	AccountID     uint64 `json:"account_id"`
	Owner         string `json:"owner"`
	Balance       uint64 `json:"balance"`
	FrozenBalance uint64 `json:"frozen_balance"`
	Currency      string `json:"currency"`
	Version       uint64 `json:"version"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type TransferRequest struct {
	FromAccountID  uint64 `json:"from_account_id"`
	ToAccountID    uint64 `json:"to_account_id"`
	Amount         uint64 `json:"amount"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	Reference      string `json:"reference,omitempty"`
	Description    string `json:"description,omitempty"`
}

type FreezeRequest struct {
	AccountID      uint64 `json:"account_id"`
	Amount         uint64 `json:"amount"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	Reference      string `json:"reference,omitempty"`
	Description    string `json:"description,omitempty"`
}

type UnfreezeRequest struct {
	AccountID      uint64 `json:"account_id"`
	Amount         uint64 `json:"amount"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	Reference      string `json:"reference,omitempty"`
	Description    string `json:"description,omitempty"`
}

type WithdrawFromFrozenRequest struct {
	AccountID             uint64 `json:"account_id"`
	CounterpartyAccountID uint64 `json:"counterparty_account_id"`
	Amount                uint64 `json:"amount"`
	IdempotencyKey        string `json:"idempotency_key,omitempty"`
	Reference             string `json:"reference,omitempty"`
	Description           string `json:"description,omitempty"`
}

type OperationResponse struct {
	OperationID           uint64  `json:"operation_id"`
	AccountID             uint64  `json:"account_id"`
	CounterpartyAccountID *uint64 `json:"counterparty_account_id,omitempty"`
	OperationType         string  `json:"operation_type"`
	Amount                uint64  `json:"amount"`
	IdempotencyKey        string  `json:"idempotency_key"`
	Status                string  `json:"status"`
	Reference             string  `json:"reference,omitempty"`
	Description           string  `json:"description,omitempty"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

type TransferResponse struct {
	TransferID     uint64 `json:"transfer_id"`
	FromAccountID  uint64 `json:"from_account_id"`
	ToAccountID    uint64 `json:"to_account_id"`
	Amount         uint64 `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
	CreatedAt      string `json:"created_at"`
}
