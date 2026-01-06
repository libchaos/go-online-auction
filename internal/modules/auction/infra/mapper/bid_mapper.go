package mapper

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/entity"
)

type BidMapper struct{}

func NewBidMapper() *BidMapper {
	return &BidMapper{}
}

func (m *BidMapper) ToDomain(e entity.BidEntity) (model.BidModel, error) {
	amount := model.NewMoneyModel(e.AmountInCents)

	return model.RestoreBidModel(
		e.ID,
		e.AuctionID,
		e.UserID,
		amount,
		e.CreatedAt,
		e.UpdatedAt,
	)
}

func (m *BidMapper) ToEntity(bid model.BidModel) entity.BidEntity {
	return entity.BidEntity{
		ID:            bid.ID(),
		AuctionID:     bid.AuctionID(),
		UserID:        bid.UserID(),
		AmountInCents: bid.Amount().AmountInCents(),
		CreatedAt:     bid.CreatedAt(),
		UpdatedAt:     bid.UpdatedAt(),
	}
}
