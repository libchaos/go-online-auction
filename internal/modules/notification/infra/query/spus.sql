-- name: FindSpuTitleByID :one
SELECT title FROM spus
WHERE id = $1;
