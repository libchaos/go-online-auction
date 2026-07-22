-- +goose Up

CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL,
    amount_cents    BIGINT       NOT NULL CHECK (amount_cents >= 0),
    currency        VARCHAR      NOT NULL DEFAULT 'CNY',
    status          VARCHAR      NOT NULL DEFAULT 'created',
    out_trade_no    VARCHAR      NOT NULL,
    qr_code_url     TEXT,
    alipay_trade_no VARCHAR,
    version         BIGINT       NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_payments_out_trade_no UNIQUE (out_trade_no)
);

CREATE INDEX idx_payments_user_id ON payments(user_id);

-- +goose Down

DROP TABLE IF EXISTS payments;
