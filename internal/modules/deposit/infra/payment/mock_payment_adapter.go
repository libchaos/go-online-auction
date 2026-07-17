package payment

import (
	"context"

	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/ports"
)

var _ ports.PaymentPort = (*MockPaymentAdapter)(nil)

type MockPaymentAdapter struct{}

func NewMockPaymentAdapter() *MockPaymentAdapter {
	return &MockPaymentAdapter{}
}

func (adapter *MockPaymentAdapter) Hold(
	_ context.Context,
	_ uint64,
	_ model.MoneyModel,
	_ string,
	reference string,
) (string, error) {
	return reference, nil
}

func (adapter *MockPaymentAdapter) Release(_ context.Context, _ string) error {
	return nil
}

func (adapter *MockPaymentAdapter) Capture(_ context.Context, _ string, _ model.MoneyModel) error {
	return nil
}

func (adapter *MockPaymentAdapter) Forfeit(_ context.Context, _ string) error {
	return nil
}
