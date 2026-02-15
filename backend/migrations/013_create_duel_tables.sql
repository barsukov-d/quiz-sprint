-- Migration: 013_create_duel_tables.sql
-- PvP Duel mode tables

-- ========================================
-- Seasons Table
-- ========================================
CREATE TABLE IF NOT EXISTS seasons (
    id VARCHAR(20) PRIMARY KEY,  -- Format: "2026-02" (year-month)
    starts_at BIGINT NOT NULL,
    ends_at BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active', 'completed'
    rewards_distributed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

-- Create initial season
INSERT INTO seasons (id, starts_at, ends_at, status)
VALUES (
    to_char(NOW(), 'YYYY-MM'),
    EXTRACT(EPOCH FROM date_trunc('month', NOW())),
    EXTRACT(EPOCH FROM (date_trunc('month', NOW()) + INTERVAL '1 month' - INTERVAL '1 second')),
    'active'
) ON CONFLICT (id) DO NOTHING;

-- ========================================
-- Player Ratings Table
-- ========================================
CREATE TABLE IF NOT EXISTS player_ratings (
    player_id VARCHAR(50) PRIMARY KEY,
    mmr INTEGER NOT NULL DEFAULT 1000,
    league VARCHAR(20) NOT NULL DEFAULT 'bronze',
    division INTEGER NOT NULL DEFAULT 4,
    peak_mmr INTEGER NOT NULL DEFAULT 1000,
    peak_league VARCHAR(20) NOT NULL DEFAULT 'bronze',
    peak_division INTEGER NOT NULL DEFAULT 4,
    games_at_rank INTEGER NOT NULL DEFAULT 0,
    season_id VARCHAR(20) NOT NULL,
    season_wins INTEGER NOT NULL DEFAULT 0,
    season_losses INTEGER NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),

    CONSTRAINT fk_player_ratings_season
        FOREIGN KEY (season_id) REFERENCES seasons(id)
);

CREATE INDEX IF NOT EXISTS idx_player_ratings_mmr ON player_ratings(mmr DESC);
CREATE INDEX IF NOT EXISTS idx_player_ratings_season ON player_ratings(season_id);
CREATE INDEX IF NOT EXISTS idx_player_ratings_league ON player_ratings(league, division);

-- ========================================
-- Duel Matches Table
-- ========================================
CREATE TABLE IF NOT EXISTS duel_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'waiting_start',  -- waiting_start, in_progress, finished, abandoned

    -- Players
    player1_id VARCHAR(50) NOT NULL,
    player2_id VARCHAR(50) NOT NULL,
    winner_id VARCHAR(50),

    -- Scores
    player1_score INTEGER NOT NULL DEFAULT 0,
    player2_score INTEGER NOT NULL DEFAULT 0,
    player1_total_time BIGINT NOT NULL DEFAULT 0,  -- milliseconds
    player2_total_time BIGINT NOT NULL DEFAULT 0,

    -- MMR tracking
    player1_mmr_before INTEGER NOT NULL DEFAULT 1000,
    player2_mmr_before INTEGER NOT NULL DEFAULT 1000,
    player1_mmr_after INTEGER,
    player2_mmr_after INTEGER,

    -- Match metadata
    win_reason VARCHAR(20),  -- 'score', 'time', 'forfeit'
    is_friend_match BOOLEAN NOT NULL DEFAULT FALSE,
    challenge_id UUID,
    current_round INTEGER NOT NULL DEFAULT 0,

    -- Questions and answers (JSONB for flexibility)
    question_ids JSONB NOT NULL DEFAULT '[]',
    round_answers JSONB NOT NULL DEFAULT '{}',

    -- Timestamps
    started_at BIGINT,
    finished_at BIGINT,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX IF NOT EXISTS idx_duel_matches_player1 ON duel_matches(player1_id);
CREATE INDEX IF NOT EXISTS idx_duel_matches_player2 ON duel_matches(player2_id);
CREATE INDEX IF NOT EXISTS idx_duel_matches_status ON duel_matches(status);
CREATE INDEX IF NOT EXISTS idx_duel_matches_created ON duel_matches(created_at DESC);

-- ========================================
-- Duel Challenges Table
-- ========================================
CREATE TABLE IF NOT EXISTS duel_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenger_id VARCHAR(50) NOT NULL,
    challenged_id VARCHAR(50),  -- NULL for link-based challenges
    challenge_type VARCHAR(20) NOT NULL DEFAULT 'direct',  -- 'direct', 'link'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'accepted', 'declined', 'expired'
    challenge_link VARCHAR(100),
    match_id UUID,
    expires_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    responded_at BIGINT,

    CONSTRAINT fk_duel_challenges_match
        FOREIGN KEY (match_id) REFERENCES duel_matches(id)
);

