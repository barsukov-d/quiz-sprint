# Direct Challenge Telegram Notification — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** When Player A sends a direct challenge to Player B, Player B receives a Telegram bot message with an "⚔️ Принять вызов" button. The message is automatically edited when the challenge status changes (accepted/declined/expired). Direct challenge TTL increases from 60s to 1 hour. Expired challenges are shown in the lobby (not deleted) and auto-deleted after 24h.

**Architecture:**
- Synchronous notification in `SendChallengeUseCase` after `repo.Save()` — same pattern as `AcceptByLinkCodeUseCase`
- `telegram_message_id` stored in DB, used to edit the message on status change
- Background goroutine runs every minute to transition expired challenges and edit their Telegram messages
- `expired` status already exists in domain; `DeleteExpired` SQL already transitions to it — no domain changes needed for status machine

**Tech Stack:** Go 1.25, Fiber v3, PostgreSQL 16, Telegram Bot API (sendMessage + editMessageText), Vue 3 + TypeScript

---

### Task B1: Increase direct challenge TTL to 1 hour

**Files:**
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go:9`

**Step 1: Change the constant**

In `duel_challenge.go`, change:
```go
DirectChallengeExpirySeconds = 60       // 60 seconds for online friend
```
to:
```go
DirectChallengeExpirySeconds = 3600     // 1 hour — time to open Telegram notification
```

**Step 2: Run domain tests**
```bash
cd backend && go test ./internal/domain/quick_duel/... -v
```
Expected: PASS (no tests depend on the specific TTL value)

**Step 3: Commit**
```bash
git add backend/internal/domain/quick_duel/duel_challenge.go
git commit -m "feat(pvp-duel): increase direct challenge TTL to 1 hour"
```

---

### Task B2: Add `telegram_message_id` to domain aggregate

**Files:**
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go`

**Step 1: Write the failing test**

In `backend/internal/domain/quick_duel/duel_challenge_test.go`, add:
```go
func TestDuelChallenge_TelegramMessageID(t *testing.T) {
    challengerID, _ := shared.NewUserID("111")
    challengedID, _ := shared.NewUserID("222")
    c, _ := NewDirectChallenge(challengerID, challengedID, 1000)

    assert.Equal(t, int64(0), c.TelegramMessageID())

    c.SetTelegramMessageID(42)
    assert.Equal(t, int64(42), c.TelegramMessageID())
}
```

**Step 2: Run test to verify it fails**
```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestDuelChallenge_TelegramMessageID -v
```
Expected: FAIL — `TelegramMessageID undefined`

**Step 3: Add field + getter + setter to `DuelChallenge`**

In `duel_challenge.go`, add field after `inviteeName`:
```go
telegramMessageID int64  // Bot message ID in invitee's chat (0 = not sent)
```

Add getter and setter after `InviteeName()`:
```go
func (dc *DuelChallenge) TelegramMessageID() int64       { return dc.telegramMessageID }
func (dc *DuelChallenge) SetTelegramMessageID(id int64)  { dc.telegramMessageID = id }
```

Update `ReconstructDuelChallenge` signature — add `telegramMessageID int64` parameter and assign it in the struct:
```go
func ReconstructDuelChallenge(
    id ChallengeID,
    challengerID UserID,
    challengedID *UserID,
    challengeType ChallengeType,
    status ChallengeStatus,
    challengeLink string,
    expiresAt int64,
    createdAt int64,
    respondedAt int64,
    matchID *GameID,
    inviteeName string,
    telegramMessageID int64,     // new
) *DuelChallenge {
    return &DuelChallenge{
        ...
        telegramMessageID: telegramMessageID,
    }
}
```

**Step 4: Fix the compile errors** — `ReconstructDuelChallenge` is called in `challenge_repository.go:273`. Add `0` as the last argument temporarily (will be fixed in B4):
```go
// In reconstructChallenge method — add 0 as last arg
return quick_duel.ReconstructDuelChallenge(
    cid, challengerUID, challengedUID,
    quick_duel.ChallengeType(challengeType),
    quick_duel.ChallengeStatus(status),
    link,
    expiresAt, createdAt, ra,
    mid,
    "",   // inviteeName (not yet in repo scan)
    0,    // telegramMessageID (will be added in B4)
)
```

**Step 5: Run tests**
```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestDuelChallenge_TelegramMessageID -v
```
Expected: PASS

**Step 6: Commit**
```bash
git add backend/internal/domain/quick_duel/duel_challenge.go \
        backend/internal/domain/quick_duel/duel_challenge_test.go \
        backend/internal/infrastructure/persistence/postgres/challenge_repository.go
git commit -m "feat(pvp-duel): add telegram_message_id to DuelChallenge domain"
```

