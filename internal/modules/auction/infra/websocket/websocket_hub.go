package websocket

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const minChannelPartsForAuctionID = 2

type SubscriptionConfirmed struct {
	Type      string `json:"type"`
	AuctionID uint64 `json:"auction_id"`
}

type EventMessage struct {
	Type      string          `json:"type"`
	EventType string          `json:"event_type"`
	EventID   string          `json:"event_id"`
	Timestamp time.Time       `json:"timestamp"`
	AuctionID uint64          `json:"auction_id"`
	Data      json.RawMessage `json:"data"`
}

type Hub struct {
	subscriberRegistry *AuctionSubscriberRegistry
	redisClient        redis.UniversalClient
	pubsub             *redis.PubSub
	register           chan *Client
	unregister         chan *Client
	logger             logger.Logger
}

func NewHub(
	subscriberRegistry *AuctionSubscriberRegistry,
	redisClient redis.UniversalClient,
	logger logger.Logger,
) *Hub {
	return &Hub{
		subscriberRegistry: subscriberRegistry,
		redisClient:        redisClient,
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		logger:             logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	h.pubsub = h.redisClient.PSubscribe(ctx, "auction:*:events")
	channel := h.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.subscriberRegistry.Add(client.auctionID, client)
			message, err := json.Marshal(SubscriptionConfirmed{
				Type:      "subscription_confirmed",
				AuctionID: client.auctionID,
			})
			if err != nil {
				h.logger.Error().Err(err).Msg("failed to marshal subscription confirmed message")
				continue
			}
			client.send <- message
		case client := <-h.unregister:
			h.subscriberRegistry.Remove(client.auctionID, client)
			client.Close()
		case message, ok := <-channel:
			if !ok || message == nil {
				return
			}
			auctionID := extractAuctionID(message.Channel)
			if auctionID == 0 {
				h.logger.Warn().Str("channel", message.Channel).Msg("failed to extract auction ID from channel")
				continue
			}
			h.subscriberRegistry.Broadcast(auctionID, []byte(message.Payload))
		}
	}
}

func (h *Hub) RegisterClient(conn *websocket.Conn, auctionID uint64) {
	client := NewClient(conn, auctionID, h)
	h.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *Hub) Shutdown() error {
	if h.pubsub != nil {
		return h.pubsub.Close()
	}
	return nil
}

func extractAuctionID(channel string) uint64 {
	parts := strings.Split(channel, ":")
	if len(parts) >= minChannelPartsForAuctionID {
		id, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return 0
		}
		return id
	}
	return 0
}
