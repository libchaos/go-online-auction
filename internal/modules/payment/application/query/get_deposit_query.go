package query

import (
	"context"

	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/ports"
)

type GetDepositQueryInput struct {
	PaymentID uint64
}

type GetDepositQueryOutput struct {
	Payment model.PaymentModel
}

type GetDepositQuery struct {
	payments ports.PaymentRepository
}

func NewGetDepositQuery(payments ports.PaymentRepository) *GetDepositQuery {
	return &GetDepositQuery{payments: payments}
}

func (query *GetDepositQuery) Execute(
	ctx context.Context,
	input GetDepositQueryInput,
) (GetDepositQueryOutput, error) {
	payment, err := query.payments.FindByID(ctx, input.PaymentID)
	if err != nil {
		return GetDepositQueryOutput{}, err
	}

	return GetDepositQueryOutput{Payment: payment}, nil
}
