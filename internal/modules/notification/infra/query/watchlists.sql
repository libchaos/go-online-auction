-- name: UpsertWatchlist :one
INSERT INTO watchlists (user_id, spu_id)
VALUES ($1, $2)
ON CONFLICT (user_id, spu_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING *;

-- name: DeleteWatchlist :one
DELETE FROM watchlists
WHERE user_id = $1 AND spu_id = $2
RETURNING *;

-- name: ListWatchlistsByUser :many
SELECT * FROM watchlists
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: FindWatcherIDsBySpuID :many
SELECT user_id FROM watchlists
WHERE spu_id = $1;
