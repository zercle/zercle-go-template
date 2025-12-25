-- name: CreateBooking :one
INSERT INTO bookings (id, user_id, service_id, start_time, end_time, status, total_price, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings
WHERE id = $1;

-- name: ListBookingsByUser :many
SELECT * FROM bookings
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListBookingsByService :many
SELECT * FROM bookings
WHERE service_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListBookingsByStatus :many
SELECT * FROM bookings
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListBookingsByDateRange :many
SELECT * FROM bookings
WHERE start_time >= $1 AND start_time <= $2
ORDER BY start_time ASC
LIMIT $3 OFFSET $4;

-- name: ListBookingsByServiceAndDateRange :many
SELECT * FROM bookings
WHERE service_id = $1 AND start_time >= $2 AND start_time <= $3
ORDER BY start_time ASC;

-- name: UpdateBookingStatus :one
UPDATE bookings
SET status = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: CancelBooking :one
UPDATE bookings
SET status = 'cancelled', cancelled_at = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings
WHERE id = $1;

-- name: CheckBookingConflict :many
SELECT * FROM bookings
WHERE service_id = $1
  AND status NOT IN ('cancelled', 'completed')
  AND (
    (start_time < $2 AND end_time > $2) OR
    (start_time < $3 AND end_time > $3) OR
    (start_time >= $2 AND end_time <= $3)
  );

-- name: GetActiveBookingsCount :one
SELECT COUNT(*) FROM bookings
WHERE service_id = $1
  AND status NOT IN ('cancelled', 'completed')
  AND (
    (start_time < $2 AND end_time > $2) OR
    (start_time < $3 AND end_time > $3) OR
    (start_time >= $2 AND end_time <= $3)
  );

-- name: GetBookingStats :one
SELECT 
    COUNT(*) as total_bookings,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    COUNT(*) FILTER (WHERE status = 'confirmed') as confirmed_count,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_count,
    COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_count,
    COALESCE(SUM(total_price), 0) as total_revenue
FROM bookings
WHERE created_at >= $1 AND created_at <= $2;
