# Design: Duel Challenge Action Buttons

**Date:** 2026-03-01
**Branch:** pvp-duel
**Status:** Approved

## Problem

Two cards in `DuelLobbyView` lack action buttons:
- **Outgoing challenge card** — no way to share again or cancel
- **Incoming challenge cards** — buttons exist but are small/inline; no challenger identity shown

## Solution: Hybrid Approach

Minimal backend change (add `challengerUsername` to DTO) + frontend UI improvements for both cards.

---

## Backend Changes

### 1. `ChallengeDTO` — add `challengerUsername`

**File:** `backend/internal/application/quick_duel/dto.go`

```go
type ChallengeDTO struct {
    ID                 string  `json:"id"`
    ChallengerID       string  `json:"challengerId"`
    ChallengedID       *string `json:"challengedId,omitempty"`
    ChallengerUsername string  `json:"challengerUsername,omitempty"`
    Type               string  `json:"type"`
    Status             string  `json:"status"`
    ChallengeLink      string  `json:"challengeLink,omitempty"`
    ExpiresAt          int64   `json:"expiresAt"`
    ExpiresIn          int     `json:"expiresIn"`
    CreatedAt          int64   `json:"createdAt"`
}
```

### 2. `ToChallengeDTO` — add username parameter

**File:** `backend/internal/application/quick_duel/mapper.go`

```go
func ToChallengeDTO(challenge *quick_duel.DuelChallenge, now int64, challengerUsername string) ChallengeDTO {
    // ... existing logic
    return ChallengeDTO{
        // ... existing fields
        ChallengerUsername: challengerUsername,
    }
}
```

### 3. `GetDuelStatusUseCase.Execute` — lookup username before DTO

**File:** `backend/internal/application/quick_duel/use_cases.go`

```go
for _, c := range pendingChallenges {
    username := c.ChallengerID().String()
    if u, err := uc.userRepo.FindByID(c.ChallengerID()); err == nil && u != nil {
        if u.Username().String() != "" {
            username = u.Username().String()
        }
    }
    challengeDTOs = append(challengeDTOs, ToChallengeDTO(c, now, username))
}
```

Outgoing challenges don't need username (challenger = current user).

### 4. `swagger_models.go` — add field to `DuelChallengeDTO`

**File:** `backend/internal/infrastructure/http/handlers/swagger_models.go`

```go
type DuelChallengeDTO struct {
    // ... existing fields
    ChallengerUsername string `json:"challengerUsername,omitempty"`
}
```

---

## Frontend Changes

### Outgoing Challenge Card

**State: Active (not expired)**
```
┌─────────────────────────────────────────┐
│  ✈ Ожидание ответа на вызов   ● ожидание│
│  Ссылка активна ещё: 23ч 40мин          │
│                                         │
│  [📤 Поделиться снова]                  │
└─────────────────────────────────────────┘
```
Button: `color="gray" variant="soft"` → calls `handleShareToTelegram()`

**State: Expired**
```
┌─────────────────────────────────────────┐
│  ✈ Вызов истёк              ○ истекла   │
│  Ссылка активна ещё: истекла            │
│                                         │
│  [📤 Создать новую ссылку]              │
└─────────────────────────────────────────┘
```
Button: `color="primary"` → also calls `handleShareToTelegram()` (creates new link)

### Incoming Challenge Cards

Current: inline small buttons, "Вызов" label without identity.
New: block buttons, challenger name + league shown.

```
┌─────────────────────────────────────────┐
│  ⚡ @Challenger  •  💍 Platinum IV       │
│                                         │
│  [          ПРИНЯТЬ          ]  green   │
│  [         ОТКЛОНИТЬ         ]  red/soft│
└─────────────────────────────────────────┘
```

Fallback: if `challengerUsername` empty → show "Вызов" (current behavior).

### i18n Keys (ru.ts + en.ts)

| Key | RU | EN |
|-----|----|----|
| `shareAgain` | Поделиться снова | Share Again |
| `createNewLink` | Создать новую ссылку | Create New Link |

---

## Data Flow

```
GetDuelStatus (backend)
  pendingChallenges[].challengerUsername ← userRepo.FindByID(challengerID)
  outgoingChallenges[] (no change needed)
        ↓
Swagger → generate:all → TypeScript types
        ↓
DuelLobbyView
  pendingChallenges[i].challengerUsername → display in card
  outgoingChallenges[0] → show share/create button
```

## Files Changed

| File | Change |
|------|--------|
| `backend/internal/application/quick_duel/dto.go` | +`ChallengerUsername` field |
| `backend/internal/application/quick_duel/mapper.go` | +username param in `ToChallengeDTO` |
| `backend/internal/application/quick_duel/use_cases.go` | lookup username before DTO |
| `backend/internal/infrastructure/http/handlers/swagger_models.go` | +`ChallengerUsername` in `DuelChallengeDTO` |
| `tma/src/i18n/locales/ru.ts` | +`shareAgain`, `createNewLink` |
| `tma/src/i18n/locales/en.ts` | +`shareAgain`, `createNewLink` |
| `tma/src/views/Duel/DuelLobbyView.vue` | outgoing + incoming card UI |
| `tma/src/api/generated/` | regenerated after swagger update |

## Out of Scope

- Cancel outgoing challenge (requires new backend endpoint — no API exists)
- Challenger league/MMR in incoming card (would require more fields in DTO)
