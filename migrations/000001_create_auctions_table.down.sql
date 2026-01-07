-- 000001_create_auctions_table.down.sql
DROP INDEX IF EXISTS idx_auctions_created_at;
DROP INDEX IF EXISTS idx_auctions_state_created_at;
DROP INDEX IF EXISTS idx_auctions_state_start_time;
DROP INDEX IF EXISTS idx_auctions_state_end_time;
DROP INDEX IF EXISTS idx_auctions_listing_id;
DROP TABLE IF EXISTS auctions;
DROP TYPE IF EXISTS auction_state;
