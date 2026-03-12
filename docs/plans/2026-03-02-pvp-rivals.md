# PvP Rivals Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `GET /api/v1/duel/rivals` endpoint returning recent unique opponents, and update DuelLobbyView to show a "Соперники" list instead of "Ожидаем ответа..." pending challenge cards.

**Architecture:** Add `FindRecentOpponents` to the DuelGameRepository interface, implement it via a single SQL GROUP BY query, wire a new `GetRivalsUseCase` through the handler/router, generate TypeScript types, then update the Vue composable and view.

**Tech Stack:** Go/Fiber backend (DDD), PostgreSQL (`duel_matches` table), Redis (`OnlineTracker`), Vue 3 + TanStack Query frontend, swaggo/swag + kubb for type generation.

---

### Task 1: Add `FindRecentOpponents` to the domain repository interface

**Files:**
- Modify: `backend/internal/domain/quick_duel/repository.go`

**Step 1: Add `RecentOpponentEntry` struct and method to the interface**

Add after the existing `DuelGameRepository` interface (after `Delete`):

```go
// RecentOpponentEntry represents a recent opponent with game count
type RecentOpponentEntry struct {
	OpponentID   UserID
	GamesCount   int
	LastPlayedAt int64
}
```

And add the method to `DuelGameRepository`:

```go
// FindRecentOpponents returns unique opponents from completed games, most recent first
FindRecentOpponents(playerID UserID, limit int) ([]RecentOpponentEntry, error)
```

**Step 2: Run tests to confirm interface break**

```bash
cd backend && go build ./...
```

Expected: compile error — `DuelGameRepository` not satisfied in `mockDuelGameRepo` and `postgres.DuelGameRepository`.

**Step 3: Commit**

```bash
git add backend/internal/domain/quick_duel/repository.go
git commit -m "feat(pvp-duel): add FindRecentOpponents to DuelGameRepository interface"
```

---

### Task 2: Implement `FindRecentOpponents` in the postgres repository

**Files:**
- Modify: `backend/internal/infrastructure/persistence/postgres/duel_game_repository.go`

**Step 1: Add the method**

```go
func (r *DuelGameRepository) FindRecentOpponents(playerID quick_duel.UserID, limit int) ([]quick_duel.RecentOpponentEntry, error) {
	query := `
		SELECT
			CASE WHEN player1_id = $1 THEN player2_id ELSE player1_id END AS opponent_id,
			COUNT(*) AS games_count,
			EXTRACT(EPOCH FROM MAX(COALESCE(finished_at, started_at)))::bigint AS last_played_at
		FROM duel_matches
		WHERE (player1_id = $1 OR player2_id = $1)
		  AND status = 'completed'
		GROUP BY opponent_id
		ORDER BY last_played_at DESC NULLS LAST
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []quick_duel.RecentOpponentEntry
	for rows.Next() {
		var opponentIDStr string
		var gamesCount int
		var lastPlayedAt int64
		if err := rows.Scan(&opponentIDStr, &gamesCount, &lastPlayedAt); err != nil {
			return nil, err
		}
		opponentID, err := shared.NewUserID(opponentIDStr)
		if err != nil {
			continue
		}
		result = append(result, quick_duel.RecentOpponentEntry{
			OpponentID:   opponentID,
			GamesCount:   gamesCount,
			LastPlayedAt: lastPlayedAt,
		})
	}
	return result, rows.Err()
}
```

**Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/infrastructure/persistence/postgres/...
```

Expected: success.

**Step 3: Commit**

```bash
git add backend/internal/infrastructure/persistence/postgres/duel_game_repository.go
git commit -m "feat(pvp-duel): implement FindRecentOpponents in postgres DuelGameRepository"
```

---

### Task 3: Add `FindRecentOpponents` to the mock repository in tests

**Files:**
- Modify: `backend/internal/application/quick_duel/testutil_test.go`

**Step 1: Add method to `mockDuelGameRepo`**

Add after the existing `Delete` method:

