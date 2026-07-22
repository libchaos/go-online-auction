-- name: InsertPaymentOutboxEvent :exec
INSERT INTO payments_outbox (event_id, event_type, schema_version, subject, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListUnpublishedPaymentOutboxEvents :many
SELECT id, event_id, event_type, schema_version, subject, payload, occurred_at, created_at, published_at
FROM payments_outbox
WHERE published_at IS NULL
ORDER BY id ASC
LIMIT $1;

-- name: MarkPaymentOutboxEventPublished :execrows
UPDATE payments_outbox
SET published_at = NOW()
WHERE id = $1 AND published_at IS NULL;
