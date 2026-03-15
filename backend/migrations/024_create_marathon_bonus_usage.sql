CREATE TABLE IF NOT EXISTS marathon_bonus_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES marathon_games(id) ON DELETE CASCADE,
    bonus_type VARCHAR(20) NOT NULL,
    question_number INT NOT NULL,
    used_at BIGINT NOT NULL
);
CREATE INDEX idx_marathon_bonus_usage_game ON marathon_bonus_usage(game_id);
