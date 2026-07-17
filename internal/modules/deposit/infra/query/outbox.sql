-- name: InsertOutboxEvent :exec
INSERT INTO event_outbox (event_id, event_type, schema_version, subject, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListUnpublishedOutboxEvents :many
SELECT id, event_id, event_type, schema_version, subject, payload, occurred_at, created_at, published_at
FROM event_outbox
WHERE published_at IS NULL
ORDER BY id ASC
LIMIT $1;

-- name: MarkOutboxEventPublished :execrows
UPDATE event_outbox
SET published_at = NOW()
WHERE id = $1 AND published_at IS NULL;
