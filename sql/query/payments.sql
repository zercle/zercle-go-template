-- name: CreatePayment :one
INSERT INTO payments (id, booking_id, amount, status, payment_method, transaction_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments
WHERE id = $1;

-- name: GetPaymentByBooking :one
SELECT * FROM payments
WHERE booking_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListPaymentsByBooking :many
SELECT * FROM payments
WHERE booking_id = $1
ORDER BY created_at DESC;

-- name: ListPaymentsByStatus :many
SELECT * FROM payments
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2, transaction_id = COALESCE(sqlc.narg('transaction_id'), transaction_id), updated_at = $3
WHERE id = $1
RETURNING *;

-- name: ConfirmPayment :one
UPDATE payments
SET status = 'completed', paid_at = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: RefundPayment :one
UPDATE payments
SET status = 'refunded', refunded_at = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: DeletePayment :exec
DELETE FROM payments
WHERE id = $1;

-- name: GetPaymentByTransactionId :one
SELECT * FROM payments
WHERE transaction_id = $1
LIMIT 1;

-- name: GetPaymentStats :one
SELECT 
    COUNT(*) as total_payments,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_count,
    COUNT(*) FILTER (WHERE status = 'refunded') as refunded_count,
    COALESCE(SUM(amount) FILTER (WHERE status = 'completed'), 0) as total_revenue,
    COALESCE(SUM(amount) FILTER (WHERE status = 'refunded'), 0) as total_refunds
FROM payments
WHERE created_at >= $1 AND created_at <= $2;
