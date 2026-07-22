package dto

type CreateDepositRequest struct {
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
}

type CreateDepositResponse struct {
	PaymentID  uint64 `json:"payment_id"`
	OutTradeNo string `json:"out_trade_no"`
	QRCodeURL  string `json:"qr_code_url"`
	Status     string `json:"status"`
}

type DepositResponse struct {
	PaymentID     uint64 `json:"payment_id"`
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	OutTradeNo    string `json:"out_trade_no"`
	QRCodeURL     string `json:"qr_code_url,omitempty"`
	AlipayTradeNo string `json:"alipay_trade_no,omitempty"`
	Version       uint64 `json:"version"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type CreateWithdrawalRequest struct {
	AmountInCents  uint64 `json:"amount_in_cents"`
	Currency       string `json:"currency"`
	AlipayAccount  string `json:"alipay_account"`
	AlipayRealName string `json:"alipay_real_name"`
}

type CreateWithdrawalResponse struct {
	WithdrawalID uint64 `json:"withdrawal_id"`
	OutBizNo     string `json:"out_biz_no"`
	Status       string `json:"status"`
}

type WithdrawalResponse struct {
	WithdrawalID    uint64 `json:"withdrawal_id"`
	UserID          uint64 `json:"user_id"`
	LedgerAccountID uint64 `json:"ledger_account_id"`
	AlipayAccount   string `json:"alipay_account"`
	AlipayRealName  string `json:"alipay_real_name"`
	AmountInCents   uint64 `json:"amount_in_cents"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
	OutBizNo        string `json:"out_biz_no"`
	AlipayOrderID   string `json:"alipay_order_id,omitempty"`
	FailReason      string `json:"fail_reason,omitempty"`
	Version         uint64 `json:"version"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}
