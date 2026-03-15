-- Migration: 022_create_inventory_tables.sql
-- Player inventory and transaction ledger tables

-- ========================================
-- User Inventory Table
-- ========================================
CREATE TABLE IF NOT EXISTS user_inventory (
    player_id TEXT PRIMARY KEY REFERENCES users(id),
    coins INT NOT NULL DEFAULT 0,
    pvp_tickets INT NOT NULL DEFAULT 3,
    shield INT NOT NULL DEFAULT 0,
    fifty_fifty INT NOT NULL DEFAULT 0,
    skip INT NOT NULL DEFAULT 0,
    freeze INT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,

    CONSTRAINT chk_inventory_non_negative CHECK (
        coins >= 0 AND
        pvp_tickets >= 0 AND
        shield >= 0 AND
        fifty_fifty >= 0 AND
        skip >= 0 AND
        freeze >= 0
    )
);

-- ========================================
-- User Transactions Table
-- ========================================
CREATE TABLE IF NOT EXISTS user_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL CHECK (type IN ('credit', 'debit')),
    source TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    created_at BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_user_transactions_player_created
    ON user_transactions(player_id, created_at DESC);
