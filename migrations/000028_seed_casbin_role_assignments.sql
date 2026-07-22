-- +goose Up
-- Backfill the grouping (g) table so every existing user keeps the permissions
-- implied by their current users.role. After this migration RBAC enforcement
-- resolves user -> role -> policy through the g relation, and roles can be
-- reassigned at runtime via the /api/v1/rbac/role-assignments API without
-- touching the denormalized users.role column.
INSERT INTO casbin_rules (ptype, v0, v1)
SELECT 'g', id::text, role::text
FROM users
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM casbin_rules WHERE ptype = 'g';
