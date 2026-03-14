-- Rooms table queries

-- name: CreateRoom :exec
INSERT INTO rooms (id, name, description, type, owner_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetRoomByID :one
SELECT r.id, r.name, r.description, r.type, r.owner_id, 
       COUNT(rm.user_id) as member_count, r.created_at, r.updated_at, r.deleted_at
FROM rooms r
LEFT JOIN room_members rm ON r.id = rm.room_id
WHERE r.id = $1 AND r.deleted_at IS NULL
GROUP BY r.id;

-- name: GetRoomsByUserID :many
SELECT r.id, r.name, r.description, r.type, r.owner_id,
       COUNT(rm.user_id) as member_count, r.created_at, r.updated_at, r.deleted_at
FROM rooms r
LEFT JOIN room_members rm ON r.id = rm.room_id
JOIN room_members my_rm ON r.id = my_rm.room_id AND my_rm.user_id = $1
WHERE r.deleted_at IS NULL
GROUP BY r.id
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountRoomsByUserID :one
SELECT COUNT(*)
FROM room_members rm
JOIN rooms r ON rm.room_id = r.id
WHERE rm.user_id = $1 AND r.deleted_at IS NULL;

-- name: UpdateRoom :exec
UPDATE rooms
SET name = $2, description = $3, updated_at = $4
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteRoom :exec
UPDATE rooms SET deleted_at = NOW() WHERE id = $1;

-- name: AddRoomMember :exec
INSERT INTO room_members (room_id, user_id, role, joined_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (room_id, user_id) DO UPDATE SET role = $3;

-- name: RemoveRoomMember :exec
DELETE FROM room_members WHERE room_id = $1 AND user_id = $2;

-- name: GetRoomMembers :many
SELECT rm.room_id, rm.user_id, u.username, u.display_name, u.avatar_url, rm.role, rm.joined_at
FROM room_members rm
JOIN users u ON rm.user_id = u.id
WHERE rm.room_id = $1;

-- name: IsRoomMember :one
SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2);

-- Messages table queries

-- name: CreateMessage :exec
INSERT INTO messages (id, room_id, sender_id, content, message_type, reply_to, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetMessageByID :one
SELECT m.id, m.room_id, m.sender_id, u.username, m.content, m.message_type, m.reply_to, m.created_at, m.updated_at, m.deleted_at
FROM messages m
LEFT JOIN users u ON m.sender_id = u.id
WHERE m.id = $1 AND m.deleted_at IS NULL;

-- name: GetMessagesByRoomID :many
SELECT m.id, m.room_id, m.sender_id, u.username, m.content, m.message_type, m.reply_to, m.created_at, m.updated_at, m.deleted_at
FROM messages m
LEFT JOIN users u ON m.sender_id = u.id
WHERE m.room_id = $1 AND m.deleted_at IS NULL
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetMessagesByRoomIDBefore :many
SELECT m.id, m.room_id, m.sender_id, u.username, m.content, m.message_type, m.reply_to, m.created_at, m.updated_at, m.deleted_at
FROM messages m
LEFT JOIN users u ON m.sender_id = u.id
WHERE m.room_id = $1 AND m.deleted_at IS NULL AND m.created_at < (
    SELECT messages.created_at FROM messages WHERE messages.id = $2
)
ORDER BY m.created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteMessage :exec
UPDATE messages SET deleted_at = NOW() WHERE id = $1;
