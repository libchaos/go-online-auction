package envelope_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/infra/event/envelope"
)

func TestBuildSubject(t *testing.T) {
	t.Run("spu subject", func(t *testing.T) {
		// Act
		subject := envelope.BuildSubject(envelope.AggregateTypeSpu, 42)

		// Assert
		require.Equal(t, "listing.evt.spu.42", subject)
	})

	t.Run("sku subject", func(t *testing.T) {
		// Act
		subject := envelope.BuildSubject(envelope.AggregateTypeSku, 7)

		// Assert
		require.Equal(t, "listing.evt.sku.7", subject)
	})
}

func TestFromSpuPublished(t *testing.T) {
	// Arrange
	evt := event.NewSpuPublishedEvent(42)

	// Act
	outboxEvent, err := envelope.FromSpuPublished(evt)

	// Assert
	require.NoError(t, err)
	require.Equal(t, event.SpuPublishedEventType, outboxEvent.EventType)
	require.Equal(t, evt.EventID(), outboxEvent.EventID)
	require.Equal(t, "listing.evt.spu.42", outboxEvent.Subject)
	require.Equal(t, envelope.SchemaVersion, outboxEvent.SchemaVersion)

	env, err := envelope.Decode(outboxEvent.Payload)
	require.NoError(t, err)
	require.Equal(t, envelope.AggregateTypeSpu, env.AggregateType)
	require.Equal(t, uint64(42), env.AggregateID)

	var data envelope.SpuPublishedData
	require.NoError(t, json.Unmarshal(env.Data, &data))
	require.Equal(t, uint64(42), data.SpuID)
}

func TestFromSkuPublished(t *testing.T) {
	// Arrange
	evt := event.NewSkuPublishedEvent(7, 42, 19900, 5)

	// Act
	outboxEvent, err := envelope.FromSkuPublished(evt)

	// Assert
	require.NoError(t, err)
	require.Equal(t, event.SkuPublishedEventType, outboxEvent.EventType)
	require.Equal(t, "listing.evt.sku.7", outboxEvent.Subject)

	env, err := envelope.Decode(outboxEvent.Payload)
	require.NoError(t, err)

	var data envelope.SkuPublishedData
	require.NoError(t, json.Unmarshal(env.Data, &data))
	require.Equal(t, uint64(7), data.SkuID)
	require.Equal(t, uint64(42), data.SpuID)
	require.Equal(t, uint64(19900), data.PriceInCents)
	require.Equal(t, uint64(5), data.Quantity)
}

func TestFromSkuOffShelf(t *testing.T) {
	// Arrange
	evt := event.NewSkuOffShelfEvent(7, 42)

	// Act
	outboxEvent, err := envelope.FromSkuOffShelf(evt)

	// Assert
	require.NoError(t, err)
	require.Equal(t, event.SkuOffShelfEventType, outboxEvent.EventType)
	require.Equal(t, "listing.evt.sku.7", outboxEvent.Subject)
}

func TestFromSpuOffShelf(t *testing.T) {
	// Arrange
	evt := event.NewSpuOffShelfEvent(42)

	// Act
	outboxEvent, err := envelope.FromSpuOffShelf(evt)

	// Assert
	require.NoError(t, err)
	require.Equal(t, event.SpuOffShelfEventType, outboxEvent.EventType)
	require.Equal(t, "listing.evt.spu.42", outboxEvent.Subject)
}

func TestDecode(t *testing.T) {
	t.Run("unsupported schema version returns error", func(t *testing.T) {
		// Arrange
		raw := []byte(`{"event_type":"listing.spu.published","event_id":"abc","schema_version":99}`)

		// Act
		_, err := envelope.Decode(raw)

		// Assert
		require.ErrorContains(t, err, "unsupported event schema version 99")
	})

	t.Run("malformed payload returns error", func(t *testing.T) {
		// Act
		_, err := envelope.Decode([]byte("not-json"))

		// Assert
		require.ErrorContains(t, err, "failed to decode event envelope")
	})
}
