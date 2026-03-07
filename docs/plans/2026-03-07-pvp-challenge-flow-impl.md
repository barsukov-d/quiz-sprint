# PvP Challenge Flow — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 4 backend bugs, add invitee waiting flow, cancel challenge feature, and UI polish for the PvP challenge system.

**Architecture:** Server-Driven UI — backend is single source of truth. Frontend renders what API returns, no business logic on client. Phase 1 delivers active-game guards + invitee waiting banner. Phase 2 adds full cancel flow + link code hardening.

**Tech Stack:** Go / Fiber v3 (backend), Vue 3 / Nuxt UI / TanStack Query (frontend), PostgreSQL (persistence).

**Design Doc:** `docs/plans/2026-03-07-pvp-challenge-flow-design.md`

---

## Phase 1: Active-game guards (B1–B4) + Invitee waiting flow (F1, F2, F5) + Frontend (U1–U3)

---

### Task 1: B3 — Active-game guard in SendChallenge

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go` (~line 282)
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Write the failing test**

In `use_cases_test.go`, add after existing SendChallenge tests:

```go
func TestSendChallenge_FailsIfChallengerInGame(t *testing.T) {
	f := setupFixture(t)
	f.startGame(t, testPlayer1ID, testPlayer2ID) // player1 is now in active game

	uc := f.newSendChallengeUC()
	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer3ID,
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestSendChallenge_FailsIfChallengerInGame -v
```
Expected: FAIL — no guard exists yet.

**Step 3: Add guard in SendChallengeUseCase.Execute**

In `use_cases.go`, after parsing `challengerID` and before the friend-busy check (~line 295):

```go
	// B3: Check if challenger is already in a game
	if activeGame, err := uc.duelGameRepo.FindActiveByPlayer(challengerID); err == nil && activeGame != nil {
		return SendChallengeOutput{}, quick_duel.ErrAlreadyInGame
	}
```

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestSendChallenge -v
```
Expected: all SendChallenge tests PASS.

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go
git commit -m "fix(pvp-duel): B3 — active game guard in SendChallenge"
```

---

### Task 2: B2 — Active-game guard in AcceptByLinkCode

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go` (AcceptByLinkCodeUseCase struct + constructor + Execute, ~line 505)
- Modify: `backend/internal/application/quick_duel/testutil_test.go` (newAcceptByLinkCodeUC)
- Modify: `backend/internal/infrastructure/http/routes/routes.go` (~line 409)
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Write the failing test**

In `use_cases_test.go`:

```go
func TestAcceptByLinkCode_FailsIfInviteeInGame(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Create a link challenge from player1
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)

	// player2 is already in an active game
	f.startGame(t, testPlayer2ID, testPlayer3ID)

	uc := f.newAcceptByLinkCodeUC()
	_, err := uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: challenge.ChallengeLink(),
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestAcceptByLinkCode_FailsIfInviteeInGame -v
```
Expected: FAIL — struct doesn't have duelGameRepo yet.

**Step 3: Add duelGameRepo to AcceptByLinkCodeUseCase**

In `use_cases.go`, update the struct and constructor:

```go
type AcceptByLinkCodeUseCase struct {
	challengeRepo quick_duel.ChallengeRepository
	duelGameRepo  quick_duel.DuelGameRepository // B2: added
	userRepo      domainUser.UserRepository
	notifier      TelegramNotifier
	eventBus      EventBus
}

func NewAcceptByLinkCodeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	duelGameRepo quick_duel.DuelGameRepository, // B2: added
	userRepo domainUser.UserRepository,
	notifier TelegramNotifier,
	eventBus EventBus,
) *AcceptByLinkCodeUseCase {
	return &AcceptByLinkCodeUseCase{
		challengeRepo: challengeRepo,
		duelGameRepo:  duelGameRepo,
		userRepo:      userRepo,
		notifier:      notifier,
		eventBus:      eventBus,
	}
}
```

Add guard in `Execute`, after `accepterID` is parsed and idempotency check is done (~line 549):

```go
	// B2: Check if accepter is already in a game
	if activeGame, err := uc.duelGameRepo.FindActiveByPlayer(accepterID); err == nil && activeGame != nil {
		return AcceptByLinkCodeOutput{}, quick_duel.ErrAlreadyInGame
	}
```

**Step 4: Fix testutil_test.go — update newAcceptByLinkCodeUC**

In `testutil_test.go`, update the constructor call:

```go
func (f *duelFixture) newAcceptByLinkCodeUC() *AcceptByLinkCodeUseCase {
	return NewAcceptByLinkCodeUseCase(
		f.challengeRepo, f.duelGameRepo, f.userRepo, &noOpNotifier{}, f.eventBus,
	)
}
```

**Step 5: Fix routes.go — add duelGameRepo**

In `routes.go` ~line 409, add `duelGameRepo` as the second arg:

```go
	acceptByLinkCodeUC = appDuel.NewAcceptByLinkCodeUseCase(
		challengeRepo,
		duelGameRepo, // B2: added
		userRepo,
		telegramNotifier,
		duelEventBus,
	)
```

**Step 6: Run tests**

```bash
cd backend && go test ./... -v 2>&1 | tail -20
```
Expected: all tests PASS.

**Step 7: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/testutil_test.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/infrastructure/http/routes/routes.go
git commit -m "fix(pvp-duel): B2 — active game guard in AcceptByLinkCode"
```

---

### Task 3: B1 — Active-game guard in StartChallenge

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go` (~line 1607)
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Write the failing test**

```go
func TestStartChallenge_FailsIfInviterInGame(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Create link challenge and have invitee accept
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer2ID), "Player2", now+10)
	f.challengeRepo.Save(challenge)

	// inviter (player1) joins another game
	f.startGame(t, testPlayer1ID, testPlayer3ID)

	uc := f.newStartChallengeUC()
	_, err := uc.Execute(StartChallengeInput{
		PlayerID:    testPlayer1ID,
		ChallengeID: challenge.ID().String(),
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}

func TestStartChallenge_FailsIfInviteeInGame(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer2ID), "Player2", now+10)
	f.challengeRepo.Save(challenge)

	// invitee (player2) joins another game
	f.startGame(t, testPlayer2ID, testPlayer3ID)

	uc := f.newStartChallengeUC()
	_, err := uc.Execute(StartChallengeInput{
		PlayerID:    testPlayer1ID,
		ChallengeID: challenge.ID().String(),
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}
```

Also add `newStartChallengeUC` helper to `testutil_test.go` if not present:

```go
func (f *duelFixture) newStartChallengeUC() *StartChallengeUseCase {
	return NewStartChallengeUseCase(
		f.challengeRepo, f.duelGameRepo, f.playerRatingRepo,
		f.seasonRepo, f.questionRepo, f.userRepo, f.eventBus,
	)
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestStartChallenge_FailsIf -v
```
Expected: FAIL.

**Step 3: Add guards in StartChallengeUseCase.Execute**

In `use_cases.go`, after `accepterID := *challenge.ChallengedID()` (~line 1635):

```go
	// B1: Guard — inviter must not be in an active game
	if active, err := uc.duelGameRepo.FindActiveByPlayer(inviterID); err == nil && active != nil {
		return StartChallengeOutput{}, quick_duel.ErrAlreadyInGame
	}
	// B1: Guard — invitee must not be in an active game
	if active, err := uc.duelGameRepo.FindActiveByPlayer(accepterID); err == nil && active != nil {
		return StartChallengeOutput{}, quick_duel.ErrAlreadyInGame
	}
```

**Step 4: Run tests**

```bash
cd backend && go test ./internal/application/quick_duel/... -v 2>&1 | tail -20
```
Expected: all PASS.

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/application/quick_duel/testutil_test.go
git commit -m "fix(pvp-duel): B1 — active game guard in StartChallenge"
```

---

### Task 4: B4 — Fix DeleteExpired for accepted_waiting_inviter

**Files:**
- Modify: `backend/internal/infrastructure/persistence/postgres/challenge_repository.go` (line 142)

**Step 1: Fix the SQL query**

Replace the `DeleteExpired` method body:

```go
func (r *ChallengeRepository) DeleteExpired(currentTime int64) error {
	query := `
		UPDATE duel_challenges
		SET status = 'expired', responded_at = $1
		WHERE (status = 'pending' AND expires_at <= $1)
		   OR (status = 'accepted_waiting_inviter' AND responded_at + 1800 <= $1)
	`
	_, err := r.db.Exec(query, currentTime)
	return err
}
```

The second clause expires `accepted_waiting_inviter` challenges 30 minutes (1800s) after the invitee accepted (`responded_at`).

**Step 2: Run all backend tests**

```bash
cd backend && go test ./... 2>&1 | tail -10
```
Expected: all PASS (this is a SQL-only change, covered by existing tests).

**Step 3: Commit**

```bash
git add backend/internal/infrastructure/persistence/postgres/challenge_repository.go
git commit -m "fix(pvp-duel): B4 — expire accepted_waiting_inviter after 30 min"
```

---

### Task 5: F5 — Add FindAcceptedWaitingForPlayer

**Files:**
- Modify: `backend/internal/domain/quick_duel/repository.go`
- Modify: `backend/internal/application/quick_duel/testutil_test.go` (mockChallengeRepo)
- Modify: `backend/internal/infrastructure/persistence/postgres/challenge_repository.go`

**Step 1: Add method to domain interface**

In `repository.go`, inside the `ChallengeRepository` interface, add after `FindPendingByChallenger`:

```go
	// FindAcceptedWaitingForPlayer retrieves challenges where the player is the invitee
	// and the challenge status is accepted_waiting_inviter (game not yet started).
	FindAcceptedWaitingForPlayer(playerID UserID) ([]*DuelChallenge, error)
```

**Step 2: The code won't compile — add mock implementation**

In `testutil_test.go`, add to `mockChallengeRepo`:

```go
func (m *mockChallengeRepo) FindAcceptedWaitingForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	var result []*quick_duel.DuelChallenge
	for _, c := range m.challenges {
		if c.Status() == quick_duel.ChallengeStatusAcceptedWaitingInviter {
			if c.ChallengedID() != nil && c.ChallengedID().Equals(playerID) {
				result = append(result, c)
			}
		}
	}
	return result, nil
}
```

**Step 3: Add postgres implementation**

In `challenge_repository.go`, add after `FindPendingByChallenger`:

```go
func (r *ChallengeRepository) FindAcceptedWaitingForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, match_id, expires_at, created_at, responded_at
		FROM duel_challenges
		WHERE challenged_id = $1 AND status = 'accepted_waiting_inviter'
		ORDER BY responded_at DESC
	`
	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}
```

