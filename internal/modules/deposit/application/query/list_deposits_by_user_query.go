package query

import (
	"context"

	"auction/internal/modules/deposit/ports"
)

type ListDepositsByUserQueryInput struct {
	UserID uint64
}

type ListDepositsByUserQueryOutput struct {
	Deposits []DepositView
}

type ListDepositsByUserQuery struct {
	repository ports.DepositRepository
}

func NewListDepositsByUserQuery(repository ports.DepositRepository) *ListDepositsByUserQuery {
	return &ListDepositsByUserQuery{repository: repository}
}

func (depositQuery *ListDepositsByUserQuery) Execute(
	ctx context.Context,
	input ListDepositsByUserQueryInput,
) (ListDepositsByUserQueryOutput, error) {
	deposits, err := depositQuery.repository.ListByUser(ctx, input.UserID)
	if err != nil {
		return ListDepositsByUserQueryOutput{}, err
	}

	views := make([]DepositView, 0, len(deposits))
	for _, deposit := range deposits {
		views = append(views, toDepositView(deposit))
	}

	return ListDepositsByUserQueryOutput{Deposits: views}, nil
}
