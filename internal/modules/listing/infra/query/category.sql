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
