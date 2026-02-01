-- ==========================================
-- Add question_started_at to daily_games
-- ==========================================
-- Adds timestamp tracking for server-side timer persistence.
-- Tracks when the current question started to calculate time remaining.
--
-- Context: Timer persistence for Daily Challenge
-- Date: 2026-02-01
-- ==========================================

-- Add question_started_at column
ALTER TABLE daily_games
ADD COLUMN IF NOT EXISTS question_started_at BIGINT;

-- Set default value for existing rows (use session started_at if available, else current time)
UPDATE daily_games
SET question_started_at = COALESCE(
    (session_state->>'started_at')::bigint,
    EXTRACT(EPOCH FROM NOW())::bigint
)
WHERE question_started_at IS NULL;

-- Make column NOT NULL after backfilling
ALTER TABLE daily_games
ALTER COLUMN question_started_at SET NOT NULL;

-- Add comment
COMMENT ON COLUMN daily_games.question_started_at IS 'Unix timestamp when current question started (for server-side timer)';
