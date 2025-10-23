-- name: CreatePost :exec
INSERT INTO posts (id, title, content, author_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetPost :one
SELECT id, title, content, author_id, created_at, updated_at
FROM posts
WHERE id = $1;

-- name: ListPosts :many
SELECT id, title, content, author_id, created_at, updated_at
FROM posts
ORDER BY created_at DESC;

-- name: UpdatePost :exec
UPDATE posts
SET title = $1, content = $2, updated_at = NOW()
WHERE id = $3;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;
