package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"auction/internal/modules/auction/infra/messaging"
	"auction/internal/modules/auction/ports"
	sharednats "auction/internal/shared/modules/nats"
	"auction/tests/mocks"
)

func newTestLogger(t *testing.T) *mocks.MockLogger {
	t.Helper()
	log := mocks.NewMockLogger(t)
	nop := zerolog.Nop()
	log.On("Debug").Return(nop.Debug()).Maybe()
	log.On("Info").Return(nop.Info()).Maybe()
	log.On("Warn").Return(nop.Warn()).Maybe()
	log.On("Error").Return(nop.Error()).Maybe()
	return log
}

// Verifies the first idempotency layer: publishing the same Nats-Msg-Id twice within
// the dedupe window results in exactly one message stored on the command stream.
func TestBidCommandPublisher_DuplicateIdempotencyKey_DeduplicatesOnStream(t *testing.T) {
	// Arrange
	js, _ := startJetStream(t)
	ctx := context.Background()
	publisher := messaging.NewJetStreamBidCommandPublisher(js, newTestLogger(t))

	cmd := ports.BidCommand{
		IdempotencyKey: "dup-key-1",
		AuctionID:      7,
		UserID:         100,
		AmountInCents:  5000,
		IssuedAt:       time.Now().UTC(),
	}

	// Act
	_, err := publisher.Publish(ctx, cmd)
	require.NoError(t, err)
	_, err = publisher.Publish(ctx, cmd)
	require.NoError(t, err)

	// Assert
	stream, err := js.Stream(ctx, sharednats.StreamCommands)
	require.NoError(t, err)
	info, err := stream.Info(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), info.State.Msgs)
}

// Verifies the publisher targets auction.cmd.bid.{id} and the command is retrievable.
func TestBidCommandPublisher_PublishesToCommandSubject(t *testing.T) {
	// Arrange
	js, _ := startJetStream(t)
	ctx := context.Background()
	publisher := messaging.NewJetStreamBidCommandPublisher(js, newTestLogger(t))

	cmd := ports.BidCommand{
		IdempotencyKey: "key-subject",
		AuctionID:      42,
		UserID:         100,
		AmountInCents:  5000,
		IssuedAt:       time.Now().UTC(),
	}

	// Act
	ack, err := publisher.Publish(ctx, cmd)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "key-subject", ack.IdempotencyKey)

	stream, err := js.Stream(ctx, sharednats.StreamCommands)
	require.NoError(t, err)
	rawMsg, err := stream.GetMsg(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "auction.cmd.bid.42", rawMsg.Subject)
}

// Verifies the end-to-end notification path: an event published by the dispatcher to
// AUCTION_EVENTS is delivered through the JetStream consumer to the handler.
func TestEventDispatcherToConsumer_DeliversEvent(t *testing.T) {
	// Arrange
	js, _ := startJetStream(t)
	publisher := messaging.NewJetStreamEventPublisher(js)
	consumer := messaging.NewJetStreamEventConsumer(js)

	received := make(chan string, 1)
	consumeCtx, consumeCancel := context.WithCancel(context.Background())
	defer consumeCancel()

	go func() {
		_ = consumer.Consume(consumeCtx, func(subject string, _ []byte) {
			select {
			case received <- subject:
			default:
			}
		})
	}()

	// Give the consumer a moment to establish its subscription before publishing,
	// since it uses DeliverNewPolicy.
	time.Sleep(500 * time.Millisecond)

	// Act
	err := publisher.Publish(context.Background(), "auction.evt.99", []byte(`{"event_type":"bid_placed"}`))
	require.NoError(t, err)

	// Assert
	select {
	case subject := <-received:
		require.Equal(t, "auction.evt.99", subject)
	case <-time.After(5 * time.Second):
		t.Fatal("event was not delivered to the consumer")
	}
}

// Verifies DLQ retention: a message published to the DLQ subject is persisted on the
// AUCTION_DLQ stream and not lost.
func TestDLQ_RetainsMessage(t *testing.T) {
	// Arrange
	js, _ := startJetStream(t)
	ctx := context.Background()

	// Act
	_, err := js.Publish(ctx, messaging.BuildDLQSubject(5), []byte(`{"idempotency_key":"failed"}`))
	require.NoError(t, err)

	// Assert
	stream, err := js.Stream(ctx, sharednats.StreamDLQ)
	require.NoError(t, err)
	info, err := stream.Info(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), info.State.Msgs)
}

// Verifies the command stream is a WorkQueue: consuming and acking a command removes it.
func TestCommandStream_WorkQueue_RemovesOnAck(t *testing.T) {
	// Arrange
	js, _ := startJetStream(t)
	ctx := context.Background()
	publisher := messaging.NewJetStreamBidCommandPublisher(js, newTestLogger(t))

	cmd := ports.BidCommand{
		IdempotencyKey: "wq-key",
		AuctionID:      1,
		UserID:         2,
		AmountInCents:  3000,
		IssuedAt:       time.Now().UTC(),
	}
	_, err := publisher.Publish(ctx, cmd)
	require.NoError(t, err)

	consumer, err := js.CreateOrUpdateConsumer(ctx, sharednats.StreamCommands, jetstream.ConsumerConfig{
		Durable:       "test-consumer",
		FilterSubject: sharednats.SubjectCommands,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
	})
	require.NoError(t, err)

	// Act
	msgs, err := consumer.Fetch(1)
	require.NoError(t, err)
	var count int
	for msg := range msgs.Messages() {
		require.NoError(t, msg.Ack())
		count++
	}
	require.NoError(t, msgs.Error())

	// Assert
	require.Equal(t, 1, count)
	stream, err := js.Stream(ctx, sharednats.StreamCommands)
	require.NoError(t, err)
	info, err := stream.Info(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), info.State.Msgs)
}