---

### Task B3: DB migration — add `telegram_message_id` column

**Files:**
- Create: `backend/migrations/019_add_telegram_message_id_to_challenges.sql`

**Step 1: Create migration file**
```sql
-- Add telegram_message_id to store the bot message ID for direct challenges
-- Needed to edit/delete the notification when challenge status changes
ALTER TABLE duel_challenges
    ADD COLUMN IF NOT EXISTS telegram_message_id BIGINT NULL;
```

**Step 2: Apply migration**
```bash
cd backend && docker compose -f docker-compose.dev.yml exec postgres \
  psql -U quiz_user -d quiz_sprint_dev \
  -f /dev/stdin < migrations/019_add_telegram_message_id_to_challenges.sql
```
Expected: `ALTER TABLE`

**Step 3: Verify column exists**
```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U quiz_user -d quiz_sprint_dev \
  -c "\d duel_challenges"
```
Expected: `telegram_message_id | bigint | nullable`

**Step 4: Commit**
```bash
git add backend/migrations/019_add_telegram_message_id_to_challenges.sql
git commit -m "feat(pvp-duel): migration 019 — add telegram_message_id to duel_challenges"
```

---

### Task B4: Update `ChallengeRepository` to persist `telegram_message_id`

**Files:**
- Modify: `backend/internal/infrastructure/persistence/postgres/challenge_repository.go`

**Step 1: Update `Save` query** — add `telegram_message_id` to INSERT and UPDATE:

