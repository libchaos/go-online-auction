-- +goose Up

-- Transactional outbox for deposit domain events: events are inserted in the
-- same database transaction as the state change that produced them, then
-- relayed to NATS JetStream by the deposit module's own outbox relay.
-- published_at IS NULL marks pending rows.
CREATE TABLE deposit_outbox (
    id             BIGSERIAL PRIMARY KEY,
    event_id       VARCHAR NOT NULL,
    event_type     VARCHAR NOT NULL,
    schema_version INT     NOT NULL DEFAULT 1,
    subject        VARCHAR NOT NULL,
    payload        JSONB   NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at   TIMESTAMPTZ,

    -- The relay uses event_id as the JetStream Nats-Msg-Id, so it must be unique
    CONSTRAINT uq_deposit_outbox_event_id UNIQUE (event_id)
);

-- Partial index: the relay only ever scans unpublished rows in insertion order
CREATE INDEX idx_deposit_outbox_unpublished ON deposit_outbox (id)
    WHERE published_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS deposit_outbox;
