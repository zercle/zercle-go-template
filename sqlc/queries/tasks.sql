-- name: CreateTask :one
INSERT INTO tasks (user_id, title, description, status, priority, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1;

-- name: GetTaskByUserAndID :one
SELECT * FROM tasks
WHERE id = $1 AND user_id = $2;

-- name: ListTasksByUser :many
SELECT * FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTasksByUserAndStatus :many
SELECT * FROM tasks
WHERE user_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    priority = COALESCE(sqlc.narg('priority'), priority),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    completed_at = COALESCE(sqlc.narg('completed_at'), completed_at),
    updated_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: CountTasksByUser :one
SELECT COUNT(*) FROM tasks
WHERE user_id = $1;
