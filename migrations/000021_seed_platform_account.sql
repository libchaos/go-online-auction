-- +goose Up

-- System (platform) ledger account used as the funding source for user
-- recharges. ledger.Transfer validates the sender balance, so the platform
-- account is pre-funded with a large reserve that is effectively unbounded for
-- practical purposes (well within the BIGINT range). The payment module
-- resolves this account via GetOrCreateAccountByOwner(PLATFORM_ACCOUNT_OWNER).
INSERT INTO ledger_accounts (owner, balance, frozen_balance, currency, version)
VALUES ('platform', 1000000000000000, 0, 'CNY', 1)
ON CONFLICT (owner, currency) DO NOTHING;

-- +goose Down

DELETE FROM ledger_accounts WHERE owner = 'platform' AND currency = 'CNY';
