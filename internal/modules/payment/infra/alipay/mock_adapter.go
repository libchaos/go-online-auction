package alipay

import (
	"context"
	"fmt"

	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/logger"
)

var _ ports.AlipayPort = (*MockAlipayAdapter)(nil)

// MockAlipayAdapter is an in-memory AlipayPort used when the provider is not
// "alipay". It lets the service compile, run, and be tested without real
// Alipay credentials (local dev, CI, unit tests, integration tests).
type MockAlipayAdapter struct {
	logger logger.Logger
}

func NewMockAlipayAdapter(logger logger.Logger) *MockAlipayAdapter {
	return &MockAlipayAdapter{logger: logger}
}

func (adapter *MockAlipayAdapter) CreateFaceToFacePayment(
	_ context.Context,
	input ports.FaceToFaceInput,
) (ports.FaceToFaceOutput, error) {
	return ports.FaceToFaceOutput{
		QRCodeURL:  fmt.Sprintf("https://mock.alipay.local/qr/%s", input.OutTradeNo),
		OutTradeNo: input.OutTradeNo,
	}, nil
}

func (adapter *MockAlipayAdapter) VerifyNotify(
	_ context.Context,
	params map[string]string,
) (ports.NotifyResult, error) {
	status := params["trade_status"]
	if status == "" {
		status = "TRADE_SUCCESS"
	}

	return ports.NotifyResult{
		TradeNo:     params["trade_no"],
		OutTradeNo:  params["out_trade_no"],
		TradeStatus: status,
	}, nil
}

func (adapter *MockAlipayAdapter) TransferToAlipayAccount(
	_ context.Context,
	input ports.TransferInput,
) (ports.TransferOutput, error) {
	return ports.TransferOutput{
		AlipayOrderID: fmt.Sprintf("mock-order-%s", input.OutBizNo),
	}, nil
}
