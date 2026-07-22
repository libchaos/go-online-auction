package sse_test

import (
	"context"
	"testing"
	"time"

	"auction/internal/modules/notification/infra/sse"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

// stubEventConsumer replays a fixed list of subject/payload pairs into the hub
// handler and then returns, so RealtimeHub.Run completes deterministically.
type stubEventConsumer struct {
	events []stubEvent
}

type stubEvent struct {
	subject string
	data    []byte
}

func (consumer *stubEventConsumer) Consume(
	_ context.Context,
	handler func(subject string, data []byte),
) error {
	for _, event := range consumer.events {
		handler(event.subject, event.data)
	}

	return nil
}

type SSEHubTestSuite struct {
	suite.Suite
	registry   *sse.SubscriberRegistry
	loggerMock *mocks.MockLogger
}

func (s *SSEHubTestSuite) SetupTest() {
	s.registry = sse.NewSubscriberRegistry()
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Warn").Return(nopLogger.Warn()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()
}

func TestSSEHubSuite(t *testing.T) {
	suite.Run(t, new(SSEHubTestSuite))
}

func (s *SSEHubTestSuite) TestRegistry_PublishToUser_FansOutToAllClients() {
	// Arrange
	clientA := sse.NewClient(42)
	clientB := sse.NewClient(42)
	other := sse.NewClient(99)
	s.registry.Add(42, clientA)
	s.registry.Add(42, clientB)
	s.registry.Add(99, other)

	// Act
	s.registry.PublishToUser(42, []byte("hello"))

	// Assert
	s.Equal([]byte("hello"), s.receive(clientA))
	s.Equal([]byte("hello"), s.receive(clientB))
	s.Empty(other.Send())
}

func (s *SSEHubTestSuite) TestRegistry_Remove_StopsDelivery() {
	// Arrange
	client := sse.NewClient(42)
	s.registry.Add(42, client)
	s.registry.Remove(42, client)

	// Act
	s.registry.PublishToUser(42, []byte("hello"))

	// Assert
	s.Empty(client.Send())
}

func (s *SSEHubTestSuite) TestRealtimeHub_Run_RoutesEventToRecipient() {
	// Arrange
	recipient := sse.NewClient(42)
	bystander := sse.NewClient(7)
	s.registry.Add(42, recipient)
	s.registry.Add(7, bystander)
	consumer := &stubEventConsumer{events: []stubEvent{
		{subject: "notification.evt.42", data: []byte("payload-42")},
	}}
	hub := sse.NewRealtimeHub(s.registry, consumer, s.loggerMock)

	// Act
	hub.Run(context.Background())

	// Assert
	s.Equal([]byte("payload-42"), s.receive(recipient))
	s.Empty(bystander.Send())
}

func (s *SSEHubTestSuite) TestRealtimeHub_Run_IgnoresUnparseableSubject() {
	// Arrange
	client := sse.NewClient(42)
	s.registry.Add(42, client)
	consumer := &stubEventConsumer{events: []stubEvent{
		{subject: "notification.evt.not-a-number", data: []byte("dropped")},
	}}
	hub := sse.NewRealtimeHub(s.registry, consumer, s.loggerMock)

	// Act
	hub.Run(context.Background())

	// Assert
	s.Empty(client.Send())
}

func (s *SSEHubTestSuite) receive(client *sse.Client) []byte {
	select {
	case message := <-client.Send():
		return message
	case <-time.After(time.Second):
		s.FailNow("timed out waiting for SSE message")

		return nil
	}
}
