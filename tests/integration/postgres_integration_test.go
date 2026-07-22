package integration_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/outbox"
	"auction/internal/modules/auction/infra/repository"
	"auction/internal/modules/auction/ports"
	sharednats "auction/internal/shared/modules/nats"
	"auction/migrations"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver used by goose
)

// startPostgres boots a real Postgres instance via testcontainers, applies every
// SQL migration from the embedded migration FS, and returns a ready connection pool.
// It skips the test (rather than failing) when no Docker daemon is reachable, so the
// suite still runs in environments without Docker.
func startPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("auction_test"),
		postgres.WithUsername("auction"),
		postgres.WithPassword("auction"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		if strings.Contains(err.Error(), "Docker") || strings.Contains(err.Error(), "daemon") {
			t.Skipf("Docker daemon not available, skipping Postgres integration test: %v", err)
		}
		require.NoError(t, err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Apply migrations exactly like the production `db:migrate` command: the
	// embedded FS is the source of truth and goose runs the SQL migrations in order.
	// goose operates on a *sql.DB (pgx stdlib driver), separate from the pgxpool.
	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, "."))

	return pool
}

// TestAuctionRepository_CreateThenRead_RoundTrip exercises the full persistence path
// (repository -> sqlc -> Postgres) against a real database, validating that the
// migrations created the auctions table and the domain<->row mapping is correct.
func TestAuctionRepository_CreateThenRead_RoundTrip(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()

	auctionMapper := mapper.NewAuctionMapper(strategy.NewDefaultResolver())
	repo := repository.NewPostgresAuctionRepository(pool, auctionMapper)

	tradingMode, err := enum.NewTradingModeEnum(enum.EnumTradingModeEnglish)
	require.NoError(t, err)

	auction, err := model.NewAuctionModelWithMode(
		1,
		time.Now().Add(time.Hour),
		tradingMode,
		nil, nil, nil,
		false,
		0,
		nil,
		strategy.NewDefaultResolver(),
	)
	require.NoError(t, err)

	created, err := repo.Create(ctx, auction)
	require.NoError(t, err)
	require.NotZero(t, created.ID(), "database should assign an ID on insert")

	read, err := repo.FindByID(ctx, created.ID())
	require.NoError(t, err)
	require.Equal(t, created.ID(), read.ID())
	require.Equal(t, auction.ListingID(), read.ListingID())
	readMode := read.TradingMode()
	require.Equal(t, tradingMode.String(), readMode.String())
}

// TestAuctionOutbox_SaveListMarkPublished_Idempotent validates the transactional
// outbox table (event_outbox) on real Postgres: an event can be persisted, listed as
// unpublished, and marked published exactly once. The second MarkPublished must report
// no row affected, which is what makes the relay safe under redelivery.
func TestAuctionOutbox_SaveListMarkPublished_Idempotent(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()

	outboxRepo := repository.NewPostgresOutboxRepository(pool)

	require.NoError(t, outboxRepo.Save(ctx, ports.OutboxEvent{
		EventID:       "evt-integration-1",
		EventType:     "auction_created",
		SchemaVersion: 1,
		Subject:       "auction.evt.integration",
		Payload:       []byte(`{"auction_id":1}`),
		OccurredAt:    time.Now().UTC(),
	}))

	pending, err := outboxRepo.ListUnpublished(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, "evt-integration-1", pending[0].EventID)

	marked, err := outboxRepo.MarkPublished(ctx, pending[0].ID)
	require.NoError(t, err)
	require.True(t, marked, "first MarkPublished should report a row was updated")

	markedAgain, err := outboxRepo.MarkPublished(ctx, pending[0].ID)
	require.NoError(t, err)
	require.False(t, markedAgain, "second MarkPublished must be idempotent (no row updated)")

	remaining, err := outboxRepo.ListUnpublished(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, remaining, "published event must no longer be listed as pending")
}

// TestAuctionOutboxRelay_PublishesToNATS is the end-to-end proof of the transactional
// outbox pattern: a row written in Postgres is drained by the relay and published to
// the real JetStream event store, then marked published so it is never redelivered.
func TestAuctionOutboxRelay_PublishesToNATS(t *testing.T) {
	pool := startPostgres(t)
	js, _ := startJetStream(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outboxRepo := repository.NewPostgresOutboxRepository(pool)
	require.NoError(t, outboxRepo.Save(ctx, ports.OutboxEvent{
		EventID:       "evt-relay-1",
		EventType:     "auction_created",
		SchemaVersion: 1,
		Subject:       "auction.evt.integration",
		Payload:       []byte(`{"auction_id":1}`),
		OccurredAt:    time.Now().UTC(),
	}))

	relay := outbox.NewRelay(outboxRepo, js, newTestLogger(t), outbox.Config{
		Interval:  100 * time.Millisecond,
		BatchSize: 10,
	})
	relay.Start(ctx)

	// Wait until the relay has drained and marked the event published.
	require.Eventually(t, func() bool {
		pending, err := outboxRepo.ListUnpublished(ctx, 10)
		if err != nil {
			return false
		}
		return len(pending) == 0
	}, 5*time.Second, 50*time.Millisecond, "outbox event was not relayed and marked published")

	// The event must now live on the real AUCTION_EVENTS stream.
	stream, err := js.Stream(ctx, sharednats.StreamEvents)
	require.NoError(t, err)
	msg, err := stream.GetLastMsgForSubject(ctx, "auction.evt.integration")
	require.NoError(t, err)
	require.Equal(t, "auction.evt.integration", msg.Subject)
	require.Contains(t, string(msg.Data), "auction_id")

	relay.Stop()
}
