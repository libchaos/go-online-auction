-- +goose Up

CREATE TABLE ledger_transfers (
    id              BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL REFERENCES ledger_accounts(id),
    to_account_id   BIGINT NOT NULL REFERENCES ledger_accounts(id),
    amount          BIGINT NOT NULL,
    idempotency_key VARCHAR NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_ledger_transfers_idempotency UNIQUE (idempotency_key)
);

CREATE INDEX idx_ledger_transfers_from ON ledger_transfers(from_account_id);
CREATE INDEX idx_ledger_transfers_to ON ledger_transfers(to_account_id);

-- +goose Down

DROP TABLE IF EXISTS ledger_transfers;
