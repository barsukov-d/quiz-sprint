-- ==========================================
-- Daily Challenge Tables Migration
-- ==========================================
-- Creates tables for Daily Challenge game mode:
-- - daily_quizzes: One quiz per day (same for all players)
-- - daily_games: Each player's attempt per day
--
-- See: docs/02_daily_challenge.md
-- ==========================================

-- ==========================================
-- Daily Quizzes Table
-- ==========================================
CREATE TABLE IF NOT EXISTS daily_quizzes (
    id UUID PRIMARY KEY,
    date DATE NOT NULL UNIQUE, -- "2026-01-25" - one quiz per day
    question_ids JSONB NOT NULL, -- Array of 10 question IDs
    expires_at BIGINT NOT NULL, -- Unix timestamp (next day 00:00 UTC)
    created_at BIGINT NOT NULL,

    CONSTRAINT daily_quizzes_question_count CHECK (jsonb_array_length(question_ids) = 10)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_daily_quizzes_date ON daily_quizzes(date DESC);
CREATE INDEX IF NOT EXISTS idx_daily_quizzes_expires_at ON daily_quizzes(expires_at);

-- Comments
COMMENT ON TABLE daily_quizzes IS 'Daily quiz content - one quiz per day, same for all players';
COMMENT ON COLUMN daily_quizzes.date IS 'Date of the quiz (one per day)';
COMMENT ON COLUMN daily_quizzes.question_ids IS 'JSONB array of 10 question IDs';
COMMENT ON COLUMN daily_quizzes.expires_at IS 'When quiz expires (next day 00:00 UTC)';

-- ==========================================
-- Daily Games Table
-- ==========================================
CREATE TABLE IF NOT EXISTS daily_games (
    id UUID PRIMARY KEY,
    player_id VARCHAR(100) NOT NULL, -- References users(id)
    daily_quiz_id UUID NOT NULL, -- References daily_quizzes(id)
    date DATE NOT NULL, -- "2026-01-25"
    status VARCHAR(20) NOT NULL, -- "in_progress", "completed"
    session_state JSONB NOT NULL, -- QuizGameplaySession state

    -- Streak System
    current_streak INT NOT NULL DEFAULT 0, -- Days in a row
    best_streak INT NOT NULL DEFAULT 0, -- All-time best
    last_played_date DATE, -- Last date player completed

    -- Leaderboard
    rank INT, -- Player's rank (calculated after completion)

    CONSTRAINT daily_games_status_check CHECK (status IN ('in_progress', 'completed')),
    CONSTRAINT daily_games_player_date_unique UNIQUE (player_id, date), -- One attempt per day
    CONSTRAINT daily_games_streak_check CHECK (current_streak >= 0 AND best_streak >= current_streak)
);

-- Foreign Keys (when tables exist)
-- ALTER TABLE daily_games ADD CONSTRAINT fk_daily_game_quiz FOREIGN KEY (daily_quiz_id) REFERENCES daily_quizzes(id) ON DELETE CASCADE;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_daily_games_player ON daily_games(player_id);
CREATE INDEX IF NOT EXISTS idx_daily_games_date ON daily_games(date DESC);
CREATE INDEX IF NOT EXISTS idx_daily_games_player_date ON daily_games(player_id, date);
CREATE INDEX IF NOT EXISTS idx_daily_games_leaderboard ON daily_games(date, status) WHERE status = 'completed';

-- Leaderboard index (score = base_score * streak_bonus)
CREATE INDEX IF NOT EXISTS idx_daily_games_score ON daily_games(
    date,
    ((session_state->>'base_score')::int * (
        CASE
            WHEN current_streak >= 100 THEN 2.0
            WHEN current_streak >= 30 THEN 1.6
            WHEN current_streak >= 14 THEN 1.4
            WHEN current_streak >= 7 THEN 1.25
            WHEN current_streak >= 3 THEN 1.1
            ELSE 1.0
        END
    )) DESC
) WHERE status = 'completed';

-- Comments
COMMENT ON TABLE daily_games IS 'Player attempts at daily challenges - one per player per day';
COMMENT ON COLUMN daily_games.session_state IS 'Serialized QuizGameplaySession with answers';
COMMENT ON COLUMN daily_games.current_streak IS 'Current daily streak (consecutive days)';
COMMENT ON COLUMN daily_games.best_streak IS 'All-time best streak';
COMMENT ON COLUMN daily_games.rank IS 'Player rank in daily leaderboard (calculated after completion)';

-- ==========================================
-- Statistics Queries (Examples)
-- ==========================================

-- Top 10 for today:
-- SELECT player_id,
--        (session_state->>'base_score')::int * (CASE WHEN current_streak >= 100 THEN 2.0 ... END) as final_score
-- FROM daily_games
-- WHERE date = '2026-01-25' AND status = 'completed'
-- ORDER BY final_score DESC
-- LIMIT 10;

-- Player's game for today:
-- SELECT * FROM daily_games
-- WHERE player_id = 'user_123' AND date = '2026-01-25';

-- Total players who played today:
-- SELECT COUNT(*) FROM daily_games
-- WHERE date = '2026-01-25' AND status = 'completed';
