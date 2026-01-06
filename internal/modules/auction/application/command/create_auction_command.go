package command

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/pkg/logger"
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
	if err := validateCreateAuctionInput(input); err != nil {
		return CreateAuctionCommandOutput{}, err
	}

	auction, err := model.NewAuctionModel(input.ListingID, input.EndTime)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create auction domain model")
		return CreateAuctionCommandOutput{}, err
	}

	err = c.auctionRepository.Create(ctx, auction)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist auction")
		return CreateAuctionCommandOutput{}, err
	}

	state := auction.State()
	return CreateAuctionCommandOutput{
		ID:        auction.ID(),
		ListingID: auction.ListingID(),
		State:     state.String(),
		EndTime:   auction.EndTime(),
		CreatedAt: auction.CreatedAt(),
	}, nil
}

func validateCreateAuctionInput(input CreateAuctionCommandInput) error {
	if input.EndTime.Before(time.Now().UTC()) || input.EndTime.Equal(time.Now().UTC()) {
		return errs.ErrEndTimeMustBeInFuture
	}
	return nil
}
