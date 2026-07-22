-- name: GetNotificationPreferences :one
SELECT user_id, preferences, updated_at
FROM notification_preferences
WHERE user_id = $1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification_preferences (user_id, preferences, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE
SET preferences = EXCLUDED.preferences, updated_at = NOW()
RETURNING user_id, preferences, updated_at;
