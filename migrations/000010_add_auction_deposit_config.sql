-- +goose Up

ALTER TABLE auctions ADD COLUMN deposit_required boolean NOT NULL DEFAULT false;
ALTER TABLE auctions ADD COLUMN deposit_amount_in_cents bigint NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE auctions DROP COLUMN IF EXISTS deposit_required;
ALTER TABLE auctions DROP COLUMN IF EXISTS deposit_amount_in_cents;
