package mapper

import (
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/sqlcgen"
)

type BidMapper struct{}

func NewBidMapper() *BidMapper {
	return &BidMapper{}
}

func (m *BidMapper) ToDomain(b sqlcgen.Bid) (model.BidModel, error) {
	amount := model.NewMoneyModel(uint64(b.AmountInCents))

	var maxAmount *model.MoneyModel
	if b.MaxAmountInCents != nil {
		restoredMax := model.NewMoneyModel(uint64(*b.MaxAmountInCents))
		maxAmount = &restoredMax
	}

	return model.RestoreBidModelWithMax(
		uint64(b.ID),
		uint64(b.AuctionID),
		uint64(b.UserID),
		amount,
		maxAmount,
		b.CreatedAt,
		b.UpdatedAt,
	)
}

func (m *BidMapper) ToCreateParams(bid model.BidModel, idempotencyKey string) sqlcgen.CreateBidParams {
	var maxAmountInCents *int64
	if bid.MaxAmount() != nil {
		cents := int64(bid.MaxAmount().AmountInCents())
		maxAmountInCents = &cents
	}

	return sqlcgen.CreateBidParams{
		AuctionID:        int64(bid.AuctionID()),
		UserID:           int64(bid.UserID()),
		AmountInCents:    int64(bid.Amount().AmountInCents()),
		MaxAmountInCents: maxAmountInCents,
		IdempotencyKey:   idempotencyKey,
		CreatedAt:        bid.CreatedAt(),
		UpdatedAt:        bid.UpdatedAt(),
	}
}
