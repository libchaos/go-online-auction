-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, role, status, oauth_provider, oauth_provider_id,
    version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :execrows
UPDATE users
SET name = @name, email = @email, password_hash = @password_hash, role = @role, status = @status,
    oauth_provider = @oauth_provider, oauth_provider_id = @oauth_provider_id, version = @version,
    updated_at = @updated_at
WHERE id = @id AND version = @previous_version;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