```go
query := `
    INSERT INTO duel_challenges (
        id, challenger_id, challenged_id, challenge_type, status,
        challenge_link, match_id, expires_at, created_at, responded_at,
        telegram_message_id
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    ON CONFLICT (id) DO UPDATE SET
        challenged_id = EXCLUDED.challenged_id,
        status = EXCLUDED.status,
        match_id = EXCLUDED.match_id,
        responded_at = EXCLUDED.responded_at,
        telegram_message_id = EXCLUDED.telegram_message_id
`
```

Add the new parameter to `r.db.Exec(...)`:
```go
_, err := r.db.Exec(query,
    challenge.ID().String(),
    challenge.ChallengerID().String(),
    challengedID,
    string(challenge.Type()),
    string(challenge.Status()),
    challenge.ChallengeLink(),
    matchID,
    challenge.ExpiresAt(),
    challenge.CreatedAt(),
    respondedAt,
    challenge.TelegramMessageID(),  // new — 0 if not set, NULL stored as 0
)
```

**Step 2: Update SELECT queries** — add `telegram_message_id` to all SELECT queries in `FindByID`, `FindByLink`, `FindByLinkCode`, `FindPendingForPlayer`, `FindPendingByChallenger`, `FindAcceptedWaitingForPlayer`:

```sql
SELECT id, challenger_id, challenged_id, challenge_type, status,
    challenge_link, match_id, expires_at, created_at, responded_at,
    telegram_message_id
FROM duel_challenges
...
```

**Step 3: Update scan variables** — in both `scanChallenge` and `scanChallenges`, add:
```go
var telegramMessageID sql.NullInt64
```

Add to `row.Scan(...)` / `rows.Scan(...)`:
```go
err := row.Scan(
    &id, &challengerID, &challengedID, &challengeType, &status,
    &challengeLink, &matchID, &expiresAt, &createdAt, &respondedAt,
    &telegramMessageID,
)
```

**Step 4: Update `reconstructChallenge`** — add `telegramMessageID sql.NullInt64` parameter and pass it:

```go
func (r *ChallengeRepository) reconstructChallenge(
    id, challengerID string,
    challengedID sql.NullString,
    challengeType, status string,
    challengeLink, matchID sql.NullString,
    expiresAt, createdAt int64,
    respondedAt sql.NullInt64,
    telegramMessageID sql.NullInt64,  // new
) (*quick_duel.DuelChallenge, error) {
    ...
    var tgMsgID int64
    if telegramMessageID.Valid {
        tgMsgID = telegramMessageID.Int64
    }

    return quick_duel.ReconstructDuelChallenge(
        cid, challengerUID, challengedUID,
        quick_duel.ChallengeType(challengeType),
        quick_duel.ChallengeStatus(status),
        link,
        expiresAt, createdAt, ra,
        mid,
        "",       // inviteeName (not stored in DB — resolved at query time)
        tgMsgID,  // new
    )
}
```

**Step 5: Add `FindPendingExpiredWithMessageID` method**

This is used by the background job to edit Telegram messages BEFORE marking challenges expired:

```go
// FindPendingExpiredWithMessageID returns pending challenges that have expired
// AND have a telegram_message_id set (need their message edited before bulk expire).
func (r *ChallengeRepository) FindPendingExpiredWithMessageID(currentTime int64) ([]*quick_duel.DuelChallenge, error) {
    query := `
        SELECT id, challenger_id, challenged_id, challenge_type, status,
            challenge_link, match_id, expires_at, created_at, responded_at,
            telegram_message_id
        FROM duel_challenges
        WHERE status = 'pending'
          AND expires_at <= $1
          AND telegram_message_id IS NOT NULL
          AND telegram_message_id > 0
    `
    rows, err := r.db.Query(query, currentTime)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return r.scanChallenges(rows)
}
```

**Step 6: Add `DeleteHardExpired` method**

Deletes expired challenges older than 24 hours (hard delete):

```go
// DeleteHardExpired removes expired challenges that have been expired for more than maxAgeSeconds.
func (r *ChallengeRepository) DeleteHardExpired(olderThan int64) error {
    query := `
        DELETE FROM duel_challenges
        WHERE status IN ('expired', 'declined')
          AND responded_at IS NOT NULL
          AND responded_at < $1
    `
    _, err := r.db.Exec(query, olderThan)
    return err
}
```

**Step 7: Add methods to domain repository interface**

In `backend/internal/domain/quick_duel/repository.go`, add to `ChallengeRepository` interface:
```go
// FindPendingExpiredWithMessageID returns pending challenges that expired and have a telegram message to edit
FindPendingExpiredWithMessageID(currentTime int64) ([]*DuelChallenge, error)
// DeleteHardExpired deletes expired/declined challenges older than the given unix timestamp
DeleteHardExpired(olderThan int64) error
```

**Step 8: Build to verify**
```bash
cd backend && go build ./...
```
Expected: no errors

**Step 9: Commit**
```bash
git add backend/internal/infrastructure/persistence/postgres/challenge_repository.go \
        backend/internal/domain/quick_duel/repository.go
git commit -m "feat(pvp-duel): persist telegram_message_id in challenge repo"
```

---

### Task B5: New notifier methods — `NotifyChallengeReceived` + `EditChallengeMessage`

**Files:**
- Modify: `backend/internal/infrastructure/telegram/notifier.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go` (TelegramNotifier interface)

**Step 1: Update application-layer interface**

In `use_cases.go` at the `TelegramNotifier` interface (line ~520), add two methods:
```go
type TelegramNotifier interface {
    NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
    NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
    // New:
    NotifyChallengeReceived(ctx context.Context, inviteeTelegramID int64, inviterName string, deepLink string) (int64, error)
    EditChallengeMessage(ctx context.Context, inviteeTelegramID int64, messageID int64, text string) error
}
```

**Step 2: Update `NoOpNotifier` in `infrastructure/telegram/notifier.go`**

Add to existing interface definition and NoOp implementation:
```go
type TelegramNotifier interface {
    NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
    NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
    NotifyChallengeReceived(ctx context.Context, inviteeTelegramID int64, inviterName string, deepLink string) (int64, error)
    EditChallengeMessage(ctx context.Context, inviteeTelegramID int64, messageID int64, text string) error
}

func (n *NoOpNotifier) NotifyChallengeReceived(_ context.Context, _ int64, _ string, _ string) (int64, error) {
    return 0, nil
}

func (n *NoOpNotifier) EditChallengeMessage(_ context.Context, _ int64, _ int64, _ string) error {
    return nil
}
```

**Step 3: Add `sendMessageWithButton` helper to `HTTPNotifier`**

Returns the Telegram message_id from the API response:
```go
type sendMessageWithButtonRequest struct {
    ChatID      int64          `json:"chat_id"`
    Text        string         `json:"text"`
    ParseMode   string         `json:"parse_mode"`
    ReplyMarkup inlineKeyboard `json:"reply_markup"`
}

type inlineKeyboard struct {
    InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

type inlineButton struct {
    Text string `json:"text"`
    URL  string `json:"url"`
}

type sendMessageResponse struct {
    OK     bool `json:"ok"`
    Result struct {
        MessageID int64 `json:"message_id"`
    } `json:"result"`
}

func (n *HTTPNotifier) sendMessageWithButton(ctx context.Context, chatID int64, text, buttonText, buttonURL string) (int64, error) {
    body, _ := json.Marshal(sendMessageWithButtonRequest{
        ChatID:    chatID,
        Text:      text,
        ParseMode: "HTML",
        ReplyMarkup: inlineKeyboard{
            InlineKeyboard: [][]inlineButton{
                {{Text: buttonText, URL: buttonURL}},
            },
        },
    })
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return 0, err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := n.client.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    var result sendMessageResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return 0, err
    }
    if !result.OK {
        return 0, fmt.Errorf("telegram API error")
    }
    return result.Result.MessageID, nil
}
```

**Step 4: Implement `NotifyChallengeReceived` in `HTTPNotifier`**

```go
func (n *HTTPNotifier) NotifyChallengeReceived(ctx context.Context, inviteeTelegramID int64, inviterName string, deepLink string) (int64, error) {
    text := fmt.Sprintf(
        "⚔️ <b>Вызов на дуэль!</b>\n\n<b>%s</b> бросает тебе вызов в Quiz Sprint.\nУ тебя есть 1 час чтобы принять.",
        inviterName,
    )
    return n.sendMessageWithButton(ctx, inviteeTelegramID, text, "⚔️ Принять вызов", deepLink)
}
```

**Step 5: Add `editMessageText` helper and implement `EditChallengeMessage`**

```go
type editMessageRequest struct {
    ChatID    int64  `json:"chat_id"`
    MessageID int64  `json:"message_id"`
    Text      string `json:"text"`
    ParseMode string `json:"parse_mode"`
}

func (n *HTTPNotifier) EditChallengeMessage(ctx context.Context, inviteeTelegramID int64, messageID int64, text string) error {
    body, _ := json.Marshal(editMessageRequest{
        ChatID:    inviteeTelegramID,
        MessageID: messageID,
        Text:      text,
        ParseMode: "HTML",
    })
    url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", n.botToken)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := n.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}
```

**Step 6: Build to verify**
```bash
cd backend && go build ./...
```
Expected: no errors

**Step 7: Commit**
```bash
git add backend/internal/infrastructure/telegram/notifier.go \
        backend/internal/application/quick_duel/use_cases.go
git commit -m "feat(pvp-duel): add NotifyChallengeReceived + EditChallengeMessage to notifier"
```

---

### Task B6: Update `SendChallengeUseCase` — send notification after save

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go`
- Modify: `backend/internal/infrastructure/http/routes/routes.go`

**Step 1: Add `notifier` and `userRepo` to `SendChallengeUseCase`**

Update struct definition (line ~283):
```go
type SendChallengeUseCase struct {
    challengeRepo quick_duel.ChallengeRepository
    duelGameRepo  quick_duel.DuelGameRepository
    userRepo      domainUser.UserRepository   // new
    notifier      TelegramNotifier             // new
    eventBus      EventBus
}
```

Update constructor:
```go
func NewSendChallengeUseCase(
    challengeRepo quick_duel.ChallengeRepository,
    duelGameRepo quick_duel.DuelGameRepository,
    userRepo domainUser.UserRepository,   // new
    notifier TelegramNotifier,             // new
    eventBus EventBus,
) *SendChallengeUseCase {
    return &SendChallengeUseCase{
        challengeRepo: challengeRepo,
        duelGameRepo:  duelGameRepo,
        userRepo:      userRepo,
        notifier:      notifier,
        eventBus:      eventBus,
    }
}
```

**Step 2: Add notification logic to `Execute` after `repo.Save`**

After the `uc.eventBus.Publish(event)` loop, add:
```go
// Send Telegram notification to invitee (best-effort)
if inviteeTgID, err := strconv.ParseInt(friendID.String(), 10, 64); err == nil && inviteeTgID > 0 {
    // Resolve challenger's display name
    inviterName := challengerID.String()
    if u, err := uc.userRepo.FindByID(challengerID); err == nil && u != nil {
        if n := u.TelegramUsername().String(); n != "" {
            inviterName = n
        } else if n := u.Username().String(); n != "" {
            inviterName = n
        }
    }
    deepLink := "https://t.me/quiz_sprint_dev_bot?startapp=challenge_" + challenge.ID().String()
    if msgID, err := uc.notifier.NotifyChallengeReceived(context.Background(), inviteeTgID, inviterName, deepLink); err == nil && msgID > 0 {
        challenge.SetTelegramMessageID(msgID)
        _ = uc.challengeRepo.Save(challenge) // save messageID (best-effort)
    }
}
```

Add import `"context"` if not already present.

**Step 3: Update `routes.go`** — pass `userRepo` and `telegramNotifier` to `SendChallengeUseCase` (line ~404):
```go
sendChallengeUC = appDuel.NewSendChallengeUseCase(
    challengeRepo,
    duelGameRepo,
    userRepo,           // new
    telegramNotifier,   // new
    duelEventBus,
)
```

**Step 4: Build to verify**
```bash
cd backend && go build ./...
```
Expected: no errors

**Step 5: Run use case tests**
```bash
cd backend && go test ./internal/application/quick_duel/... -v -run TestSendChallenge
```
Expected: PASS (NoOpNotifier used in tests)

**Step 6: Commit**
```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): send Telegram notification on direct challenge"
```

---

### Task B7: Edit Telegram message when invitee responds (accept/decline)

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go`
- Modify: `backend/internal/infrastructure/http/routes/routes.go`

