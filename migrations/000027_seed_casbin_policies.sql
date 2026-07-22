-- +goose Up
-- admin is a superuser: every path under /api/v1 and every action.
-- keyMatch4 turns "/*" into "/.*", so "/api/v1/*" matches any depth.
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES ('p', 'admin', '/api/v1/*', '*')
    ON CONFLICT DO NOTHING;

-- seller manages auctions and the catalog (SPU/SKU).
-- Path params use "{id}" (keyMatch4 syntax), not ":id".
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES
    ('p', 'seller', '/api/v1/auctions', 'POST'),
    ('p', 'seller', '/api/v1/auctions/{id}/start', 'PUT'),
    ('p', 'seller', '/api/v1/auctions/{id}/cancel', 'PUT'),
    ('p', 'seller', '/api/v1/spus', 'POST'),
    ('p', 'seller', '/api/v1/spus/{id}', 'PUT'),
    ('p', 'seller', '/api/v1/spus/{id}/publish', 'PUT'),
    ('p', 'seller', '/api/v1/spus/{id}/off-shelf', 'PUT'),
    ('p', 'seller', '/api/v1/spus/{id}/skus', 'POST'),
    ('p', 'seller', '/api/v1/skus/{id}', 'PUT'),
    ('p', 'seller', '/api/v1/skus/{id}/publish', 'PUT'),
    ('p', 'seller', '/api/v1/skus/{id}/off-shelf', 'PUT')
    ON CONFLICT DO NOTHING;

-- bidder may open deposits and check eligibility.
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES
    ('p', 'bidder', '/api/v1/deposits', 'POST'),
    ('p', 'bidder', '/api/v1/deposits/eligibility', 'GET')
    ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM casbin_rules WHERE ptype = 'p';
