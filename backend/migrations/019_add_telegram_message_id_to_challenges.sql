-- Add telegram_message_id to store the bot message ID for direct challenges
-- Needed to edit/delete the notification when challenge status changes
ALTER TABLE duel_challenges
    ADD COLUMN IF NOT EXISTS telegram_message_id BIGINT NULL;