**Step 1: Add `notifier` to `RespondChallengeUseCase`**

Update struct:
```go
type RespondChallengeUseCase struct {
    challengeRepo    quick_duel.ChallengeRepository
    duelGameRepo     quick_duel.DuelGameRepository
    playerRatingRepo quick_duel.PlayerRatingRepository
    questionRepo     QuestionRepository
    seasonRepo       quick_duel.SeasonRepository
    userRepo         domainUser.UserRepository
    notifier         TelegramNotifier  // new
    eventBus         EventBus
}
```

Update constructor — add `notifier TelegramNotifier` parameter and assign it.

**Step 2: Edit message on decline**

After `uc.challengeRepo.Save(challenge)` for the decline path, add:
```go
// Edit Telegram notification (best-effort)
if challenge.TelegramMessageID() > 0 {
    if tgID, err := strconv.ParseInt(challenge.ChallengerID().String(), 10, 64); err == nil {
        _ = uc.notifier.EditChallengeMessage(context.Background(), tgID, challenge.TelegramMessageID(), "❌ Вызов отклонён")
    }
}
```

Wait — the notification was sent to the **invitee** (challengedID), not the challenger. So we edit the message in the **invitee's** chat, not the challenger's. Correct the approach:

```go
if challenge.TelegramMessageID() > 0 {
    if tgID, err := strconv.ParseInt(playerID.String(), 10, 64); err == nil {
        _ = uc.notifier.EditChallengeMessage(context.Background(), tgID, challenge.TelegramMessageID(), "❌ Вызов отклонён")
    }
}
```

