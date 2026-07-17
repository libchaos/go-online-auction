-- name: GetAuctionDepositConfig :one
SELECT deposit_required, deposit_amount_in_cents
FROM auctions
WHERE id = $1;

-- name: GetAuctionWinner :one
SELECT winner_user_id
FROM auctions
WHERE id = $1;
