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

-- name: CreateUser :exec
INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateUser :exec
UPDATE users
SET username = $2, email = $3, password_hash = $4, display_name = $5, avatar_url = $6, status = $7, updated_at = $8
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;

-- name: CreateRoom :exec
INSERT INTO rooms (id, name, description, type, owner_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetRoomByID :one
SELECT id, name, description, type, owner_id, created_at, updated_at, deleted_at
FROM rooms
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListRooms :many
SELECT id, name, description, type, owner_id, created_at, updated_at, deleted_at
FROM rooms
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateRoom :exec
UPDATE rooms
SET name = $2, description = $3, type = $4, updated_at = $5
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteRoom :exec
UPDATE rooms SET deleted_at = NOW() WHERE id = $1;

-- name: AddRoomMember :exec
INSERT INTO room_members (room_id, user_id, role, joined_at)
VALUES ($1, $2, $3, $4);

-- name: RemoveRoomMember :exec
DELETE FROM room_members WHERE room_id = $1 AND user_id = $2;

-- name: GetRoomMembers :many
SELECT rm.room_id, rm.user_id, rm.role, rm.joined_at, u.username, u.email, u.display_name, u.avatar_url, u.status
FROM room_members rm
JOIN users u ON rm.user_id = u.id
WHERE rm.room_id = $1 AND u.deleted_at IS NULL;

-- name: GetRoomMember :one
SELECT room_id, user_id, role, joined_at
FROM room_members
WHERE room_id = $1 AND user_id = $2;

-- name: CreateMessage :exec
INSERT INTO messages (id, room_id, sender_id, content, message_type, reply_to, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetMessageByID :one
SELECT id, room_id, sender_id, content, message_type, reply_to, created_at, updated_at, deleted_at
FROM messages
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetMessageHistory :many
SELECT id, room_id, sender_id, content, message_type, reply_to, created_at, updated_at, deleted_at
FROM messages
WHERE room_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateMessage :exec
UPDATE messages
SET content = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteMessage :exec
UPDATE messages SET deleted_at = NOW() WHERE id = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, created_at, revoked_at
FROM refresh_tokens
WHERE token = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1;

-- name: DeleteRefreshTokenByUserID :exec
DELETE FROM refresh_tokens WHERE user_id = $1;
