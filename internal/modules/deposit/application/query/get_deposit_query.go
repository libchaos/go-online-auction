package query

import (
	"context"

	"auction/internal/modules/deposit/ports"
)

type GetDepositQueryInput struct {
	DepositID uint64
}

type GetDepositQueryOutput struct {
	Deposit DepositView
}

type GetDepositQuery struct {
	repository ports.DepositRepository
}

func NewGetDepositQuery(repository ports.DepositRepository) *GetDepositQuery {
	return &GetDepositQuery{repository: repository}
}

func (depositQuery *GetDepositQuery) Execute(
	ctx context.Context,
	input GetDepositQueryInput,
) (GetDepositQueryOutput, error) {
	deposit, err := depositQuery.repository.FindByID(ctx, input.DepositID)
	if err != nil {
		return GetDepositQueryOutput{}, err
	}

	return GetDepositQueryOutput{Deposit: toDepositView(deposit)}, nil
}