**Step 4: Verify compilation**

```bash
cd backend && go build ./...
```
Expected: no errors.

**Step 5: Commit**

```bash
git add backend/internal/domain/quick_duel/repository.go \
        backend/internal/application/quick_duel/testutil_test.go \
        backend/internal/infrastructure/persistence/postgres/challenge_repository.go
git commit -m "feat(pvp-duel): F5 — FindAcceptedWaitingForPlayer repo method"
```

---

### Task 6: F1 — acceptedChallenges in /duel/status

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go` (GetDuelStatusUseCase.Execute)
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Write the failing test**

```go
func TestGetDuelStatus_WithAcceptedChallenge(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Create link challenge and have player1 accept it as invitee
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer1ID), "Player1", now+10)
	f.challengeRepo.Save(challenge)

	uc := f.newGetDuelStatusUC()
	output, err := uc.Execute(GetDuelStatusInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.AcceptedChallenges) != 1 {
		t.Fatalf("AcceptedChallenges = %d, want 1", len(output.AcceptedChallenges))
	}
	if output.AcceptedChallenges[0].ChallengerID != testPlayer2ID {
		t.Errorf("ChallengerID = %s, want %s", output.AcceptedChallenges[0].ChallengerID, testPlayer2ID)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetDuelStatus_WithAcceptedChallenge -v
```
Expected: FAIL — field doesn't exist yet.

**Step 3: Add AcceptedChallenges to GetDuelStatusOutput**

In `dto.go`, add field to `GetDuelStatusOutput`:

```go
type GetDuelStatusOutput struct {
	HasActiveDuel      bool           `json:"hasActiveDuel"`
	ActiveGameID       *string        `json:"activeGameId,omitempty"`
	Player             PlayerRatingDTO `json:"player"`
	Tickets            int            `json:"tickets"`
	FriendsOnline      []FriendDTO    `json:"friendsOnline"`
	PendingChallenges  []ChallengeDTO `json:"pendingChallenges"`
	OutgoingChallenges []ChallengeDTO `json:"outgoingChallenges"`
	AcceptedChallenges []ChallengeDTO `json:"acceptedChallenges"` // F1: added
	SeasonID           string         `json:"seasonId"`
	SeasonEndsAt       int64          `json:"seasonEndsAt"`
}
```

**Step 4: Populate AcceptedChallenges in GetDuelStatusUseCase.Execute**

In `use_cases.go`, in `GetDuelStatusUseCase.Execute`, after building `outgoingDTOs` (around line 119), add:

```go
	// F1: accepted challenges — invitee is waiting for inviter to start
	acceptedChallenges, err := uc.challengeRepo.FindAcceptedWaitingForPlayer(playerID)
	if err != nil {
		acceptedChallenges = []*quick_duel.DuelChallenge{}
	}
	acceptedDTOs := make([]ChallengeDTO, 0, len(acceptedChallenges))
	for _, c := range acceptedChallenges {
		username := c.ChallengerID().String()
		if u, err := uc.userRepo.FindByID(c.ChallengerID()); err == nil && u != nil {
			if n := u.TelegramUsername().String(); n != "" {
				username = n
			} else if n := u.Username().String(); n != "" {
				username = n
			}
		}
		acceptedDTOs = append(acceptedDTOs, ToChallengeDTO(c, now, username))
	}
```

Then in the return statement, add `AcceptedChallenges: acceptedDTOs`.

**Step 5: Update swagger model**

In `swagger_models.go`, update `GetDuelStatusResponse.Data`:

```go
type GetDuelStatusResponse struct {
	Data struct {
		HasActiveDuel      bool               `json:"hasActiveDuel"`
		ActiveGameID       *string            `json:"activeGameId,omitempty"`
		Player             DuelPlayerRatingDTO `json:"player"`
		Tickets            int                `json:"tickets"`
		FriendsOnline      []DuelFriendDTO    `json:"friendsOnline"`
		PendingChallenges  []DuelChallengeDTO `json:"pendingChallenges"`
		OutgoingChallenges []DuelChallengeDTO `json:"outgoingChallenges"`
		AcceptedChallenges []DuelChallengeDTO `json:"acceptedChallenges"` // F1: added
		SeasonID           string             `json:"seasonId"`
		SeasonEndsAt       int64              `json:"seasonEndsAt"`
	} `json:"data"`
}
```

**Step 6: Run tests**

```bash
cd backend && go test ./... 2>&1 | tail -10
```
Expected: all PASS.

**Step 7: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/infrastructure/http/handlers/swagger_models.go
git commit -m "feat(pvp-duel): F1 — acceptedChallenges in /duel/status"
```

---

### Task 7: F2 — inviterName in accept-by-code response

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go` (AcceptByLinkCodeUseCase.Execute)
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Write the failing test**

```go
func TestAcceptByLinkCode_ReturnsInviterName(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// player1 (username "Player1") creates link challenge
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newAcceptByLinkCodeUC()
	output, err := uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: challenge.ChallengeLink(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.InviterName != "Player1" {
		t.Errorf("InviterName = %q, want %q", output.InviterName, "Player1")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestAcceptByLinkCode_ReturnsInviterName -v
```
Expected: FAIL — field doesn't exist.

**Step 3: Add InviterName to AcceptByLinkCodeOutput**

In `dto.go`:

```go
type AcceptByLinkCodeOutput struct {
	Success     bool   `json:"success"`
	ChallengeID string `json:"challengeId"`
	Status      string `json:"status"`
	InviterName string `json:"inviterName"` // F2: added
}
```

**Step 4: Populate InviterName in AcceptByLinkCodeUseCase.Execute**

In `use_cases.go`, at the end of `AcceptByLinkCodeUseCase.Execute`, before the final return, add inviter name resolution:

```go
	// F2: Resolve inviter's display name
	inviterName := challenge.ChallengerID().String()
	if u, err := uc.userRepo.FindByID(challenge.ChallengerID()); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			inviterName = n
		} else if n := u.Username().String(); n != "" {
			inviterName = n
		}
	}

	return AcceptByLinkCodeOutput{
		Success:     true,
		ChallengeID: challenge.ID().String(),
		Status:      string(quick_duel.ChallengeStatusAcceptedWaitingInviter),
		InviterName: inviterName, // F2: added
	}, nil
```

Also update the idempotency early-return to include `InviterName` (when challenge is already `accepted_waiting_inviter` for this player):

```go
	if challenge.Status() == quick_duel.ChallengeStatusAcceptedWaitingInviter {
		if challenged := challenge.ChallengedID(); challenged != nil && challenged.Equals(accepterID) {
			// Resolve inviter name for idempotent response
			inviterName := challenge.ChallengerID().String()
			if u, err := uc.userRepo.FindByID(challenge.ChallengerID()); err == nil && u != nil {
				if n := u.TelegramUsername().String(); n != "" {
					inviterName = n
				} else if n := u.Username().String(); n != "" {
					inviterName = n
				}
			}
			return AcceptByLinkCodeOutput{
				Success:     true,
				ChallengeID: challenge.ID().String(),
				Status:      string(quick_duel.ChallengeStatusAcceptedWaitingInviter),
				InviterName: inviterName,
			}, nil
		}
		return AcceptByLinkCodeOutput{}, quick_duel.ErrChallengeNotPending
	}
```

**Step 5: Update swagger model**

In `swagger_models.go`, update `AcceptByLinkCodeResponse`:

```go
type AcceptByLinkCodeResponse struct {
	Data struct {
		Success     bool    `json:"success"`
		ChallengeID string  `json:"challengeId"`
		Status      string  `json:"status"`
		InviterName string  `json:"inviterName"` // F2: added
	} `json:"data"`
}
```

**Step 6: Run tests**

```bash
cd backend && go test ./... 2>&1 | tail -10
```
Expected: all PASS.

**Step 7: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/infrastructure/http/handlers/swagger_models.go
git commit -m "feat(pvp-duel): F2 — inviterName in accept-by-code response"
```

---

### Task 8: Generate TypeScript types

**Files:**
- Generated: `tma/src/api/generated/` (auto-generated, do not edit manually)

**Step 1: Regenerate Swagger + TypeScript**

```bash
cd tma && pnpm run generate:all
```
Expected output: `swagger.json` updated, `tma/src/api/generated/` files updated.

**Step 2: Verify new types exist**

```bash
grep -r "acceptedChallenges\|inviterName" tma/src/api/generated/
```
Expected: matches in generated types/schemas files.

**Step 3: Commit**

```bash
git add tma/src/api/generated/ backend/docs/
git commit -m "chore: regenerate API types after Phase 1 backend changes"
```

---

### Task 9: U1 — Render outgoingPendingChallenges cards

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Context:** `outgoingPendingChallenges` computed is already in `usePvPDuel.ts` (line 141). This is pending LINK challenges created by the current user that nobody has accepted yet. We need to render them.

**Step 1: Add formatExpiresIn helper to DuelLobbyView script**

In the `<script setup>` section, after the existing helpers:

```typescript
const formatExpiresIn = (seconds: number): string => {
	const hours = Math.floor(seconds / 3600)
	const minutes = Math.floor((seconds % 3600) / 60)
	if (hours > 0) return `${hours}ч ${minutes}мин`
	return `${minutes}мин`
}
```

**Step 2: Add outgoingPendingChallenges to destructuring**

In the `usePvPDuel` destructuring at the top of the script, add:

```typescript
	outgoingPendingChallenges,
```

**Step 3: Add cards to template**

In the Play Tab section (`<div v-if="activeTab === 'play'">`), after the existing `outgoingReadyChallenges` block (around line 480), add:

```html
<!-- Outgoing pending link challenges (inviter waiting for someone to click) -->
<div v-if="outgoingPendingChallenges.length > 0" class="space-y-2">
    <UCard
        v-for="challenge in outgoingPendingChallenges"
        :key="challenge.id"
        class="border-blue-200 dark:border-blue-800"
    >
        <div class="flex items-center gap-2 mb-1">
            <span class="text-blue-500">⏳</span>
            <p class="font-medium">{{ t('duel.waitingForResponse') }}</p>
        </div>
        <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('duel.linkExpiresIn', { time: formatExpiresIn(challenge.expiresIn ?? 0) }) }}
        </p>
    </UCard>
</div>
```

**Step 4: Add i18n keys**

In the i18n locale files (find with `grep -r "duel\." tma/src/i18n/ --include="*.json" -l`), add:

```json
"duel.waitingForResponse": "Ожидаем ответа...",
"duel.linkExpiresIn": "Ссылка истекает через: {time}"
```

**Step 5: Verify in dev**

```bash
cd tma && pnpm dev
```
Open `https://dev.quiz-sprint-tma.online` in Telegram and verify pending link challenges appear.

**Step 6: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue tma/src/i18n/
git commit -m "feat(pvp-duel): U1 — render outgoingPendingChallenges cards"
```

---

### Task 10: U2 — Invitee waiting banner + polling

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts`
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Add acceptedChallenges computed to usePvPDuel.ts**

In `usePvPDuel.ts`, after `outgoingPendingChallenges` computed (~line 143):

```typescript
// Accepted challenges: invitee accepted a link challenge, waiting for inviter to start
const acceptedChallenges = computed(() => statusData.value?.data?.acceptedChallenges ?? [])
```

**Step 2: Extend poll to cover invitee side**

Replace the existing `watch(outgoingChallenges, ...)` block with:

```typescript
watch(
    [outgoingChallenges, acceptedChallenges],
    ([outgoing, accepted], prev) => {
        const prevOutgoing = prev?.[0]
        if (outgoing.length > 0 || accepted.length > 0) {
            startOutgoingPoll()
        } else {
            stopOutgoingPoll()
        }
        if (prevOutgoing !== undefined && outgoing.length < prevOutgoing.length) {
            refetchRivals()
        }
    },
    { immediate: true },
)
```

**Step 3: Export acceptedChallenges from usePvPDuel**

In the return object at the bottom of `usePvPDuel.ts`:

```typescript
return {
    // ... existing exports ...
    acceptedChallenges, // U2: added
    // ...
}
```

**Step 4: Add waiting banner to DuelLobbyView template**

In `DuelLobbyView.vue`, add `acceptedChallenges` to the `usePvPDuel` destructuring.

In the template, before the Player Rating Card (around line 289), add the accepted challenge banner:

```html
<!-- Invitee waiting banner: accepted link challenge, waiting for inviter to start -->
<UCard
    v-for="challenge in acceptedChallenges"
    :key="challenge.id"
    class="mb-4 border-purple-200 dark:border-purple-800"
>
    <div class="flex items-center gap-3">
        <UIcon
            name="i-heroicons-clock"
            class="size-6 text-purple-500 animate-pulse flex-shrink-0"
        />
        <div class="flex-1">
            <p class="font-semibold">
                {{ t('duel.acceptedChallengeFrom', { name: challenge.challengerUsername || t('duel.opponent') }) }}
            </p>
            <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('duel.waitingForGameStart') }}
            </p>
        </div>
    </div>
</UCard>
```

**Step 5: Add i18n keys**

```json
"duel.acceptedChallengeFrom": "Вы приняли вызов {name}",
"duel.waitingForGameStart": "Ждём начала игры...",
"duel.opponent": "соперника"
```

**Step 6: Verify the polling works**

The poll triggers automatically via the watch. When `acceptedChallenges.length > 0`, `startOutgoingPoll` runs and refetches status every 5s. When `hasActiveDuel` becomes true, `goToActiveDuel()` navigates to the game.

**Step 7: Commit**

```bash
git add tma/src/composables/usePvPDuel.ts \
        tma/src/views/Duel/DuelLobbyView.vue \
        tma/src/i18n/
git commit -m "feat(pvp-duel): U2 — invitee waiting banner + polling"
```

---

### Task 11: U3 — Inviter name in post-accept display

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Context:** The `inviterName` is now available in the `accept-by-code` API response (F2) and also in `acceptedChallenges[].challengerUsername` from `/duel/status` (F1). U3 is satisfied by the U2 banner already showing `challenge.challengerUsername`.

For the confirmation modal, we can show the inviter name only AFTER accepting (since we get the name from the response). We update the modal text to display the name when available.

**Step 1: Add inviterName ref to DuelLobbyView script**

```typescript
const inviterName = ref<string | null>(null)
```

**Step 2: Store inviterName after accept-by-code response**

In `handleAcceptByLinkCode`, after a successful response:

```typescript
const handleAcceptByLinkCode = async (linkCode: string) => {
    // ...existing checks...
    try {
        const response = await acceptByLinkCode({
            data: { playerId: playerId.value, linkCode },
        })

        if (response.data?.success) {
            // U3: store inviter name from response
            inviterName.value = response.data.inviterName ?? null

            deepLinkChallenge.value = null
            if (response.data.gameId) {
                router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
            } else {
                router.replace({ name: 'duel-lobby' })
                await refetchStatus()
                if (hasActiveDuel.value && activeGameId.value) {
                    goToActiveDuel()
                }
            }
        }
    } catch (error: unknown) {
        // ...existing error handling...
    }
}
```

**Step 3: Update confirmation modal text to show inviter name when available**

In the confirmation modal (`<UModal v-model:open="showConfirmModal">`), update the description paragraph:

```html
<p class="text-gray-600 dark:text-gray-400 mb-6">
    <template v-if="inviterName">
        {{ t('duel.wantsToFightNamed', { name: inviterName }) }}
    </template>
    <template v-else>
        {{ t('duel.wantsToFight') }}
    </template>
</p>
```

**Step 4: Add i18n key**

```json
"duel.wantsToFightNamed": "{name} вызывает вас на дуэль!"
```

**Step 5: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue tma/src/i18n/
git commit -m "feat(pvp-duel): U3 — inviter name in post-accept display"
```

---

## Phase 2: Cancel flow (F3, F4) + UI (U4, U5) + B5 link code hardening

---

### Task 12: B5 — Extend link code to 12 hex chars

**Files:**
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go` (line 116)
- Test: `backend/internal/domain/quick_duel/duel_challenge_test.go`

**Step 1: Write the failing test**

In `duel_challenge_test.go`:

```go
func TestNewLinkChallenge_LinkCodeIs12HexChars(t *testing.T) {
	challengerID, _ := shared.NewUserID("user-challenger-001")
	now := int64(1700000000)

	challenge, err := quick_duel.NewLinkChallenge(challengerID, now)
	if err != nil {
		t.Fatal(err)
	}

	link := challenge.ChallengeLink()
	// Link format: "https://t.me/quiz_sprint_dev_bot?startapp=duel_<code>"
	parts := strings.Split(link, "duel_")
	if len(parts) != 2 {
		t.Fatalf("unexpected link format: %s", link)
	}
	code := parts[1]
	if len(code) != 12 {
		t.Errorf("link code length = %d, want 12", len(code))
	}
	// All chars must be hex
	for _, c := range code {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("link code contains non-hex char %q in %q", c, code)
		}
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestNewLinkChallenge_LinkCodeIs12HexChars -v
```
Expected: FAIL (current code uses 8 chars with possible hyphen at position 8).

**Step 3: Fix link code generation in duel_challenge.go**

In `NewLinkChallenge`, update the link generation:

```go
// B5: Use 12 pure hex chars (strip hyphens from UUID to get 32 hex chars, take first 12)
// 16^12 = 281 trillion combos vs 16^8 = 4.3 billion previously
rawHex := strings.ReplaceAll(challengeID.String(), "-", "")
link := "https://t.me/quiz_sprint_dev_bot?startapp=duel_" + rawHex[:12]
```

Add `"strings"` to imports if not present.

**Step 4: Run tests**

```bash
cd backend && go test ./internal/domain/quick_duel/... -v 2>&1 | tail -15
```
Expected: all PASS.

**Step 5: Commit**

```bash
git add backend/internal/domain/quick_duel/duel_challenge.go \
        backend/internal/domain/quick_duel/duel_challenge_test.go
git commit -m "fix(pvp-duel): B5 — extend link code to 12 hex chars (281T combos)"
```

---

### Task 13: F3 — pendingChallengeId in RivalDTO

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go` (GetRivalsUseCase.Execute)
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Add PendingChallengeID to RivalDTO**

In `dto.go`:

```go
type RivalDTO struct {
	ID                  string  `json:"id"`
	Username            string  `json:"username"`
	MMR                 int     `json:"mmr"`
	League              string  `json:"league"`
	LeagueIcon          string  `json:"leagueIcon"`
	IsOnline            bool    `json:"isOnline"`
	GamesCount          int     `json:"gamesCount"`
	HasPendingChallenge bool    `json:"hasPendingChallenge"`
	PendingChallengeID  *string `json:"pendingChallengeId,omitempty"` // F3: added
}
```

**Step 2: Populate PendingChallengeID in GetRivalsUseCase.Execute**

In `use_cases.go`, find `GetRivalsUseCase.Execute`. When building each `RivalDTO`, look up the pending challenge and include its ID:

```go
// For each rival, find pending challenge ID
var pendingChallengeID *string
outgoing, _ := uc.challengeRepo.FindPendingByChallenger(playerID)
for _, c := range outgoing {
    if c.ChallengedID() != nil && c.ChallengedID().Equals(rivalID) &&
        c.Status() == quick_duel.ChallengeStatusPending {
        id := c.ID().String()
        pendingChallengeID = &id
        break
    }
}
```

Set `PendingChallengeID: pendingChallengeID` in the `RivalDTO` construction.

Note: `FindPendingByChallenger` is called once per `GetRivals` call and cached in a local slice. Then per-rival, scan the slice for a match. Avoid calling the repo per rival.

**Step 3: Update swagger model**

Find `RivalDTO` in `swagger_models.go` (or create it if it's only in the app DTO). The Swagger model for rivals is likely in swagger_models.go or inline in the handler annotation. Update or create:

```go
// DuelRivalDTO represents a recent opponent
type DuelRivalDTO struct {
	ID                  string  `json:"id"`
	Username            string  `json:"username"`
	MMR                 int     `json:"mmr"`
	League              string  `json:"league"`
	LeagueIcon          string  `json:"leagueIcon"`
	IsOnline            bool    `json:"isOnline"`
	GamesCount          int     `json:"gamesCount"`
	HasPendingChallenge bool    `json:"hasPendingChallenge"`
	PendingChallengeID  *string `json:"pendingChallengeId,omitempty"` // F3: added
}
// @name DuelRivalDTO
```

**Step 4: Run tests**

```bash
cd backend && go test ./... 2>&1 | tail -10
```
Expected: all PASS.

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go \
        backend/internal/infrastructure/http/handlers/swagger_models.go
git commit -m "feat(pvp-duel): F3 — pendingChallengeId in RivalDTO"
```

---

### Task 14: F4 — CancelChallenge domain methods + errors

**Files:**
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go`
- Modify: `backend/internal/domain/quick_duel/errors.go`
- Test: `backend/internal/domain/quick_duel/duel_challenge_test.go`

**Step 1: Add new status and errors**

In `duel_challenge.go`, add to status constants:

```go
ChallengeStatusCancelled ChallengeStatus = "cancelled"
```

In `errors.go`, add:

```go
ErrCancelNotAuthorized    = errors.New("not authorized to cancel this challenge")
ErrChallengeNotCancellable = errors.New("challenge cannot be cancelled in its current state")
```

**Step 2: Write the failing tests**

In `duel_challenge_test.go`:

```go
func TestCancelByChallenger_CancelsFromPending(t *testing.T) {
	challengerID, _ := shared.NewUserID("challenger-001")
	friendID, _ := shared.NewUserID("friend-001")
	now := int64(1700000000)

	challenge, _ := quick_duel.NewDirectChallenge(challengerID, friendID, now)
	err := challenge.CancelByChallenger(challengerID, now+5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if challenge.Status() != quick_duel.ChallengeStatusCancelled {
		t.Errorf("status = %s, want cancelled", challenge.Status())
	}
}

func TestCancelByChallenger_CancelsFromAcceptedWaiting(t *testing.T) {
	challengerID, _ := shared.NewUserID("challenger-001")
	inviteeID, _ := shared.NewUserID("invitee-001")
	now := int64(1700000000)

	challenge, _ := quick_duel.NewLinkChallenge(challengerID, now)
	_ = challenge.AcceptWaiting(inviteeID, "invitee", now+10)

	err := challenge.CancelByChallenger(challengerID, now+20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if challenge.Status() != quick_duel.ChallengeStatusCancelled {
		t.Errorf("status = %s, want cancelled", challenge.Status())
	}
}

func TestCancelByChallenger_FailsIfNotChallenger(t *testing.T) {
	challengerID, _ := shared.NewUserID("challenger-001")
	friendID, _ := shared.NewUserID("friend-001")
	now := int64(1700000000)

	challenge, _ := quick_duel.NewDirectChallenge(challengerID, friendID, now)
	err := challenge.CancelByChallenger(friendID, now+5)
	if !errors.Is(err, quick_duel.ErrCancelNotAuthorized) {
		t.Errorf("expected ErrCancelNotAuthorized, got %v", err)
	}
}

func TestCancelByInvitee_CancelsFromAcceptedWaiting(t *testing.T) {
	challengerID, _ := shared.NewUserID("challenger-001")
	inviteeID, _ := shared.NewUserID("invitee-001")
	now := int64(1700000000)

	challenge, _ := quick_duel.NewLinkChallenge(challengerID, now)
	_ = challenge.AcceptWaiting(inviteeID, "invitee", now+10)

	err := challenge.CancelByInvitee(inviteeID, now+20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if challenge.Status() != quick_duel.ChallengeStatusCancelled {
		t.Errorf("status = %s, want cancelled", challenge.Status())
	}
}

func TestCancelByInvitee_FailsIfNotAcceptedWaiting(t *testing.T) {
	challengerID, _ := shared.NewUserID("challenger-001")
	friendID, _ := shared.NewUserID("friend-001")
	now := int64(1700000000)

	challenge, _ := quick_duel.NewDirectChallenge(challengerID, friendID, now)
	err := challenge.CancelByInvitee(friendID, now+5)
	if !errors.Is(err, quick_duel.ErrChallengeNotCancellable) {
		t.Errorf("expected ErrChallengeNotCancellable, got %v", err)
	}
}
```

**Step 3: Run tests to verify they fail**

```bash
cd backend && go test ./internal/domain/quick_duel/... -run "TestCancel" -v
```
Expected: FAIL.

**Step 4: Implement CancelByChallenger and CancelByInvitee**

In `duel_challenge.go`, add after `Decline`:

```go
// CancelByChallenger cancels the challenge. Works for pending and accepted_waiting_inviter.
func (dc *DuelChallenge) CancelByChallenger(cancellerID UserID, cancelledAt int64) error {
	if !dc.challengerID.Equals(cancellerID) {
		return ErrCancelNotAuthorized
	}
	if dc.status != ChallengeStatusPending && dc.status != ChallengeStatusAcceptedWaitingInviter {
		return ErrChallengeNotCancellable
	}
	dc.status = ChallengeStatusCancelled
	dc.respondedAt = cancelledAt
	return nil
}

// CancelByInvitee cancels the challenge after invitee accepted but inviter hasn't started.
// Works only for accepted_waiting_inviter.
func (dc *DuelChallenge) CancelByInvitee(cancellerID UserID, cancelledAt int64) error {
	if dc.challengedID == nil || !dc.challengedID.Equals(cancellerID) {
		return ErrCancelNotAuthorized
	}
	if dc.status != ChallengeStatusAcceptedWaitingInviter {
		return ErrChallengeNotCancellable
	}
	dc.status = ChallengeStatusCancelled
	dc.respondedAt = cancelledAt
	return nil
}
```

**Step 5: Run tests**

```bash
cd backend && go test ./internal/domain/quick_duel/... -v 2>&1 | tail -15
```
Expected: all PASS.

**Step 6: Commit**

```bash
git add backend/internal/domain/quick_duel/duel_challenge.go \
        backend/internal/domain/quick_duel/errors.go \
        backend/internal/domain/quick_duel/duel_challenge_test.go
git commit -m "feat(pvp-duel): F4 — CancelByChallenger + CancelByInvitee domain methods"
```

---

### Task 15: F4 — CancelChallengeUseCase + HTTP endpoint

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go`
- Modify: `backend/internal/application/quick_duel/testutil_test.go`
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Modify: `backend/internal/infrastructure/http/routes/routes.go`
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Step 1: Add DTOs**

In `dto.go`:

```go
// ========================================
// CancelChallenge Use Case
// ========================================

type CancelChallengeInput struct {
	PlayerID    string `json:"playerId"`
	ChallengeID string `json:"challengeId"`
}

type CancelChallengeOutput struct {
	Success        bool `json:"success"`
	TicketRefunded bool `json:"ticketRefunded"`
}
```

**Step 2: Write the failing test**

In `use_cases_test.go`:

```go
func TestCancelChallenge_ChallengerCancelsPending(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	challenge, _ := quick_duel.NewDirectChallenge(mustUserID(testPlayer1ID), mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newCancelChallengeUC()
	output, err := uc.Execute(CancelChallengeInput{
		PlayerID:    testPlayer1ID,
		ChallengeID: challenge.ID().String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected Success=true")
	}

	// Verify challenge is cancelled in repo
	updated, _ := f.challengeRepo.FindByID(challenge.ID())
	if updated.Status() != quick_duel.ChallengeStatusCancelled {
		t.Errorf("status = %s, want cancelled", updated.Status())
	}
}

func TestCancelChallenge_InviteeCancelsAcceptedWaiting(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer2ID), "Player2", now+10)
	f.challengeRepo.Save(challenge)

	uc := f.newCancelChallengeUC()
	_, err := uc.Execute(CancelChallengeInput{
		PlayerID:    testPlayer2ID,
		ChallengeID: challenge.ID().String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelChallenge_FailsIfUnauthorized(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	challenge, _ := quick_duel.NewDirectChallenge(mustUserID(testPlayer1ID), mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newCancelChallengeUC()
	_, err := uc.Execute(CancelChallengeInput{
		PlayerID:    testPlayer3ID, // not challenger or invitee
		ChallengeID: challenge.ID().String(),
	})
	if !errors.Is(err, quick_duel.ErrCancelNotAuthorized) {
		t.Errorf("expected ErrCancelNotAuthorized, got %v", err)
	}
}
```

**Step 3: Run tests to verify they fail**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestCancelChallenge -v
```
Expected: FAIL.

**Step 4: Implement CancelChallengeUseCase**

In `use_cases.go`, add:

```go
// ========================================
// CancelChallenge Use Case
// ========================================

type CancelChallengeUseCase struct {
	challengeRepo quick_duel.ChallengeRepository
	eventBus      EventBus
}

func NewCancelChallengeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	eventBus EventBus,
) *CancelChallengeUseCase {
	return &CancelChallengeUseCase{
		challengeRepo: challengeRepo,
		eventBus:      eventBus,
	}
}

func (uc *CancelChallengeUseCase) Execute(input CancelChallengeInput) (CancelChallengeOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return CancelChallengeOutput{}, err
	}

	challengeID := quick_duel.NewChallengeIDFromString(input.ChallengeID)
	challenge, err := uc.challengeRepo.FindByID(challengeID)
	if err != nil {
		return CancelChallengeOutput{}, err
	}

	// Try as challenger first, then as invitee
	err = challenge.CancelByChallenger(playerID, now)
	if err != nil {
		err = challenge.CancelByInvitee(playerID, now)
		if err != nil {
			return CancelChallengeOutput{}, quick_duel.ErrCancelNotAuthorized
		}
	}

	if err := uc.challengeRepo.Save(challenge); err != nil {
		return CancelChallengeOutput{}, err
	}

	return CancelChallengeOutput{
		Success:        true,
		TicketRefunded: true,
	}, nil
}
```

**Step 5: Add helper to testutil_test.go**

```go
func (f *duelFixture) newCancelChallengeUC() *CancelChallengeUseCase {
	return NewCancelChallengeUseCase(f.challengeRepo, f.eventBus)
}
```

**Step 6: Add HTTP handler**

In `duel_handlers.go`, add `cancelChallengeUC` to the `DuelHandler` struct and constructor. Then add the handler:

```go
// CancelChallenge handles DELETE /api/v1/duel/challenge/:challengeId
// @Summary Cancel a challenge
// @Description Cancel a pending or accepted_waiting_inviter challenge
// @Tags duel
// @Accept json
// @Produce json
// @Param challengeId path string true "Challenge ID"
// @Param request body CancelChallengeRequest true "Cancel request"
// @Success 200 {object} CancelChallengeResponse "Cancelled"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 403 {object} ErrorResponse "Not authorized"
// @Failure 404 {object} ErrorResponse "Challenge not found"
// @Failure 409 {object} ErrorResponse "Cannot cancel in current state"
// @Router /duel/challenge/{challengeId} [delete]
func (h *DuelHandler) CancelChallenge(c fiber.Ctx) error {
	challengeID := c.Params("challengeId")
	if _, err := uuid.Parse(challengeID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID format")
	}

	var req CancelChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.cancelChallengeUC.Execute(appDuel.CancelChallengeInput{
		PlayerID:    req.PlayerID,
		ChallengeID: challengeID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}
```

**Step 7: Add DTOs to swagger_models.go**

```go
// CancelChallengeRequest is the request for cancelling a challenge
type CancelChallengeRequest struct {
	PlayerID string `json:"playerId" validate:"required"`
}
// @name CancelChallengeRequest

// CancelChallengeResponse wraps the cancel challenge response
type CancelChallengeResponse struct {
	Data struct {
		Success        bool `json:"success"`
		TicketRefunded bool `json:"ticketRefunded"`
	} `json:"data"`
}
// @name CancelChallengeResponse
```

**Step 8: Add error mappings to mapDuelError**

In `duel_handlers.go`, add to the switch in `mapDuelError`:

```go
case domainDuel.ErrCancelNotAuthorized:
    return fiber.NewError(fiber.StatusForbidden, "Not authorized to cancel this challenge")
case domainDuel.ErrChallengeNotCancellable:
    return fiber.NewError(fiber.StatusConflict, "Challenge cannot be cancelled in its current state")
```

**Step 9: Register route in routes.go**

In `routes.go`, in the duel routes section, add:

```go
duelGroup.Delete("/challenge/:challengeId", duelHandler.CancelChallenge)
```

Also wire the use case in the handler constructor section:

```go
cancelChallengeUC := appDuel.NewCancelChallengeUseCase(challengeRepo, duelEventBus)
```

**Step 10: Run all tests**

```bash
cd backend && go test ./... 2>&1 | tail -15
```
Expected: all PASS.

**Step 11: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/application/quick_duel/testutil_test.go \
        backend/internal/infrastructure/http/handlers/duel_handlers.go \
        backend/internal/infrastructure/http/handlers/swagger_models.go \
        backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): F4 — CancelChallenge use case + DELETE endpoint"
```

---

### Task 16: Generate TypeScript types (after Phase 2 backend)

```bash
cd tma && pnpm run generate:all
```

```bash
git add tma/src/api/generated/ backend/docs/
git commit -m "chore: regenerate API types after Phase 2 backend changes"
```

---

### Task 17: U4 — Cancel buttons in DuelLobbyView

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts`
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Add cancelChallenge action to usePvPDuel.ts**

After the `pnpm run generate:all` in Task 16, a hook `useDeleteDuelChallengeChallengeid` should exist. Import and use it:

```typescript
import { useDeleteDuelChallengeChallengeid } from '@/api/generated'

// In usePvPDuel function body:
const cancelChallengeMutation = useDeleteDuelChallengeChallengeid()

const cancelChallenge = async (challengeId: string) => {
    try {
        await cancelChallengeMutation.mutateAsync({
            challengeId,
            data: { playerId },
        })
        await refetchStatus()
        await refetchRivals()
    } catch (error) {
        console.error('[usePvPDuel] Failed to cancel challenge:', error)
        throw error
    }
}
```

Export `cancelChallenge` in the return object.

**Step 2: Add cancelChallenge to DuelLobbyView destructuring**

```typescript
const {
    // ...existing...
    cancelChallenge, // U4: added
} = usePvPDuel(playerId.value)
```

**Step 3: Add cancel button to outgoing pending link cards**

In the outgoing pending challenges template block (Task 9), add a cancel button:

```html
<UCard
    v-for="challenge in outgoingPendingChallenges"
    :key="challenge.id"
    class="border-blue-200 dark:border-blue-800"
>
    <div class="flex items-center justify-between">
        <div>
            <div class="flex items-center gap-2 mb-1">
                <span class="text-blue-500">⏳</span>
                <p class="font-medium">{{ t('duel.waitingForResponse') }}</p>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('duel.linkExpiresIn', { time: formatExpiresIn(challenge.expiresIn ?? 0) }) }}
            </p>
        </div>
        <UButton
            size="xs"
            color="gray"
            variant="ghost"
            icon="i-heroicons-x-mark"
            @click="() => cancelChallenge(challenge.id!)"
        />
    </div>
</UCard>
```

**Step 4: Add cancel button to invitee waiting banner**

In the accepted challenge banner (Task 10), add cancel button:

```html
<UCard
    v-for="challenge in acceptedChallenges"
    :key="challenge.id"
    class="mb-4 border-purple-200 dark:border-purple-800"
>
    <div class="flex items-center gap-3">
        <UIcon name="i-heroicons-clock" class="size-6 text-purple-500 animate-pulse flex-shrink-0" />
        <div class="flex-1">
            <p class="font-semibold">
                {{ t('duel.acceptedChallengeFrom', { name: challenge.challengerUsername || t('duel.opponent') }) }}
            </p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('duel.waitingForGameStart') }}</p>
        </div>
        <UButton
            size="xs"
            color="gray"
            variant="ghost"
            icon="i-heroicons-x-mark"
            @click="() => cancelChallenge(challenge.id!)"
        />
    </div>
</UCard>
```

**Step 5: Add cancel button to rival "Отправлено" state**

In the rivals list template, update the action button for `hasPendingChallenge`:

```html
<div class="flex items-center gap-1">
    <UButton
        size="sm"
        :disabled="rival.hasPendingChallenge || sendingChallengeId === rival.id"
        :loading="sendingChallengeId === rival.id"
        :color="rival.hasPendingChallenge ? 'gray' : 'primary'"
        :variant="rival.hasPendingChallenge ? 'soft' : 'solid'"
        @click="() => {
            if (!rival.hasPendingChallenge && !sendingChallengeId)
                handleChallengeFriend(rival.id!)
        }"
    >
        {{ rival.hasPendingChallenge ? t('duel.challengeSent') : t('duel.challenge') }}
    </UButton>
    <!-- Cancel button for sent challenge -->
    <UButton
        v-if="rival.hasPendingChallenge && rival.pendingChallengeId"
        size="xs"
        color="gray"
        variant="ghost"
        icon="i-heroicons-x-mark"
        @click="() => cancelChallenge(rival.pendingChallengeId!)"
    />
