-- name: InsertListingOutboxEvent :exec
INSERT INTO event_outbox (event_id, event_type, schema_version, subject, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6);
