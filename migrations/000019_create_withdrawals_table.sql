-- +goose Up

CREATE TABLE withdrawals (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT       NOT NULL,
    ledger_account_id BIGINT       NOT NULL,
    alipay_account    VARCHAR      NOT NULL,
    alipay_real_name  VARCHAR      NOT NULL,
    amount_cents      BIGINT       NOT NULL CHECK (amount_cents >= 0),
    currency          VARCHAR      NOT NULL DEFAULT 'CNY',
    status            VARCHAR      NOT NULL DEFAULT 'created',
    out_biz_no        VARCHAR      NOT NULL,
    frozen_op_id      VARCHAR,
    alipay_order_id   VARCHAR,
    fail_reason       TEXT,
    version           BIGINT       NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_withdrawals_out_biz_no UNIQUE (out_biz_no)
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);

-- +goose Down

DROP TABLE IF EXISTS withdrawals;
