-- +goose Up

CREATE TABLE ledger_operations (
    id                       BIGSERIAL PRIMARY KEY,
    account_id               BIGINT NOT NULL REFERENCES ledger_accounts(id),
    counterparty_account_id  BIGINT REFERENCES ledger_accounts(id),
    operation_type           VARCHAR NOT NULL,
    amount                   BIGINT NOT NULL,
    idempotency_key          VARCHAR NOT NULL,
    status                   VARCHAR NOT NULL DEFAULT 'committed',
    reference                VARCHAR,
    description              VARCHAR,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_ledger_operations_type CHECK (operation_type IN ('freeze', 'unfreeze', 'withdraw_from_frozen')),
    CONSTRAINT chk_ledger_operations_status CHECK (status IN ('pending', 'committed', 'failed')),
    CONSTRAINT uq_ledger_operations_idempotency UNIQUE (idempotency_key)
);

CREATE INDEX idx_ledger_operations_account_id ON ledger_operations(account_id);
CREATE INDEX idx_ledger_operations_reference ON ledger_operations(reference);

-- +goose Down

DROP TABLE IF EXISTS ledger_operations;
