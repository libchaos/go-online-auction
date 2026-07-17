-- +goose Up
CREATE TYPE user_role AS ENUM ('admin', 'seller', 'bidder');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'blocked');

CREATE TABLE users (
    id                BIGSERIAL PRIMARY KEY,
    name              TEXT NOT NULL,
    email             TEXT NOT NULL,
    password_hash     TEXT,
    role              user_role NOT NULL DEFAULT 'bidder',
    status            user_status NOT NULL DEFAULT 'active',
    oauth_provider    TEXT,
    oauth_provider_id TEXT,
    version           BIGINT NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT chk_users_credential CHECK (
        password_hash IS NOT NULL
        OR (oauth_provider IS NOT NULL AND oauth_provider_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_users_oauth ON users (oauth_provider, oauth_provider_id)
    WHERE oauth_provider IS NOT NULL;

CREATE INDEX idx_users_role ON users (role);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
