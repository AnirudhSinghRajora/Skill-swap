-- Migration 006: Add encrypted flag to messages table for E2EE support.
-- Existing messages default to false (plaintext). New E2EE messages set this to true.

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS encrypted BOOLEAN NOT NULL DEFAULT false;