**Step 3: Edit message on accept**

After `uc.challengeRepo.Save(challenge)` for the accept path, add:
```go
if challenge.TelegramMessageID() > 0 {
    if tgID, err := strconv.ParseInt(playerID.String(), 10, 64); err == nil {
        _ = uc.notifier.EditChallengeMessage(context.Background(), tgID, challenge.TelegramMessageID(), "✅ Вызов принят — удачи!")
    }
}
```

**Step 4: Update `routes.go`** — pass `telegramNotifier` to `RespondChallengeUseCase`:
```go
respondChallengeUC = appDuel.NewRespondChallengeUseCase(
    challengeRepo,
    duelGameRepo,
    playerRatingRepo,
    seasonRepo,
    duelQuestionRepo,
    userRepo,
    telegramNotifier,   // new
    duelEventBus,
)
```

**Step 5: Build to verify**
```bash
cd backend && go build ./...
```
Expected: no errors

**Step 6: Run tests**
```bash
cd backend && go test ./internal/application/quick_duel/... -v -run TestRespondChallenge
```
Expected: PASS (known pre-existing failures only: `TestRespondChallenge_Accept` StartsIn issue)

**Step 7: Commit**
```bash
git add backend/internal/application/quick_duel/use_cases.go \
        backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): edit Telegram message on challenge accept/decline"
```

---

### Task B8: Background scheduler — expire challenges, edit messages, hard-delete old ones

**Files:**
- Modify: `backend/internal/infrastructure/http/routes/routes.go`

**Step 1: Add scheduler goroutine**

After the use case wiring block in `routes.go`, add a background scheduler that:
1. Every minute: finds challenges about to expire with message IDs → edits messages → calls `DeleteExpired`
2. Every hour: calls `DeleteHardExpired` to remove 24h+ old expired/declined challenges

