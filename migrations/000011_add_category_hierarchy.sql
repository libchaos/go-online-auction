-- +goose Up
-- Materialized path + depth turn the category adjacency list into a queryable
-- multi-level tree: `path` is "/<root>/<child>/<id>" and enables prefix-based
-- subtree scans; `depth` is the 0-based level (root = 0).
ALTER TABLE categories
    ADD COLUMN depth INT  NOT NULL DEFAULT 0,
    ADD COLUMN path  TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_categories_path ON categories (path);

-- Backfill existing rows with a recursive walk over the existing parent_id links.
WITH RECURSIVE tree AS (
    SELECT id, parent_id, 0 AS d, '/' || id::text AS p
    FROM categories
    WHERE parent_id IS NULL
    UNION ALL
    SELECT c.id, c.parent_id, t.d + 1, t.p || '/' || c.id::text
    FROM categories c
    JOIN tree t ON c.parent_id = t.id
)
UPDATE categories AS cat
SET depth = tree.d,
    path  = tree.p
FROM tree
WHERE cat.id = tree.id;

-- +goose Down
ALTER TABLE categories DROP COLUMN path;
ALTER TABLE categories DROP COLUMN depth;
