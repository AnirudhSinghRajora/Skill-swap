-- Migration 007: Update swap_requests for two-sided completion flow
-- Swap is not "done" until both parties confirm completion

-- Add completion tracking columns
ALTER TABLE swap_requests
    ADD COLUMN IF NOT EXISTS requester_completed BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS responder_completed BOOLEAN DEFAULT FALSE;

-- Update swap_status enum to include 'completed'
-- PostgreSQL doesn't allow direct ALTER TYPE ADD VALUE inside transactions,
-- so we use a DO block with exception handling
DO $$ BEGIN
    ALTER TYPE swap_status ADD VALUE IF NOT EXISTS 'completed';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
