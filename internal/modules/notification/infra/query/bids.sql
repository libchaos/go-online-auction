-- name: GetPreviousHighestBidder :one
SELECT user_id
FROM bids
WHERE auction_id = $1
  AND user_id <> $2
ORDER BY amount_in_cents DESC, created_at DESC
LIMIT 1;

-- name: GetBidderIDByBidID :one
SELECT user_id
FROM bids
WHERE id = $1;
