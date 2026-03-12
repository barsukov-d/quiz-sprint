-- Fix stale accepted_waiting_inviter challenges that already have a match_id set.
-- These were created before MarkStarted() was introduced and never had their
-- status updated to accepted after the game was created.
UPDATE duel_challenges
SET status = 'accepted'
WHERE status = 'accepted_waiting_inviter'
  AND match_id IS NOT NULL;
