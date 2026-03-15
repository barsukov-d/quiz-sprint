-- Add streak tracking columns to marathon_games
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS streak_count INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS best_streak INT NOT NULL DEFAULT 0;
ALTER TABLE marathon_games ADD COLUMN IF NOT EXISTS lives_restored INT NOT NULL DEFAULT 0;
