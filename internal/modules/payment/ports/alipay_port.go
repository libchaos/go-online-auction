package ports

import "context"

// AlipayPort abstracts the Alipay operations the payment module depends on.
// Implementations may call the real Alipay gateway or a mock for local/CI use.
type AlipayPort interface {
	// CreateFaceToFacePayment opens a face-to-face prepay order and returns a
	// QR code the user scans to pay. OutTradeNo is the caller-supplied idempotency key.
	CreateFaceToFacePayment(ctx context.Context, input FaceToFaceInput) (FaceToFaceOutput, error)
	// VerifyNotify verifies and decodes an asynchronous notification from Alipay.
	VerifyNotify(ctx context.Context, params map[string]string) (NotifyResult, error)
	// TransferToAlipayAccount pays out to a user's Alipay account. OutBizNo is
	// the caller-supplied idempotency key honoured by Alipay.
	TransferToAlipayAccount(ctx context.Context, input TransferInput) (TransferOutput, error)
}

type FaceToFaceInput struct {
	OutTradeNo    string
	Subject       string
	NotifyURL     string
	AmountInCents uint64
	Currency      string
}

type FaceToFaceOutput struct {
	QRCodeURL  string
	OutTradeNo string
}

type NotifyResult struct {
	TradeNo     string
	OutTradeNo  string
	TradeStatus string
}

type TransferInput struct {
	OutBizNo      string
	PayeeAccount  string
	PayeeRealName string
	AmountInCents uint64
	Currency      string
	Remark        string
}

type TransferOutput struct {
	AlipayOrderID string
}
