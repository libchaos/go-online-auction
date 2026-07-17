-- +goose Up
CREATE TYPE listing_status AS ENUM ('draft', 'published', 'off_shelf');

-- Categories form an adjacency-list tree: parent_id IS NULL marks root nodes.
CREATE TABLE categories (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    parent_id  BIGINT REFERENCES categories(id),
    sort_order INT NOT NULL DEFAULT 0,
    version    BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);

-- Sibling names must be unique; NULL parent_id needs its own partial index
-- because Postgres treats NULLs as distinct in unique constraints.
CREATE UNIQUE INDEX uq_categories_child_name ON categories (parent_id, name)
    WHERE parent_id IS NOT NULL;
CREATE UNIQUE INDEX uq_categories_root_name ON categories (name)
    WHERE parent_id IS NULL;

-- SPU (Standard Product Unit): the product listing shell shared by its SKUs.
CREATE TABLE spus (
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category_id BIGINT NOT NULL REFERENCES categories(id),
    brand       TEXT,
    images      JSONB NOT NULL DEFAULT '[]',
    status      listing_status NOT NULL DEFAULT 'draft',
    version     BIGINT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_spus_category_id ON spus (category_id);
CREATE INDEX idx_spus_status_created_at ON spus (status, created_at DESC);
CREATE INDEX idx_spus_created_at ON spus (created_at DESC);

-- SKU (Stock Keeping Unit): a concrete spec combination of an SPU.
-- auctions.listing_id points at skus.id semantically; no FK is added there so
-- the auction module stays decoupled (validated via a port at the app layer).
CREATE TABLE skus (
    id             BIGSERIAL PRIMARY KEY,
    spu_id         BIGINT NOT NULL REFERENCES spus(id),
    spec_values    JSONB NOT NULL DEFAULT '{}',
    price_in_cents BIGINT NOT NULL,
    quantity       BIGINT NOT NULL DEFAULT 0,
    status         listing_status NOT NULL DEFAULT 'draft',
    version        BIGINT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_skus_price_positive CHECK (price_in_cents > 0),
    CONSTRAINT chk_skus_quantity_nonneg CHECK (quantity >= 0)
);

CREATE INDEX idx_skus_spu_id ON skus (spu_id);
CREATE INDEX idx_skus_status ON skus (status);

-- +goose Down
DROP TABLE IF EXISTS skus;
DROP TABLE IF EXISTS spus;
DROP TABLE IF EXISTS categories;
DROP TYPE IF EXISTS listing_status;
