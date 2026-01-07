-- 000001_create_auctions_table.up.sql

-- Create enum type for auction state
CREATE TYPE auction_state AS ENUM ('draft', 'active', 'closed', 'cancelled');

CREATE TABLE auctions (
    id             BIGSERIAL PRIMARY KEY,
    listing_id     BIGINT NOT NULL,
    start_time     TIMESTAMPTZ,
    end_time       TIMESTAMPTZ NOT NULL,
    state          auction_state NOT NULL DEFAULT 'draft',
    highest_bid_id BIGINT,
    highest_bid_amount_in_cents BIGINT,
    version        BIGINT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_end_after_start CHECK (end_time > start_time)
);

-- Index for listing-based lookups (find auction by listing)
CREATE INDEX idx_auctions_listing_id ON auctions(listing_id);

-- Composite index for scheduler queries: find active auctions ending soon
-- Covers: WHERE state = 'active' ORDER BY end_time
CREATE INDEX idx_auctions_state_end_time ON auctions(state, end_time);

-- Composite index for scheduler queries: find draft auctions ready to start
-- Covers: WHERE state = 'draft' AND start_time <= now()
CREATE INDEX idx_auctions_state_start_time ON auctions(state, start_time);
