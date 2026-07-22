package authz

import (
	_ "embed"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/shared/modules/logger"
)

//go:embed model.conf
var modelConfig string

// NewEnforcer builds the Casbin enforcer backed by the Postgres policy store.
// The model is embedded; policies are loaded from the casbin_rules table via
// the pgx adapter. The enforcer is a singleton wired through the fx graph.
func NewEnforcer(pool *pgxpool.Pool, logger logger.Logger) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(modelConfig)
	if err != nil {
		return nil, fmt.Errorf("authz: parse model: %w", err)
	}

	adapter := NewPgxAdapter(pool, "")

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("authz: create enforcer: %w", err)
	}

	if err = enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("authz: load policy: %w", err)
	}

	if logger != nil {
		policies, policyErr := enforcer.GetPolicy()
		if policyErr != nil {
			return nil, fmt.Errorf("authz: read policy count: %w", policyErr)
		}
		logger.Info().Int("rules", len(policies)).Msg("authz enforcer initialized")
	}

	return enforcer, nil
}
