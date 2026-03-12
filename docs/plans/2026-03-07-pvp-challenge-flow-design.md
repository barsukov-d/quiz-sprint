# PvP Challenge Flow — Design Document

**Date:** 2026-03-07
**Branch:** pvp-duel
**Status:** Approved

## Context

PvP Duel has two ways to challenge someone:
1. **Link invite** — share Telegram deep link, invitee opens TMA
2. **Rivals list** — challenge a recent opponent from in-app list

Both flows have gaps: missing UI states, backend bugs, and unimplemented features.

## Architecture Principle

**Server-Driven UI (SDUI):** Backend is the single source of truth. Frontend only renders what API returns. No business state in local refs.

---

## Current State & Gaps

### Bugs (backend)

| # | Issue | Location | Risk |
|---|-------|----------|------|
| B1 | `StartChallenge` doesn't check if either player is in active game | `use_cases.go:1607` | Two simultaneous games per player |
| B2 | `AcceptByLinkCode` doesn't check if invitee is in active game | `use_cases.go:526` | Same |
| B3 | `SendChallenge` doesn't check if challenger is in active game | `use_cases.go:282` | Same |
| B4 | `DeleteExpired` ignores `accepted_waiting_inviter` | `challenge_repository.go:144` | Zombie challenges forever |
| B5 | Link code = 8 hex chars of UUID, collision possible | `duel_challenge.go:116` | Wrong challenge matched |

### Missing Features

| # | Feature | Needed For |
|---|---------|-----------|
| F1 | `acceptedChallenges[]` in `/duel/status` | Invitee waiting banner |
| F2 | `inviterName` in `accept-by-code` response | Confirmation modal shows who challenges |
| F3 | `pendingChallengeId` in `RivalDTO` | Cancel button on rivals |
| F4 | `DELETE /duel/challenge/:id` + `CancelChallenge` use case | Cancel from invitee and challenger |
| F5 | `FindAcceptedWaitingForPlayer` repo method | Query for F1 |

### Missing Frontend

| # | Feature | Screen |
|---|---------|--------|
| U1 | Render `outgoingPendingChallenges` cards | DuelLobbyView (inviter) |
| U2 | Waiting banner for invitee after accepting link | DuelLobbyView (invitee) |
| U3 | Inviter name in confirmation modal | DuelLobbyView modal |
| U4 | Cancel button (invitee waiting + rival challenge) | DuelLobbyView |
| U5 | Specific error messages for 409/400 codes | DuelLobbyView deep link |

### Doc Issues

| # | Issue |
|---|-------|
| D1 | `01_concept.md` says "instant start" but link flow is two-step |
| D2 | Invitee ticket cost unclear for link challenges |
| D3 | `PushChallengeExpirySeconds` (300s) unused — direct = always 60s |

### Stubs (not blocking, track separately)

- Tickets hardcoded to 10 everywhere
- `inviteeName` not persisted in DB (works via userRepo re-resolution)
- `FindByLinkCode` uses `LIKE '%'` (no index, OK at current scale)

---

## Design Decisions

### 1. Invitee waiting state: banner in lobby

After invitee accepts link challenge, lobby shows a banner:

```
+-------------------------------------+
|  Vы приняли вызов                   |
|  Ждём пока @username начнёт игру... |
|  (pulsating indicator)              |
|                               [x]   |
+-------------------------------------+
```

- Data source: `acceptedChallenges[]` from `/duel/status`
- [x] = full cancel (`DELETE /duel/challenge/:id`)
- Polling every 5s; when `hasActiveDuel` becomes true, auto-navigate to game

### 2. Cancel = full cancellation

Both invitee and challenger can cancel:
- Invitee: cancel after accepting link challenge (status `accepted_waiting_inviter` -> `cancelled`)
- Challenger: cancel pending challenge to rival

New domain status: `ChallengeStatusCancelled = "cancelled"`

Domain methods:
- `DuelChallenge.CancelByChallenger(cancellerID, cancelledAt)` — works for `pending` and `accepted_waiting_inviter`
- `DuelChallenge.CancelByInvitee(cancellerID, cancelledAt)` — works only for `accepted_waiting_inviter`

