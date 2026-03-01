# PvP Scoring Bug Fix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix two bugs in PvP Duel mode: (1) scores shown to wrong player because hub WS-slot order differs from domain player order; (2) dark screen for challenger between navigation and first WS message.

**Architecture:** Bug 1 — add `GetDomainPlayerOrder` to `StartGameUseCase`, use it in `notifyBothPlayersReady` so `game_ready.player1Id` always matches domain player1 (challenger). Bug 2 — add explicit loading/connecting state to `DuelPlayView` when `game === null`.

**Tech Stack:** Go (backend), Vue 3 + TypeScript (frontend), WebSocket, Vitest

---

## Root Cause Summary

### Bug 1: Swapped scores

Two independent orderings of player1/player2 exist:

| Source | player1 |
|--------|---------|
| Domain (`DuelGame` aggregate) | Challenger (set at game creation in `AcceptByLinkCode`) |
| WS Hub (`DuelWebSocketHub.DuelGame`) | First to open WebSocket connection |

Accepter gets the `gameId` immediately from HTTP response → connects to WS first.
Challenger finds out via polling (up to 5s later) → connects second.
Result: Hub.Player1 = accepter = domain.Player2. Scores broadcast in domain order but IDs broadcast in hub order → **every player sees opponent's score as their own**.

### Bug 2: Dark screen

`DuelPlayView` has no loading state when `game === null`. Between navigation to `/duel/play/:id` and receiving the first `game_ready` WS message, all `v-if` conditions are false → blank content area.

---

## Task 1: Add `GetDomainPlayerOrder` to `StartGameUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go` (after `GetRoundQuestion`)
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

### Step 1: Write the failing test

Add to `use_cases_test.go` (after existing StartGame tests, search for `TestStartGame`):

```go
func TestGetDomainPlayerOrder_ReturnsCorrectOrder(t *testing.T) {
	f := setupFixture(t)

	// Start game: player1=testPlayer1ID (challenger), player2=testPlayer2ID (accepter)
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newStartGameUC()
	p1ID, p2ID, err := uc.GetDomainPlayerOrder(gameOutput.GameID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p1ID != testPlayer1ID {
		t.Errorf("Player1ID = %s, want %s", p1ID, testPlayer1ID)
	}
	if p2ID != testPlayer2ID {
		t.Errorf("Player2ID = %s, want %s", p2ID, testPlayer2ID)
	}
}

func TestGetDomainPlayerOrder_GameNotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartGameUC()

	_, _, err := uc.GetDomainPlayerOrder("nonexistent-game-id")
	if err == nil {
		t.Error("expected error for nonexistent game, got nil")
	}
}
```

### Step 2: Run to verify FAIL

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetDomainPlayerOrder -v
```

Expected: `FAIL` — `uc.GetDomainPlayerOrder undefined`

### Step 3: Implement the method

In `backend/internal/application/quick_duel/use_cases.go`, add after `GetRoundQuestion` (around line 904):

```go
// GetDomainPlayerOrder returns the domain's canonical player1ID and player2ID
// for a given game. This is the authoritative order (set at game creation):
// Player1 = challenger, Player2 = accepter.
// Used by the WS hub to send consistent player IDs in game_ready messages.
func (uc *StartGameUseCase) GetDomainPlayerOrder(gameIDStr string) (player1ID, player2ID string, err error) {
	gameID := quick_duel.NewGameIDFromString(gameIDStr)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return "", "", err
	}
	return game.Player1().UserID().String(), game.Player2().UserID().String(), nil
}
```

### Step 4: Run tests to verify PASS

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetDomainPlayerOrder -v
```

Expected: both tests `PASS`

### Step 5: Run full backend test suite

```bash
cd backend && go test ./...
```

Expected: all tests pass, no regressions.

### Step 6: Commit

```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go
git commit -m "feat(pvp-duel): add GetDomainPlayerOrder to StartGameUseCase"
```

---