CREATE INDEX IF NOT EXISTS idx_duel_challenges_challenger ON duel_challenges(challenger_id);
CREATE INDEX IF NOT EXISTS idx_duel_challenges_challenged ON duel_challenges(challenged_id);
CREATE INDEX IF NOT EXISTS idx_duel_challenges_status ON duel_challenges(status);
CREATE INDEX IF NOT EXISTS idx_duel_challenges_expires ON duel_challenges(expires_at);
CREATE INDEX IF NOT EXISTS idx_duel_challenges_link ON duel_challenges(challenge_link) WHERE challenge_link IS NOT NULL;

-- ========================================
-- Referrals Table
-- ========================================
CREATE TABLE IF NOT EXISTS referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inviter_id VARCHAR(50) NOT NULL,
    invitee_id VARCHAR(50) NOT NULL,

    -- Milestones
    milestone_registered BOOLEAN NOT NULL DEFAULT TRUE,
    milestone_played_5 BOOLEAN NOT NULL DEFAULT FALSE,
    milestone_reached_silver BOOLEAN NOT NULL DEFAULT FALSE,
    milestone_reached_gold BOOLEAN NOT NULL DEFAULT FALSE,
    milestone_reached_platinum BOOLEAN NOT NULL DEFAULT FALSE,

    -- Claimed rewards (JSONB for flexibility)
    inviter_rewards_claimed JSONB NOT NULL DEFAULT '{}',
    invitee_rewards_claimed JSONB NOT NULL DEFAULT '{}',

    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),

    CONSTRAINT unique_referral UNIQUE (inviter_id, invitee_id),
    CONSTRAINT no_self_referral CHECK (inviter_id != invitee_id)
);

CREATE INDEX IF NOT EXISTS idx_referrals_inviter ON referrals(inviter_id);
CREATE INDEX IF NOT EXISTS idx_referrals_invitee ON referrals(invitee_id);

-- ========================================
-- Ticket Transactions Table (for tracking ticket consumption)
-- ========================================
CREATE TABLE IF NOT EXISTS duel_ticket_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id VARCHAR(50) NOT NULL,
    amount INTEGER NOT NULL,  -- Positive = earned, Negative = spent
    reason VARCHAR(50) NOT NULL,  -- 'duel_entry', 'duel_refund', 'daily_reward', 'referral', 'purchase'
    reference_id VARCHAR(100),  -- Match ID, challenge ID, etc.
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX IF NOT EXISTS idx_ticket_transactions_player ON duel_ticket_transactions(player_id);
CREATE INDEX IF NOT EXISTS idx_ticket_transactions_created ON duel_ticket_transactions(created_at DESC);

-- ========================================
-- Add ticket balance to users table
-- ========================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS duel_tickets INTEGER NOT NULL DEFAULT 10;

-- ========================================
-- Views for common queries
-- ========================================

-- Leaderboard view
CREATE OR REPLACE VIEW v_duel_leaderboard AS
SELECT
    pr.player_id,
    u.username,
    pr.mmr,
    pr.league,
    pr.division,
    pr.season_wins,
    pr.season_losses,
    CASE WHEN (pr.season_wins + pr.season_losses) > 0
         THEN ROUND(pr.season_wins::numeric / (pr.season_wins + pr.season_losses) * 100, 1)
         ELSE 0
    END as win_rate,
    RANK() OVER (ORDER BY pr.mmr DESC) as rank
FROM player_ratings pr
LEFT JOIN users u ON u.id = pr.player_id
WHERE pr.season_id = (SELECT id FROM seasons WHERE status = 'active' ORDER BY starts_at DESC LIMIT 1);

-- Referral leaderboard view
CREATE OR REPLACE VIEW v_referral_leaderboard AS
SELECT
    r.inviter_id as player_id,
    u.username,
    COUNT(*) as total_referrals,
    COUNT(*) FILTER (WHERE r.milestone_played_5 = TRUE) as active_referrals,
    RANK() OVER (ORDER BY COUNT(*) DESC) as rank
FROM referrals r
LEFT JOIN users u ON u.id = r.inviter_id
GROUP BY r.inviter_id, u.username
ORDER BY total_referrals DESC;

-- Match history view
CREATE OR REPLACE VIEW v_match_history AS
SELECT
    m.id,
    m.player1_id,
    m.player2_id,
    m.winner_id,
    m.player1_score,
    m.player2_score,
    m.player1_mmr_before,
    m.player2_mmr_before,
    m.player1_mmr_after,
    m.player2_mmr_after,
    m.is_friend_match,
    m.finished_at,
    u1.username as player1_username,
    u2.username as player2_username
FROM duel_matches m
LEFT JOIN users u1 ON u1.id = m.player1_id
LEFT JOIN users u2 ON u2.id = m.player2_id
WHERE m.status = 'finished'
ORDER BY m.finished_at DESC;
