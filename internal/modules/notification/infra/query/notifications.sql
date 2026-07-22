-- name: InsertNotification :one
INSERT INTO notifications (
    user_id, category, type, title, body, payload, channels, idempotency_key, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING id, user_id, category, type, title, body, payload, channels, idempotency_key, read_at, created_at;

-- name: GetNotificationByID :one
SELECT id, user_id, category, type, title, body, payload, channels, idempotency_key, read_at, created_at
FROM notifications
WHERE id = $1;

-- name: ListNotificationsByUser :many
SELECT id, user_id, category, type, title, body, payload, channels, idempotency_key, read_at, created_at
FROM notifications
WHERE user_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: ListUnreadNotificationsByUser :many
SELECT id, user_id, category, type, title, body, payload, channels, idempotency_key, read_at, created_at
FROM notifications
WHERE user_id = $1 AND read_at IS NULL
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1 AND read_at IS NULL;

-- name: MarkNotificationRead :execrows
UPDATE notifications
SET read_at = NOW()
WHERE id = $1 AND user_id = $2 AND read_at IS NULL;

-- name: MarkAllNotificationsRead :execrows
UPDATE notifications
SET read_at = NOW()
WHERE user_id = $1 AND read_at IS NULL;

-- name: DeleteNotification :execrows
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;
