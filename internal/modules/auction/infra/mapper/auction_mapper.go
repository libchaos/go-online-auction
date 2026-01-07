package mapper

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/entity"
)

type AuctionMapper struct{}

func NewAuctionMapper() *AuctionMapper {
	return &AuctionMapper{}
}

func (m *AuctionMapper) ToDomain(e entity.AuctionEntity) (model.AuctionModel, error) {
	state, err := enum.NewAuctionStateEnum(e.State)
	if err != nil {
		return model.AuctionModel{}, err
	}

	return model.RestoreAuctionModel(
		e.ID,
		e.ListingID,
		e.StartTime,
		e.EndTime,
		state,
		e.HighestBidID,
		nil, // TODO: Map highestBidAmount in Task 2.0
		e.Version,
		e.CreatedAt,
		e.UpdatedAt,
	)
}

func (m *AuctionMapper) ToEntity(auction model.AuctionModel) entity.AuctionEntity {
	state := auction.State()
	return entity.AuctionEntity{
		ID:           auction.ID(),
		ListingID:    auction.ListingID(),
		StartTime:    auction.StartTime(),
		EndTime:      auction.EndTime(),
		State:        state.String(),
		HighestBidID: auction.HighestBidID(),
		Version:      auction.Version(),
		CreatedAt:    auction.CreatedAt(),
		UpdatedAt:    auction.UpdatedAt(),
	}
}
