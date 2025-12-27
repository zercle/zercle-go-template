-- name: CreateAvailabilitySlot :one
INSERT INTO availability_slots (service_id, day_of_week, start_time, end_time, max_bookings, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAvailabilitySlot :one
SELECT * FROM availability_slots
WHERE id = $1;

-- name: ListAvailabilitySlotsByService :many
SELECT * FROM availability_slots
WHERE service_id = $1 AND is_active = $2
ORDER BY day_of_week ASC, start_time ASC;

-- name: GetAvailabilitySlotForDayAndTime :one
SELECT * FROM availability_slots
WHERE service_id = $1 AND day_of_week = $2 AND is_active = $3
LIMIT 1;

-- name: UpdateAvailabilitySlot :one
UPDATE availability_slots
SET start_time = COALESCE(sqlc.narg('start_time'), start_time),
    end_time = COALESCE(sqlc.narg('end_time'), end_time),
    max_bookings = COALESCE(sqlc.narg('max_bookings'), max_bookings),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAvailabilitySlot :exec
DELETE FROM availability_slots
WHERE id = $1;

-- name: ListAllAvailabilitySlots :many
SELECT * FROM availability_slots
ORDER BY service_id ASC, day_of_week ASC, start_time ASC;
