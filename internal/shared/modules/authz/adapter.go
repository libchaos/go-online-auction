package authz

import (
	"context"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultRuleTable = "casbin_rules"

// PgxAdapter persists Casbin policy rules in a Postgres table using the pool
// already wired through the application. It implements persist.Adapter so the
// enforcer can load and mutate policies at runtime without an extra DB driver.
type PgxAdapter struct {
	pool  *pgxpool.Pool
	table string
}

func NewPgxAdapter(pool *pgxpool.Pool, table string) *PgxAdapter {
	if table == "" {
		table = defaultRuleTable
	}
	return &PgxAdapter{pool: pool, table: table}
}

func (a *PgxAdapter) LoadPolicy(m model.Model) error {
	ctx := context.Background()

	query := fmt.Sprintf("SELECT ptype, v0, v1, v2, v3, v4, v5 FROM %s", a.table)
	rows, err := a.pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("authz: load policy: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 string
		if err = rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return fmt.Errorf("authz: scan policy row: %w", err)
		}

		line := buildPolicyLine(ptype, v0, v1, v2, v3, v4, v5)
		if err = persist.LoadPolicyLine(line, m); err != nil {
			return fmt.Errorf("authz: parse policy line: %w", err)
		}
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("authz: iterate policy rows: %w", err)
	}

	return nil
}

func (a *PgxAdapter) SavePolicy(m model.Model) error {
	ctx := context.Background()

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("authz: begin save policy: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err = tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s", a.table)); err != nil {
		return fmt.Errorf("authz: clear policy: %w", err)
	}

	for _, assertionMap := range m {
		for _, assertion := range assertionMap {
			for _, rule := range assertion.Policy {
				if err = a.insertRule(ctx, tx, assertion.Key, rule); err != nil {
					return err
				}
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("authz: commit save policy: %w", err)
	}

	return nil
}

func (a *PgxAdapter) AddPolicy(_ string, ptype string, rule []string) error {
	ctx := context.Background()

	values := padRule(rule)
	query := fmt.Sprintf(
		"INSERT INTO %s (ptype, v0, v1, v2, v3, v4, v5) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING",
		a.table,
	)
	if _, err := a.pool.Exec(ctx, query, ptype, values[0], values[1], values[2], values[3], values[4], values[5]); err != nil {
		return fmt.Errorf("authz: add policy rule: %w", err)
	}

	return nil
}

func (a *PgxAdapter) RemovePolicy(_ string, ptype string, rule []string) error {
	ctx := context.Background()

	values := padRule(rule)
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6 AND v5=$7",
		a.table,
	)
	if _, err := a.pool.Exec(ctx, query, ptype, values[0], values[1], values[2], values[3], values[4], values[5]); err != nil {
		return fmt.Errorf("authz: remove policy rule: %w", err)
	}

	return nil
}

func (a *PgxAdapter) RemoveFilteredPolicy(_ string, ptype string, fieldIndex int, fieldValues ...string) error {
	if len(fieldValues) == 0 {
		return nil
	}

	ctx := context.Background()

	conditions := []string{"ptype = $1"}
	args := []interface{}{ptype}
	cursor := 2

	for i, value := range fieldValues {
		column := fmt.Sprintf("v%d", fieldIndex+i)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, cursor))
		args = append(args, value)
		cursor++
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", a.table, strings.Join(conditions, " AND "))
	if _, err := a.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("authz: remove filtered policy: %w", err)
	}

	return nil
}

func (a *PgxAdapter) insertRule(ctx context.Context, tx pgx.Tx, ptype string, rule []string) error {
	values := padRule(rule)
	query := fmt.Sprintf(
		"INSERT INTO %s (ptype, v0, v1, v2, v3, v4, v5) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING",
		a.table,
	)
	if _, err := tx.Exec(ctx, query, ptype, values[0], values[1], values[2], values[3], values[4], values[5]); err != nil {
		return fmt.Errorf("authz: insert policy rule: %w", err)
	}

	return nil
}

func buildPolicyLine(ptype, v0, v1, v2, v3, v4, v5 string) string {
	parts := []string{ptype}
	for _, value := range []string{v0, v1, v2, v3, v4, v5} {
		if value == "" {
			break
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, ", ")
}

func padRule(rule []string) []string {
	values := []string{"", "", "", "", "", ""}
	for i := 0; i < 6 && i < len(rule); i++ {
		values[i] = rule[i]
	}
	return values
}
