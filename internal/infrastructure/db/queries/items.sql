-- name: CreateItem :exec
INSERT INTO items (id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4);

-- name: GetItem :one
SELECT * FROM items WHERE id = $1;

-- name: ListItems :many
SELECT * FROM items
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;
