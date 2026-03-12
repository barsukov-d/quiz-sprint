# PvP Rivals — Design Doc

**Date:** 2026-03-02
**Branch:** pvp-duel

## Problem

The "Мои вызовы" section in DuelLobbyView shows pending link-challenges ("Ожидаем ответа..."). These are viral invite links sent to people not yet in the app — there is no way to track if they accepted. The section is misleading and adds noise.

The "Вызвать друга" section only has Telegram share and copy-link buttons — no way to challenge someone you already know from the app.

## Solution

1. **Remove** `outgoingPendingChallenges` cards ("Ожидаем ответа...") from DuelLobbyView.
2. **Add** "Соперники" list — recent unique opponents from game history — with a [Вызвать] button per row.
3. Keep "Пригласить в Telegram" as a secondary action below the rivals list (for onboarding new users).
4. Keep `outgoingReadyChallenges` ("✅ Готов к дуэли") — these are actionable and must stay.

## Naming

"Друзья" is inaccurate — these are people you played against, not personal friends.
→ Section label: **"Соперники"**

## Architecture

```
GET /api/v1/duel/rivals
  Auth: TMA middleware
  Logic:
    1. Query duel_games WHERE player1_id=me OR player2_id=me, ORDER BY started_at DESC
    2. Deduplicate opponent IDs, take last 20 unique
    3. For each: username from users, mmr+league from player_ratings
    4. Online status from Redis: duel:online:{playerId}
  Response: { rivals: []RivalDTO }
```

## Backend Changes

### New files
- `backend/internal/application/duel/get_rivals_use_case.go` — use case
- `backend/internal/infrastructure/http/handlers/duel_handlers.go` — add `GetRivals` handler

### Modified files
- `backend/internal/infrastructure/http/handlers/swagger_models.go` — add `RivalDTO`, `GetRivalsResponse`
- Router: register `GET /api/v1/duel/rivals`

### DTO
```go
type RivalDTO struct {
    ID          string `json:"id"`
    Username    string `json:"username"`
    MMR         int    `json:"mmr"`
    League      string `json:"league"`
    LeagueIcon  string `json:"leagueIcon"`
    IsOnline    bool   `json:"isOnline"`
    GamesCount  int    `json:"gamesCount"`  // total games played together
}

type GetRivalsResponse struct {
    Rivals []RivalDTO `json:"rivals"`
}
```

## Frontend Changes

### DuelLobbyView.vue

**Remove:**
- `outgoingPendingChallenges` display block (the "Ожидаем ответа..." cards)
- `outgoingPendingChallenges` from template (keep it in composable for now, just hide from UI)

**Replace** "Вызвать друга" UCard with:

```
[ Мои вызовы ]               ← v-if="outgoingReadyChallenges.length > 0"
  ✅ @Vasya готов!  [Начать дуэль →]

[ Соперники ]
  🟢 @ProGamer  🥇 1720 MMR   [Вызвать]     ← green dot = isOnline
  ⚫ @BestQuiz  🥈 1230 MMR   [Вызвать]
  ─── или ───
  [Пригласить нового в Telegram]

  (empty state if rivals is empty):
  "Сыграй несколько дуэлей — здесь появятся соперники"
  [Пригласить в Telegram]
```

**Add:**
- `useGetDuelRivals` hook (auto-generated after `pnpm run generate:all`)
- Call on `onMounted` alongside `refetchStatus`

### Reuse
- `handleChallengeFriend(rivalId)` — already exists, sends direct challenge

## Edge Cases

| Case | Behaviour |
|------|-----------|
| Rival is in game | Show grey [Вызвать] disabled, badge "В игре" |
| No rivals yet | Empty state with invite button |
| Rival already has pending challenge from me | Disable [Вызвать], show "Ожидает..." |
| Self in rivals (impossible by query, but guard) | Skip |

## Out of Scope

- Online status push updates (polling on mount is enough for MVP)
- Rivals pagination (20 is enough for MVP)
- Referral-based friends list (separate feature)
