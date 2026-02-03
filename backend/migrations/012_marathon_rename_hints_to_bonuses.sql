-- ==========================================
-- Marathon: Rename hints → bonuses, add shield/freeze/continue
-- ==========================================
-- Changes:
-- 1. Rename hint columns → bonus columns (with shield + freeze)
-- 2. Add shield_active, continue_count columns
-- 3. Rename scoring columns (current_streak/max_streak/base_score → score/total_questions)
-- 4. Update status constraint (add 'game_over', rename 'finished'→'completed')
-- 5. Update difficulty constraint (remove 'expert')
-- ==========================================

-- Step 1: Add new columns
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS bonus_shield INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS bonus_fifty_fifty INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS bonus_skip INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS bonus_freeze INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS shield_active BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS continue_count INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS score INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS total_questions INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS personal_best_score INT;

-- Step 2: Migrate data from old columns to new
UPDATE marathon_games SET
    bonus_fifty_fifty = hints_fifty_fifty,
    bonus_freeze = hints_extra_time,
    bonus_skip = hints_skip,
    score = max_streak,
    total_questions = max_streak,
    personal_best_score = personal_best_streak
WHERE TRUE;

-- Step 3: Drop old constraints
ALTER TABLE marathon_games DROP CONSTRAINT IF EXISTS marathon_games_status_check;
ALTER TABLE marathon_games DROP CONSTRAINT IF EXISTS marathon_games_difficulty_check;
ALTER TABLE marathon_games DROP CONSTRAINT IF EXISTS marathon_games_streak_check;

-- Step 4: Update status values ('finished' → 'completed')
UPDATE marathon_games SET status = 'completed' WHERE status = 'finished';

-- Step 5: Add new constraints
ALTER TABLE marathon_games ADD CONSTRAINT marathon_games_status_check
    CHECK (status IN ('in_progress', 'game_over', 'completed', 'abandoned'));
ALTER TABLE marathon_games ADD CONSTRAINT marathon_games_difficulty_check
    CHECK (difficulty_level IN ('beginner', 'medium', 'hard', 'master'));

-- Step 6: Update index for active games (include game_over as "active")
DROP INDEX IF EXISTS idx_marathon_player_active;
CREATE INDEX IF NOT EXISTS idx_marathon_player_active ON marathon_games(player_id, status) WHERE status IN ('in_progress', 'game_over');

-- Step 7: Drop old columns (keeping them temporarily for rollback safety)
-- Uncomment these after verifying the migration works:
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS hints_fifty_fifty;
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS hints_extra_time;
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS hints_skip;
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS current_streak;
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS max_streak;
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS base_score;
-- ALTER TABLE marathon_games DROP COLUMN IF EXISTS personal_best_streak;

-- Comments
COMMENT ON COLUMN marathon_games.bonus_shield IS 'Available shield bonuses (protect from wrong answer)';
COMMENT ON COLUMN marathon_games.bonus_fifty_fifty IS 'Available 50/50 bonuses (remove 2 wrong answers)';
COMMENT ON COLUMN marathon_games.bonus_skip IS 'Available skip bonuses (skip question)';
COMMENT ON COLUMN marathon_games.bonus_freeze IS 'Available freeze bonuses (+10 seconds)';
COMMENT ON COLUMN marathon_games.shield_active IS 'Whether shield is currently active for the current question';
COMMENT ON COLUMN marathon_games.continue_count IS 'Number of times player has used continue';
COMMENT ON COLUMN marathon_games.score IS 'Total correct answers (primary score metric)';
COMMENT ON COLUMN marathon_games.total_questions IS 'Total questions attempted';
COMMENT ON COLUMN marathon_games.personal_best_score IS 'Player personal best score when game started';
