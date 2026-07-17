package websocket

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"auction/internal/shared/modules/logger"
)

const (
	userIDTokenPosition      = 2
	minSubjectPartsForID     = userIDTokenPosition + 1
	eventBroadcastBufferSize = 256
)

type EventConsumer interface {
	Consume(ctx context.Context, handler func(subject string, data []byte)) error
}

type EventMessage struct {
	Type      string          `json:"type"`
	EventType string          `json:"event_type"`
	EventID   string          `json:"event_id"`
	Timestamp time.Time       `json:"timestamp"`
	UserID    uint64          `json:"user_id"`
	AuctionID uint64          `json:"auction_id"`
	Data      json.RawMessage `json:"data"`
}

type broadcastEvent struct {
	subject string
	data    []byte
}

type Hub struct {
	registry      *UserSubscriberRegistry
	eventConsumer EventConsumer
	register      chan *Client
	unregister    chan *Client
	events        chan broadcastEvent
	logger        logger.Logger
}

func NewHub(
	registry *UserSubscriberRegistry,
	eventConsumer EventConsumer,
	logger logger.Logger,
) *Hub {
	return &Hub{
		registry:      registry,
		eventConsumer: eventConsumer,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		events:        make(chan broadcastEvent, eventBroadcastBufferSize),
		logger:        logger,
	}
}

func (hub *Hub) Run(ctx context.Context) {
	go func() {
		consumeErr := hub.eventConsumer.Consume(ctx, func(subject string, data []byte) {
			select {
			case hub.events <- broadcastEvent{subject: subject, data: data}:
			case <-ctx.Done():
			}
		})
		if consumeErr != nil {
			hub.logger.Error().Err(consumeErr).Msg("failed to consume deposit events")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-hub.register:
			hub.registry.Add(client.userID, client)
		case client := <-hub.unregister:
			hub.registry.Remove(client.userID, client)
			client.Close()
		case evt := <-hub.events:
			userID := extractUserID(evt.subject)
			if userID == 0 {
				hub.logger.Warn().
					Str("subject", evt.subject).
					Msg("failed to extract user id from deposit event subject")

				continue
			}

			hub.registry.PublishToUser(userID, evt.data)
		}
	}
}

func (hub *Hub) RegisterClient(conn *websocket.Conn, userID uint64) {
	client := NewClient(conn, userID, hub)
	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (hub *Hub) Shutdown() error {
	return nil
}

func extractUserID(subject string) uint64 {
	parts := strings.Split(subject, ".")
	if len(parts) >= minSubjectPartsForID {
		id, err := strconv.ParseUint(parts[userIDTokenPosition], 10, 64)
		if err != nil {
			return 0
		}

		return id
	}

	return 0
}
