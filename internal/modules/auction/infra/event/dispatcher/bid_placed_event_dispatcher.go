package dispatcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/ports"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/redis/go-redis/v9"
)

type BidPlacedPayload struct {
	EventType string               `json:"event_type"`
	EventID   string               `json:"event_id"`
	Timestamp time.Time            `json:"timestamp"`
	AuctionID uint64               `json:"auction_id"`
	Data      BidPlacedPayloadData `json:"data"`
}

type BidPlacedPayloadData struct {
	BidID  uint64       `json:"bid_id"`
	UserID uint64       `json:"user_id"`
	Amount MoneyPayload `json:"amount"`
}

type MoneyPayload struct {
	AmountInCents uint64 `json:"amount_in_cents"`
}

type RedisBidPlacedEventDispatcher struct {
	redisClient redis.UniversalClient
	logger      logger.Logger
}

func NewRedisBidPlacedEventDispatcher(
	redisClient redis.UniversalClient,
	logger logger.Logger,
) *RedisBidPlacedEventDispatcher {
	return &RedisBidPlacedEventDispatcher{
		redisClient: redisClient,
		logger:      logger,
	}
}

var _ ports.BidPlacedEventDispatcher = (*RedisBidPlacedEventDispatcher)(nil)

func (d *RedisBidPlacedEventDispatcher) Dispatch(ctx context.Context, evt event.BidPlacedEvent) error {
	payload := BidPlacedPayload{
		EventType: event.BidPlacedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: evt.AuctionID(),
		Data: BidPlacedPayloadData{
			BidID:  evt.BidID(),
			UserID: evt.UserID(),
			Amount: MoneyPayload{
				AmountInCents: evt.Amount().AmountInCents(),
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		d.logger.Error().
			Err(err).
			Uint64("auction_id", evt.AuctionID()).
			Str("event_id", evt.EventID()).
			Msg("failed to marshal BidPlacedEvent")
		return err
	}

	channel := BuildAuctionEventChannel(evt.AuctionID())
	if pubErr := d.redisClient.Publish(ctx, channel, jsonData).Err(); pubErr != nil {
		d.logger.Error().
			Err(pubErr).
			Str("channel", channel).
			Uint64("auction_id", evt.AuctionID()).
			Str("event_id", evt.EventID()).
			Msg("failed to publish BidPlacedEvent to Redis")
		return pubErr
	}

	d.logger.Debug().
		Str("channel", channel).
		Uint64("auction_id", evt.AuctionID()).
		Str("event_id", evt.EventID()).
		Uint64("bid_id", evt.BidID()).
		Msg("BidPlacedEvent dispatched successfully")

	return nil
}
