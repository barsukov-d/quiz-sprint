CREATE TABLE IF NOT EXISTS player_bonus_wallet (
    player_id VARCHAR(255) PRIMARY KEY REFERENCES users(id),
    bonus_shield INT NOT NULL DEFAULT 0,
    bonus_fifty_fifty INT NOT NULL DEFAULT 0,
    bonus_skip INT NOT NULL DEFAULT 0,
    bonus_freeze INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
