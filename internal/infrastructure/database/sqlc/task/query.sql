-- name: CreateTask :one
INSERT INTO tasks (
    id,
    title,
    description,
    status,
    priority,
    user_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM tasks
WHERE ($1::uuid IS NULL OR user_id = $1)
  AND ($2::varchar IS NULL OR status = $2)
  AND ($3::varchar IS NULL OR priority = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: UpdateTask :one
UPDATE tasks SET
    title = $2,
    description = $3,
    status = $4,
    priority = $5,
    updated_at = $6
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;

-- name: CountTasks :one
SELECT COUNT(*) FROM tasks
WHERE ($1::uuid IS NULL OR user_id = $1)
  AND ($2::varchar IS NULL OR status = $2)
  AND ($3::varchar IS NULL OR priority = $3);

-- name: ExistsTaskByID :one
SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1);
