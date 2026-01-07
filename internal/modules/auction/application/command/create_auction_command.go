package command

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
)

type CreateAuctionCommandInput struct {
	ListingID uint64
	EndTime   time.Time
}

type CreateAuctionCommandOutput struct {
	ID        uint64
	ListingID uint64
	State     string
	EndTime   time.Time
	CreatedAt time.Time
}

type CreateAuctionCommand struct {
	auctionRepository ports.AuctionRepository
	logger            logger.Logger
}

func NewCreateAuctionCommand(
	auctionRepository ports.AuctionRepository,
	logger logger.Logger,
) *CreateAuctionCommand {
	return &CreateAuctionCommand{
		auctionRepository: auctionRepository,
		logger:            logger,
	}
}

func (c *CreateAuctionCommand) Execute(
	ctx context.Context,
	input CreateAuctionCommandInput,
) (CreateAuctionCommandOutput, error) {
	auction, err := model.NewAuctionModel(input.ListingID, input.EndTime)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create auction domain model")
		return CreateAuctionCommandOutput{}, err
	}

	persistedAuction, err := c.auctionRepository.Create(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist auction")
		return CreateAuctionCommandOutput{}, err
	}

	state := persistedAuction.State()
	return CreateAuctionCommandOutput{
		ID:        persistedAuction.ID(),
		ListingID: persistedAuction.ListingID(),
		State:     state.String(),
		EndTime:   persistedAuction.EndTime(),
		CreatedAt: persistedAuction.CreatedAt(),
	}, nil
}
