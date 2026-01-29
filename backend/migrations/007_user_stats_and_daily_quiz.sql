-- ==========================================
-- User Stats and Daily Quiz Migration
-- ==========================================
-- Purpose: Add user statistics tracking for streaks and daily quiz feature
-- Date: 2026-01-21
-- ==========================================

-- ==========================================
-- User Stats Table
-- ==========================================

CREATE TABLE IF NOT EXISTS user_stats (
    user_id VARCHAR(100) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_streak INT DEFAULT 0 NOT NULL,
    longest_streak INT DEFAULT 0 NOT NULL,
    last_daily_quiz_at BIGINT,
    last_daily_quiz_date DATE,
    total_quizzes_completed INT DEFAULT 0 NOT NULL,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX idx_user_stats_streak ON user_stats(current_streak DESC);
CREATE INDEX idx_user_stats_last_daily ON user_stats(last_daily_quiz_date);

-- ==========================================
-- Initialize user_stats for existing users
-- ==========================================

INSERT INTO user_stats (user_id, total_quizzes_completed, created_at, updated_at)
SELECT
    u.id,
    COALESCE(COUNT(DISTINCT qs.id) FILTER (WHERE qs.status = 'completed'), 0) as total_completed,
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW())
FROM users u
LEFT JOIN quiz_sessions qs ON u.id = qs.user_id
GROUP BY u.id
ON CONFLICT (user_id) DO NOTHING;

-- ==========================================
-- Daily Quiz Selection Table
-- ==========================================
-- Stores the daily quiz selection (one row per day)

CREATE TABLE IF NOT EXISTS daily_quiz_selection (
    date DATE PRIMARY KEY,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX idx_daily_quiz_date ON daily_quiz_selection(date DESC);

-- ==========================================
-- Notes
-- ==========================================

-- 1. User Stats:
--    - current_streak: consecutive days user completed daily quiz
--    - longest_streak: best streak ever achieved
--    - last_daily_quiz_at: timestamp of last daily quiz completion
--    - last_daily_quiz_date: date (not timestamp) for easy streak calculation
--    - total_quizzes_completed: count of all completed quizzes (any type)
--
-- 2. Daily Quiz Selection:
--    - One row per day
--    - Application will select quiz_id using deterministic algorithm (e.g., hash of date)
--    - Allows manual override for special events
--
-- 3. Streak Calculation Logic:
--    - If last_daily_quiz_date = yesterday: increment current_streak
--    - If last_daily_quiz_date = today: no change (already completed today)
--    - Otherwise: reset current_streak to 1
--    - Update longest_streak if current_streak > longest_streak
--
-- 4. Future Enhancements:
--    - Add weekly/monthly stats
--    - Add achievements table referencing user_stats
