-- Migration: 026_create_marathon_milestone_claims.sql
-- Deduplication table for marathon milestone rewards.
-- Ensures each player receives each milestone reward at most once.

CREATE TABLE IF NOT EXISTS marathon_milestone_claims (
    player_id  TEXT    NOT NULL REFERENCES users(id),
    milestone  INT     NOT NULL,
    claimed_at BIGINT  NOT NULL,
    PRIMARY KEY (player_id, milestone)
);
