DROP INDEX IF EXISTS idx_bookings_users_booking_id;
DROP INDEX IF EXISTS idx_bookings_users_user_id;

DROP INDEX IF EXISTS idx_bookings_created_at;
DROP INDEX IF EXISTS idx_bookings_end_time;
DROP INDEX IF EXISTS idx_bookings_start_time;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_service_id;
DROP INDEX IF EXISTS idx_bookings_user_id;

DROP INDEX IF EXISTS idx_services_created_at;
DROP INDEX IF EXISTS idx_services_is_active;
DROP INDEX IF EXISTS idx_services_name;

DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS bookings_users;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
