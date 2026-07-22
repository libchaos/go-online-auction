-- +goose Up

CREATE TABLE ledger_accounts (
    id             BIGSERIAL PRIMARY KEY,
    owner          VARCHAR NOT NULL,
    balance        BIGINT NOT NULL DEFAULT 0,
    frozen_balance BIGINT NOT NULL DEFAULT 0,
    currency       VARCHAR NOT NULL DEFAULT 'CNY',
    version        BIGINT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_ledger_accounts_balance_nonneg CHECK (balance >= 0),
    CONSTRAINT chk_ledger_accounts_frozen_nonneg CHECK (frozen_balance >= 0),
    CONSTRAINT uq_ledger_accounts_owner_currency UNIQUE (owner, currency)
);

CREATE INDEX idx_ledger_accounts_owner ON ledger_accounts(owner);

-- +goose Down

DROP TABLE IF EXISTS ledger_accounts;
