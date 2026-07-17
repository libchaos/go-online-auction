-- name: CreateSku :one
INSERT INTO skus (spu_id, spec_values, price_in_cents, quantity, status, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetSkuByID :one
SELECT * FROM skus
WHERE id = $1;

-- name: GetSkuByIDForUpdate :one
SELECT * FROM skus
WHERE id = $1
FOR UPDATE NOWAIT;

-- name: UpdateSku :execrows
UPDATE skus
SET spec_values = @spec_values, price_in_cents = @price_in_cents, quantity = @quantity,
    status = @status, version = @version, updated_at = @updated_at
WHERE id = @id AND version = @previous_version;

-- name: ListSkusBySpuID :many
SELECT * FROM skus
WHERE spu_id = $1
ORDER BY id ASC;

-- name: ListPublishedSkusBySpuIDForUpdate :many
SELECT * FROM skus
WHERE spu_id = $1 AND status = 'published'
ORDER BY id ASC
FOR UPDATE NOWAIT;
