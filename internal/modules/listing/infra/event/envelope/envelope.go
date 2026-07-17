// Package envelope serializes listing domain events into versioned envelopes
// recorded in the transactional outbox and published to the LISTING_EVENTS
// stream by the shared outbox relay.
package envelope

import (
	"encoding/json"
	"fmt"
	"time"

	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/ports"
)

// SchemaVersion is the current version of every event payload schema. Bump it
// when a payload changes shape incompatibly; consumers switch on the
// schema_version field to decode historical events (see Decode).
const SchemaVersion = 1

// Aggregate types carried in the envelope and the subject.
const (
	AggregateTypeSpu = "spu"
	AggregateTypeSku = "sku"
)

// Envelope is the wire format of every listing domain event. The envelope
// fields are stable across schema versions; only Data evolves.
type Envelope struct {
	EventType     string          `json:"event_type"`
	EventID       string          `json:"event_id"`
	SchemaVersion int             `json:"schema_version"`
	Timestamp     time.Time       `json:"timestamp"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uint64          `json:"aggregate_id"`
	Data          json.RawMessage `json:"data"`
}

type SpuPublishedData struct {
	SpuID uint64 `json:"spu_id"`
}

type SpuOffShelfData struct {
	SpuID uint64 `json:"spu_id"`
}

type SkuPublishedData struct {
	SkuID        uint64 `json:"sku_id"`
	SpuID        uint64 `json:"spu_id"`
	PriceInCents uint64 `json:"price_in_cents"`
	Quantity     uint64 `json:"quantity"`
}

type SkuOffShelfData struct {
	SkuID uint64 `json:"sku_id"`
	SpuID uint64 `json:"spu_id"`
}

// BuildSubject builds the JetStream subject for a listing event,
// e.g. listing.evt.sku.42
func BuildSubject(aggregateType string, aggregateID uint64) string {
	return fmt.Sprintf("listing.evt.%s.%d", aggregateType, aggregateID)
}

// FromSpuPublished converts a SpuPublishedEvent into an outbox record
func FromSpuPublished(evt event.SpuPublishedEvent) (ports.OutboxEvent, error) {
	data := SpuPublishedData{SpuID: evt.SpuID()}
	return build(event.SpuPublishedEventType, evt.EventID(), evt.Timestamp(), AggregateTypeSpu, evt.SpuID(), data)
}

// FromSpuOffShelf converts a SpuOffShelfEvent into an outbox record
func FromSpuOffShelf(evt event.SpuOffShelfEvent) (ports.OutboxEvent, error) {
	data := SpuOffShelfData{SpuID: evt.SpuID()}
	return build(event.SpuOffShelfEventType, evt.EventID(), evt.Timestamp(), AggregateTypeSpu, evt.SpuID(), data)
}

// FromSkuPublished converts a SkuPublishedEvent into an outbox record
func FromSkuPublished(evt event.SkuPublishedEvent) (ports.OutboxEvent, error) {
	data := SkuPublishedData{
		SkuID:        evt.SkuID(),
		SpuID:        evt.SpuID(),
		PriceInCents: evt.PriceInCents(),
		Quantity:     evt.Quantity(),
	}
	return build(event.SkuPublishedEventType, evt.EventID(), evt.Timestamp(), AggregateTypeSku, evt.SkuID(), data)
}

// FromSkuOffShelf converts a SkuOffShelfEvent into an outbox record
func FromSkuOffShelf(evt event.SkuOffShelfEvent) (ports.OutboxEvent, error) {
	data := SkuOffShelfData{SkuID: evt.SkuID(), SpuID: evt.SpuID()}
	return build(event.SkuOffShelfEventType, evt.EventID(), evt.Timestamp(), AggregateTypeSku, evt.SkuID(), data)
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

func build(
	eventType, eventID string,
	timestamp time.Time,
	aggregateType string,
	aggregateID uint64,
	data any,
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
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Data:          rawData,
	})
	if err != nil {
		return ports.OutboxEvent{}, fmt.Errorf("failed to marshal %s envelope: %w", eventType, err)
	}

	return ports.OutboxEvent{
		EventID:       eventID,
		EventType:     eventType,
		SchemaVersion: SchemaVersion,
		Subject:       BuildSubject(aggregateType, aggregateID),
		Payload:       payload,
		OccurredAt:    timestamp,
	}, nil
}