</div>
```

**Step 6: Commit**

```bash
git add tma/src/composables/usePvPDuel.ts tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): U4 — cancel buttons for pending/accepted challenges"
```

---

### Task 18: U5 — Specific error messages for deep link errors

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Update handleAcceptByLinkCode error handling**

The backend returns HTTP 409/400 with a message string. Map the message to specific user-facing text:

```typescript
const handleAcceptByLinkCode = async (linkCode: string) => {
    if (!playerId.value) {
        deepLinkError.value = t('duel.pleaseLogin')
        return
    }

    try {
        const response = await acceptByLinkCode({
            data: { playerId: playerId.value, linkCode },
        })

        if (response.data?.success) {
            inviterName.value = response.data.inviterName ?? null
            deepLinkChallenge.value = null
            if (response.data.gameId) {
                router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
            } else {
                router.replace({ name: 'duel-lobby' })
                await refetchStatus()
                if (hasActiveDuel.value && activeGameId.value) {
                    goToActiveDuel()
                }
            }
        }
    } catch (error: unknown) {
        console.error('[DuelLobby] Failed to accept challenge:', error)
        const msg: string =
            (error as { response?: { data?: { message?: string } } })?.response?.data?.message ?? ''
        if (msg.includes('expired') || msg.includes('Expired')) {
            deepLinkError.value = t('duel.errorLinkExpired')
        } else if (msg.includes('no longer pending') || msg.includes('already accepted')) {
            deepLinkError.value = t('duel.errorAlreadyAccepted')
        } else if (msg.includes('yourself') || msg.includes('Cannot challenge')) {
            deepLinkError.value = t('duel.errorSelfChallenge')
        } else if (msg.includes('active game') || msg.includes('Already in')) {
            deepLinkError.value = t('duel.errorAlreadyInGame')
        } else {
            deepLinkError.value = t('duel.acceptFailed')
        }
    }
}
```

**Step 2: Add i18n keys**

```json
"duel.errorLinkExpired": "Ссылка устарела. Попроси друга прислать новую",
"duel.errorAlreadyAccepted": "Вызов уже принят другим игроком",
"duel.errorSelfChallenge": "Нельзя вызвать самого себя",
"duel.errorAlreadyInGame": "Вы уже в игре"
```

**Step 3: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue tma/src/i18n/
git commit -m "feat(pvp-duel): U5 — specific error messages for deep link failures"
```

