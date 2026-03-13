-- Add link_code column for indexed lookup instead of LIKE on challenge_link
ALTER TABLE duel_challenges
    ADD COLUMN IF NOT EXISTS link_code VARCHAR(16);

-- Backfill: extract code from challenge_link (part after last '_')
UPDATE duel_challenges
SET link_code = SUBSTRING(challenge_link FROM '[^_]+$')
WHERE challenge_link IS NOT NULL AND link_code IS NULL;

-- Unique partial index for fast exact-match lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_duel_challenges_link_code
    ON duel_challenges (link_code) WHERE link_code IS NOT NULL;
