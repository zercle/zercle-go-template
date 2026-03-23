-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    password_hash,
    first_name,
    last_name,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users SET
    email = $2,
    password_hash = $3,
    first_name = $4,
    last_name = $5,
    status = $6,
    updated_at = $7
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE ($1::varchar IS NULL OR email ILIKE '%' || $1 || '%')
  AND ($2::varchar IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE ($1::varchar IS NULL OR email ILIKE '%' || $1 || '%')
  AND ($2::varchar IS NULL OR status = $2);

-- name: ExistsUserByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: ExistsUserByID :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);
