-- name: CreateBookingUser :exec
INSERT INTO bookings_users (booking_id, user_id, is_primary)
VALUES ($1, $2, $3);

-- name: GetBookingUsers :many
SELECT u.* 
FROM users u
JOIN bookings_users bu ON u.id = bu.user_id
WHERE bu.booking_id = $1
ORDER BY bu.is_primary DESC, u.created_at ASC;

-- name: DeleteBookingUsers :exec
DELETE FROM bookings_users
WHERE booking_id = $1;

-- name: DeleteBookingUser :exec
DELETE FROM bookings_users
WHERE booking_id = $1 AND user_id = $2;

-- name: GetPrimaryBookingUser :one
SELECT u.* 
FROM users u
JOIN bookings_users bu ON u.id = bu.user_id
WHERE bu.booking_id = $1 AND bu.is_primary = TRUE
LIMIT 1;

-- name: SetPrimaryBookingUser :exec
UPDATE bookings_users
SET is_primary = CASE WHEN user_id = $2 THEN TRUE ELSE FALSE END
WHERE booking_id = $1;
