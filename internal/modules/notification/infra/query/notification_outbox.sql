-- name: InsertNotificationOutboxEvent :exec
INSERT INTO notification_outbox (event_id, event_type, schema_version, subject, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertEmailRequestOutboxEvent :exec
INSERT INTO notification_outbox (event_id, event_type, schema_version, subject, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (event_id) DO NOTHING;

-- name: ListUnpublishedNotificationOutboxEvents :many
SELECT id, event_id, event_type, schema_version, subject, payload, occurred_at, created_at, published_at
FROM notification_outbox
WHERE published_at IS NULL
ORDER BY id ASC
LIMIT $1;

-- name: MarkNotificationOutboxEventPublished :execrows
UPDATE notification_outbox
SET published_at = NOW()
WHERE id = $1 AND published_at IS NULL;
