-- Fix duel_challenges.status column to support longer status values
-- 'accepted_waiting_inviter' is 24 chars, exceeds VARCHAR(20)
ALTER TABLE duel_challenges ALTER COLUMN status TYPE VARCHAR(30);
