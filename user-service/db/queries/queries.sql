-- name: GetUserByEmailForAuth :one
SELECT id, first_name, last_name, phone, email, nik, address, birth_date, gender,
       password_hash, role, is_active, email_verified, created_at, updated_at
FROM users WHERE email = sqlc.arg('email') LIMIT 1;

-- name: GetUserByID :one
SELECT id, first_name, last_name, phone, email, nik, address, birth_date, gender,
       role, is_active, created_at, updated_at
FROM users WHERE id = sqlc.arg('id') LIMIT 1;

-- name: CheckEmailExists :one
SELECT id FROM users WHERE email = sqlc.arg('email') LIMIT 1;

-- name: CheckPhoneExists :one
SELECT id FROM users WHERE phone = sqlc.arg('phone') LIMIT 1;

-- name: CheckNIKExists :one
SELECT id FROM users WHERE nik = sqlc.arg('nik') LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    first_name, last_name, phone, email, nik,
    address, birth_date, gender, password_hash
) VALUES (
    sqlc.arg('first_name'),
    sqlc.arg('last_name'),
    sqlc.arg('phone'),
    sqlc.arg('email'),
    sqlc.arg('nik'),
    sqlc.narg('address'),
    sqlc.narg('birth_date'),
    sqlc.narg('gender'),
    sqlc.arg('password_hash')
)
RETURNING id, first_name, last_name, phone, email, nik,
          address, birth_date, gender, role, is_active, created_at, updated_at;

-- name: UpdateUserProfile :one
UPDATE users
SET
    first_name = sqlc.arg('first_name'),
    last_name  = sqlc.arg('last_name'),
    phone      = sqlc.arg('phone'),
    address    = sqlc.narg('address'),
    birth_date = sqlc.narg('birth_date'),
    gender     = sqlc.narg('gender')
WHERE id = sqlc.arg('id')
RETURNING id, first_name, last_name, phone, email, nik,
          address, birth_date, gender, role, is_active, created_at, updated_at;

-- name: UpdateUserRole :one
UPDATE users SET role = sqlc.arg('role')
WHERE id = sqlc.arg('id')
RETURNING id, first_name, last_name, email, role, is_active, updated_at;

-- name: DeactivateUser :exec
UPDATE users SET is_active = false WHERE id = sqlc.arg('id');

-- name: BumpFavoritesTimestamp :exec
UPDATE users SET favorites_updated_at = NOW() WHERE id = sqlc.arg('id');

-- name: GetFavoritesUpdatedAt :one
SELECT favorites_updated_at FROM users WHERE id = sqlc.arg('id');

-- name: ListUsers :many
SELECT id, first_name, last_name, phone, email, nik, role, is_active, created_at
FROM users
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- =============================================================================
-- Refresh Token Queries
-- =============================================================================

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('token_hash'),
    sqlc.arg('expires_at')
);

-- name: GetRefreshToken :one
SELECT id, user_id, token_hash, expires_at, created_at, is_revoked
FROM refresh_tokens
WHERE token_hash = sqlc.arg('token_hash') LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET is_revoked = true
WHERE token_hash = sqlc.arg('token_hash');

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens SET is_revoked = true
WHERE user_id = sqlc.arg('user_id');

-- name: DeleteExpiredTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < NOW() OR is_revoked = true;

-- =============================================================================
-- User Category Preferences
-- =============================================================================

-- name: GetUserPreferences :many
SELECT category_id
FROM user_category_preferences
WHERE user_id = sqlc.arg('user_id');

-- name: DeleteUserPreferences :exec
DELETE FROM user_category_preferences
WHERE user_id = sqlc.arg('user_id');

-- name: InsertUserPreference :exec
INSERT INTO user_category_preferences (user_id, category_id)
VALUES (sqlc.arg('user_id'), sqlc.arg('category_id'))
ON CONFLICT DO NOTHING;

-- name: DeleteUserPreference :exec
DELETE FROM user_category_preferences
WHERE user_id = sqlc.arg('user_id')
  AND category_id = sqlc.arg('category_id');

-- =============================================================================
-- Email Verification
-- =============================================================================

-- name: SetEmailVerificationToken :exec
UPDATE users
SET email_verify_token = sqlc.arg('email_verify_token'),
    email_verify_expires_at = sqlc.arg('email_verify_expires_at')
WHERE id = sqlc.arg('id');

-- name: GetUserByEmailVerifyToken :one
SELECT id, first_name, last_name, email, email_verified
FROM users
WHERE email_verify_token = sqlc.arg('token')
  AND email_verify_expires_at > NOW()
  AND email_verified = false
LIMIT 1;

-- name: MarkEmailVerified :exec
UPDATE users
SET email_verified = true,
    email_verify_token = NULL,
    email_verify_expires_at = NULL
WHERE id = sqlc.arg('id');

-- =============================================================================
-- Service Favorites
-- =============================================================================

-- name: GetUserFavorites :many
SELECT service_id, added_at AS updated_at FROM user_service_favorites
WHERE user_id = sqlc.arg('user_id')
ORDER BY added_at DESC;

-- name: AddUserFavorite :exec
INSERT INTO user_service_favorites (user_id, service_id)
VALUES (sqlc.arg('user_id'), sqlc.arg('service_id'))
ON CONFLICT DO NOTHING;

-- name: RemoveUserFavorite :exec
DELETE FROM user_service_favorites
WHERE user_id = sqlc.arg('user_id') AND service_id = sqlc.arg('service_id');

-- name: GetUserFavoriteCount :one
SELECT COUNT(*) FROM user_service_favorites WHERE user_id = sqlc.arg('user_id');
