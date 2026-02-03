-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    first_name,
    last_name,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users SET
    email = $2,
    first_name = $3,
    last_name = $4,
    status = $5,
    updated_at = $6
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
