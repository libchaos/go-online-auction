-- +goose Up

-- notifications is the authoritative store for the in-app notification centre.
-- SSE only delivers a real-time copy; this table backs the list, unread filter,
-- unread count, mark-read and delete endpoints. Rows are deduplicated by
-- idempotency_key so redelivered source events never create duplicates.
CREATE TABLE notifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL,
    category        VARCHAR      NOT NULL,
    type            VARCHAR      NOT NULL,
    title           VARCHAR      NOT NULL,
    body            TEXT         NOT NULL,
    payload         JSONB        NOT NULL DEFAULT '{}',
    channels        VARCHAR[]    NOT NULL,
    idempotency_key VARCHAR      NOT NULL,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_notifications_idempotency UNIQUE (idempotency_key)
);

-- Partial index for the unread list / unread count (the hot path).
CREATE INDEX idx_notifications_user_unread ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;

-- Full list ordered newest-first.
CREATE INDEX idx_notifications_user_created ON notifications (user_id, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS notifications;
