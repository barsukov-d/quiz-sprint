# PvP Duel — запрет дублей и UI состояние "Вызов отправлен"

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Запретить повторный вызов одному сопернику; в списке соперников показать кнопку "Вызов отправлен" (disabled) если вызов уже отправлен.

**Architecture:** Thin Client — бэкенд добавляет `hasPendingChallenge: bool` в `RivalItemDTO`, фронт только рендерит. Дублепроверка — в `SendChallengeUseCase` до создания challenge.

**Tech Stack:** Go/Fiber (backend), Vue 3 + Nuxt UI (frontend), swaggo/swag + kubb (codegen)

---

### Task 1: Добавить `ErrChallengeAlreadySent` в domain errors

**Files:**
- Modify: `backend/internal/domain/quick_duel/errors.go`

**Step 1: Добавить ошибку**

В блок `// Challenge errors` добавить строку:

```go
ErrChallengeAlreadySent = errors.New("challenge already sent to this player")
```

После `ErrFriendBusy = errors.New("friend is already in a game")`.

**Step 2: Убедиться что компилируется**

```bash
cd backend && go build ./...
```
Expected: no errors

**Step 3: Commit**

```bash
git add backend/internal/domain/quick_duel/errors.go
git commit -m "feat(pvp-duel): add ErrChallengeAlreadySent domain error"
```

---

### Task 2: Добавить дублепроверку в `SendChallengeUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go:274-316`
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Написать failing test**

В `use_cases_test.go` после `TestSendChallenge_FriendInGame` (строка ~294) добавить:

```go
func TestSendChallenge_DuplicateChallenge(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSendChallengeUC()

	// Send first challenge
	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != nil {
		t.Fatalf("first challenge failed: %v", err)
	}

	// Send second challenge to same friend — must fail
	_, err = uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != quick_duel.ErrChallengeAlreadySent {
		t.Errorf("expected ErrChallengeAlreadySent, got %v", err)
	}
}
```

**Step 2: Запустить тест — убедиться что FAIL**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestSendChallenge_DuplicateChallenge -v
```
Expected: FAIL (тест упадёт, т.к. проверки ещё нет)

**Step 3: Реализовать проверку в `SendChallengeUseCase.Execute`**

В `use_cases.go`, после строки `// Check if friend is already in a game` (строка ~287), добавить блок:

```go
// Check for existing pending challenge to same friend
existingChallenges, err := uc.challengeRepo.FindPendingByChallenger(challengerID)
if err == nil {
    for _, c := range existingChallenges {
        if c.ChallengedID() != nil && c.ChallengedID().Equals(friendID) {
            return SendChallengeOutput{}, quick_duel.ErrChallengeAlreadySent
        }
    }
}
```

**Step 4: Запустить тест — убедиться что PASS**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestSendChallenge -v
```
Expected: все TestSendChallenge_* PASS

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go backend/internal/application/quick_duel/use_cases_test.go
git commit -m "feat(pvp-duel): prevent duplicate challenges in SendChallengeUseCase"
```

---

### Task 3: Добавить `HasPendingChallenge` в `RivalDTO`

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go:139-148`

**Step 1: Добавить поле**

В структуру `RivalDTO` добавить поле после `GamesCount`:

```go
type RivalDTO struct {
    ID                 string `json:"id"`
    Username           string `json:"username"`
    MMR                int    `json:"mmr"`
    League             string `json:"league"`
    LeagueIcon         string `json:"leagueIcon"`
    IsOnline           bool   `json:"isOnline"`
    GamesCount         int    `json:"gamesCount"`
    HasPendingChallenge bool  `json:"hasPendingChallenge"`
}
```

**Step 2: Убедиться что компилируется**

```bash
cd backend && go build ./...
```
Expected: no errors

**Step 3: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go
git commit -m "feat(pvp-duel): add HasPendingChallenge field to RivalDTO"
```

---

### Task 4: Обновить `GetRivalsUseCase` — добавить `challengeRepo` и заполнить `HasPendingChallenge`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go:1628-1694`
- Modify: `backend/internal/application/quick_duel/testutil_test.go:781-785`

**Step 1: Написать failing test**

В `use_cases_test.go` добавить тест (найди секцию GetRivals tests или добавить в конец):

```go
// ========================================
// GetRivals Tests
// ========================================

func TestGetRivals_HasPendingChallenge(t *testing.T) {
    f := setupFixture(t)

    // Create a game between player1 and player2 so player2 appears as rival
    f.startGame(t, testPlayer1ID, testPlayer2ID)

    // Player1 sends challenge to player2
    sendUC := f.newSendChallengeUC()
    _, err := sendUC.Execute(SendChallengeInput{
        PlayerID: testPlayer1ID,
        FriendID: testPlayer2ID,
    })
    if err != nil {
        t.Fatalf("sendChallenge failed: %v", err)
    }

    uc := f.newGetRivalsUC()
    output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(output.Rivals) == 0 {
        t.Fatal("expected at least one rival")
    }

    rival := output.Rivals[0]
    if !rival.HasPendingChallenge {
        t.Error("expected HasPendingChallenge=true for rival with pending challenge")
    }
}

func TestGetRivals_NoPendingChallenge(t *testing.T) {
    f := setupFixture(t)

    f.startGame(t, testPlayer1ID, testPlayer2ID)

    uc := f.newGetRivalsUC()
    output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(output.Rivals) == 0 {
        t.Fatal("expected at least one rival")
    }

    if output.Rivals[0].HasPendingChallenge {
        t.Error("expected HasPendingChallenge=false when no challenge sent")
    }
}
```

