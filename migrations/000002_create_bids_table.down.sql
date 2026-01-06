-- 000002_create_bids_table.down.sql
ALTER TABLE auctions DROP CONSTRAINT IF EXISTS fk_auctions_highest_bid;
DROP TABLE IF EXISTS bids;
