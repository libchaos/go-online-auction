-- +goose Up

CREATE TABLE ledger_entries (
    id           BIGSERIAL PRIMARY KEY,
    account_id   BIGINT NOT NULL REFERENCES ledger_accounts(id),
    amount       BIGINT NOT NULL,
    entry_type   VARCHAR NOT NULL,
    operation_id BIGINT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_operation_id ON ledger_entries(operation_id);

-- +goose Down

DROP TABLE IF EXISTS ledger_entries;
