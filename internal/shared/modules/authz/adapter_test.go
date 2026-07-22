package authz

import (
	"context"
	"os"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AdapterIntegrationSuite exercises the Postgres-backed PgxAdapter against a real
// database. It is skipped unless TEST_DATABASE_URL is provided, so the unit suite
// stays hermetic.
type AdapterIntegrationSuite struct {
	suite.Suite
	pool    *pgxpool.Pool
	adapter *PgxAdapter
}

func TestAdapterIntegrationSuite(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PgxAdapter integration test")
	}

	suite.Run(t, new(AdapterIntegrationSuite))
}

func (s *AdapterIntegrationSuite) SetupSuite() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(s.T(), err)
	s.pool = pool
	s.adapter = NewPgxAdapter(pool, "")

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS casbin_rules (
			id    BIGSERIAL PRIMARY KEY,
			ptype TEXT NOT NULL,
			v0    TEXT NOT NULL DEFAULT '',
			v1    TEXT NOT NULL DEFAULT '',
			v2    TEXT NOT NULL DEFAULT '',
			v3    TEXT NOT NULL DEFAULT '',
			v4    TEXT NOT NULL DEFAULT '',
			v5    TEXT NOT NULL DEFAULT ''
		);
		CREATE UNIQUE INDEX IF NOT EXISTS uniq_casbin_rule ON casbin_rules (ptype, v0, v1, v2, v3, v4, v5);
	`)
	require.NoError(s.T(), err)
}

func (s *AdapterIntegrationSuite) TearDownSuite() {
	if s.pool != nil {
		_, _ = s.pool.Exec(context.Background(), "DROP TABLE IF EXISTS casbin_rules")
		s.pool.Close()
	}
}

func (s *AdapterIntegrationSuite) SetupTest() {
	_, err := s.pool.Exec(context.Background(), "DELETE FROM casbin_rules")
	require.NoError(s.T(), err)
}

func (s *AdapterIntegrationSuite) TestLoadPolicy_RoundTripsStoredRules() {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx,
		"INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES ('p', 'admin', '/api/v1/**', '*')")
	require.NoError(s.T(), err)

	m, err := model.NewModelFromString(modelConfig)
	require.NoError(s.T(), err)

	require.NoError(s.T(), s.adapter.LoadPolicy(m))

	enforcer, err := casbin.NewEnforcer(m)
	require.NoError(s.T(), err)

	allowed, err := enforcer.Enforce("admin", "/api/v1/users", "GET")
	require.NoError(s.T(), err)
	require.True(s.T(), allowed)
}

func (s *AdapterIntegrationSuite) TestAddPolicy_InsertsRow() {
	ctx := context.Background()

	require.NoError(s.T(), s.adapter.AddPolicy("", "p", []string{"seller", "/api/v1/spus", "POST"}))

	var count int
	require.NoError(s.T(), s.pool.QueryRow(ctx,
		"SELECT count(*) FROM casbin_rules WHERE ptype='p' AND v0='seller' AND v1='/api/v1/spus' AND v2='POST'").Scan(&count))
	require.Equal(s.T(), 1, count)
}

func (s *AdapterIntegrationSuite) TestRemovePolicy_DeletesRow() {
	ctx := context.Background()

	_, err := s.pool.Exec(ctx,
		"INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES ('p', 'seller', '/api/v1/spus', 'POST')")
	require.NoError(s.T(), err)

	require.NoError(s.T(), s.adapter.RemovePolicy("", "p", []string{"seller", "/api/v1/spus", "POST"}))

	var count int
	require.NoError(s.T(), s.pool.QueryRow(ctx,
		"SELECT count(*) FROM casbin_rules WHERE ptype='p' AND v0='seller'").Scan(&count))
	require.Equal(s.T(), 0, count)
}