**Step 2: Запустить тесты — убедиться что FAIL**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetRivals -v
```
Expected: FAIL или compile error

**Step 3: Обновить `GetRivalsUseCase`**

Изменить структуру и конструктор:

```go
type GetRivalsUseCase struct {
    duelGameRepo     quick_duel.DuelGameRepository
    playerRatingRepo quick_duel.PlayerRatingRepository
    userRepo         domainUser.UserRepository
    onlineTracker    OnlineTracker
    challengeRepo    quick_duel.ChallengeRepository
}

func NewGetRivalsUseCase(
    duelGameRepo quick_duel.DuelGameRepository,
    playerRatingRepo quick_duel.PlayerRatingRepository,
    userRepo domainUser.UserRepository,
    onlineTracker OnlineTracker,
    challengeRepo quick_duel.ChallengeRepository,
) *GetRivalsUseCase {
    return &GetRivalsUseCase{
        duelGameRepo:     duelGameRepo,
        playerRatingRepo: playerRatingRepo,
        userRepo:         userRepo,
        onlineTracker:    onlineTracker,
        challengeRepo:    challengeRepo,
    }
}
```

Обновить `Execute` — перед `rivals := make(...)` загрузить pending challenges и собрать Set challenged IDs:

```go
func (uc *GetRivalsUseCase) Execute(input GetRivalsInput) (GetRivalsOutput, error) {
    playerID, err := shared.NewUserID(input.PlayerID)
    if err != nil {
        return GetRivalsOutput{}, err
    }

    const limit = 20
    opponents, err := uc.duelGameRepo.FindRecentOpponents(playerID, limit)
    if err != nil {
        return GetRivalsOutput{Rivals: []RivalDTO{}}, nil
    }

    // Build set of rival IDs that already have a pending challenge from this player
    pendingChallengedIDs := map[string]bool{}
    if pending, err := uc.challengeRepo.FindPendingByChallenger(playerID); err == nil {
        for _, c := range pending {
            if c.ChallengedID() != nil {
                pendingChallengedIDs[c.ChallengedID().String()] = true
            }
        }
    }

    rivals := make([]RivalDTO, 0, len(opponents))
    for _, opp := range opponents {
        user, err := uc.userRepo.FindByID(opp.OpponentID)
        if err != nil {
            continue
        }

        mmr := 1000
        leagueStr := "bronze"
        leagueIcon := "🥉"
        if rating, err := uc.playerRatingRepo.FindByPlayerID(opp.OpponentID); err == nil && rating != nil {
            mmr = rating.MMR()
            leagueStr = rating.League().String()
            leagueIcon = rating.League().Icon()
        }

        isOnline := false
        if uc.onlineTracker != nil {
            isOnline, _ = uc.onlineTracker.IsOnline(opp.OpponentID.String())
        }

        rivals = append(rivals, RivalDTO{
            ID:                  opp.OpponentID.String(),
            Username:            user.Username().String(),
            MMR:                 mmr,
            League:              leagueStr,
            LeagueIcon:          leagueIcon,
            IsOnline:            isOnline,
            GamesCount:          opp.GamesCount,
            HasPendingChallenge: pendingChallengedIDs[opp.OpponentID.String()],
        })
    }

    return GetRivalsOutput{Rivals: rivals}, nil
}
```

**Step 4: Обновить `testutil_test.go` — `newGetRivalsUC`**

Строка 781-784:
```go
func (f *duelFixture) newGetRivalsUC() *GetRivalsUseCase {
    return NewGetRivalsUseCase(
        f.duelGameRepo, f.playerRatingRepo, f.userRepo, f.onlineTracker, f.challengeRepo,
    )
}
```

**Step 5: Запустить тесты**

```bash
cd backend && go test ./internal/application/quick_duel/... -v
```
Expected: все тесты PASS

**Step 6: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go backend/internal/application/quick_duel/use_cases_test.go backend/internal/application/quick_duel/testutil_test.go
git commit -m "feat(pvp-duel): GetRivalsUseCase returns hasPendingChallenge per rival"
```

---

### Task 5: Обновить `swagger_models.go` и error mapper

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go:1486-1495`
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go` — `mapDuelError`

**Step 1: Добавить поле в `RivalItemDTO`**

Строки 1487-1495 — добавить `HasPendingChallenge`:

