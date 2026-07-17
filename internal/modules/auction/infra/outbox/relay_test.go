package outbox_test

import (
	"context"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"auction/internal/modules/auction/infra/outbox"
	"auction/internal/modules/auction/ports"
	sharednats "auction/internal/shared/modules/nats"
	"auction/tests/mocks"
)

const serverReadyTimeout = 10 * time.Second

func startEmbeddedJetStream(t *testing.T) (jetstream.JetStream, func()) {
	t.Helper()

	opts := &natsserver.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
		NoLog:     true,
		NoSigs:    true,
	}

	server, err := natsserver.NewServer(opts)
	require.NoError(t, err)

	go server.Start()
	if !server.ReadyForConnections(serverReadyTimeout) {
		server.Shutdown()
		t.Fatal("embedded nats server did not become ready")
	}

	conn, err := natsgo.Connect(server.ClientURL())
	require.NoError(t, err)

	js, err := jetstream.New(conn)
	require.NoError(t, err)

	require.NoError(t, sharednats.CreateOrUpdateStreams(context.Background(), js, 2*time.Minute))

	return js, func() {
		conn.Close()
		server.Shutdown()
	}
}

func pendingEvent(id uint64, eventID string) ports.OutboxEvent {
	return ports.OutboxEvent{
		ID:            id,
		EventID:       eventID,
		EventType:     "bid_placed",
		SchemaVersion: 1,
		Subject:       "auction.evt.1",
		Payload:       []byte(`{"event_type":"bid_placed","event_id":"` + eventID + `","schema_version":1,"auction_id":1,"data":{}}`),
		OccurredAt:    time.Now().UTC(),
	}
}

func TestDrain_PublishesPendingEventsAndMarksThem(t *testing.T) {
	// Arrange
	js, shutdown := startEmbeddedJetStream(t)
	defer shutdown()

	outboxRepo := mocks.NewMockOutboxRepository(t)
	logger := mocks.NewMockLogger(t)

	events := []ports.OutboxEvent{pendingEvent(1, "evt-1"), pendingEvent(2, "evt-2")}
	outboxRepo.On("ListUnpublished", mock.Anything, 100).Return(events, nil).Once()
	outboxRepo.On("MarkPublished", mock.Anything, uint64(1)).Return(true, nil).Once()
	outboxRepo.On("MarkPublished", mock.Anything, uint64(2)).Return(true, nil).Once()

	relay := outbox.NewRelay(outboxRepo, js, logger, outbox.Config{Interval: time.Minute, BatchSize: 100})

	// Act
	relay.Drain(context.Background())

	// Assert: both events are in the stream
	stream, err := js.Stream(context.Background(), sharednats.StreamEvents)
	require.NoError(t, err)
	info, err := stream.Info(context.Background())
	require.NoError(t, err)
	require.Equal(t, uint64(2), info.State.Msgs)
}

func TestDrain_RedeliveredEventIsDeduplicatedByMsgID(t *testing.T) {
	// Arrange: the same event is listed twice, simulating a crash after publish
	// but before MarkPublished succeeded.
	js, shutdown := startEmbeddedJetStream(t)
	defer shutdown()

	outboxRepo := mocks.NewMockOutboxRepository(t)
	logger := mocks.NewMockLogger(t)

	evt := pendingEvent(1, "evt-dup")
	outboxRepo.On("ListUnpublished", mock.Anything, 100).Return([]ports.OutboxEvent{evt}, nil).Once()
	outboxRepo.On("MarkPublished", mock.Anything, uint64(1)).Return(true, nil).Twice()

	relay := outbox.NewRelay(outboxRepo, js, logger, outbox.Config{Interval: time.Minute, BatchSize: 100})
	relay.Drain(context.Background())

	outboxRepo.On("ListUnpublished", mock.Anything, 100).Return([]ports.OutboxEvent{evt}, nil).Once()

	// Act: second drain republishes the same event ID
	relay.Drain(context.Background())

	// Assert: the duplicate window absorbed the redelivery
	stream, err := js.Stream(context.Background(), sharednats.StreamEvents)
	require.NoError(t, err)
	info, err := stream.Info(context.Background())
	require.NoError(t, err)
	require.Equal(t, uint64(1), info.State.Msgs)
}
