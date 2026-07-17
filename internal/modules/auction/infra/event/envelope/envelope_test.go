package envelope_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/event/envelope"
)

func TestFromBidPlaced_BuildsVersionedEnvelope(t *testing.T) {
	// Arrange
	money := model.NewMoneyModel(5000)
	evt := event.NewBidPlacedEvent(10, 1, 100, money)

	// Act
	outboxEvent, err := envelope.FromBidPlaced(evt)

	// Assert
	require.NoError(t, err)
	require.Equal(t, evt.EventID(), outboxEvent.EventID)
	require.Equal(t, event.BidPlacedEventType, outboxEvent.EventType)
	require.Equal(t, envelope.SchemaVersion, outboxEvent.SchemaVersion)
	require.Equal(t, "auction.evt.1", outboxEvent.Subject)

	env, err := envelope.Decode(outboxEvent.Payload)
	require.NoError(t, err)
	require.Equal(t, evt.EventID(), env.EventID)
	require.Equal(t, uint64(1), env.AuctionID)

	var data envelope.BidPlacedData
	require.NoError(t, json.Unmarshal(env.Data, &data))
	require.Equal(t, uint64(10), data.BidID)
	require.Equal(t, uint64(100), data.UserID)
	require.Equal(t, uint64(5000), data.Amount.AmountInCents)
}

func TestFromAuctionEnded_NoWinner_OmitsWinnerFields(t *testing.T) {
	// Arrange
	evt := event.NewAuctionEndedEvent(2, nil, nil)

	// Act
	outboxEvent, err := envelope.FromAuctionEnded(evt)

	// Assert
	require.NoError(t, err)

	env, err := envelope.Decode(outboxEvent.Payload)
	require.NoError(t, err)

	var data envelope.AuctionEndedData
	require.NoError(t, json.Unmarshal(env.Data, &data))
	require.Nil(t, data.WinningBidID)
	require.Nil(t, data.FinalAmount)
}

func TestDecode_UnsupportedSchemaVersion_ReturnsError(t *testing.T) {
	// Arrange: an event written by a hypothetical future (or corrupted) producer
	raw, err := json.Marshal(envelope.Envelope{
		EventType:     "bid_placed",
		EventID:       "evt-1",
		SchemaVersion: 999,
		Timestamp:     time.Now().UTC(),
		AuctionID:     1,
		Data:          json.RawMessage(`{}`),
	})
	require.NoError(t, err)

	// Act
	_, err = envelope.Decode(raw)

	// Assert
	require.ErrorContains(t, err, "unsupported event schema version 999")
}

func TestDecode_InvalidJSON_ReturnsError(t *testing.T) {
	_, err := envelope.Decode([]byte("{not json"))
	require.Error(t, err)
}
