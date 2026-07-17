-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked_at, replaced_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshTokenByTokenHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1;

-- name: UpdateRefreshToken :execrows
UPDATE refresh_tokens
SET revoked_at = $1, replaced_by = $2
WHERE id = $3;

-- name: RevokeAllRefreshTokensForUser :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