```go
func (m *mockDuelGameRepo) FindRecentOpponents(playerID quick_duel.UserID, limit int) ([]quick_duel.RecentOpponentEntry, error) {
	seen := make(map[string]int)      // opponentID -> count
	lastPlayed := make(map[string]int64) // opponentID -> lastPlayedAt

	for _, g := range m.games {
		if g.Status() != quick_duel.GameStatusCompleted {
			continue
		}
		var opponentID string
		if g.Player1().UserID().Equals(playerID) {
			opponentID = g.Player2().UserID().String()
		} else if g.Player2().UserID().Equals(playerID) {
			opponentID = g.Player1().UserID().String()
		} else {
			continue
		}
		seen[opponentID]++
		if g.StartedAt() > lastPlayed[opponentID] {
			lastPlayed[opponentID] = g.StartedAt()
		}
	}

	var result []quick_duel.RecentOpponentEntry
	for idStr, count := range seen {
		id, err := shared.NewUserID(idStr)
		if err != nil {
			continue
		}
		result = append(result, quick_duel.RecentOpponentEntry{
			OpponentID:   id,
			GamesCount:   count,
			LastPlayedAt: lastPlayed[idStr],
		})
	}

	// Sort by LastPlayedAt desc
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].LastPlayedAt > result[i].LastPlayedAt {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
```

**Step 2: Verify tests compile**

```bash
cd backend && go test ./internal/application/quick_duel/... 2>&1 | head -20
```

Expected: tests pass (no new failures).

**Step 3: Commit**

```bash
git add backend/internal/application/quick_duel/testutil_test.go
git commit -m "test(pvp-duel): add FindRecentOpponents to mockDuelGameRepo"
```

---

### Task 4: Add `RivalDTO` + `GetRivalsInput/Output` to dto.go

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go`

**Step 1: Add DTO types**

Add after `FriendDTO` (around line 137):

```go
// RivalDTO represents a recent opponent
type RivalDTO struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	MMR         int    `json:"mmr"`
	League      string `json:"league"`
	LeagueIcon  string `json:"leagueIcon"`
	IsOnline    bool   `json:"isOnline"`
	GamesCount  int    `json:"gamesCount"`
}
```

Add at the end of the file:

```go
// ========================================
// GetRivals Use Case
// ========================================

type GetRivalsInput struct {
	PlayerID string `json:"playerId"`
}

type GetRivalsOutput struct {
	Rivals []RivalDTO `json:"rivals"`
}
```

**Step 2: Verify**

```bash
cd backend && go build ./internal/application/quick_duel/...
```

Expected: success.

**Step 3: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go
git commit -m "feat(pvp-duel): add RivalDTO and GetRivals DTOs"
```

---

### Task 5: Write failing test for `GetRivalsUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/testutil_test.go`
- Modify: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Add fixture constructor to testutil_test.go**

Add at the end of the fixture constructors section:

```go
func (f *duelFixture) newGetRivalsUC() *GetRivalsUseCase {
	return NewGetRivalsUseCase(
		f.duelGameRepo, f.playerRatingRepo, f.userRepo, f.onlineTracker,
	)
}
```

**Step 2: Add test to use_cases_test.go**

```go
// ========================================
// GetRivals Tests
// ========================================

func TestGetRivals_EmptyWhenNoGames(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetRivalsUC()

	output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Rivals) != 0 {
		t.Errorf("Rivals = %d, want 0", len(output.Rivals))
	}
}

func TestGetRivals_ReturnsOpponentsFromCompletedGames(t *testing.T) {
	f := setupFixture(t)

	// Play a game between player1 and player2
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	// Force-complete the game by submitting all answers
	submitUC := f.newSubmitDuelAnswerUC()
	for i := 0; i < 7; i++ {
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer1ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  3000,
		})
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer2ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  4000,
		})
	}

	uc := f.newGetRivalsUC()
	output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Rivals) != 1 {
		t.Fatalf("Rivals = %d, want 1", len(output.Rivals))
	}
	if output.Rivals[0].ID != testPlayer2ID {
		t.Errorf("Rivals[0].ID = %s, want %s", output.Rivals[0].ID, testPlayer2ID)
	}
	if output.Rivals[0].GamesCount != 1 {
		t.Errorf("Rivals[0].GamesCount = %d, want 1", output.Rivals[0].GamesCount)
	}
}
```

