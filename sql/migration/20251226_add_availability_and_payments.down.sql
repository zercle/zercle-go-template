DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_transaction_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_booking_id;

DROP INDEX IF EXISTS idx_availability_is_active;
DROP INDEX IF EXISTS idx_availability_day_of_week;
DROP INDEX IF EXISTS idx_availability_service_id;

DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS availability_slots;
