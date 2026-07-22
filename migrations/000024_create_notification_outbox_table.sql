-- +goose Up

-- Transactional outbox for notification domain events: a notification.evt.created
-- row is inserted in the same transaction as the notification itself, then
-- relayed to the NOTIFICATION_EVENTS stream by the notification module's own
-- relay. The SSE realtime hub consumes notification.evt.created and fans it out
-- to the owning user's connected clients. Delivery is at-least-once and
-- deduplicated by the stream's duplicate window keyed on event_id.
CREATE TABLE notification_outbox (
    id             BIGSERIAL PRIMARY KEY,
    event_id       VARCHAR NOT NULL,
    event_type     VARCHAR NOT NULL,
    schema_version INT     NOT NULL DEFAULT 1,
    subject        VARCHAR NOT NULL,
    payload        JSONB   NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at   TIMESTAMPTZ,

    CONSTRAINT uq_notification_outbox_event_id UNIQUE (event_id)
);

CREATE INDEX idx_notification_outbox_unpublished ON notification_outbox (id)
    WHERE published_at IS NULL;

-- +goose Down

DROP TABLE IF EXISTS notification_outbox;