**Step 3: Run tests to confirm they fail**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetRivals -v 2>&1
```

Expected: compile error — `GetRivalsUseCase` undefined.

---

### Task 6: Implement `GetRivalsUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go`

**Step 1: Add use case at end of file**

```go
// ========================================
// GetRivals Use Case
// ========================================

type GetRivalsUseCase struct {
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	userRepo         domainUser.UserRepository
	onlineTracker    OnlineTracker // may be nil
}

func NewGetRivalsUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	userRepo domainUser.UserRepository,
	onlineTracker OnlineTracker,
) *GetRivalsUseCase {
	return &GetRivalsUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		userRepo:         userRepo,
		onlineTracker:    onlineTracker,
	}
}

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
			ID:         opp.OpponentID.String(),
			Username:   user.Username().String(),
			MMR:        mmr,
			League:     leagueStr,
			LeagueIcon: leagueIcon,
			IsOnline:   isOnline,
			GamesCount: opp.GamesCount,
		})
	}

	return GetRivalsOutput{Rivals: rivals}, nil
}
```

**Step 2: Run tests**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetRivals -v 2>&1
```

Expected: both tests PASS.

**Step 3: Run full test suite**

```bash
cd backend && go test ./... 2>&1
```

Expected: all pass.

**Step 4: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/application/quick_duel/testutil_test.go
git commit -m "feat(pvp-duel): add GetRivalsUseCase with tests"
```

---

### Task 7: Add Swagger model, HTTP handler, and route

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- Modify: `backend/internal/infrastructure/http/routes/routes.go`

**Step 1: Add Swagger model to swagger_models.go**

Add at end of file (follow existing pattern):

```go
// RivalItemDTO represents a rival in Swagger docs
type RivalItemDTO struct {
	ID         string `json:"id" validate:"required"`
	Username   string `json:"username" validate:"required"`
	MMR        int    `json:"mmr" validate:"required"`
	League     string `json:"league" validate:"required"`
	LeagueIcon string `json:"leagueIcon" validate:"required"`
	IsOnline   bool   `json:"isOnline" validate:"required"`
	GamesCount int    `json:"gamesCount" validate:"required"`
}

// @name RivalItemDTO

// GetRivalsResponse is the Swagger response model for GET /duel/rivals
type GetRivalsResponse struct {
	Rivals []RivalItemDTO `json:"rivals" validate:"required"`
}

