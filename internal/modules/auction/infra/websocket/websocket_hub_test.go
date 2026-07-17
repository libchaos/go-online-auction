package websocket_test

import (
	"context"
	"testing"
	"time"

	"auction/internal/modules/auction/infra/websocket"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

// stubEventConsumer replays a fixed set of events into the handler, then blocks
// until the context is cancelled to mimic a live JetStream consumer.
type stubEventConsumer struct {
	events []stubEvent
}

type stubEvent struct {
	subject string
	data    []byte
}

func (c *stubEventConsumer) Consume(ctx context.Context, handler func(subject string, data []byte)) error {
	for _, e := range c.events {
		handler(e.subject, e.data)
	}
	<-ctx.Done()
	return nil
}

type HubTestSuite struct {
	suite.Suite
	registry *websocket.AuctionSubscriberRegistry
	logger   *mocks.MockLogger
}

func (s *HubTestSuite) SetupTest() {
	s.registry = websocket.NewAuctionSubscriberRegistry()
	s.logger = mocks.NewMockLogger(s.T())
}

func TestHubSuite(t *testing.T) {
	suite.Run(t, new(HubTestSuite))
}

func (s *HubTestSuite) TestRun_BroadcastsDottedSubjectEventToSubscriber() {
	// Arrange
	auctionID := uint64(42)
	payload := []byte(`{"event_type":"bid_placed"}`)
	consumer := &stubEventConsumer{
		events: []stubEvent{{subject: "auction.evt.42", data: payload}},
	}
	hub := websocket.NewHub(s.registry, consumer, s.logger)

	client := createTestClient(auctionID)
	defer client.Close()
	s.registry.Add(auctionID, client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	go hub.Run(ctx)

	// Assert
	received := readFromClientChannel(client)
	s.Equal(payload, received)
}

func (s *HubTestSuite) TestRun_InvalidSubjectIsIgnored() {
	// Arrange
	nopLogger := zerolog.Nop()
	s.logger.On("Warn").Return(nopLogger.Warn())
	auctionID := uint64(7)
	consumer := &stubEventConsumer{
		events: []stubEvent{
			{subject: "auction.evt.not-a-number", data: []byte(`{}`)},
			{subject: "auction.evt.7", data: []byte(`{"ok":true}`)},
		},
	}
	hub := websocket.NewHub(s.registry, consumer, s.logger)

	client := createTestClient(auctionID)
	defer client.Close()
	s.registry.Add(auctionID, client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	go hub.Run(ctx)

	// Assert only the valid-subject message is broadcast to the subscriber.
	received := readFromClientChannel(client)
	s.Equal([]byte(`{"ok":true}`), received)
}

func (s *HubTestSuite) TestRun_StopsOnContextCancel() {
	// Arrange
	consumer := &stubEventConsumer{}
	hub := websocket.NewHub(s.registry, consumer, s.logger)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	// Act
	cancel()

	// Assert
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		s.Fail("hub did not stop after context cancellation")
	}
}
