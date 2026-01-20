-- ==========================================
-- Users Table Migration
-- ==========================================

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(100) PRIMARY KEY, -- Telegram user ID (string)
    username VARCHAR(100) NOT NULL, -- Display name
    telegram_username VARCHAR(32), -- @username (optional)
    email VARCHAR(255), -- Email (optional)
    avatar_url TEXT, -- Avatar URL (optional)
    language_code VARCHAR(2) NOT NULL DEFAULT 'en', -- ISO 639-1 code
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Indexes
CREATE INDEX idx_users_telegram_username ON users(telegram_username) WHERE telegram_username IS NOT NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_users_is_blocked ON users(is_blocked) WHERE is_blocked = TRUE;

-- Update quiz_sessions to use users table (if needed in future)
-- Note: Currently quiz_sessions.user_id is VARCHAR(100) which matches users.id
-- Add foreign key constraint when ready:
-- ALTER TABLE quiz_sessions ADD CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
