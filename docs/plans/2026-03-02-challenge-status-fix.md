# Challenge Status Fix: Remove stale card after duel starts

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** After inviter starts a challenge duel, the "✅ готов к дуэли" card disappears from the lobby.

**Architecture:** Bug is a missing status transition in the domain. `SetMatchID()` only stores the game ID but never changes `status` from `accepted_waiting_inviter` → `accepted`. Because `FindPendingByChallenger` queries for `status IN ('pending', 'accepted_waiting_inviter')`, the challenge keeps reappearing. Fix: add `MarkStarted()` to the domain aggregate that atomically sets status → `accepted` and stores the matchID. Replace the call in `StartChallengeUseCase`. Backfill existing stale rows via migration.

**Tech Stack:** Go 1.25 (domain/application layers), PostgreSQL (migration), no frontend changes needed.

---

### Task 1: Domain — add `MarkStarted()` method

**Files:**
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go:263-266`
- Test: `backend/internal/domain/quick_duel/duel_challenge_test.go` (create if absent)

**Step 1: Write the failing test**

Add to `backend/internal/domain/quick_duel/duel_challenge_test.go`:

```go
package quick_duel_test

import (
    "testing"
    "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
    "github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

func TestMarkStarted_TransitionsToAccepted(t *testing.T) {
    challengerID, _ := shared.NewUserID("user-challenger-001")
    now := int64(1700000000)

    challenge, err := quick_duel.NewLinkChallenge(challengerID, now)
    if err != nil {
        t.Fatal(err)
    }

    inviteeID, _ := shared.NewUserID("user-invitee-001")
    err = challenge.AcceptWaiting(inviteeID, "invitee_name", now+10)
    if err != nil {
        t.Fatal(err)
    }

    // Pre-condition
    if challenge.Status() != quick_duel.ChallengeStatusAcceptedWaitingInviter {
        t.Fatalf("expected accepted_waiting_inviter, got %s", challenge.Status())
    }

    gameID := quick_duel.NewGameIDFromString("game-001")
    err = challenge.MarkStarted(gameID)
    if err != nil {
        t.Fatal(err)
    }

    // Status must be accepted
    if challenge.Status() != quick_duel.ChallengeStatusAccepted {
        t.Errorf("expected accepted, got %s", challenge.Status())
    }

    // MatchID must be set
    if challenge.MatchID() == nil || challenge.MatchID().String() != "game-001" {
        t.Errorf("expected matchID=game-001, got %v", challenge.MatchID())
    }
}

func TestMarkStarted_FailsIfNotWaitingInviter(t *testing.T) {
    challengerID, _ := shared.NewUserID("user-challenger-001")
    now := int64(1700000000)

    challenge, err := quick_duel.NewLinkChallenge(challengerID, now)
    if err != nil {
        t.Fatal(err)
    }

    gameID := quick_duel.NewGameIDFromString("game-001")
    err = challenge.MarkStarted(gameID)
    if err == nil {
        t.Error("expected error when status is pending")
    }
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestMarkStarted -v
```

Expected: FAIL — `challenge.MarkStarted undefined`

**Step 3: Implement `MarkStarted` in the domain**

In `backend/internal/domain/quick_duel/duel_challenge.go`, replace the existing `SetMatchID` usage — add a new method right after `SetMatchID` (line ~266):

```go
// MarkStarted transitions challenge from accepted_waiting_inviter to accepted.
// Called when the inviter confirms game start via StartChallenge.
func (dc *DuelChallenge) MarkStarted(matchID GameID) error {
    if dc.status != ChallengeStatusAcceptedWaitingInviter {
        return ErrChallengeNotPending
    }
    dc.matchID = &matchID
    dc.status = ChallengeStatusAccepted
    return nil
}
```

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestMarkStarted -v
```

Expected: PASS (2 tests)

**Step 5: Run all domain tests**

```bash
cd backend && go test ./internal/domain/quick_duel/... -v
```

Expected: all PASS

**Step 6: Commit**

```bash
git add backend/internal/domain/quick_duel/duel_challenge.go backend/internal/domain/quick_duel/duel_challenge_test.go
git commit -m "feat(pvp-duel): add MarkStarted() domain method — transitions challenge to accepted"
```

---

### Task 2: Application — use `MarkStarted` in `StartChallengeUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go:1601-1603`

**Step 1: Replace `SetMatchID` call**

Find the block at line ~1601 in `use_cases.go`:

```go
// Mark challenge as accepted (game started)
challenge.SetMatchID(game.ID())
_ = uc.challengeRepo.Save(challenge)
```

Replace with:

```go
// Transition challenge to accepted — removes it from lobby cards
if err := challenge.MarkStarted(game.ID()); err != nil {
    return StartChallengeOutput{}, err
}
if err := uc.challengeRepo.Save(challenge); err != nil {
    return StartChallengeOutput{}, err
}
```

**Step 2: Verify it compiles**

```bash
cd backend && go build ./...
```

Expected: no errors

**Step 3: Run application-layer tests**

```bash
cd backend && go test ./internal/application/quick_duel/... -v
```

Expected: all PASS

**Step 4: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go
git commit -m "fix(pvp-duel): use MarkStarted() in StartChallengeUseCase — status now transitions to accepted"
```

---

### Task 3: Migration — backfill stale rows

**Files:**
- Create: `backend/migrations/018_fix_stale_challenge_status.sql`

**Step 1: Create the migration file**

```sql
-- Fix stale accepted_waiting_inviter challenges that already have a match_id set.
-- These were created before MarkStarted() was introduced and never had their
-- status updated to accepted after the game was created.
UPDATE duel_challenges
SET status = 'accepted'
WHERE status = 'accepted_waiting_inviter'
  AND match_id IS NOT NULL;
```

**Step 2: Apply the migration**

```bash
cd backend && docker compose -f docker-compose.dev.yml exec postgres \
  psql -U quiz_user -d quiz_sprint_dev \
  -f /migrations/018_fix_stale_challenge_status.sql
```

Or via adminer at http://localhost:8080 if Docker isn't running.

**Step 3: Verify the fix**

```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U quiz_user -d quiz_sprint_dev \
  -c "SELECT id, status, match_id FROM duel_challenges WHERE match_id IS NOT NULL;"
```

Expected: all rows with `match_id IS NOT NULL` now have `status = 'accepted'`

**Step 4: Commit**

```bash
git add backend/migrations/018_fix_stale_challenge_status.sql
git commit -m "fix(pvp-duel): migration 018 — backfill stale accepted_waiting_inviter → accepted"
```

---

### Task 4: Manual smoke test

1. Open the lobby in dev (`https://dev.quiz-sprint-tma.online`)
2. Create a challenge link → share to another account
3. Second account accepts via the link
4. Back on the first account — card "✅ готов к дуэли" appears — click "Начать дуэль"
5. Both players complete the duel
6. Navigate back to the lobby
7. **Expected:** the "✅ готов к дуэли" card is **gone**
8. Also check: `GET /api/v1/duel/status?playerId=<id>` — `outgoingChallenges` should be empty (or only contain pending ones)
