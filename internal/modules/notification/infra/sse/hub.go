package sse

import (
	"context"
	"strconv"
	"strings"

	"auction/internal/shared/modules/logger"
)

const (
	userIDTokenPosition  = 2
	minSubjectPartsForID = userIDTokenPosition + 1
)

// EventConsumer streams notification.evt.{userID} events out of JetStream. The
// concrete implementation lives in the messaging package.
type EventConsumer interface {
	Consume(ctx context.Context, handler func(subject string, data []byte)) error
}

// RealtimeHub bridges the notification event stream to connected SSE clients. It
// consumes notification.evt.{userID} events, extracts the recipient id from the
// subject, and fans each event out to that user's clients. Client membership is
// managed directly through the mutex-guarded registry so registration never
// blocks on the consume loop and shutdown can never deadlock a request.
type RealtimeHub struct {
	registry      *SubscriberRegistry
	eventConsumer EventConsumer
	logger        logger.Logger
}

func NewRealtimeHub(
	registry *SubscriberRegistry,
	eventConsumer EventConsumer,
	logger logger.Logger,
) *RealtimeHub {
	return &RealtimeHub{
		registry:      registry,
		eventConsumer: eventConsumer,
		logger:        logger,
	}
}

// Run blocks consuming notification events until the context is cancelled. It is
// launched in a background goroutine by the fx lifecycle hook.
func (hub *RealtimeHub) Run(ctx context.Context) {
	consumeErr := hub.eventConsumer.Consume(ctx, func(subject string, data []byte) {
		userID := extractUserID(subject)
		if userID == 0 {
			hub.logger.Warn().
				Str("subject", subject).
				Msg("failed to extract user id from notification event subject")

			return
		}

		hub.registry.PublishToUser(userID, data)
	})
	if consumeErr != nil && ctx.Err() == nil {
		hub.logger.Error().Err(consumeErr).Msg("failed to consume notification events")
	}
}

func (hub *RealtimeHub) AddClient(client *Client) {
	hub.registry.Add(client.UserID(), client)
}

func (hub *RealtimeHub) RemoveClient(client *Client) {
	hub.registry.Remove(client.UserID(), client)
	client.Close()
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
