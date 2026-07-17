// Package envelope serializes deposit domain events into versioned envelopes
// recorded in the shared transactional outbox and published to the
// DEPOSIT_EVENTS stream.
package envelope

import (
	"encoding/json"
	"fmt"
	"time"

	"auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/ports"
)

const SchemaVersion = 1

type Envelope struct {
	EventType     string          `json:"event_type"`
	EventID       string          `json:"event_id"`
	SchemaVersion int             `json:"schema_version"`
	Timestamp     time.Time       `json:"timestamp"`
	UserID        uint64          `json:"user_id"`
	AuctionID     uint64          `json:"auction_id"`
	Data          json.RawMessage `json:"data"`
}

type DepositData struct {
	DepositID         uint64 `json:"deposit_id"`
	UserID            uint64 `json:"user_id"`
	AuctionID         uint64 `json:"auction_id"`
	AmountInCents     uint64 `json:"amount_in_cents"`
	Currency          string `json:"currency"`
	Status            string `json:"status"`
	ExternalReference string `json:"external_reference,omitempty"`
}

func BuildSubject(userID uint64) string {
	return fmt.Sprintf("deposit.evt.%d", userID)
}

func FromDepositHeld(evt event.DepositHeldEvent) (ports.OutboxEvent, error) {
	data := DepositData{
		DepositID:         evt.DepositID(),
		UserID:            evt.UserID(),
		AuctionID:         evt.AuctionID(),
		AmountInCents:     evt.Amount().AmountInCents(),
		Currency:          evt.Currency(),
		Status:            "held",
		ExternalReference: evt.ExternalReference(),
	}

	return build(event.DepositHeldEventType, evt.EventID(), evt.Timestamp(), evt.UserID(), evt.AuctionID(), data)
}

func FromDepositReleased(evt event.DepositReleasedEvent) (ports.OutboxEvent, error) {
	data := DepositData{
		DepositID:     evt.DepositID(),
		UserID:        evt.UserID(),
		AuctionID:     evt.AuctionID(),
		AmountInCents: evt.Amount().AmountInCents(),
		Currency:      evt.Currency(),
		Status:        "released",
	}

	return build(event.DepositReleasedEventType, evt.EventID(), evt.Timestamp(), evt.UserID(), evt.AuctionID(), data)
}

func FromDepositApplied(evt event.DepositAppliedEvent) (ports.OutboxEvent, error) {
	data := DepositData{
		DepositID:     evt.DepositID(),
		UserID:        evt.UserID(),
		AuctionID:     evt.AuctionID(),
		AmountInCents: evt.Amount().AmountInCents(),
		Currency:      evt.Currency(),
		Status:        "applied",
	}

	return build(event.DepositAppliedEventType, evt.EventID(), evt.Timestamp(), evt.UserID(), evt.AuctionID(), data)
}

func FromDepositForfeited(evt event.DepositForfeitedEvent) (ports.OutboxEvent, error) {
	data := DepositData{
		DepositID:     evt.DepositID(),
		UserID:        evt.UserID(),
		AuctionID:     evt.AuctionID(),
		AmountInCents: evt.Amount().AmountInCents(),
		Currency:      evt.Currency(),
		Status:        "forfeited",
	}

	return build(event.DepositForfeitedEventType, evt.EventID(), evt.Timestamp(), evt.UserID(), evt.AuctionID(), data)
}

func build(
	eventType string,
	eventID string,
	timestamp time.Time,
	userID uint64,
	auctionID uint64,
	data DepositData,
) (ports.OutboxEvent, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return ports.OutboxEvent{}, fmt.Errorf("failed to marshal %s data: %w", eventType, err)
	}

	payload, err := json.Marshal(Envelope{
		EventType:     eventType,
		EventID:       eventID,
		SchemaVersion: SchemaVersion,
		Timestamp:     timestamp,
		UserID:        userID,
		AuctionID:     auctionID,
		Data:          rawData,
	})
	if err != nil {
		return ports.OutboxEvent{}, fmt.Errorf("failed to marshal %s envelope: %w", eventType, err)
	}

	return ports.OutboxEvent{
		EventID:       eventID,
		EventType:     eventType,
		SchemaVersion: SchemaVersion,
		Subject:       BuildSubject(userID),
		Payload:       payload,
		OccurredAt:    timestamp,
	}, nil
}
