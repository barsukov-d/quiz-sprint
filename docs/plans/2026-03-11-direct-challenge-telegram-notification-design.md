# Direct Challenge Telegram Notification — Design

**Date:** 2026-03-11
**Branch:** pvp-duel
**Status:** Approved

## Problem

When Player A sends a direct challenge to Player B, Player B has no way of knowing about it unless they're actively watching the TMA. We need a Telegram bot notification so Player B can react in time.

## Assumptions

- Invitee has previously interacted with the bot (started it) — guaranteed because direct challenges only happen between players who have already played together in the TMA.
- Best-effort notification: failure to send does NOT fail the use case.

## Design

### 1. TTL Change

Direct challenge TTL: **60 seconds → 1 hour**

### 2. `expired` Status

Instead of deleting expired challenges, transition them to `expired` status.

**Status machine (direct challenges):**
```
pending → accepted   (invitee accepted in TMA)
        → declined   (invitee declined in TMA)
        → expired    (1 hour TTL passed)
```

**Cleanup:** background job deletes `expired` challenges after 24 hours.

### 3. Telegram Message

Sent via `sendMessage` Bot API call with `parse_mode=MarkdownV2` and one inline URL button.

**Text:**
```
⚔️ *Вызов на дуэль\!*

*{InviterName}* бросает тебе вызов в Quiz Sprint\.
У тебя есть 1 час чтобы принять\.
```

**Inline keyboard:** one button
`[ ⚔️ Принять вызов ]` → deep link to TMA: `https://t.me/{bot}?startapp=challenge_{challengeId}`

### 4. Message Lifecycle (Edit on Status Change)

Store `telegram_message_id` in the challenge record. Edit the message when status changes:

| Status change | Message updated to |
|---------------|--------------------|
| `accepted`    | `✅ Вызов принят — удачи!` (buttons removed) |
| `declined`    | `❌ Вызов отклонён` (buttons removed) |
| `expired`     | `⏰ Время истекло` (buttons removed) |

### 5. Frontend — `expired` State

**Inviter sees:**
- Label: "⏰ Время истекло"
- Buttons: `[Повторить вызов]` `[Удалить]`

**Invitee sees (if they open TMA via deep link or naturally):**
- Label: "⏰ Вызов устарел"
- Button: `[Удалить]`

If invitee opens expired deep link from Telegram: TMA shows "Вызов устарел — попроси друга отправить новый".

## Implementation Plan

### Backend

1. **Domain:** Add `expired` status to `DuelChallengeStatus`
2. **Domain:** Add `telegram_message_id` field to `DuelChallenge`
3. **DB migration:** Add column `telegram_message_id BIGINT NULL` to challenges table
4. **TTL constant:** `DirectChallengeTTL = 1 * time.Hour`
5. **Notifier interface:** Add two methods:
   - `NotifyChallengeReceived(ctx, inviteeTelegramID int64, inviterName, deepLink string) (messageID int64, error)`
   - `EditChallengeMessage(ctx, inviteeTelegramID int64, messageID int64, text string) error`
6. **HTTPNotifier:** Implement both methods
7. **NoOpNotifier:** Implement both methods (return 0, nil)
8. **SendChallengeUseCase:** Call `NotifyChallengeReceived` after save, store returned `messageID`
9. **Status transitions:** When challenge moves to `accepted`/`declined`/`expired`, call `EditChallengeMessage`
10. **Background job:** Update existing `DeleteExpired` to transition to `expired` status; add separate job to hard-delete `expired` after 24h

### Frontend

11. **`/duel/status` response:** Include `expired` challenges for both inviter and invitee
12. **`DuelLobbyView.vue`:** Render expired state with correct labels and buttons
13. **Deep link handler:** Parse `startapp=challenge_{id}`, handle expired state gracefully

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Invitee blocked bot | `NotifyChallengeReceived` error ignored, challenge still created |
| "Повторить вызов" when another pending exists | Blocked by existing duplicate-pending guard |
| Telegram API timeout | best-effort, challenge still created, `telegram_message_id` stays null |
| `EditChallengeMessage` fails | Ignored (best-effort), stale button remains in Telegram |
| Multiple simultaneous challenges to same invitee | Each creates a separate notification |
