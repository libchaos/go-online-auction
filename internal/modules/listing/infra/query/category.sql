-- name: CreateCategory :one
INSERT INTO categories (name, parent_id, sort_order, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1;

-- name: UpdateCategory :execrows
UPDATE categories
SET name = @name, sort_order = @sort_order, version = @version, updated_at = @updated_at
WHERE id = @id AND version = @previous_version;

-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = $1;

-- name: ListRootCategories :many
SELECT * FROM categories
WHERE parent_id IS NULL
ORDER BY sort_order ASC, id ASC;

-- name: ListCategoriesByParent :many
SELECT * FROM categories
WHERE parent_id = $1::BIGINT
ORDER BY sort_order ASC, id ASC;

-- name: CountCategoryChildren :one
SELECT COUNT(*) FROM categories
WHERE parent_id = $1::BIGINT;

-- name: CountSpusByCategory :one
SELECT COUNT(*) FROM spus
WHERE category_id = $1;

-- name: FinalizeCategoryHierarchy :one
-- Computes depth and materialized path from the (possibly new) parent after the
-- row already exists, so the auto-generated id can be embedded in the path.
UPDATE categories AS c
SET depth = CASE WHEN c.parent_id IS NULL THEN 0 ELSE COALESCE(p.depth, 0) + 1 END,
    path  = CASE WHEN c.parent_id IS NULL THEN '/' || c.id::text
                 ELSE COALESCE(p.path, '') || '/' || c.id::text END
FROM (SELECT 1 AS one) AS dummy
LEFT JOIN categories AS p ON c.parent_id = p.id
WHERE c.id = $1
RETURNING c.*;

-- name: ListAllCategories :many
SELECT * FROM categories
ORDER BY depth ASC, sort_order ASC, id ASC;

-- name: ListCategoryDescendants :many
-- All nodes below the given category (excluding itself) via path prefix.
SELECT c.* FROM categories AS c
WHERE c.path LIKE (
    SELECT p.path FROM categories AS p WHERE p.id = $1
) || '/%'
ORDER BY c.depth ASC, c.sort_order ASC, c.id ASC;
