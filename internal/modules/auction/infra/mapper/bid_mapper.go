package mapper

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/entity"
)

type BidMapper struct{}

func NewBidMapper() *BidMapper {
	return &BidMapper{}
}

func (m *BidMapper) ToDomain(e entity.BidEntity) (model.BidModel, error) {
	amount, err := model.NewMoneyModel(e.AmountInCents, e.Currency)
	if err != nil {
		return model.BidModel{}, err
	}

	status, err := enum.NewBidStatusEnum(e.Status)
	if err != nil {
		return model.BidModel{}, err
	}

	return model.RestoreBidModel(
		e.ID,
		e.AuctionID,
		e.UserID,
		amount,
		status,
		e.CreatedAt,
		e.UpdatedAt,
	)
}

func (m *BidMapper) ToEntity(bid model.BidModel) entity.BidEntity {
	status := bid.Status()
	return entity.BidEntity{
		ID:            bid.ID(),
		AuctionID:     bid.AuctionID(),
		UserID:        bid.UserID(),
		AmountInCents: bid.Amount().AmountInCents(),
		Currency:      bid.Amount().Currency(),
		Status:        status.String(),
		CreatedAt:     bid.CreatedAt(),
		UpdatedAt:     bid.UpdatedAt(),
	}
}
