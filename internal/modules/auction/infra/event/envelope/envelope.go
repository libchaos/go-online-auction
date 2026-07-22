// Package envelope serializes domain events into versioned envelopes recorded
// in the transactional outbox and published to the AUCTION_EVENTS stream.
package envelope

import (
	"encoding/json"
	"fmt"
	"time"

	"auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/ports"
)

// SchemaVersion is the current version of every event payload schema. Bump it
// when a payload changes shape incompatibly; consumers switch on the
// schema_version field to decode historical events (see Decode).
const SchemaVersion = 1

// Envelope is the wire format of every auction domain event. The envelope
// fields are stable across schema versions; only Data evolves.
type Envelope struct {
	EventType     string          `json:"event_type"`
	EventID       string          `json:"event_id"`
	SchemaVersion int             `json:"schema_version"`
	Timestamp     time.Time       `json:"timestamp"`
	AuctionID     uint64          `json:"auction_id"`
	Data          json.RawMessage `json:"data"`
}

type BidPlacedData struct {
	BidID  uint64       `json:"bid_id"`
	UserID uint64       `json:"user_id"`
	Amount MoneyPayload `json:"amount"`
}

type AuctionStartedData struct {
	ListingID uint64    `json:"listing_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type AuctionEndedData struct {
	WinningBidID *uint64       `json:"winning_bid_id,omitempty"`
	FinalAmount  *MoneyPayload `json:"final_amount,omitempty"`
}

type AuctionCreatedData struct {
	ListingID          uint64     `json:"listing_id"`
	TradingMode        string     `json:"trading_mode"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            time.Time  `json:"end_time"`
	StartingPrice      *uint64    `json:"starting_price,omitempty"`
	PriceStep          *uint64    `json:"price_step,omitempty"`
	ReservePrice       *uint64    `json:"reserve_price,omitempty"`
	AntiSnipeEnabled   bool       `json:"anti_snipe_enabled"`
	ExtensionWindowSec int64      `json:"extension_window_sec"`
}

type MoneyPayload struct {
	AmountInCents uint64 `json:"amount_in_cents"`
}

// BuildSubject builds the JetStream subject for an auction event
func BuildSubject(auctionID uint64) string {
	return fmt.Sprintf("auction.evt.%d", auctionID)
}

// FromBidPlaced converts a BidPlacedEvent into an outbox record
func FromBidPlaced(evt event.BidPlacedEvent) (ports.OutboxEvent, error) {
	data := BidPlacedData{
		BidID:  evt.BidID(),
		UserID: evt.UserID(),
		Amount: MoneyPayload{AmountInCents: evt.Amount().AmountInCents()},
	}
	return build(event.BidPlacedEventType, evt.EventID(), evt.Timestamp(), evt.AuctionID(), data)
}

// FromAuctionStarted converts an AuctionStartedEvent into an outbox record
func FromAuctionStarted(evt event.AuctionStartedEvent) (ports.OutboxEvent, error) {
	data := AuctionStartedData{
		ListingID: evt.ListingID(),
		StartTime: evt.StartTime(),
		EndTime:   evt.EndTime(),
	}
	return build(event.AuctionStartedEventType, evt.EventID(), evt.Timestamp(), evt.AuctionID(), data)
}

// FromAuctionEnded converts an AuctionEndedEvent into an outbox record
func FromAuctionEnded(evt event.AuctionEndedEvent) (ports.OutboxEvent, error) {
	data := AuctionEndedData{
		WinningBidID: evt.WinningBidID(),
	}
	if evt.FinalAmount() != nil {
		data.FinalAmount = &MoneyPayload{AmountInCents: evt.FinalAmount().AmountInCents()}
	}
	return build(event.AuctionEndedEventType, evt.EventID(), evt.Timestamp(), evt.AuctionID(), data)
}

// FromAuctionCreated converts an AuctionCreatedEvent into an outbox record so
// the auction lifecycle event stream is complete (create/start/close/place_bid).
func FromAuctionCreated(evt event.AuctionCreatedEvent) (ports.OutboxEvent, error) {
	data := AuctionCreatedData{
		ListingID:          evt.ListingID(),
		TradingMode:        evt.TradingMode(),
		StartTime:          evt.StartTime(),
		EndTime:            evt.EndTime(),
		StartingPrice:      evt.StartingPrice(),
		PriceStep:          evt.PriceStep(),
		ReservePrice:       evt.ReservePrice(),
		AntiSnipeEnabled:   evt.AntiSnipeEnabled(),
		ExtensionWindowSec: evt.ExtensionWindowSec(),
	}
	return build(event.AuctionCreatedEventType, evt.EventID(), evt.Timestamp(), evt.AuctionID(), data)
}

// Decode parses an envelope from its wire format. Unknown schema versions are
// rejected so consumers fail loudly instead of misinterpreting payloads; when
// SchemaVersion is bumped, add upcasting from older versions here.
func Decode(raw []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return Envelope{}, fmt.Errorf("failed to decode event envelope: %w", err)
	}

	switch env.SchemaVersion {
	case SchemaVersion:
		return env, nil
	default:
		return Envelope{}, fmt.Errorf("unsupported event schema version %d for event %s",
			env.SchemaVersion, env.EventID)
	}
}

func build(eventType, eventID string, timestamp time.Time, auctionID uint64, data any) (ports.OutboxEvent, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return ports.OutboxEvent{}, fmt.Errorf("failed to marshal %s data: %w", eventType, err)
	}

	payload, err := json.Marshal(Envelope{
		EventType:     eventType,
		EventID:       eventID,
		SchemaVersion: SchemaVersion,
		Timestamp:     timestamp,
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
		Subject:       BuildSubject(auctionID),
		Payload:       payload,
		OccurredAt:    timestamp,
	}, nil
}
