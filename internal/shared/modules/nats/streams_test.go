package nats_test

import (
	"context"
	"testing"
	"time"

	sharednats "auction/internal/shared/modules/nats"
	natsserver "github.com/nats-io/nats-server/v2/server"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

const serverReadyTimeout = 10 * time.Second

func startEmbeddedJetStream(t *testing.T) *natsserver.Server {
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

	t.Cleanup(server.Shutdown)

	return server
}

func TestCreateOrUpdateStreams_Idempotent(t *testing.T) {
	// Arrange
	server := startEmbeddedJetStream(t)

	conn, err := natsgo.Connect(server.ClientURL())
	require.NoError(t, err)
	t.Cleanup(conn.Close)

	js, err := jetstream.New(conn)
	require.NoError(t, err)

	ctx := context.Background()
	dedupeWindow := 2 * time.Minute

	// Act
	firstErr := sharednats.CreateOrUpdateStreams(ctx, js, dedupeWindow)
	secondErr := sharednats.CreateOrUpdateStreams(ctx, js, dedupeWindow)

	// Assert
	require.NoError(t, firstErr)
	require.NoError(t, secondErr)

	streamNames := []string{
		sharednats.StreamCommands,
		sharednats.StreamEvents,
		sharednats.StreamDLQ,
	}
	for _, name := range streamNames {
		stream, streamErr := js.Stream(ctx, name)
		require.NoError(t, streamErr)

		info, infoErr := stream.Info(ctx)
		require.NoError(t, infoErr)
		require.Equal(t, name, info.Config.Name)
	}
}

func TestCreateOrUpdateStreams_DefaultsDedupeWindow(t *testing.T) {
	// Arrange
	server := startEmbeddedJetStream(t)

	conn, err := natsgo.Connect(server.ClientURL())
	require.NoError(t, err)
	t.Cleanup(conn.Close)

	js, err := jetstream.New(conn)
	require.NoError(t, err)

	ctx := context.Background()

	// Act
	err = sharednats.CreateOrUpdateStreams(ctx, js, 0)

	// Assert
	require.NoError(t, err)

	stream, err := js.Stream(ctx, sharednats.StreamCommands)
	require.NoError(t, err)

	info, err := stream.Info(ctx)
	require.NoError(t, err)
	require.Equal(t, 2*time.Minute, info.Config.Duplicates)
}
