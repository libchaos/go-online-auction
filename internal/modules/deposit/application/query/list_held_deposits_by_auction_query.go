package query

import (
	"context"

	"auction/internal/modules/deposit/ports"
)

type ListHeldDepositsByAuctionQueryInput struct {
	AuctionID uint64
}

type ListHeldDepositsByAuctionQueryOutput struct {
	Deposits []DepositView
}

type ListHeldDepositsByAuctionQuery struct {
	repository ports.DepositRepository
}

func NewListHeldDepositsByAuctionQuery(repository ports.DepositRepository) *ListHeldDepositsByAuctionQuery {
	return &ListHeldDepositsByAuctionQuery{repository: repository}
}

func (depositQuery *ListHeldDepositsByAuctionQuery) Execute(
	ctx context.Context,
	input ListHeldDepositsByAuctionQueryInput,
) (ListHeldDepositsByAuctionQueryOutput, error) {
	deposits, err := depositQuery.repository.ListHeldByAuction(ctx, input.AuctionID)
	if err != nil {
		return ListHeldDepositsByAuctionQueryOutput{}, err
	}

	views := make([]DepositView, 0, len(deposits))
	for _, deposit := range deposits {
		views = append(views, toDepositView(deposit))
	}

	return ListHeldDepositsByAuctionQueryOutput{Deposits: views}, nil
}
