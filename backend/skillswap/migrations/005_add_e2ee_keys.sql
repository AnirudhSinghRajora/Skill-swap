-- Migration 005: Add E2EE key storage fields to users table
-- Adds fields for storing public keys and password-encrypted private key backups.
-- The server NEVER sees plaintext private keys — only the encrypted backup blob.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS public_key TEXT,
ADD COLUMN IF NOT EXISTS encrypted_key_backup TEXT;

-- Partial index: efficiently find users who have published a public key
CREATE INDEX IF NOT EXISTS idx_users_has_public_key
    ON users (user_id) WHERE public_key IS NOT NULL;
