-- name: CreateAuction :one
INSERT INTO auctions (listing_id, end_time, state, trading_mode, starting_price, price_step, reserve_price,
    current_price, highest_bid_amount_in_cents, winner_user_id, winning_bid_id, winning_bid_amount,
    anti_snipe_enabled, extension_window_sec, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetAuctionByID :one
SELECT * FROM auctions
WHERE id = $1;

-- name: GetAuctionByIDForUpdate :one
SELECT * FROM auctions
WHERE id = $1
FOR UPDATE NOWAIT;

-- name: UpdateAuction :execrows
UPDATE auctions
SET listing_id = @listing_id, start_time = @start_time, end_time = @end_time, state = @state,
    trading_mode = @trading_mode, starting_price = @starting_price, price_step = @price_step,
    reserve_price = @reserve_price, current_price = @current_price,
    highest_bid_amount_in_cents = @highest_bid_amount_in_cents, winner_user_id = @winner_user_id,
    winning_bid_id = @winning_bid_id, winning_bid_amount = @winning_bid_amount,
    anti_snipe_enabled = @anti_snipe_enabled, extension_window_sec = @extension_window_sec,
    version = @version, updated_at = @updated_at
WHERE id = @id AND version = @previous_version;

-- name: ListAuctions :many
SELECT * FROM auctions
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuctionsByState :many
SELECT * FROM auctions
WHERE state = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAuctions :one
SELECT COUNT(*) FROM auctions;

-- name: CountAuctionsByState :one
SELECT COUNT(*) FROM auctions
WHERE state = $1;

-- name: ListAuctionIDsDueToStart :many
SELECT id FROM auctions
WHERE state = 'draft' AND start_time IS NOT NULL AND start_time <= NOW()
ORDER BY start_time ASC
LIMIT $1;

-- name: ListAuctionIDsDueToClose :many
SELECT id FROM auctions
WHERE state = 'active' AND end_time <= NOW()
ORDER BY end_time ASC
LIMIT $1;
