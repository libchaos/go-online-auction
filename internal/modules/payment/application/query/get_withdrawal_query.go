package query

import (
	"context"

	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/ports"
)

type GetWithdrawalQueryInput struct {
	WithdrawalID uint64
}

type GetWithdrawalQueryOutput struct {
	Withdrawal model.WithdrawalModel
}

type GetWithdrawalQuery struct {
	withdrawals ports.WithdrawalRepository
}

func NewGetWithdrawalQuery(withdrawals ports.WithdrawalRepository) *GetWithdrawalQuery {
	return &GetWithdrawalQuery{withdrawals: withdrawals}
}

func (query *GetWithdrawalQuery) Execute(
	ctx context.Context,
	input GetWithdrawalQueryInput,
) (GetWithdrawalQueryOutput, error) {
	withdrawal, err := query.withdrawals.FindByID(ctx, input.WithdrawalID)
	if err != nil {
		return GetWithdrawalQueryOutput{}, err
	}

	return GetWithdrawalQueryOutput{Withdrawal: withdrawal}, nil
}
