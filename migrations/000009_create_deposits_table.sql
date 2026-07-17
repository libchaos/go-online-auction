-- +goose Up

CREATE TABLE deposits (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL,
    auction_id       BIGINT NOT NULL,
    amount_in_cents  BIGINT NOT NULL,
    currency         VARCHAR NOT NULL DEFAULT 'CNY',
    status           VARCHAR NOT NULL DEFAULT 'pending',
    external_reference VARCHAR,
    reference        VARCHAR,
    version          BIGINT NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_deposits_status CHECK (status IN ('pending', 'held', 'released', 'applied', 'forfeited')),
    CONSTRAINT uq_deposits_user_auction UNIQUE (user_id, auction_id)
);

CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_deposits_auction_id ON deposits(auction_id);
CREATE INDEX idx_deposits_status ON deposits(status);

-- +goose Down

DROP TABLE IF EXISTS deposits;
