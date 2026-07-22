-- +goose Up

-- watchlists records a user's explicit interest in a catalogue product (SPU).
-- When that SPU (or any SKU under it) emits a listing event, every watcher
-- receives an in-app notification. Uniqueness on (user_id, spu_id) makes the
-- "watch" action idempotent.
CREATE TABLE watchlists (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    spu_id     BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_watchlists_user_spu UNIQUE (user_id, spu_id)
);

CREATE INDEX idx_watchlists_user_id ON watchlists (user_id);
CREATE INDEX idx_watchlists_spu_id ON watchlists (spu_id);

-- +goose Down

DROP TABLE IF EXISTS watchlists;