## Task 2: Fix `notifyBothPlayersReady` to use domain player order

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/duel_websocket_handler.go`

The `notifyBothPlayersReady` function (line ~299) sends `game_ready` with `player1Id: game.Player1ID` (hub WS-slot order). Fix it to use domain order from `startGameUC.GetDomainPlayerOrder`.

### Step 1: Update `notifyBothPlayersReady`

Replace the existing `notifyBothPlayersReady` method with:

```go
func (h *DuelWebSocketHub) notifyBothPlayersReady(game *DuelGame) {
	// Use domain player order so game_ready.player1Id matches answer_result.player1Score.
	// Domain order: player1 = challenger (set at game creation), player2 = accepter.
	// Hub order (WS connection order) may differ — accepter usually connects first.
	domainPlayer1ID := game.Player1ID
	domainPlayer2ID := game.Player2ID
	if h.startGameUC != nil {
		if p1, p2, err := h.startGameUC.GetDomainPlayerOrder(game.ID); err == nil {
			domainPlayer1ID = p1
			domainPlayer2ID = p2
		} else {
			log.Printf("[DuelWS] GetDomainPlayerOrder failed for %s: %v (falling back to hub order)", game.ID, err)
		}
	}

	readyMsg := map[string]interface{}{
		"type": "game_ready",
		"data": map[string]interface{}{
			"gameId":      game.ID,
			"player1Id":   domainPlayer1ID,
			"player2Id":   domainPlayer2ID,
			"startsIn":    3,
			"totalRounds": quick_duel.QuestionsPerDuel,
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(readyMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(readyMsg)
	}

	go func() {
		time.Sleep(3 * time.Second)
		h.startRound(game, 1)
	}()
}
```

### Step 2: Verify build

```bash
cd backend && go build ./...
```

Expected: compiles without errors.

### Step 3: Run full test suite

```bash
cd backend && go test ./...
```

Expected: all tests pass.

### Step 4: Commit

```bash
git add backend/internal/infrastructure/http/handlers/duel_websocket_handler.go
git commit -m "fix(pvp-duel): use domain player order in game_ready to fix swapped scores

Accepter connects to WS before challenger (gets gameId from HTTP immediately,
challenger waits up to 5s via polling). Hub was assigning player1 slot by
connection order while answer_result broadcasts scores in domain order
(challenger=player1). This caused every player to see opponent's score as
their own. Fix: load domain player order from DB in notifyBothPlayersReady."
```

---

## Task 3: Add loading state to `DuelPlayView`

**Files:**
- Modify: `tma/src/views/Duel/DuelPlayView.vue`

### Step 1: Add loading state to template

In `DuelPlayView.vue`, the template has:
```html
<div v-if="isWaiting" ...>        <!-- status = 'waiting' -->
<div v-else-if="isCountdown" ...> <!-- status = 'countdown' -->
<template v-else-if="isPlaying && currentQuestion"> <!-- playing + question ready -->
<div v-else-if="isFinished" ...>  <!-- status = 'finished' -->
```

When `game === null` (initial state before first WS message), **none** of these match → blank content area. Add a connecting/loading state before the first `v-if`:

Replace the block starting with `<!-- Waiting for Opponent -->` through `<!-- Match Finished -->` — specifically, add a new first block:

```html
<!-- Initial connecting / loading state -->
<div v-if="!game" class="flex-1 flex items-center justify-center p-4">
    <div class="text-center">
        <div class="animate-pulse mb-4">
            <UIcon name="i-heroicons-bolt" class="size-16 text-primary" />
        </div>
        <h2 class="text-xl font-bold mb-2">{{ t('duel.connecting') }}</h2>
    </div>
</div>

<!-- Waiting for Opponent -->
<div v-else-if="isWaiting" ...>
```

Note: change `v-if="isWaiting"` to `v-else-if="isWaiting"` since we added a leading `v-if`.

### Step 2: Verify i18n key exists

Check `tma/src/i18n/locales/ru.ts` and `en.ts` — `duel.connecting` should already exist (used in the reconnecting banner). If it doesn't, add:

```ts
// in ru.ts under duel:
connecting: 'Подключение...',

// in en.ts under duel:
connecting: 'Connecting...',
```

### Step 3: Verify in browser

1. Start dev server: `cd tma && pnpm dev`
2. Navigate to `/duel/play/any-id`
3. Should immediately show the bolt icon + "Connecting..." instead of blank screen
4. Yellow banner still shows at top (from `!isConnected` check)

### Step 4: Commit

```bash
git add tma/src/views/Duel/DuelPlayView.vue \
        tma/src/i18n/locales/ru.ts \
        tma/src/i18n/locales/en.ts
git commit -m "fix(pvp-duel): add loading state when game not yet initialized

Challenger is redirected to duel-play via polling (up to 5s delay).
Between navigation and receiving the first game_ready WS message,
game === null and all v-if conditions were false → blank dark screen.
Fix: show connecting state while game is null."
```

---

## Task 4: Write a regression test for score ordering

Add a test that verifies player1 in domain order (challenger) is correctly preserved when the accepter connects to WS first (simulated via the use case layer — we can't easily test the WS handler directly, so we test the invariant at the use case level).

**Files:**
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

### Step 1: Write the test

```go
// TestScoreOrdering_AccepterAnswersFirst verifies that when the accepter (domain player2)
// answers correctly and the challenger (domain player1) answers wrong,
// player1Score reflects domain player1's score (not the accepter's).
// This catches the bug where hub WS-slot order could shadow domain player order.
func TestScoreOrdering_AccepterAnswersFirst(t *testing.T) {
	f := setupFixture(t)

	// Challenger = player1 (domain), Accepter = player2 (domain)
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)
	uc := f.newSubmitDuelAnswerUC()

	// Accepter (domain player2) answers CORRECTLY — simulates accepter connecting to WS first
	out, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer2ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.correctAnswerID(0),
		TimeTaken: 2000,
	})
	if err != nil {
		t.Fatalf("accepter answer error: %v", err)
	}

	// Player2Score (accepter = domain player2) should be > 0
	if out.Player2Score == 0 {
		t.Error("accepter answered correctly: Player2Score should be > 0")
	}
	// Player1Score (challenger = domain player1) should still be 0
	if out.Player1Score != 0 {
		t.Errorf("challenger hasn't answered: Player1Score should be 0, got %d", out.Player1Score)
	}

	// Now challenger (domain player1) answers WRONG
	out2, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer1ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.wrongAnswerID(0),
		TimeTaken: 3000,
	})
	if err != nil {
		t.Fatalf("challenger answer error: %v", err)
	}

	// After both answered: Player1Score = challenger (0, wrong), Player2Score = accepter (> 0, correct)
	if out2.Player1Score != 0 {
		t.Errorf("challenger answered wrong: Player1Score should be 0, got %d", out2.Player1Score)
	}
	if out2.Player2Score == 0 {
		t.Error("accepter answered correctly: Player2Score should be > 0")
	}
}
```

### Step 2: Run test

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestScoreOrdering -v
```