// @name GetRivalsResponse
```

**Step 2: Add `getRivalsUC` field and method to DuelHandler**

In `duel_handlers.go`, add to the `DuelHandler` struct:

```go
getRivalsUC *appDuel.GetRivalsUseCase
```

Add to `NewDuelHandler` parameter list:

```go
getRivalsUC *appDuel.GetRivalsUseCase,
```

Add to the `return &DuelHandler{...}`:

```go
getRivalsUC: getRivalsUC,
```

Add the handler method:

```go
// GetRivals handles GET /api/v1/duel/rivals
// @Summary Get recent rivals
// @Description Get list of recent unique opponents the player has faced
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetRivalsResponse "Rivals list"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/rivals [get]
func (h *DuelHandler) GetRivals(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.getRivalsUC.Execute(appDuel.GetRivalsInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}
```

**Step 3: Wire use case in routes.go**

Find the `var` block for duel use cases (around line 356) and add:

```go
getRivalsUC *appDuel.GetRivalsUseCase
```

In the `if duelGameRepo != nil && ...` block, add after other use case instantiations:

```go
getRivalsUC = appDuel.NewGetRivalsUseCase(
    duelGameRepo,
    playerRatingRepo,
    userRepo,
    nil, // onlineTracker — wired below if Redis available
)
```

Then find where `redisClient` is successfully created (the `else` branch after Redis connection):

```go
} else {
    log.Println("✅ Connected to Redis for matchmaking queue and round-answer cache")
    matchmakingQueue = redisStore.NewMatchmakingQueue(redisClient)
    duelRoundCache = redisStore.NewDuelRoundCache(redisClient)
    // Add this:
    onlineTracker := redisStore.NewOnlineTracker(redisClient)
    if getRivalsUC != nil {
        getRivalsUC = appDuel.NewGetRivalsUseCase(
            duelGameRepo,
            playerRatingRepo,
            userRepo,
            onlineTracker,
        )
    }
}
```

Find `NewDuelHandler(` call and add `getRivalsUC` as the last parameter.

Register the route in the `duel` group (after line 675):

```go
duel.Get("/rivals", duelHandler.GetRivals)
```

**Step 4: Build and verify**

```bash
cd backend && go build ./...
```

Expected: success.

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/http/handlers/swagger_models.go \
        backend/internal/infrastructure/http/handlers/duel_handlers.go \
        backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): add GET /duel/rivals handler and route"
```

---

### Task 8: Generate Swagger docs and TypeScript types

**Step 1: Generate Swagger**

```bash
cd backend && make swagger
```

Expected: `docs/swagger.json` updated with `/duel/rivals` endpoint.

**Step 2: Generate TypeScript types**

```bash
cd tma && pnpm run generate:all
```

Expected: new hook `useGetDuelRivals` appears in `tma/src/api/generated/`.

**Step 3: Verify generated hook exists**

```bash
grep -r "useGetDuelRivals\|getDuelRivals" tma/src/api/generated/ | head -5
```

Expected: matches found.

**Step 4: Commit**

```bash
git add backend/docs/ tma/src/api/generated/
git commit -m "chore(pvp-duel): regenerate Swagger docs and TypeScript types for rivals endpoint"
```

---

### Task 9: Update `usePvPDuel.ts` — add rivals

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts`

**Step 1: Import the new hook**

Add to the import from `@/api/generated`:

```typescript
useGetDuelRivals,
```

**Step 2: Add rivals query inside `usePvPDuel`**

After the `historyData` query block (around line 110):

```typescript
// Rivals (recent opponents)
const { data: rivalsData, refetch: refetchRivals } = useGetDuelRivals(
    computed(() => ({ playerId })),
    {
        query: {
            enabled: computed(() => !!playerId),
            staleTime: 30000,
        },
    },
)
```

**Step 3: Add computed `rivals`**

After `gameHistory` computed (near other computeds):

```typescript
const rivals = computed(() => rivalsData.value?.data?.rivals ?? [])
```

**Step 4: Add `refetchRivals` to `onMounted` call in the composable return**

Add `refetchRivals` to the returned object.

**Step 5: Verify TypeScript**

```bash
cd tma && pnpm run type-check
```

Expected: no errors.

**Step 6: Commit**

```bash
git add tma/src/composables/usePvPDuel.ts
git commit -m "feat(pvp-duel): add rivals query to usePvPDuel composable"
```

---

### Task 10: Update `DuelLobbyView.vue` — replace "Ожидаем ответа" with "Соперники"

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Remove `outgoingPendingChallenges` import from composable destructure**

In the `usePvPDuel(...)` destructure, remove `outgoingPendingChallenges` and add `rivals` and `refetchRivals`.

**Step 2: Add `refetchRivals` call to `onMounted`**

```typescript
onMounted(async () => {
    await refetchStatus()
    await refetchLeaderboard()
    await refetchHistory()
    await refetchRivals()  // add this
    ...
})
```

**Step 3: In the template, remove the `outgoingPendingChallenges` block**

Remove this entire block (lines ~481–495):

```html
<!-- Waiting for response -->
<UCard v-for="challenge in outgoingPendingChallenges" :key="challenge.id">
    ...
</UCard>
```

Also update the wrapping `v-if` condition from:
```html
v-if="outgoingReadyChallenges.length > 0 || outgoingPendingChallenges.length > 0"
```
to:
```html
v-if="outgoingReadyChallenges.length > 0"
```

**Step 4: Replace the "Invite Friend" UCard with "Соперники" section**

Replace the entire `<!-- Invite Friend -->` UCard (lines ~499–539):

```html
<!-- Rivals Section -->
<UCard>
    <h3 class="font-semibold mb-3">{{ t('duel.rivals') }}</h3>

    <!-- Rivals list -->
    <div v-if="rivals.length > 0" class="space-y-2 mb-4">
        <div
            v-for="rival in rivals"
            :key="rival.id"
            class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded"
        >
            <div class="flex items-center gap-2">
                <div
                    class="w-2 h-2 rounded-full"
                    :class="rival.isOnline ? 'bg-green-500' : 'bg-gray-400'"
                />
                <div>
                    <span class="font-medium">{{ rival.username }}</span>
                    <span class="text-xs text-gray-500 ml-2">
                        {{ rival.leagueIcon }} {{ rival.mmr }} MMR
                    </span>
                </div>
            </div>
            <UButton
                size="xs"
                @click="() => handleChallengeFriend(rival.id)"
            >
                {{ t('duel.challenge') }}
            </UButton>
        </div>
    </div>

    <!-- Empty state -->
    <p v-else class="text-sm text-gray-500 dark:text-gray-400 mb-4">
        {{ t('duel.noRivalsYet') }}
    </p>

    <!-- Divider -->
    <div class="flex items-center gap-3 my-3">
        <div class="flex-1 h-px bg-gray-200 dark:bg-gray-700" />
        <span class="text-xs text-gray-400">{{ t('duel.orInvite') }}</span>
        <div class="flex-1 h-px bg-gray-200 dark:bg-gray-700" />
    </div>

    <!-- Invite via Telegram -->
    <UButton
        icon="i-heroicons-paper-airplane"
        color="primary"
        block
        :loading="isSharing"
        @click="handleShareToTelegram"
    >
        {{ t('duel.inviteFriend') }}
    </UButton>
</UCard>
```

**Step 5: Also remove the old "Friends Online" card** (lines ~542–566) since "Соперники" replaces it:

```html
<!-- Friends Online -->
<UCard v-if="friendsOnline.length > 0">
    ...
</UCard>
```

**Step 6: Add i18n keys**

Find the locale file(s) and add:
- `duel.rivals` → `"Соперники"` / `"Rivals"`
- `duel.noRivalsYet` → `"Сыграй несколько дуэлей — здесь появятся соперники"` / `"Play a few duels and your rivals will appear here"`
- `duel.orInvite` → `"или"` / `"or"`

**Step 7: Verify TypeScript**

```bash
cd tma && pnpm run type-check
```

Expected: no errors.

**Step 8: Lint**

```bash
cd tma && pnpm lint
```

Expected: no errors.

**Step 9: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue tma/src/locales/
git commit -m "feat(pvp-duel): replace pending challenges with Соперники list in DuelLobbyView"
```

---

### Task 11: Manual smoke test

**Step 1: Start backend**

```bash
cd backend && docker compose -f docker-compose.dev.yml up
```

**Step 2: Start frontend + tunnel**

```bash
cd tma && pnpm dev
# In another terminal:
cloudflared tunnel run quiz-sprint-dev
```

**Step 3: Open app in Telegram**

Open `https://dev.quiz-sprint-tma.online` → PvP Дуэль tab.

**Expected:**
- No "Ожидаем ответа..." cards visible
- "Соперники" section appears
- If no games played: empty state message + "Пригласить в Telegram" button
- After playing a duel: opponent appears in Соперники with [Вызвать] button

**Step 4: Final commit if any fixes needed, then push**

```bash
git push origin pvp-duel
```
