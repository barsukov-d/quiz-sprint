-- ==========================================
-- Global Leaderboard Migration
-- ==========================================
-- Purpose: Create global leaderboard VIEW that aggregates user scores
-- across all quizzes (sum of best scores per quiz)
-- Author: Claude
-- Date: 2026-01-21
-- ==========================================

-- ==========================================
-- Performance Indexes
-- ==========================================

-- Index for filtering completed sessions efficiently
CREATE INDEX IF NOT EXISTS idx_sessions_completed
ON quiz_sessions(status, completed_at DESC)
WHERE status = 'completed';

-- Index for grouping by user and quiz with score sorting
CREATE INDEX IF NOT EXISTS idx_sessions_user_quiz_score
ON quiz_sessions(user_id, quiz_id, score DESC);

-- ==========================================
-- Add Missing Columns (if not exist)
-- ==========================================

-- Streak tracking column (added in migration 005, but ensure it exists)
ALTER TABLE quiz_sessions
ADD COLUMN IF NOT EXISTS correct_answer_streak INT DEFAULT 0;

-- Points breakdown columns for user_answers table
ALTER TABLE user_answers
    ADD COLUMN IF NOT EXISTS base_points INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS time_bonus INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS streak_bonus INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS time_spent BIGINT DEFAULT 0;

-- Backfill base_points from existing points column
UPDATE user_answers
SET base_points = points
WHERE base_points = 0 AND points > 0;

-- ==========================================
-- Global Leaderboard VIEW
-- ==========================================

CREATE OR REPLACE VIEW global_leaderboard AS
WITH user_best_scores AS (
    -- For each user and quiz, find their BEST score (MAX)
    SELECT
        user_id,
        quiz_id,
        MAX(score) as best_score,
        MAX(completed_at) as last_completed_at
    FROM quiz_sessions
    WHERE status = 'completed'
    GROUP BY user_id, quiz_id
),
user_total_scores AS (
    -- For each user, sum all their best scores across all quizzes
    SELECT
        user_id,
        SUM(best_score) as total_score,
        COUNT(DISTINCT quiz_id) as quizzes_completed,
        MAX(last_completed_at) as last_activity_at
    FROM user_best_scores
    GROUP BY user_id
)
SELECT
    user_id,
    total_score,
    quizzes_completed,
    last_activity_at,
    ROW_NUMBER() OVER (ORDER BY total_score DESC, last_activity_at ASC) as rank
FROM user_total_scores
ORDER BY rank;

-- ==========================================
-- Verification Query (commented out)
-- ==========================================

-- To verify the VIEW works correctly, uncomment and run:
-- SELECT * FROM global_leaderboard LIMIT 10;

-- ==========================================
-- Notes
-- ==========================================

-- 1. Scoring Logic:
--    - For each quiz, take user's BEST score (if completed multiple times)
--    - Sum all best scores to get total_score
--    - Example: Quiz A (100, 150) + Quiz B (200) = 150 + 200 = 350 total
--
-- 2. Ranking:
--    - Primary sort: total_score DESC (highest first)
--    - Tiebreaker: last_activity_at ASC (earliest completion wins)
--
-- 3. Performance:
--    - Indexed on (user_id, quiz_id, score) for fast grouping
--    - Filtered on status = 'completed' with partial index
--    - VIEW is computed on-the-fly (no materialization)
--
-- 4. Future Optimization:
--    - If query becomes slow, consider MATERIALIZED VIEW
--    - Would require REFRESH MATERIALIZED VIEW after quiz completion
--
-- 5. Backward Compatibility:
--    - Existing 'leaderboard' VIEW (per-quiz) remains unchanged
--    - This is an additive change, no breaking changes
