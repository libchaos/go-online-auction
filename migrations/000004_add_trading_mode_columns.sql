-- +goose Up

-- Trading mode selects which strategy resolves bid validation, pricing and winner
-- determination at runtime (english, dutch, sealed_bid, vickrey, fixed_price, ebay_proxy).
ALTER TABLE auctions ADD COLUMN trading_mode varchar NOT NULL DEFAULT 'english';

-- Price configuration for the various trading modes.
ALTER TABLE auctions ADD COLUMN starting_price bigint;
ALTER TABLE auctions ADD COLUMN price_step bigint;
ALTER TABLE auctions ADD COLUMN reserve_price bigint;
ALTER TABLE auctions ADD COLUMN current_price bigint;

-- Winner snapshot captured when an auction closes.
ALTER TABLE auctions ADD COLUMN winner_user_id bigint;
ALTER TABLE auctions ADD COLUMN winning_bid_id bigint;
ALTER TABLE auctions ADD COLUMN winning_bid_amount bigint;

-- Anti-snipe: when enabled, a bid placed within the extension window pushes the
-- close time out by extension_window_sec seconds.
ALTER TABLE auctions ADD COLUMN anti_snipe_enabled boolean NOT NULL DEFAULT false;
ALTER TABLE auctions ADD COLUMN extension_window_sec bigint NOT NULL DEFAULT 300;

-- Ensure only supported trading modes are persisted.
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_trading_mode
    CHECK (trading_mode IN ('english', 'dutch', 'sealed_bid', 'vickrey', 'fixed_price', 'ebay_proxy'));

-- Proxy bidding (eBay-style): the maximum amount a bidder is willing to pay.
ALTER TABLE bids ADD COLUMN max_amount_in_cents bigint;

-- +goose Down

ALTER TABLE auctions DROP CONSTRAINT IF EXISTS chk_auctions_trading_mode;

ALTER TABLE auctions DROP COLUMN IF EXISTS trading_mode;
ALTER TABLE auctions DROP COLUMN IF EXISTS starting_price;
ALTER TABLE auctions DROP COLUMN IF EXISTS price_step;
ALTER TABLE auctions DROP COLUMN IF EXISTS reserve_price;
ALTER TABLE auctions DROP COLUMN IF EXISTS current_price;
ALTER TABLE auctions DROP COLUMN IF EXISTS winner_user_id;
ALTER TABLE auctions DROP COLUMN IF EXISTS winning_bid_id;
ALTER TABLE auctions DROP COLUMN IF EXISTS winning_bid_amount;
ALTER TABLE auctions DROP COLUMN IF EXISTS anti_snipe_enabled;
ALTER TABLE auctions DROP COLUMN IF EXISTS extension_window_sec;

ALTER TABLE bids DROP COLUMN IF EXISTS max_amount_in_cents;
