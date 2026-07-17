-- name: CreateDeposit :one
INSERT INTO deposits (
    user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at;

-- name: GetDepositByID :one
SELECT id, user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at
FROM deposits
WHERE id = $1;

-- name: GetDepositByUserAndAuction :one
SELECT id, user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at
FROM deposits
WHERE user_id = $1 AND auction_id = $2;

-- name: ListDepositsByUser :many
SELECT id, user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at
FROM deposits
WHERE user_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: ListHeldDepositsByAuction :many
SELECT id, user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at
FROM deposits
WHERE auction_id = $1 AND status = 'held';

-- name: UpdateDeposit :one
UPDATE deposits
SET user_id = $2, auction_id = $3, amount_in_cents = $4, currency = $5, status = $6,
    external_reference = $7, reference = $8, version = $9, updated_at = $10
WHERE id = $1 AND version = $11
RETURNING id, user_id, auction_id, amount_in_cents, currency, status, external_reference, reference, version, created_at, updated_at;