```go
// Start background challenge cleanup scheduler
go func() {
    minuteTicker := time.NewTicker(1 * time.Minute)
    hourTicker := time.NewTicker(1 * time.Hour)
    defer minuteTicker.Stop()
    defer hourTicker.Stop()

    for {
        select {
        case <-minuteTicker.C:
            now := time.Now().UTC().Unix()
            // Edit Telegram messages for challenges about to be marked expired
            if challengeRepo != nil {
                expiring, err := challengeRepo.FindPendingExpiredWithMessageID(now)
                if err == nil {
                    for _, c := range expiring {
                        if c.TelegramMessageID() > 0 && c.ChallengedID() != nil {
                            if tgID, err := strconv.ParseInt(c.ChallengedID().String(), 10, 64); err == nil {
                                _ = telegramNotifier.EditChallengeMessage(context.Background(), tgID, c.TelegramMessageID(), "⏰ Время истекло")
                            }
                        }
                    }
                }
                _ = challengeRepo.DeleteExpired(now)
            }
        case <-hourTicker.C:
            if challengeRepo != nil {
                oneDayAgo := time.Now().UTC().Unix() - 86400
                _ = challengeRepo.DeleteHardExpired(oneDayAgo)
            }
        }
    }
}()
```

Note: `telegramNotifier` and `challengeRepo` are already in scope from the use case wiring above.

Add imports `"time"` and `"context"` to routes.go if not already present.

**Step 2: Build to verify**
```bash
cd backend && go build ./...
```
Expected: no errors

**Step 3: Commit**
```bash
git add backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): background scheduler for challenge expiry + cleanup"
```

---

### Task B9: Return expired challenges in `GetDuelStatus`

**Files:**
- Modify: `backend/internal/domain/quick_duel/repository.go`
- Modify: `backend/internal/infrastructure/persistence/postgres/challenge_repository.go`
- Modify: `backend/internal/application/quick_duel/dto.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go`

**Step 1: Add `FindExpiredForPlayer` to repo interface**

In `repository.go`, add to `ChallengeRepository`:
```go
// FindExpiredForPlayer returns expired challenges visible to this player (as inviter or invitee)
// within the last 24 hours (still shown in UI before auto-deletion)
FindExpiredForPlayer(playerID UserID) ([]*DuelChallenge, error)
```

**Step 2: Implement in postgres repo**

```go
func (r *ChallengeRepository) FindExpiredForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
    query := `
        SELECT id, challenger_id, challenged_id, challenge_type, status,
            challenge_link, match_id, expires_at, created_at, responded_at,
            telegram_message_id
        FROM duel_challenges
        WHERE status = 'expired'
          AND (challenger_id = $1 OR challenged_id = $1)
        ORDER BY responded_at DESC
        LIMIT 20
    `
    rows, err := r.db.Query(query, playerID.String())
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return r.scanChallenges(rows)
}
```

**Step 3: Add `ExpiredChallenges` to `GetDuelStatusOutput` DTO**

In `dto.go`:
```go
type GetDuelStatusOutput struct {
    ...
    ExpiredChallenges  []ChallengeDTO  `json:"expiredChallenges"`  // new
}
```

**Step 4: Populate in `GetDuelStatusUseCase.Execute`**

After the `acceptedChallenges` block, add:
```go
expiredChallenges, err := uc.challengeRepo.FindExpiredForPlayer(playerID)
if err != nil {
    expiredChallenges = []*quick_duel.DuelChallenge{}
}
expiredDTOs := make([]ChallengeDTO, 0, len(expiredChallenges))
for _, c := range expiredChallenges {
    dto := ToChallengeDTO(c, now, "")
    // Resolve the other player's name
    otherID := c.ChallengerID()
    if c.ChallengerID().Equals(playerID) && c.ChallengedID() != nil {
        otherID = *c.ChallengedID()
    }
    if u, err := uc.userRepo.FindByID(otherID); err == nil && u != nil {
        name := u.TelegramUsername().String()
        if name == "" {
            name = u.Username().String()
        }
        if c.ChallengerID().Equals(playerID) {
            dto.InviteeName = name
        } else {
            dto.ChallengerUsername = name
        }
    }
    expiredDTOs = append(expiredDTOs, dto)
}
```

Update return:
```go
return GetDuelStatusOutput{
    ...
    ExpiredChallenges:  expiredDTOs,  // new
}, nil
```

**Step 5: Build to verify**
```bash
cd backend && go build ./...
```
Expected: no errors

**Step 6: Commit**
```bash
git add backend/internal/domain/quick_duel/repository.go \
        backend/internal/infrastructure/persistence/postgres/challenge_repository.go \
        backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go
git commit -m "feat(pvp-duel): return expired challenges in GetDuelStatus"
```

---

### Task B10: `DeleteChallenge` endpoint

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go`
- Modify: `backend/internal/application/quick_duel/use_cases.go`
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- Modify: `backend/internal/infrastructure/http/routes/routes.go`

**Step 1: Add DTOs**

In `dto.go`:
```go
// ========================================
// DeleteChallenge Use Case
// ========================================