---

## Phase 3: Documentation (D1–D3)

---

### Task 19: D1 — Fix 01_concept.md link flow description

**Files:**
- Modify: `docs/game_modes/pvp_duel/01_concept.md`

Find and correct the "instant start" description for link challenges. The actual flow is two-step:
1. Inviter creates link → sends to friend
2. Friend clicks → `accepted_waiting_inviter`
3. Inviter sees banner → clicks "Start"
4. Game begins

Replace any text suggesting "instant start" for link flow with the two-step description.

```bash
git add docs/game_modes/pvp_duel/01_concept.md
git commit -m "docs(pvp-duel): D1 — fix link flow description in 01_concept.md"
```

---

### Task 20: D2 — Clarify invitee ticket cost for link challenges

**Files:**
- Modify: `docs/game_modes/pvp_duel/03_rules.md` or `04_rewards.md` (whichever mentions tickets)

Find the ticket cost section. Clarify: link challenges currently deduct tickets from the **inviter** only at invite creation. The invitee does NOT pay a ticket for accepting a link challenge (tickets hardcoded to 10; invitee cost = stub pending product decision).

```bash
git add docs/game_modes/pvp_duel/
git commit -m "docs(pvp-duel): D2 — clarify invitee ticket cost for link challenges"
```

---

### Task 21: D3 — Note PushChallengeExpirySeconds unused

**Files:**
- Modify: `docs/game_modes/pvp_duel/03_rules.md` (or `06_domain.md`)
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go` (comment)

In docs, note that `PushChallengeExpirySeconds = 300` is defined but currently unused. Direct challenges always use `DirectChallengeExpirySeconds = 60`. The push notification path is not yet implemented.

In `duel_challenge.go`, add a comment to the constant:

```go
PushChallengeExpirySeconds = 300 // Defined but unused — push notification path not implemented
```

```bash
git add docs/game_modes/pvp_duel/ backend/internal/domain/quick_duel/duel_challenge.go
git commit -m "docs(pvp-duel): D3 — note PushChallengeExpirySeconds is unused"
```

---

## Final verification

After all phases complete:

```bash
# Backend: all tests pass
cd backend && go test ./... -v 2>&1 | grep -E "^(ok|FAIL|---)"

# Frontend: type-check + lint
cd tma && pnpm run type-check && pnpm lint

# Push
git push origin pvp-duel
```
