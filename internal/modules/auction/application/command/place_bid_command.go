package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

type PlaceBidCommandInput struct {
	AuctionID        uint64
	UserID           uint64
	AmountInCents    uint64
	MaxAmountInCents *uint64
	IdempotencyKey   string
}

type PlaceBidCommandOutput struct {
	IdempotencyKey string
	Status         string
}

const bidAcceptedStatus = "accepted"

type PlaceBidCommand struct {
	bidCommandPublisher ports.BidCommandPublisher
	logger              logger.Logger
}

func NewPlaceBidCommand(
	bidCommandPublisher ports.BidCommandPublisher,
	logger logger.Logger,
) *PlaceBidCommand {
	return &PlaceBidCommand{
		bidCommandPublisher: bidCommandPublisher,
		logger:              logger,
	}
}

func (c *PlaceBidCommand) Execute(
	ctx context.Context,
	input PlaceBidCommandInput,
) (PlaceBidCommandOutput, error) {
	if input.AmountInCents == 0 {
		return PlaceBidCommandOutput{}, errs.ErrFirstBidMustBePositive
	}

	idempotencyKey := input.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}

	cmd := ports.BidCommand{
		IdempotencyKey:   idempotencyKey,
		AuctionID:        input.AuctionID,
		UserID:           input.UserID,
		AmountInCents:    input.AmountInCents,
		MaxAmountInCents: input.MaxAmountInCents,
		IssuedAt:         time.Now().UTC(),
	}

	ack, err := c.bidCommandPublisher.Publish(ctx, cmd)
	if err != nil {
		c.logger.Error().Err(err).
			Uint64("auction_id", input.AuctionID).
			Uint64("user_id", input.UserID).
			Str("idempotency_key", idempotencyKey).
			Msg("failed to publish bid command")
		return PlaceBidCommandOutput{}, err
	}

	return PlaceBidCommandOutput{
		IdempotencyKey: ack.IdempotencyKey,
		Status:         bidAcceptedStatus,
	}, nil
}
