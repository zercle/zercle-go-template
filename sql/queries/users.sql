-- name: GetUserByID :one
SELECT id, email, password_hash, first_name, last_name, is_active, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, is_active, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT id, email, password_hash, first_name, last_name, is_active, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :exec
INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW());

-- name: UpdateUser :exec
UPDATE users
SET email = $1, password_hash = $2, first_name = $3, last_name = $4, is_active = $5, updated_at = NOW()
WHERE id = $6 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: HardDeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) as count
FROM users
WHERE deleted_at IS NULL;
