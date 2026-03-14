-- Users table queries

-- name: CreateUser :exec
INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetUserByID :one
SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at, deleted_at
FROM users
WHERE username = $1 AND deleted_at IS NULL;

-- name: UpdateUser :exec
UPDATE users
SET username = $2, email = $3, password_hash = $4, display_name = $5, avatar_url = $6, status = $7, updated_at = $8
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, created_at, revoked_at
FROM refresh_tokens
WHERE token = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1;