### 3. Link code: use 12 chars instead of 8

Change `duel_challenge.go:116`:
```go
// Before: challengeID.String()[:8]  — 16^8 = 4.3B combinations
// After:  challengeID.String()[:12] — 16^12 = 281T combinations
```

Also add `link_code` column to DB for indexed lookup instead of `LIKE '%'`.

### 4. `accepted_waiting_inviter` expiry: 30 minutes

Add to `DeleteExpired` or a separate cleanup:
```sql
UPDATE duel_challenges
SET status = 'expired'
WHERE status = 'accepted_waiting_inviter'
  AND responded_at + 1800 <= $1
```

### 5. Active game guard on all entry points

Add `FindActiveByPlayer` check to:
- `SendChallenge` (for challenger)
- `AcceptByLinkCode` (for invitee)
- `StartChallenge` (for both players)

---

## API Changes

### `GET /duel/status` — add `acceptedChallenges`

```json
{
  "data": {
    "acceptedChallenges": [
      {
        "id": "ch_xyz",
        "challengerId": "user_456",
        "challengerUsername": "@inviter_name",
        "status": "accepted_waiting_inviter",
        "expiresAt": 1706430600,
        "expiresIn": 1500,
        "createdAt": 1706428800
      }
    ]
  }
}
```

### `POST /duel/challenge/accept-by-code` — add `inviterName`

```json
{
  "data": {
    "success": true,
    "challengeId": "ch_xyz",
    "status": "accepted_waiting_inviter",
    "inviterName": "@username"
  }
}
```

### `DELETE /duel/challenge/:challengeId` — new endpoint

Request: `{ "playerId": "user_123" }`

Response 200:
```json
{
  "data": {
    "success": true,
    "ticketRefunded": true
  }
}
```

Errors:
- 403: Not authorized to cancel this challenge
- 404: Challenge not found
- 409: Challenge not in cancellable state

### `RivalDTO` — add `pendingChallengeId`

```json
{
  "id": "user_456",
  "username": "rival",
  "hasPendingChallenge": true,
  "pendingChallengeId": "ch_abc123"
}
```

---

## Frontend Changes

### DuelLobbyView

1. **Outgoing pending cards** (inviter created link, waiting for someone to click):
   ```
   +--------------------------------------+
   |  Ожидаем ответа...                   |
   |  Ссылка истекает через: 23ч 45мин    |
   |                          [Отменить]  |
   +--------------------------------------+
   ```

2. **Accepted challenge banner** (invitee accepted, waiting for inviter):
   ```
   +--------------------------------------+
   |  Вы приняли вызов @username          |
   |  Ждём начала игры...                 |
   |                          [Отменить]  |
   +--------------------------------------+
   ```

3. **Confirmation modal** — show inviter name from `accept-by-code` response

4. **Rival cancel** — "Отправлено" button gets small "x" that calls `DELETE /duel/challenge/:id`

5. **Error mapping** for deep link:
   - `CHALLENGE_EXPIRED` -> "Ссылка устарела. Попроси друга прислать новую"
   - `CHALLENGE_NOT_PENDING` -> "Вызов уже принят другим игроком"
   - `SELF_CHALLENGE` -> "Нельзя вызвать самого себя"
   - `ALREADY_IN_GAME` -> "Вы уже в игре"

### usePvPDuel composable

- Add `acceptedChallenges` computed from status
- Add `cancelChallenge(challengeId)` action
- Invitee polling: when `acceptedChallenges.length > 0`, poll every 5s, auto-navigate on `hasActiveDuel`

---

## Out of Scope

- Ticket system (stays as stub)
- WebSocket for real-time updates (polling is sufficient for now)
- `PushChallengeExpirySeconds` logic (direct challenges stay at 60s)
- `inviteeName` DB persistence (resolved from userRepo)
- `FindByLinkCode` index optimization (OK at current scale, but fixed by `link_code` column)
