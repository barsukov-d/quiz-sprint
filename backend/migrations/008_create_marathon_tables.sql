-- ==========================================
-- Marathon Mode Tables Migration
-- ==========================================
-- Creates tables for Solo Marathon game mode:
-- - marathon_games: Active and completed marathon games
-- - marathon_personal_bests: Personal best records per category
--
-- Architecture: Marathon V2 with dynamic question loading
-- See: backend/MARATHON_V2_SUMMARY.md
-- ==========================================

-- ==========================================
-- Marathon Games Table
-- ==========================================
CREATE TABLE IF NOT EXISTS marathon_games (
    id UUID PRIMARY KEY,
    player_id VARCHAR(100) NOT NULL, -- References users(id)
    category_id UUID, -- References categories(id), NULL = all categories
    status VARCHAR(20) NOT NULL, -- "in_progress", "finished", "abandoned"
    started_at BIGINT NOT NULL,
    finished_at BIGINT,

    -- V2 Architecture: Dynamic Question Loading
    current_question_id UUID, -- References questions(id), current question
    answered_question_ids JSONB NOT NULL DEFAULT '[]', -- Array of answered question IDs (all)
    recent_question_ids JSONB NOT NULL DEFAULT '[]', -- Array of recent question IDs (last 20 for exclusion)

    -- Scoring & Progress
    current_streak INT NOT NULL DEFAULT 0,
    max_streak INT NOT NULL DEFAULT 0,
    base_score INT NOT NULL DEFAULT 0, -- Total base score (direct storage, no session)

    -- Lives System
    current_lives INT NOT NULL DEFAULT 3, -- 0-3
    lives_last_update BIGINT NOT NULL, -- Timestamp for regeneration calculation

    -- Hints System
    hints_fifty_fifty INT NOT NULL DEFAULT 3, -- Remove 2 incorrect answers
    hints_extra_time INT NOT NULL DEFAULT 2, -- Add 10 seconds
    hints_skip INT NOT NULL DEFAULT 1, -- Skip question without losing life

    -- Difficulty Progression
    difficulty_level VARCHAR(20) NOT NULL DEFAULT 'beginner', -- beginner, medium, hard, expert, master

    -- Personal Best Context
    personal_best_streak INT, -- Player's personal best when game started (for comparison)

    CONSTRAINT marathon_games_status_check CHECK (status IN ('in_progress', 'finished', 'abandoned')),
    CONSTRAINT marathon_games_difficulty_check CHECK (difficulty_level IN ('beginner', 'medium', 'hard', 'expert', 'master')),
    CONSTRAINT marathon_games_lives_check CHECK (current_lives >= 0 AND current_lives <= 3),
    CONSTRAINT marathon_games_streak_check CHECK (current_streak >= 0 AND max_streak >= current_streak)
);

-- Foreign Keys (if tables exist)
-- Note: Uncomment these when categories and questions tables are created
-- ALTER TABLE marathon_games ADD CONSTRAINT fk_marathon_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
-- ALTER TABLE marathon_games ADD CONSTRAINT fk_marathon_current_question FOREIGN KEY (current_question_id) REFERENCES questions(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_marathon_player_active ON marathon_games(player_id, status) WHERE status = 'in_progress';
CREATE INDEX IF NOT EXISTS idx_marathon_player ON marathon_games(player_id);
CREATE INDEX IF NOT EXISTS idx_marathon_category ON marathon_games(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_marathon_current_question ON marathon_games(current_question_id) WHERE current_question_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_marathon_started_at ON marathon_games(started_at DESC);

-- Comments
COMMENT ON TABLE marathon_games IS 'Marathon game sessions with dynamic question loading';
COMMENT ON COLUMN marathon_games.current_question_id IS 'Currently active question, NULL when no question loaded';
COMMENT ON COLUMN marathon_games.answered_question_ids IS 'JSONB array of all answered question IDs for statistics';
COMMENT ON COLUMN marathon_games.recent_question_ids IS 'JSONB array of last 20 question IDs for exclusion logic';
COMMENT ON COLUMN marathon_games.base_score IS 'Total base score from all correct answers (stored directly, not via session)';

-- ==========================================
-- Marathon Personal Bests Table
-- ==========================================
CREATE TABLE IF NOT EXISTS marathon_personal_bests (
    id UUID PRIMARY KEY,
    player_id VARCHAR(100) NOT NULL, -- References users(id)
    category_id UUID, -- References categories(id), NULL = all categories
    best_streak INT NOT NULL DEFAULT 0, -- Longest streak achieved
    best_score INT NOT NULL DEFAULT 0, -- Highest score achieved
    achieved_at BIGINT NOT NULL, -- When the record was achieved
    updated_at BIGINT NOT NULL, -- Last update timestamp

    CONSTRAINT personal_bests_streak_check CHECK (best_streak >= 0),
    CONSTRAINT personal_bests_score_check CHECK (best_score >= 0),
    UNIQUE (player_id, category_id) -- One record per player per category
);

-- Foreign Keys (if tables exist)
-- ALTER TABLE marathon_personal_bests ADD CONSTRAINT fk_personal_best_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_personal_bests_player ON marathon_personal_bests(player_id);
CREATE INDEX IF NOT EXISTS idx_personal_bests_category ON marathon_personal_bests(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_personal_bests_leaderboard ON marathon_personal_bests(category_id, best_streak DESC, best_score DESC, achieved_at ASC);
CREATE INDEX IF NOT EXISTS idx_personal_bests_global ON marathon_personal_bests(best_streak DESC, best_score DESC, achieved_at ASC) WHERE category_id IS NULL;

-- Comments
COMMENT ON TABLE marathon_personal_bests IS 'Personal best records for marathon mode per category';
COMMENT ON COLUMN marathon_personal_bests.category_id IS 'Category for this record, NULL = all categories';
COMMENT ON COLUMN marathon_personal_bests.best_streak IS 'Longest streak achieved (primary ranking metric)';
COMMENT ON COLUMN marathon_personal_bests.best_score IS 'Highest score achieved (secondary ranking metric)';

-- ==========================================
-- Statistics Queries (Examples)
-- ==========================================

-- Top 10 players in a category (leaderboard):
-- SELECT pb.*, u.username
-- FROM marathon_personal_bests pb
-- JOIN users u ON pb.player_id = u.id
-- WHERE pb.category_id = '...'
-- ORDER BY pb.best_streak DESC, pb.best_score DESC, pb.achieved_at ASC
-- LIMIT 10;

-- Player's active game:
-- SELECT * FROM marathon_games
-- WHERE player_id = '...' AND status = 'in_progress'
-- LIMIT 1;

-- Player's personal bests across all categories:
-- SELECT pb.*, c.name as category_name
-- FROM marathon_personal_bests pb
-- LEFT JOIN categories c ON pb.category_id = c.id
-- WHERE pb.player_id = '...'
-- ORDER BY pb.best_streak DESC;
