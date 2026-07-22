package websocket

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"auction/internal/modules/auction/infra/messaging"
	"auction/internal/shared/modules/logger"

	"github.com/gorilla/websocket"
)

const (
	auctionIDTokenPosition   = 2
	minSubjectPartsForID     = auctionIDTokenPosition + 1
	eventBroadcastBufferSize = 256
)

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

type broadcastEvent struct {
	subject string
	data    []byte
}

type Hub struct {
	subscriberRegistry *AuctionSubscriberRegistry
	eventConsumer      messaging.EventConsumer
	register           chan *Client
	unregister         chan *Client
	events             chan broadcastEvent
	logger             logger.Logger
}

func NewHub(
	subscriberRegistry *AuctionSubscriberRegistry,
	eventConsumer messaging.EventConsumer,
	logger logger.Logger,
) *Hub {
	return &Hub{
		subscriberRegistry: subscriberRegistry,
		eventConsumer:      eventConsumer,
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		events:             make(chan broadcastEvent, eventBroadcastBufferSize),
		logger:             logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	go func() {
		consumeErr := h.eventConsumer.Consume(ctx, func(subject string, data []byte) {
			select {
			case h.events <- broadcastEvent{subject: subject, data: data}:
			case <-ctx.Done():
			}
		})
		if consumeErr != nil {
			h.logger.Error().Err(consumeErr).Msg("failed to consume auction events")
		}
	}()

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
		case evt := <-h.events:
			auctionID := extractAuctionID(evt.subject)
			if auctionID == 0 {
				h.logger.Warn().Str("subject", evt.subject).Msg("failed to extract auction ID from subject")
				continue
			}
			h.subscriberRegistry.Broadcast(auctionID, evt.data)
			messaging.IncWebsocketBroadcast()
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
	return nil
}

func extractAuctionID(subject string) uint64 {
	parts := strings.Split(subject, ".")
	if len(parts) >= minSubjectPartsForID {
		id, err := strconv.ParseUint(parts[auctionIDTokenPosition], 10, 64)
		if err != nil {
			return 0
		}
		return id
	}
	return 0
}
