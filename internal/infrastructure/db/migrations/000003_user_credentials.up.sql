-- Migration: Create user_credentials table
-- Separates credential storage from user data for better security

CREATE TABLE user_credentials (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for credential lookups by user
CREATE INDEX idx_user_credentials_user_id ON user_credentials(user_id);

-- Remove password_hash from users table (migrated to credentials)
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
