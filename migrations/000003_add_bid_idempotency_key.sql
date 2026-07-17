-- +goose Up

-- Backfill existing rows with a unique value so the unique index below can be
-- created without collisions; new bids always carry a client/server idempotency key.
ALTER TABLE bids ADD COLUMN idempotency_key VARCHAR NOT NULL DEFAULT gen_random_uuid()::text;

CREATE UNIQUE INDEX ux_bid_auction_idempotency ON bids(auction_id, idempotency_key);

-- +goose Down
DROP INDEX IF EXISTS ux_bid_auction_idempotency;

ALTER TABLE bids DROP COLUMN IF EXISTS idempotency_key;
