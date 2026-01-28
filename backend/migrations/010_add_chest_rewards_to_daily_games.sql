-- ==========================================
-- Add Chest Rewards to Daily Games
-- ==========================================
-- Migration 010: Adds chest reward columns to daily_games table
-- and updates streak multipliers to match documentation.
--
-- See: docs/game_modes/daily_challenge/04_rewards.md
-- See: docs/GLOSSARY.md (streak multipliers)
-- ==========================================

-- ==========================================
-- Step 1: Add Chest Reward Columns
-- ==========================================

-- Chest type: wooden, silver, golden
ALTER TABLE daily_games ADD COLUMN IF NOT EXISTS chest_type VARCHAR(20);

-- Chest rewards
ALTER TABLE daily_games ADD COLUMN IF NOT EXISTS chest_coins INT;
ALTER TABLE daily_games ADD COLUMN IF NOT EXISTS chest_pvp_tickets INT;
ALTER TABLE daily_games ADD COLUMN IF NOT EXISTS chest_bonuses JSONB; -- Array of bonus types

-- Add constraints
ALTER TABLE daily_games ADD CONSTRAINT chest_type_check
    CHECK (chest_type IS NULL OR chest_type IN ('wooden', 'silver', 'golden'));

ALTER TABLE daily_games ADD CONSTRAINT chest_coins_check
    CHECK (chest_coins IS NULL OR chest_coins >= 0);

ALTER TABLE daily_games ADD CONSTRAINT chest_pvp_tickets_check
    CHECK (chest_pvp_tickets IS NULL OR chest_pvp_tickets >= 0);

-- Comments
COMMENT ON COLUMN daily_games.chest_type IS 'Type of chest earned: wooden (0-4 correct), silver (5-7), golden (8-10)';
COMMENT ON COLUMN daily_games.chest_coins IS 'Coins from chest (with streak multiplier applied)';
COMMENT ON COLUMN daily_games.chest_pvp_tickets IS 'PvP tickets from chest';
COMMENT ON COLUMN daily_games.chest_bonuses IS 'JSONB array of marathon bonus types: ["shield", "fifty_fifty", "skip", "freeze"]';

-- ==========================================
-- Step 2: Update Leaderboard Index with Correct Streak Multipliers
-- ==========================================

-- Drop old index with incorrect multipliers
DROP INDEX IF EXISTS idx_daily_games_score;

-- Create new index with correct multipliers per docs/GLOSSARY.md:
-- 0-2 days: 1.0 | 3-6 days: 1.1 | 7-13 days: 1.25 | 14-29 days: 1.4 | 30+ days: 1.5
CREATE INDEX idx_daily_games_score ON daily_games(
    date,
    ((session_state->>'base_score')::int * (
        CASE
            WHEN current_streak >= 30 THEN 1.5
            WHEN current_streak >= 14 THEN 1.4
            WHEN current_streak >= 7 THEN 1.25
            WHEN current_streak >= 3 THEN 1.1
            ELSE 1.0
        END
    )) DESC
) WHERE status = 'completed';

-- ==========================================
-- Step 3: Add Retry Tracking (for Second Attempt feature)
-- ==========================================

-- Track how many times player retried this day
ALTER TABLE daily_games ADD COLUMN IF NOT EXISTS attempt_number INT NOT NULL DEFAULT 1;

-- Update unique constraint to allow multiple attempts
ALTER TABLE daily_games DROP CONSTRAINT IF EXISTS daily_games_player_date_unique;

-- New constraint: unique (player_id, date, attempt_number)
ALTER TABLE daily_games ADD CONSTRAINT daily_games_player_date_attempt_unique
    UNIQUE (player_id, date, attempt_number);

-- Index for finding all attempts for a player on a date
CREATE INDEX IF NOT EXISTS idx_daily_games_player_date_attempts
    ON daily_games(player_id, date, attempt_number);

-- Comments
COMMENT ON COLUMN daily_games.attempt_number IS 'Attempt number for this date (1 = free, 2+ = paid retry)';

-- ==========================================
-- Notes
-- ==========================================

-- Chest rewards are calculated on game completion and stored here.
-- Chest opening endpoint is idempotent (just reads stored values).
--
-- Retry logic:
-- - Free users: 1 free + 1 paid retry (max 2 attempts)
-- - Premium users: 1 free + unlimited paid retries
-- - Best score counts for leaderboard (query uses MAX(score))
-- - Streak updated only on FIRST completion (attempt_number = 1)
