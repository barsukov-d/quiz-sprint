-- Migration: 025_create_marathon_weekly_distribution.sql
-- Tracks which weeks have had marathon rewards distributed (idempotency guard)

CREATE TABLE IF NOT EXISTS marathon_weekly_distribution (
    week_id TEXT PRIMARY KEY,          -- e.g. "2026-W11"
    distributed_at BIGINT NOT NULL     -- Unix timestamp when distribution ran
);
