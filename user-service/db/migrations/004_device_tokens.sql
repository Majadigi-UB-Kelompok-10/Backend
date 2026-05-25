CREATE TABLE IF NOT EXISTS device_tokens (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT         NOT NULL,
    platform   VARCHAR(10)  NOT NULL CHECK (platform IN ('android', 'ios')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, platform)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user_id ON device_tokens(user_id);

CREATE TRIGGER set_timestamp_device_tokens
    BEFORE UPDATE ON device_tokens
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
