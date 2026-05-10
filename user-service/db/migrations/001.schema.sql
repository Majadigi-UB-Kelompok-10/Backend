CREATE TYPE gender_type AS ENUM ('LAKI_LAKI', 'PEREMPUAN');
CREATE TYPE user_role   AS ENUM ('user', 'admin', 'superadmin');

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE IF NOT EXISTS users (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name              VARCHAR(100)    NOT NULL
                                CHECK (length(trim(first_name)) > 0),
    last_name               VARCHAR(100)    NOT NULL
                                CHECK (length(trim(last_name)) > 0),
    phone                   VARCHAR(25)     NOT NULL UNIQUE
                                CHECK (phone ~ '^\+62[0-9]{8,13}$'),
    email                   VARCHAR(150)    NOT NULL UNIQUE
                                CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+\.[A-Za-z]+$'),
    nik                     CHAR(16)        NOT NULL UNIQUE
                                CHECK (nik ~ '^[0-9]{16}$'),
    address                 TEXT,
    birth_date              DATE,
    gender                  gender_type,
    password_hash           VARCHAR(255)    NOT NULL,
    role                    user_role       NOT NULL DEFAULT 'user',
    is_active               BOOLEAN         NOT NULL DEFAULT true,
    email_verified          BOOLEAN         NOT NULL DEFAULT false,
    email_verify_token      VARCHAR(64),
    email_verify_expires_at TIMESTAMPTZ,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email              ON users(email);
CREATE INDEX idx_users_phone              ON users(phone);
CREATE INDEX idx_users_nik                ON users(nik);
CREATE INDEX idx_users_role               ON users(role);
CREATE INDEX idx_users_active             ON users(is_active);
CREATE INDEX idx_users_email_verify_token ON users(email_verify_token);

CREATE TRIGGER set_timestamp_users
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();


CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_revoked  BOOLEAN     NOT NULL DEFAULT false
);

CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);


CREATE TABLE IF NOT EXISTS user_category_preferences (
    user_id     UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID    NOT NULL,
    PRIMARY KEY (user_id, category_id)
);

CREATE INDEX idx_user_prefs_user_id ON user_category_preferences(user_id);


CREATE TABLE IF NOT EXISTS user_service_favorites (
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id UUID        NOT NULL,
    added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, service_id)
);

CREATE INDEX idx_user_favorites_user_id ON user_service_favorites(user_id);
