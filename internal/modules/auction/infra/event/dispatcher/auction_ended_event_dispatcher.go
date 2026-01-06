package dispatcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type AuctionEndedPayload struct {
	EventType string                  `json:"event_type"`
	EventID   string                  `json:"event_id"`
	Timestamp time.Time               `json:"timestamp"`
	AuctionID uint64                  `json:"auction_id"`
	Data      AuctionEndedPayloadData `json:"data"`
}

type AuctionEndedPayloadData struct {
	WinningBidID *uint64       `json:"winning_bid_id,omitempty"`
	FinalAmount  *MoneyPayload `json:"final_amount,omitempty"`
}

type RedisAuctionEndedEventDispatcher struct {
	redisClient redis.UniversalClient
	logger      logger.Logger
}

func NewRedisAuctionEndedEventDispatcher(
	redisClient redis.UniversalClient,
	logger logger.Logger,
) *RedisAuctionEndedEventDispatcher {
	return &RedisAuctionEndedEventDispatcher{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (d *RedisAuctionEndedEventDispatcher) Dispatch(ctx context.Context, evt event.AuctionEndedEvent) error {
	payloadData := AuctionEndedPayloadData{
		WinningBidID: evt.WinningBidID(),
	}

	if finalAmount := evt.FinalAmount(); finalAmount != nil {
		payloadData.FinalAmount = &MoneyPayload{
			AmountInCents: finalAmount.AmountInCents(),
			Currency:      finalAmount.Currency(),
		}
	}

	payload := AuctionEndedPayload{
		EventType: event.AuctionEndedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: evt.AuctionID(),
		Data:      payloadData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		d.logger.Error().
			Err(err).
			Uint64("auction_id", evt.AuctionID()).
			Str("event_id", evt.EventID()).
			Msg("failed to marshal AuctionEndedEvent")
		return err
	}

	channel := BuildAuctionEventChannel(evt.AuctionID())
	if pubErr := d.redisClient.Publish(ctx, channel, jsonData).Err(); pubErr != nil {
		d.logger.Error().
			Err(pubErr).
			Str("channel", channel).
			Uint64("auction_id", evt.AuctionID()).
			Str("event_id", evt.EventID()).
			Msg("failed to publish AuctionEndedEvent to Redis")
		return pubErr
	}

	d.logger.Debug().
		Str("channel", channel).
		Uint64("auction_id", evt.AuctionID()).
		Str("event_id", evt.EventID()).
		Msg("AuctionEndedEvent dispatched successfully")

	return nil
}
