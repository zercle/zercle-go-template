-- name: CreateService :one
INSERT INTO services (name, description, duration_minutes, price, max_capacity, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetService :one
SELECT * FROM services
WHERE id = $1;

-- name: ListServices :many
SELECT * FROM services
WHERE is_active = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateService :one
UPDATE services
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    duration_minutes = COALESCE(sqlc.narg('duration_minutes'), duration_minutes),
    price = COALESCE(sqlc.narg('price'), price),
    max_capacity = COALESCE(sqlc.narg('max_capacity'), max_capacity),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = $1;

-- name: SearchServicesByName :many
SELECT * FROM services
WHERE name ILIKE $1 AND is_active = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: GetServicesByIds :many
SELECT * FROM services
WHERE id = ANY($1::uuid[])
ORDER BY created_at DESC;