```go
type RivalItemDTO struct {
    ID                  string `json:"id" validate:"required"`
    Username            string `json:"username" validate:"required"`
    MMR                 int    `json:"mmr" validate:"required"`
    League              string `json:"league" validate:"required"`
    LeagueIcon          string `json:"leagueIcon" validate:"required"`
    IsOnline            bool   `json:"isOnline" validate:"required"`
    GamesCount          int    `json:"gamesCount" validate:"required"`
    HasPendingChallenge bool   `json:"hasPendingChallenge" validate:"required"`
}
```

**Step 2: Добавить `ErrChallengeAlreadySent` в `mapDuelError`**

В функции `mapDuelError` после `case domainDuel.ErrFriendBusy:` добавить:

```go
case domainDuel.ErrChallengeAlreadySent:
    return fiber.NewError(fiber.StatusConflict, "Challenge already sent to this player")
```

**Step 3: Обновить wire в `routes.go`**

Найти строку `getRivalsUC = appDuel.NewGetRivalsUseCase(` (~строка 474) и добавить `challengeRepo` последним аргументом:

```go
getRivalsUC = appDuel.NewGetRivalsUseCase(
    duelGameRepo,
    playerRatingRepo,
    userRepo,
    duelOnlineTracker,
    challengeRepo,  // добавить
)
```

**Step 4: Убедиться что компилируется**

```bash
cd backend && go build ./...
```
Expected: no errors

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/http/handlers/swagger_models.go backend/internal/infrastructure/http/handlers/duel_handlers.go backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): add hasPendingChallenge to RivalItemDTO swagger, wire challengeRepo, map 409"
```

---

### Task 6: Регенерировать Swagger + TypeScript

**Step 1: Сгенерировать Swagger docs**

```bash
cd backend && make swagger
```
Expected: `swagger.json` обновлён, `RivalItemDTO` содержит `hasPendingChallenge`

Проверить:
```bash
grep "hasPendingChallenge" backend/docs/swagger.json
```
Expected: найдено

**Step 2: Сгенерировать TypeScript типы**

```bash
cd tma && pnpm run generate:all
```
Expected: `tma/src/api/generated/` обновлён

**Step 3: Убедиться что тип содержит поле**

```bash
grep "hasPendingChallenge" tma/src/api/generated/schemas/internalInfrastructureHttpHandlers/rivalItemDTOSchema.ts
```
Expected: найдено

**Step 4: Запустить все тесты бэкенда**

```bash
cd backend && go test ./...
```
Expected: все PASS

**Step 5: Commit**

```bash
git add backend/docs/ tma/src/api/generated/
git commit -m "chore(pvp-duel): regenerate Swagger docs and TypeScript types with hasPendingChallenge"
```

---

### Task 7: Обновить i18n ключи

**Files:**
- Modify: `tma/src/i18n/locales/ru.ts`
- Modify: `tma/src/i18n/locales/en.ts`

**Step 1: Добавить ключ в `ru.ts`**

В блок `duel:` после `challenge: 'Вызов',` добавить:

```typescript
challengeSent: 'Вызов отправлен',
```

**Step 2: Добавить ключ в `en.ts`**

В блок `duel:` после `challenge: 'Challenge',` добавить:

```typescript
challengeSent: 'Challenge Sent',
```

**Step 3: Убедиться что TypeScript не ругается**

```bash
cd tma && pnpm run type-check
```
Expected: no errors

**Step 4: Commit**

```bash
git add tma/src/i18n/locales/ru.ts tma/src/i18n/locales/en.ts
git commit -m "feat(pvp-duel): add challengeSent i18n key"
```

---

### Task 8: Обновить кнопку в `DuelLobbyView.vue`

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue:485`

**Step 1: Найти кнопку**

В шаблоне найти блок `<UButton size="xs" @click="() => handleChallengeFriend(rival.id!)">` (~строка 485).

**Step 2: Заменить кнопку**

```html
<UButton
    size="xs"
    :disabled="rival.hasPendingChallenge"
    :color="rival.hasPendingChallenge ? 'gray' : 'primary'"
    @click="() => !rival.hasPendingChallenge && handleChallengeFriend(rival.id!)"
>
    {{ rival.hasPendingChallenge ? t('duel.challengeSent') : t('duel.challenge') }}
</UButton>
```

**Step 3: Проверить type-check и lint**

```bash
cd tma && pnpm run type-check && pnpm lint
```
Expected: no errors

**Step 4: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): show disabled 'Вызов отправлен' button for rivals with pending challenge"
```

---

### Task 9: Финальная проверка

**Step 1: Запустить все backend тесты**

```bash
cd backend && go test ./...
```
Expected: все PASS

**Step 2: Запустить frontend тесты**

```bash
cd tma && pnpm test:unit
```
Expected: все PASS

**Step 3: Проверить сборку фронта**

```bash
cd tma && pnpm build
```
Expected: build successful, no TypeScript errors

**Step 4: Push**

```bash
git push
```
