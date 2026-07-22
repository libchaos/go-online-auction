-- name: GetUserEmailByID :one
SELECT email FROM users WHERE id = $1;
