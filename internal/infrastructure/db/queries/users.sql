-- name: GetUserByID :one
-- Get a user by their ID
SELECT id, email, password_hash, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
-- Get a user by their email address
SELECT id, email, password_hash, name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
-- Create a new user and return the created record
INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, password_hash, name, created_at, updated_at;

-- name: UpdateUser :one
-- Update an existing user and return the updated record
UPDATE users
SET email = $2,
    password_hash = $3,
    name = $4,
    updated_at = $5
WHERE id = $1
RETURNING id, email, password_hash, name, created_at, updated_at;

-- name: DeleteUser :exec
-- Delete a user by their ID
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
-- Get all users with pagination support
SELECT id, email, password_hash, name, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
-- Get the total number of users
SELECT COUNT(*)
FROM users;

-- name: CheckUserExists :one
-- Check if a user with the given email exists
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
) AS exists;
