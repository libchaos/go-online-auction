-- +goose Up

-- notification_preferences stores per-user channel opt-in as a JSONB map keyed
-- by category, e.g. {"payment":{"in_app":true,"email":true},
-- "auction":{"in_app":true,"email":false}}. A missing row means the user has
-- never customised preferences; the application layer then applies
-- DefaultPreferences(), so no seed row is required.
CREATE TABLE notification_preferences (
    user_id     BIGINT       PRIMARY KEY,
    preferences JSONB        NOT NULL DEFAULT '{}',
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS notification_preferences;