type DeleteChallengeInput struct {
    PlayerID    string `json:"playerId"`
    ChallengeID string `json:"challengeId"`
}

type DeleteChallengeOutput struct {
    Success bool `json:"success"`
}
```

**Step 2: Add use case**

In `use_cases.go`:
```go
// ========================================
// DeleteChallenge Use Case
// ========================================

type DeleteChallengeUseCase struct {
    challengeRepo quick_duel.ChallengeRepository
}

func NewDeleteChallengeUseCase(challengeRepo quick_duel.ChallengeRepository) *DeleteChallengeUseCase {
    return &DeleteChallengeUseCase{challengeRepo: challengeRepo}
}

func (uc *DeleteChallengeUseCase) Execute(input DeleteChallengeInput) (DeleteChallengeOutput, error) {
    playerID, err := shared.NewUserID(input.PlayerID)
    if err != nil {
        return DeleteChallengeOutput{}, err
    }
    challengeID := quick_duel.NewChallengeIDFromString(input.ChallengeID)
    challenge, err := uc.challengeRepo.FindByID(challengeID)
    if err != nil {
        return DeleteChallengeOutput{}, err
    }
    // Only allow inviter or invitee to delete
    isInviter := challenge.ChallengerID().Equals(playerID)
    isInvitee := challenge.ChallengedID() != nil && challenge.ChallengedID().Equals(playerID)
    if !isInviter && !isInvitee {
        return DeleteChallengeOutput{}, quick_duel.ErrNotChallengedPlayer
    }
    // Only allow deleting expired or declined challenges
    if challenge.Status() == quick_duel.ChallengeStatusPending || challenge.Status() == quick_duel.ChallengeStatusAccepted {
        return DeleteChallengeOutput{}, quick_duel.ErrChallengeNotPending
    }
    if err := uc.challengeRepo.Delete(challengeID); err != nil {
        return DeleteChallengeOutput{}, err
    }
    return DeleteChallengeOutput{Success: true}, nil
}
```

**Step 3: Add HTTP handler**

In `duel_handlers.go`, add:
```go
// @Summary      Delete a challenge
// @Description  Player deletes an expired or declined challenge from their lobby
// @Tags         duel
// @Accept       json
// @Produce      json
// @Param        challengeId  path      string  true  "Challenge ID"
// @Success      200          {object}  swagger_models.DeleteChallengeResponseSwagger
// @Failure      403          {object}  swagger_models.ErrorResponseSwagger
// @Failure      404          {object}  swagger_models.ErrorResponseSwagger
// @Router       /duel/challenge/{challengeId} [delete]
func HandleDeleteChallenge(uc *appDuel.DeleteChallengeUseCase) fiber.Handler {
    return func(c fiber.Ctx) error {
        playerID, ok := middleware.GetValidatedPlayerID(c)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
        }
        challengeID := c.Params("challengeId")
        output, err := uc.Execute(appDuel.DeleteChallengeInput{
            PlayerID:    playerID,
            ChallengeID: challengeID,
        })
        if err != nil {
            switch err {
            case quick_duel.ErrChallengeNotFound:
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "challenge not found"})
            case quick_duel.ErrNotChallengedPlayer:
                return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not your challenge"})
            default:
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
            }
        }
        return c.Status(fiber.StatusOK).JSON(output)
    }
}
```

Add swagger model in `swagger_models.go`:
```go
type DeleteChallengeResponseSwagger struct {
    Success bool `json:"success" example:"true"`
}
```

**Step 4: Register route in `routes.go`**

Find where duel routes are registered and add:
```go
duelGroup.Delete("/challenge/:challengeId", handlers.HandleDeleteChallenge(deleteChallengeUC))
```

Also wire the use case:
```go
deleteChallengeUC = appDuel.NewDeleteChallengeUseCase(challengeRepo)
```

**Step 5: Build to verify**
```bash
cd backend && go build ./...
```

**Step 6: Commit**
```bash
git add backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go \
        backend/internal/infrastructure/http/handlers/duel_handlers.go \
        backend/internal/infrastructure/http/handlers/swagger_models.go \
        backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): DeleteChallenge endpoint for expired/declined challenges"
```

---

### Task F1: Generate TypeScript types

**Step 1: Generate Swagger docs + TypeScript types**
```bash
cd backend && make swagger
cd ../tma && pnpm run generate:all
```
Expected: new types in `tma/src/api/generated/` including `expiredChallenges` in status response

**Step 2: Commit**
```bash
git add backend/docs/ tma/src/api/generated/
git commit -m "chore: regenerate API types after Telegram notification backend changes"
```

---

### Task F2: Frontend — show expired challenges + delete button

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts`
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Expose `expiredChallenges` from composable**

