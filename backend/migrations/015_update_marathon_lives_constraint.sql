-- Update marathon_games_lives_check to allow MaxLives=5
ALTER TABLE marathon_games
    DROP CONSTRAINT marathon_games_lives_check,
    ADD CONSTRAINT marathon_games_lives_check CHECK (current_lives >= 0 AND current_lives <= 5);
