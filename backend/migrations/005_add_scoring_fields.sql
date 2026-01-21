-- Migration: Add scoring system fields
-- Date: 2026-01-20
-- Description: Add time bonus, streak bonus, and per-question time limit fields

-- Add new scoring fields to quizzes table
ALTER TABLE quizzes
ADD COLUMN IF NOT EXISTS base_points INTEGER DEFAULT 50 CHECK (base_points >= 0 AND base_points <= 1000),
ADD COLUMN IF NOT EXISTS time_limit_per_question INTEGER DEFAULT 30 CHECK (time_limit_per_question >= 5 AND time_limit_per_question <= 60),
ADD COLUMN IF NOT EXISTS max_time_bonus INTEGER DEFAULT 50 CHECK (max_time_bonus >= 0 AND max_time_bonus <= 1000),
ADD COLUMN IF NOT EXISTS streak_threshold INTEGER DEFAULT 3 CHECK (streak_threshold > 0),
ADD COLUMN IF NOT EXISTS streak_bonus INTEGER DEFAULT 100 CHECK (streak_bonus >= 0 AND streak_bonus <= 1000);

-- Add streak tracking to quiz_sessions table
ALTER TABLE quiz_sessions
ADD COLUMN IF NOT EXISTS correct_answer_streak INTEGER DEFAULT 0 CHECK (correct_answer_streak >= 0);

-- Add detailed points breakdown to user_answers table
ALTER TABLE user_answers
ADD COLUMN IF NOT EXISTS base_points INTEGER DEFAULT 0 CHECK (base_points >= 0),
ADD COLUMN IF NOT EXISTS time_bonus INTEGER DEFAULT 0 CHECK (time_bonus >= 0),
ADD COLUMN IF NOT EXISTS streak_bonus INTEGER DEFAULT 0 CHECK (streak_bonus >= 0),
ADD COLUMN IF NOT EXISTS time_spent BIGINT CHECK (time_spent >= 0);

-- Update existing quizzes with default values
UPDATE quizzes
SET
    base_points = 50,
    time_limit_per_question = 30,
    max_time_bonus = 50,
    streak_threshold = 3,
    streak_bonus = 100
WHERE base_points IS NULL;

-- Update existing user_answers to migrate old points to base_points
UPDATE user_answers
SET base_points = points
WHERE base_points = 0 AND points > 0;

-- Add comment explaining the scoring system
COMMENT ON COLUMN quizzes.base_points IS 'Base points awarded for a correct answer (default per question)';
COMMENT ON COLUMN quizzes.time_limit_per_question IS 'Time limit for each question in seconds';
COMMENT ON COLUMN quizzes.max_time_bonus IS 'Maximum bonus points for answering quickly';
COMMENT ON COLUMN quizzes.streak_threshold IS 'Number of correct answers needed to trigger streak bonus';
COMMENT ON COLUMN quizzes.streak_bonus IS 'Bonus points awarded when streak threshold is reached';

COMMENT ON COLUMN quiz_sessions.correct_answer_streak IS 'Current streak of consecutive correct answers';

COMMENT ON COLUMN user_answers.base_points IS 'Base points earned for this answer';
COMMENT ON COLUMN user_answers.time_bonus IS 'Speed bonus points earned';
COMMENT ON COLUMN user_answers.streak_bonus IS 'Streak bonus points earned';
COMMENT ON COLUMN user_answers.time_spent IS 'Time spent on question in milliseconds';
