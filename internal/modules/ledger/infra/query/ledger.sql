-- name: CreateLedgerAccount :one
INSERT INTO ledger_accounts (
    owner, balance, frozen_balance, currency, version, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, owner, balance, frozen_balance, currency, version, created_at, updated_at;

-- name: GetLedgerAccountByID :one
SELECT id, owner, balance, frozen_balance, currency, version, created_at, updated_at
FROM ledger_accounts
WHERE id = $1;

-- name: GetLedgerAccountByOwner :one
SELECT id, owner, balance, frozen_balance, currency, version, created_at, updated_at
FROM ledger_accounts
WHERE owner = $1 AND currency = $2;

-- name: GetLedgerAccountByIDForUpdate :one
SELECT id, owner, balance, frozen_balance, currency, version, created_at, updated_at
FROM ledger_accounts
WHERE id = $1 FOR UPDATE;

-- name: GetLedgerAccountByOwnerForUpdate :one
SELECT id, owner, balance, frozen_balance, currency, version, created_at, updated_at
FROM ledger_accounts
WHERE owner = $1 AND currency = $2 FOR UPDATE;

-- name: UpdateLedgerAccountBalance :one
UPDATE ledger_accounts
SET balance = $2, frozen_balance = $3, version = $4, updated_at = $5
WHERE id = $1 AND version = $6
RETURNING id, owner, balance, frozen_balance, currency, version, created_at, updated_at;

-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (
    account_id, amount, entry_type, operation_id, created_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, account_id, amount, entry_type, operation_id, created_at;

-- name: CreateLedgerTransfer :one
INSERT INTO ledger_transfers (
    from_account_id, to_account_id, amount, idempotency_key, created_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, from_account_id, to_account_id, amount, idempotency_key, created_at;

-- name: GetLedgerTransferByIdempotency :one
SELECT id, from_account_id, to_account_id, amount, idempotency_key, created_at
FROM ledger_transfers
WHERE idempotency_key = $1;

-- name: CreateLedgerOperation :one
INSERT INTO ledger_operations (
    account_id, counterparty_account_id, operation_type, amount, idempotency_key,
    status, reference, description, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, account_id, counterparty_account_id, operation_type, amount,
    idempotency_key, status, reference, description, created_at, updated_at;

-- name: GetLedgerOperationByIdempotency :one
SELECT id, account_id, counterparty_account_id, operation_type, amount,
    idempotency_key, status, reference, description, created_at, updated_at
FROM ledger_operations
WHERE idempotency_key = $1;

-- name: ListLedgerAccounts :many
SELECT id, owner, balance, frozen_balance, currency, version, created_at, updated_at
FROM ledger_accounts
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: ListLedgerEntriesByAccount :many
SELECT id, account_id, amount, entry_type, operation_id, created_at
FROM ledger_entries
WHERE account_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: ListLedgerOperationsByAccount :many
SELECT id, account_id, counterparty_account_id, operation_type, amount,
    idempotency_key, status, reference, description, created_at, updated_at
FROM ledger_operations
WHERE account_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;
