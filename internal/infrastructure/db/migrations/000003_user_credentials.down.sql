-- Migration: Drop user_credentials table

-- Add password_hash back to users table
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

-- Drop user_credentials table and indexes
DROP INDEX IF EXISTS idx_user_credentials_user_id;
DROP TABLE IF EXISTS user_credentials;
