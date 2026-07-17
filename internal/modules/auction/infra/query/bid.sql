-- name: CreateBid :one
INSERT INTO bids (auction_id, user_id, amount_in_cents, max_amount_in_cents, idempotency_key, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetBidByID :one
SELECT * FROM bids
WHERE id = $1;

-- name: ListBidsByAuctionID :many
SELECT * FROM bids
WHERE auction_id = $1
ORDER BY created_at ASC;

-- name: UpdateBid :execrows
UPDATE bids
SET updated_at = $1
WHERE id = $2;

-- name: ListTopBidsByAuctionID :many
SELECT * FROM bids
WHERE auction_id = $1
ORDER BY amount_in_cents DESC
LIMIT $2;
