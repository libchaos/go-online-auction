-- 000002_create_bids_table.up.sql

CREATE TABLE bids (
    id              BIGSERIAL PRIMARY KEY,
    auction_id      BIGINT NOT NULL REFERENCES auctions(id),
    user_id         BIGINT NOT NULL,
    amount_in_cents BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_positive_amount CHECK (amount_in_cents > 0)
);

-- Composite index for FindByAuctionID with ORDER BY created_at
-- Covers: WHERE auction_id = $1 ORDER BY created_at ASC
CREATE INDEX idx_bids_auction_id_created_at ON bids(auction_id, created_at);

-- Index for user bid history queries
-- Covers: WHERE user_id = $1 ORDER BY created_at DESC
CREATE INDEX idx_bids_user_id_created_at ON bids(user_id, created_at DESC);

-- Composite index for finding highest bid per auction efficiently
-- Covers: WHERE auction_id = $1 ORDER BY amount_in_cents DESC LIMIT 1
CREATE INDEX idx_bids_auction_id_amount ON bids(auction_id, amount_in_cents DESC);

-- Add foreign key from auctions.highest_bid_id to bids.id
ALTER TABLE auctions 
    ADD CONSTRAINT fk_auctions_highest_bid 
    FOREIGN KEY (highest_bid_id) REFERENCES bids(id);
