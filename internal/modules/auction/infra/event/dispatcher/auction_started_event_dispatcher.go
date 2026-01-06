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

type AuctionStartedPayload struct {
	EventType string                    `json:"event_type"`
	EventID   string                    `json:"event_id"`
	Timestamp time.Time                 `json:"timestamp"`
	AuctionID uint64                    `json:"auction_id"`
	Data      AuctionStartedPayloadData `json:"data"`
}

type AuctionStartedPayloadData struct {
	ListingID uint64    `json:"listing_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type RedisAuctionStartedEventDispatcher struct {
	redisClient redis.UniversalClient
	logger      logger.Logger
}

func NewRedisAuctionStartedEventDispatcher(
	redisClient redis.UniversalClient,
	logger logger.Logger,
) *RedisAuctionStartedEventDispatcher {
	return &RedisAuctionStartedEventDispatcher{
		redisClient: redisClient,
		logger:      logger,
	}
}

var _ ports.AuctionStartedEventDispatcher = (*RedisAuctionStartedEventDispatcher)(nil)

func (d *RedisAuctionStartedEventDispatcher) Dispatch(ctx context.Context, evt event.AuctionStartedEvent) error {
	payload := AuctionStartedPayload{
		EventType: event.AuctionStartedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: evt.AuctionID(),
		Data: AuctionStartedPayloadData{
			ListingID: evt.ListingID(),
			StartTime: evt.StartTime(),
			EndTime:   evt.EndTime(),
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		d.logger.Error().
			Err(err).
			Uint64("auction_id", evt.AuctionID()).
			Str("event_id", evt.EventID()).
			Msg("failed to marshal AuctionStartedEvent")
		return err
	}

	channel := BuildAuctionEventChannel(evt.AuctionID())
	if pubErr := d.redisClient.Publish(ctx, channel, jsonData).Err(); pubErr != nil {
		d.logger.Error().
			Err(pubErr).
			Str("channel", channel).
			Uint64("auction_id", evt.AuctionID()).
			Str("event_id", evt.EventID()).
			Msg("failed to publish AuctionStartedEvent to Redis")
		return pubErr
	}

	d.logger.Debug().
		Str("channel", channel).
		Uint64("auction_id", evt.AuctionID()).
		Str("event_id", evt.EventID()).
		Msg("AuctionStartedEvent dispatched successfully")

	return nil
}
