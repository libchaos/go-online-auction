package integration_test

import (
	"context"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	sharednats "auction/internal/shared/modules/nats"
)

const serverReadyTimeout = 10 * time.Second

// startJetStream boots an in-process NATS server with JetStream enabled and returns
// a connected JetStream context with the auction streams already declared.
func startJetStream(t *testing.T) (jetstream.JetStream, *natsgo.Conn) {
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

	conn, err := natsgo.Connect(server.ClientURL())
	require.NoError(t, err)
	t.Cleanup(conn.Close)

	js, err := jetstream.New(conn)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), serverReadyTimeout)
	defer cancel()
	require.NoError(t, sharednats.CreateOrUpdateStreams(ctx, js, 2*time.Minute))

	return js, conn
}