Expected: `PASS` — this documents the invariant that `Player1Score` always = domain player1's score regardless of answer order.

### Step 3: Commit

```bash
git add backend/internal/application/quick_duel/use_cases_test.go
git commit -m "test(pvp-duel): add regression test for score ordering invariant

Documents that Player1Score always reflects domain player1 (challenger),
not WS connection order. This is the invariant that notifyBothPlayersReady
now preserves by using GetDomainPlayerOrder."
```

---

## Task 5: Verify end-to-end fix

### Manual verification checklist

1. Run backend: `cd backend && docker compose -f docker-compose.dev.yml up`
2. Run frontend: `cd tma && pnpm dev`
3. Open tunnel: `cloudflared tunnel run quiz-sprint-dev`
4. Player A (challenger): creates challenge link → shares to Player B
5. Player B (accepter): clicks link → accepts → navigates to play view immediately
6. Player A: sees loading state in play view (bolt icon) until WS connects → sees countdown
7. Both answer question 1:
   - Player A (challenger = domain player1) answers correctly
   - Player B (accepter = domain player2) answers wrong
8. **Expected**: Player A sees their own score increase, Player B sees 0
9. **Previously**: Player A saw 0 (opponent's score shown as theirs)

### Final test run

```bash
cd backend && go test ./...
cd tma && pnpm test:unit
```

---

## Summary of Changes

| File | Change |
|------|--------|
| `backend/internal/application/quick_duel/use_cases.go` | Add `GetDomainPlayerOrder` method |
| `backend/internal/application/quick_duel/use_cases_test.go` | Add 3 new tests |
| `backend/internal/infrastructure/http/handlers/duel_websocket_handler.go` | Fix `notifyBothPlayersReady` to use domain order |
| `tma/src/views/Duel/DuelPlayView.vue` | Add `v-if="!game"` loading state |
| `tma/src/i18n/locales/ru.ts` / `en.ts` | Add `connecting` key if missing |