In `usePvPDuel.ts`, find where `pendingChallenges`, `outgoingChallenges`, etc. are computed from the status response, and add:
```typescript
const expiredChallenges = computed(() => duelStatus.value?.expiredChallenges ?? [])
```

Add `deleteChallenge` function:
```typescript
const { mutateAsync: deleteChallengeApi } = useDeleteDuelChallengeChallengeid()

const deleteChallenge = async (challengeId: string) => {
    await deleteChallengeApi({ challengeId })
    await refetchStatus()
}
```

Export `expiredChallenges` and `deleteChallenge` in the return object.

**Step 2: Destructure in `DuelLobbyView.vue`**

Add to the destructuring from `usePvPDuel`:
```typescript
const {
    ...
    expiredChallenges,
    deleteChallenge,
} = usePvPDuel(playerId.value)
```

Add handler:
```typescript
const handleDeleteChallenge = async (challengeId: string) => {
    await deleteChallenge(challengeId)
}
```

**Step 3: Add expired challenges section to template**

After the `outgoingPendingChallenges` section, add:
```html
<!-- Expired Challenges -->
<UCard v-if="expiredChallenges.length > 0" class="mb-4">
    <template #header>
        <div class="flex items-center gap-2.5">
            <UIcon name="i-heroicons-clock" class="size-5 text-gray-400" />
            <h3 class="text-base font-semibold text-gray-500">{{ t('duel.expiredChallenges') }}</h3>
        </div>
    </template>

    <div class="space-y-3">
        <div v-for="challenge in expiredChallenges" :key="challenge.id">
            <div class="flex items-center justify-between">
                <div>
                    <!-- Inviter view: shows inviteeName -->
                    <p v-if="challenge.challengerId === playerId" class="text-sm text-gray-500">
                        {{ t('duel.expiredChallengeToInvitee', { name: challenge.inviteeName || t('duel.unknownPlayer') }) }}
                    </p>
                    <!-- Invitee view: shows challenger name -->
                    <p v-else class="text-sm text-gray-500">
                        {{ t('duel.expiredChallengeFromInviter', { name: challenge.challengerUsername || t('duel.unknownPlayer') }) }}
                    </p>
                    <p class="text-xs text-gray-400">⏰ {{ t('duel.challengeExpired') }}</p>
                </div>
                <div class="flex items-center gap-2">
                    <!-- Re-challenge button only for inviter -->
                    <UButton
                        v-if="challenge.challengerId === playerId && challenge.challengedId"
                        size="xs"
                        color="primary"
                        variant="outline"
                        @click="handleChallengeFriend(challenge.challengedId!)"
                    >
                        {{ t('duel.rechallenge') }}
                    </UButton>
                    <UButton
                        size="xs"
                        color="gray"
                        variant="ghost"
                        icon="i-heroicons-trash"
                        @click="handleDeleteChallenge(challenge.id)"
                    />
                </div>
            </div>
        </div>
    </div>
</UCard>
```

**Step 4: Add i18n keys**

In the relevant locale files (search for existing duel keys location), add:
```
duel.expiredChallenges: "Истёкшие вызовы"
duel.expiredChallengeToInvitee: "Вызов {name} истёк"
duel.expiredChallengeFromInviter: "Вызов от {name} истёк"
duel.challengeExpired: "Время истекло"
duel.rechallenge: "Повторить вызов"
```

**Step 5: Verify build**
```bash
cd tma && pnpm build
```
Expected: no TypeScript errors

**Step 6: Commit**
```bash
git add tma/src/composables/usePvPDuel.ts \
        tma/src/views/Duel/DuelLobbyView.vue \
        tma/src/locales/
git commit -m "feat(pvp-duel): show expired challenges with re-challenge and delete buttons"
```

---

### Task F3: Update Swagger annotations

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`

**Step 1: Update `GetDuelStatusResponseSwagger`** to include `expired_challenges` field.

Find the Swagger model that maps to `GetDuelStatusOutput` and add:
```go
ExpiredChallenges []ChallengeDTOSwagger `json:"expiredChallenges" example:"[]"`
```

**Step 2: Regenerate**
```bash
cd backend && make swagger && cd ../tma && pnpm run generate:all
```

**Step 3: Commit**
```bash
git add backend/docs/ backend/internal/infrastructure/http/handlers/swagger_models.go \
        tma/src/api/generated/
git commit -m "chore: update Swagger annotations for expiredChallenges field"
```
