-- +goose Up
CREATE TABLE casbin_rules (
    id    BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(10) NOT NULL,
    v0    VARCHAR(255) NOT NULL DEFAULT '',
    v1    VARCHAR(255) NOT NULL DEFAULT '',
    v2    VARCHAR(255) NOT NULL DEFAULT '',
    v3    VARCHAR(255) NOT NULL DEFAULT '',
    v4    VARCHAR(255) NOT NULL DEFAULT '',
    v5    VARCHAR(255) NOT NULL DEFAULT '',
    CONSTRAINT uq_casbin_rule UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
);

CREATE INDEX idx_casbin_rules_ptype ON casbin_rules (ptype);

-- +goose Down
DROP TABLE IF EXISTS casbin_rules;
