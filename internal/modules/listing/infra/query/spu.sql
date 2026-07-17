-- name: CreateSpu :one
INSERT INTO spus (title, description, category_id, brand, images, status, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetSpuByID :one
SELECT * FROM spus
WHERE id = $1;

-- name: GetSpuByIDForUpdate :one
SELECT * FROM spus
WHERE id = $1
FOR UPDATE NOWAIT;

-- name: UpdateSpu :execrows
UPDATE spus
SET title = @title, description = @description, category_id = @category_id, brand = @brand,
    images = @images, status = @status, version = @version, updated_at = @updated_at
WHERE id = @id AND version = @previous_version;

-- name: ListSpus :many
SELECT * FROM spus
WHERE (sqlc.narg('status')::listing_status IS NULL OR status = sqlc.narg('status')::listing_status)
  AND (sqlc.narg('category_id')::BIGINT IS NULL OR category_id = sqlc.narg('category_id')::BIGINT)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountSpus :one
SELECT COUNT(*) FROM spus
WHERE (sqlc.narg('status')::listing_status IS NULL OR status = sqlc.narg('status')::listing_status)
  AND (sqlc.narg('category_id')::BIGINT IS NULL OR category_id = sqlc.narg('category_id')::BIGINT);
