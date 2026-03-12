# PvP Challenge Flow Redesign — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Реализовать асинхронный challenge lifecycle: подтверждение при принятии вызова, исходящие карточки, Telegram уведомления, кнопка "Начать" у инвайтера.

**Architecture:** Добавляем статус `accepted_waiting_inviter` в domain. `AcceptByLinkCode` теперь НЕ создаёт игру — только уведомляет инвайтера. Новый `StartChallenge` use case создаёт игру когда инвайтер подтверждает. Frontend показывает модал перед принятием и карточки исходящих вызовов.

**Tech Stack:** Go (Fiber, DDD), Vue 3 + TypeScript, Telegram Bot API HTTP, Vitest

---

## Task 1: Domain — новый статус `accepted_waiting_inviter`

**Files:**
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go`
- Test: `backend/internal/domain/quick_duel/duel_game_aggregate_test.go`

**Контекст:** Сейчас у `DuelChallenge` есть статусы `pending`, `accepted`, `declined`, `expired`. Нам нужен промежуточный статус — инвайти принял, но инвайтер ещё не подтвердил старт.

**Step 1: Write failing test**

В `duel_game_aggregate_test.go` добавить в конец файла:
```go
func TestDuelChallenge_AcceptWaiting(t *testing.T) {
    challengerID, _ := shared.NewUserID("challenger-uuid-1234")
    now := int64(1706429000)
    challenge, _ := NewLinkChallenge(challengerID, now)

    accepterID, _ := shared.NewUserID("accepter-uuid-5678")
    err := challenge.AcceptWaiting(accepterID, "Vasya", now+10)

    assert.NoError(t, err)
    assert.Equal(t, ChallengeStatusAcceptedWaitingInviter, challenge.Status())
    assert.Equal(t, "Vasya", challenge.InviteeName())
    assert.NotNil(t, challenge.ChallengedID())
}
```

**Step 2: Run test to verify it fails**
```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestDuelChallenge_AcceptWaiting -v
```
Expected: FAIL — `AcceptWaiting undefined`

**Step 3: Implement**

В `duel_challenge.go` добавить:

```go
// В блок const ChallengeStatus:
ChallengeStatusAcceptedWaitingInviter ChallengeStatus = "accepted_waiting_inviter"

// Новое поле в DuelChallenge struct (после respondedAt):
inviteeName string

// Новый метод:
// AcceptWaiting sets status to accepted_waiting_inviter (invitee accepted, waiting for inviter to start).
// Used only for link-based challenges.
func (dc *DuelChallenge) AcceptWaiting(accepterID UserID, inviteeName string, acceptedAt int64) error {
    if dc.status != ChallengeStatusPending {
        return ErrChallengeNotPending
    }
    if dc.IsExpired(acceptedAt) {
        dc.status = ChallengeStatusExpired
        return ErrChallengeExpired
    }
    if dc.challengeType != ChallengeTypeLink {
        return ErrNotChallengedPlayer
    }
    if dc.challengerID.Equals(accepterID) {
        return ErrCannotChallengeSelf
    }
    dc.challengedID = &accepterID
    dc.inviteeName = inviteeName
    dc.status = ChallengeStatusAcceptedWaitingInviter
    dc.respondedAt = acceptedAt
    dc.events = append(dc.events, NewChallengeAcceptedEvent(
        dc.id, dc.challengerID, accepterID, acceptedAt,
    ))
    return nil
}

// Новый getter:
func (dc *DuelChallenge) InviteeName() string { return dc.inviteeName }
```

Обновить `ReconstructDuelChallenge` — добавить параметр `inviteeName string` (последним перед закрывающей скобкой) и присвоить `inviteeName: inviteeName`.

**Step 4: Run test**
```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestDuelChallenge_AcceptWaiting -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add backend/internal/domain/quick_duel/duel_challenge.go backend/internal/domain/quick_duel/duel_game_aggregate_test.go
git commit -m "feat(pvp-duel): add accepted_waiting_inviter status to DuelChallenge"
```

---

## Task 2: TelegramNotifier — интерфейс и реализации

**Files:**
- Create: `backend/internal/infrastructure/telegram/notifier.go`
- Create: `backend/internal/infrastructure/telegram/notifier_test.go`

**Контекст:** Бот должен уведомлять инвайтера когда друг принял вызов. `TELEGRAM_BOT_TOKEN` уже есть в `.env`. HTTP-запрос к `api.telegram.org/bot{TOKEN}/sendMessage`.

**Step 1: Write failing test**

Создать `backend/internal/infrastructure/telegram/notifier_test.go`:
```go
package telegram_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/barsukov/quiz-sprint/backend/internal/infrastructure/telegram"
)

func TestNoOpNotifier_DoesNotError(t *testing.T) {
    n := telegram.NewNoOpNotifier()
    err := n.NotifyChallengeAccepted(context.Background(), 123456, "@friend", "https://t.me/bot")
    assert.NoError(t, err)
}
```

**Step 2: Run test to verify it fails**
```bash
cd backend && go test ./internal/infrastructure/telegram/... -v
```
Expected: FAIL — package not found

**Step 3: Implement**

Создать `backend/internal/infrastructure/telegram/notifier.go`:
```go
package telegram

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

// TelegramNotifier sends notifications via Telegram Bot API.
type TelegramNotifier interface {
    NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
    NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
}

// NoOpNotifier does nothing (used in tests / when bot token is absent).
type NoOpNotifier struct{}

func NewNoOpNotifier() TelegramNotifier { return &NoOpNotifier{} }

func (n *NoOpNotifier) NotifyChallengeAccepted(_ context.Context, _ int64, _ string, _ string) error {
    return nil
}
func (n *NoOpNotifier) NotifyInviterWaiting(_ context.Context, _ int64, _ string, _ string) error {
    return nil
}

// HTTPNotifier sends real Telegram messages.
type HTTPNotifier struct {
    botToken string
    client   *http.Client
}

func NewHTTPNotifier(botToken string) TelegramNotifier {
    return &HTTPNotifier{botToken: botToken, client: &http.Client{}}
}

type sendMessageRequest struct {
    ChatID    int64  `json:"chat_id"`
    Text      string `json:"text"`
    ParseMode string `json:"parse_mode"`
}

func (n *HTTPNotifier) sendMessage(ctx context.Context, chatID int64, text string) error {
    body, _ := json.Marshal(sendMessageRequest{
        ChatID:    chatID,
        Text:      text,
        ParseMode: "HTML",
    })
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
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

func (n *HTTPNotifier) NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error {
    text := fmt.Sprintf("⚔️ <b>%s</b> принял твой вызов и готов к дуэли!\n\n<a href=\"%s\">Зайти в лобби →</a>", inviteeName, lobbyURL)
    return n.sendMessage(ctx, inviterTelegramID, text)
}

func (n *HTTPNotifier) NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error {
    text := fmt.Sprintf("⚔️ <b>%s</b> ждёт тебя в лобби!\n\n<a href=\"%s\">Зайти →</a>", inviterName, lobbyURL)
    return n.sendMessage(ctx, inviteeTelegramID, text)
}
```

**Step 4: Run tests**
```bash
cd backend && go test ./internal/infrastructure/telegram/... -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add backend/internal/infrastructure/telegram/
git commit -m "feat(pvp-duel): add TelegramNotifier with NoOp and HTTP implementations"
```

---

## Task 3: Postgres — FindPendingByChallenger включает `accepted_waiting_inviter`

**Files:**
- Modify: `backend/internal/infrastructure/persistence/postgres/challenge_repository.go`
- Modify: `backend/internal/domain/quick_duel/duel_challenge.go` (ReconstructDuelChallenge — добавить inviteeName из joined users)

**Контекст:** Сейчас `FindPendingByChallenger` фильтрует только `status = 'pending'`. Нам нужно также получать `accepted_waiting_inviter` вызовы чтобы инвайтер видел что друг принял.

**Step 1: Update query in `challenge_repository.go`**

Найти метод `FindPendingByChallenger` (строка ~118) и изменить SQL:
```go
// БЫЛО:
WHERE challenger_id = $1 AND status = 'pending'
// СТАЛО:
WHERE challenger_id = $1 AND status IN ('pending', 'accepted_waiting_inviter')
```

**Step 2: Add inviteeName resolution to GetDuelStatus**

В `backend/internal/application/quick_duel/use_cases.go` в методе `GetDuelStatusUseCase.Execute()`, в блоке где строятся `outgoingDTOs` (~строка 96-99):
```go
// БЫЛО:
outgoingDTOs = append(outgoingDTOs, ToChallengeDTO(c, now, ""))

// СТАЛО:
inviteeName := ""
if c.ChallengedID() != nil {
    if u, err := uc.userRepo.FindByID(*c.ChallengedID()); err == nil && u != nil {
        inviteeName = u.TelegramUsername().String()
        if inviteeName == "" {
            inviteeName = u.Username().String()
        }
    }
}
outgoingDTOs = append(outgoingDTOs, ToChallengeDTO(c, now, inviteeName))
```

В `dto.go` в `ChallengeDTO` добавить поле (после `ChallengerUsername`):
```go
InviteeName string `json:"inviteeName,omitempty"`
```

Найти функцию `ToChallengeDTO` в `use_cases.go` (или вынести в отдельный маппер) и убедиться что второй аргумент username используется как `InviteeName` в outgoing context. Фактически `ToChallengeDTO` принимает `username` и кладёт в `ChallengerUsername`. Нам нужно либо:
- Поменять сигнатуру чтобы принимать обоих, или
- Установить `InviteeName` отдельно после вызова функции

Проще всего: после `ToChallengeDTO` вызова присвоить поле напрямую:
```go
dto := ToChallengeDTO(c, now, "")
if c.ChallengedID() != nil {
    if u, err := uc.userRepo.FindByID(*c.ChallengedID()); err == nil && u != nil {
        name := u.TelegramUsername().String()
        if name == "" { name = u.Username().String() }
        dto.InviteeName = name
    }
}
outgoingDTOs = append(outgoingDTOs, dto)
```

**Step 3: Run all duel tests**
```bash
cd backend && go test ./internal/... -v 2>&1 | grep -E "PASS|FAIL|ok"
```
Expected: all PASS

**Step 4: Commit**
```bash
git add backend/internal/infrastructure/persistence/postgres/challenge_repository.go
git add backend/internal/application/quick_duel/use_cases.go
git add backend/internal/application/quick_duel/dto.go
git commit -m "feat(pvp-duel): include accepted_waiting_inviter in outgoing challenges + add inviteeName"
```

---

## Task 4: AcceptByLinkCodeUseCase — убрать авто-создание игры

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go`
- Modify: `backend/internal/application/quick_duel/dto.go`

**Контекст:** Сейчас `AcceptByLinkCodeUseCase.Execute()` (строка ~435-540) создаёт игру сразу. Нам нужно: установить статус `accepted_waiting_inviter`, уведомить инвайтера через Telegram, вернуть `challengeId` (без `gameId`).

**Step 1: Обновить `AcceptByLinkCodeOutput` в `dto.go`**

```go
// БЫЛО:
type AcceptByLinkCodeOutput struct {
    Success        bool    `json:"success"`
    GameID         *string `json:"gameId,omitempty"`
    TicketConsumed bool    `json:"ticketConsumed"`
    StartsIn       *int    `json:"startsIn,omitempty"`
    ChallengerID   string  `json:"challengerId"`
}

// СТАЛО:
type AcceptByLinkCodeOutput struct {
    Success     bool   `json:"success"`
    ChallengeID string `json:"challengeId"`
    Status      string `json:"status"` // "accepted_waiting_inviter"
}
```

**Step 2: Добавить `TelegramNotifier` в `AcceptByLinkCodeUseCase`**

```go
type AcceptByLinkCodeUseCase struct {
    challengeRepo    quick_duel.ChallengeRepository
    userRepo         domainUser.UserRepository
    notifier         TelegramNotifier  // добавить
    eventBus         EventBus
    // убрать: duelGameRepo, playerRatingRepo, seasonRepo, questionRepo
}
```

Упростить конструктор соответственно.

**Step 3: Переписать `Execute`**

```go
func (uc *AcceptByLinkCodeUseCase) Execute(input AcceptByLinkCodeInput) (AcceptByLinkCodeOutput, error) {
    now := time.Now().UTC().Unix()

    accepterID, err := shared.NewUserID(input.PlayerID)
    if err != nil {
        return AcceptByLinkCodeOutput{}, err
    }

    // Find challenge by link code
    challenge, err := uc.challengeRepo.FindByLinkCode(input.LinkCode)
    if err != nil {
        return AcceptByLinkCodeOutput{}, err
    }

    // Get accepter's display name
    inviteeName := accepterID.String()
    if u, err := uc.userRepo.FindByID(accepterID); err == nil && u != nil {
        if n := u.TelegramUsername().String(); n != "" {
            inviteeName = n
        } else if n := u.Username().String(); n != "" {
            inviteeName = n
        }
    }

    // Set status to accepted_waiting_inviter
    if err := challenge.AcceptWaiting(accepterID, inviteeName, now); err != nil {
        return AcceptByLinkCodeOutput{}, err
    }

    if err := uc.challengeRepo.Save(challenge); err != nil {
        return AcceptByLinkCodeOutput{}, err
    }

    for _, event := range challenge.Events() {
        uc.eventBus.Publish(event)
    }

    // Notify inviter via Telegram (best-effort — do not fail if notification errors)
    challengerID := challenge.ChallengerID()
    if u, err := uc.userRepo.FindByID(challengerID); err == nil && u != nil {
        if tgID := u.TelegramID(); tgID > 0 {
            lobbyURL := "https://t.me/quiz_sprint_dev_bot?startapp=lobby"
            _ = uc.notifier.NotifyChallengeAccepted(context.Background(), tgID, inviteeName, lobbyURL)
        }
    }

    return AcceptByLinkCodeOutput{
        Success:     true,
        ChallengeID: challenge.ID().String(),
        Status:      string(quick_duel.ChallengeStatusAcceptedWaitingInviter),
    }, nil
}
```

> Примечание: `u.TelegramID()` — нужно проверить что такой метод есть у `User` entity. Если нет — добавить getter который возвращает `int64` из `telegramID` поля.

**Step 4: Добавить `TelegramNotifier` интерфейс в application layer**

В пакете `application/quick_duel` создать интерфейс (можно в `use_cases.go` вверху):
```go
// TelegramNotifier sends Telegram notifications (defined in infrastructure/telegram).
type TelegramNotifier interface {
    NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
    NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
}
```

**Step 5: Update routes.go — упростить инициализацию AcceptByLinkCodeUseCase**

В `routes.go` найти `acceptByLinkCodeUC = appDuel.NewAcceptByLinkCodeUseCase(...)` и обновить аргументы (убрать `duelGameRepo`, `playerRatingRepo`, `seasonRepo`, `questionRepo`; добавить `telegramNotifier`).

`telegramNotifier` объявить в начале duel-секции:
```go
var telegramNotifier appDuel.TelegramNotifier = telegram.NewNoOpNotifier()
if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
    telegramNotifier = telegram.NewHTTPNotifier(token)
}
```

Добавить импорт `"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/telegram"` и `"os"`.

**Step 6: Run tests**
```bash
cd backend && go test ./internal/... 2>&1 | grep -E "PASS|FAIL|ok|error"
```
Expected: all PASS, no compile errors

**Step 7: Commit**
```bash
git add backend/internal/application/quick_duel/use_cases.go
git add backend/internal/application/quick_duel/dto.go
git add backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): AcceptByLinkCode now sets accepted_waiting_inviter + fires Telegram notification"
```

---

## Task 5: StartChallengeUseCase — инвайтер запускает игру

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go` (добавить новый use case)
- Modify: `backend/internal/application/quick_duel/dto.go` (добавить Input/Output)
- Test: `backend/internal/application/quick_duel/use_cases_test.go`

**Контекст:** Новый use case. Инвайтер нажимает "Начать дуэль" → вызываем этот use case → создаётся `DuelGame`. Логика создания игры переносится из `AcceptByLinkCode`.

**Step 1: Добавить DTOs в `dto.go`**

```go
// ========================================
// StartChallenge Use Case
// ========================================

type StartChallengeInput struct {
    PlayerID    string `json:"playerId"`
    ChallengeID string `json:"challengeId"`
}

type StartChallengeOutput struct {
    GameID string `json:"gameId"`
}
```

**Step 2: Write failing test в `use_cases_test.go`**

```go
func TestStartChallengeUseCase_CreatesGame(t *testing.T) {
    // Проверяем что use case существует и принимает правильные параметры
    // (интеграционный тест — пропускаем без БД)
    t.Skip("integration test — requires DB")
}
```

**Step 3: Implement `StartChallengeUseCase`**

В конце `use_cases.go` добавить:

```go
// ========================================
// StartChallenge Use Case
// ========================================

type StartChallengeUseCase struct {
    challengeRepo    quick_duel.ChallengeRepository
    duelGameRepo     quick_duel.DuelGameRepository
    playerRatingRepo quick_duel.PlayerRatingRepository
    seasonRepo       quick_duel.SeasonRepository
    questionRepo     QuestionRepository
    userRepo         domainUser.UserRepository
    eventBus         EventBus
}

func NewStartChallengeUseCase(
    challengeRepo quick_duel.ChallengeRepository,
    duelGameRepo quick_duel.DuelGameRepository,
    playerRatingRepo quick_duel.PlayerRatingRepository,
    seasonRepo quick_duel.SeasonRepository,
    questionRepo QuestionRepository,
    userRepo domainUser.UserRepository,
    eventBus EventBus,
) *StartChallengeUseCase {
    return &StartChallengeUseCase{
        challengeRepo:    challengeRepo,
        duelGameRepo:     duelGameRepo,
        playerRatingRepo: playerRatingRepo,
        seasonRepo:       seasonRepo,
        questionRepo:     questionRepo,
        userRepo:         userRepo,
        eventBus:         eventBus,
    }
}

func (uc *StartChallengeUseCase) Execute(input StartChallengeInput) (StartChallengeOutput, error) {
    now := time.Now().UTC().Unix()

    inviterID, err := shared.NewUserID(input.PlayerID)
    if err != nil {
        return StartChallengeOutput{}, err
    }

    challengeID := quick_duel.NewChallengeIDFromString(input.ChallengeID)
    challenge, err := uc.challengeRepo.FindByID(challengeID)
    if err != nil {
        return StartChallengeOutput{}, err
    }

    // Validate inviter is the challenger
    if !challenge.ChallengerID().Equals(inviterID) {
        return StartChallengeOutput{}, quick_duel.ErrNotChallengedPlayer
    }

    // Validate status
    if challenge.Status() != quick_duel.ChallengeStatusAcceptedWaitingInviter {
        return StartChallengeOutput{}, quick_duel.ErrChallengeNotPending
    }

    if challenge.ChallengedID() == nil {
        return StartChallengeOutput{}, quick_duel.ErrChallengeNotFound
    }

    accepterID := *challenge.ChallengedID()
    seasonID, _ := uc.seasonRepo.GetCurrentSeason()

    rating1, err := uc.playerRatingRepo.FindOrCreate(inviterID, seasonID, now)
    if err != nil {
        return StartChallengeOutput{}, err
    }
    rating2, err := uc.playerRatingRepo.FindOrCreate(accepterID, seasonID, now)
    if err != nil {
        return StartChallengeOutput{}, err
    }

    // Get usernames
    inviterName := inviterID.String()
    if u, err := uc.userRepo.FindByID(inviterID); err == nil && u != nil {
        if n := u.TelegramUsername().String(); n != "" {
            inviterName = n
        }
    }
    accepterName := accepterID.String()
    if u, err := uc.userRepo.FindByID(accepterID); err == nil && u != nil {
        if n := u.TelegramUsername().String(); n != "" {
            accepterName = n
        }
    }

    // Select random questions
    questions, err := uc.questionRepo.FindRandomByDifficulty(quick_duel.QuestionsPerDuel, "medium")
    if err != nil {
        return StartChallengeOutput{}, err
    }

    questionIDs := make([]quick_duel.QuestionID, 0, len(questions))
    for _, q := range questions {
        qid, _ := quiz.NewQuestionIDFromString(q.ID)
        questionIDs = append(questionIDs, qid)
    }

    player1 := quick_duel.NewDuelPlayer(inviterID, inviterName, quick_duel.ReconstructEloRating(rating1.MMR(), 0))
    player2 := quick_duel.NewDuelPlayer(accepterID, accepterName, quick_duel.ReconstructEloRating(rating2.MMR(), 0))

    game, err := quick_duel.NewDuelGame(player1, player2, questionIDs, now)
    if err != nil {
        return StartChallengeOutput{}, err
    }
    if err := game.Start(now); err != nil {
        return StartChallengeOutput{}, err
    }
    if err := uc.duelGameRepo.Save(game); err != nil {
        return StartChallengeOutput{}, err
    }

    // Mark challenge as accepted (game started)
    challenge.SetMatchID(game.ID())
    _ = uc.challengeRepo.Save(challenge)

    return StartChallengeOutput{GameID: game.ID().String()}, nil
}
```

**Step 4: Run tests**
```bash
cd backend && go test ./internal/... 2>&1 | grep -E "PASS|FAIL|ok"
```
Expected: all PASS

**Step 5: Commit**
```bash
git add backend/internal/application/quick_duel/use_cases.go backend/internal/application/quick_duel/dto.go
git commit -m "feat(pvp-duel): add StartChallengeUseCase — inviter confirms game start"
```

---

## Task 6: HTTP Handler и Route для StartChallenge

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go`
- Modify: `backend/internal/infrastructure/http/routes/routes.go`

**Step 1: Добавить `startChallengeUC` в `DuelHandler`**

В `duel_handlers.go`:
```go
// В структуре DuelHandler добавить:
startChallengeUC *appDuel.StartChallengeUseCase

// В NewDuelHandler добавить параметр и присвоение
```

**Step 2: Добавить handler**

```go
// StartChallenge handles POST /api/v1/duel/challenge/:challengeId/start
// @Summary Start the duel after invitee accepted
// @Description Inviter confirms game start after invitee accepted via link
// @Tags duel
// @Accept json
// @Produce json
// @Param challengeId path string true "Challenge ID"
// @Param request body StartChallengeRequest true "Start request"
// @Success 200 {object} StartChallengeResponse "Game started"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Challenge not found"
// @Failure 409 {object} ErrorResponse "Challenge not in accepted_waiting_inviter state"
// @Router /duel/challenge/{challengeId}/start [post]
func (h *DuelHandler) StartChallenge(c fiber.Ctx) error {
    challengeID := c.Params("challengeId")
    if _, err := uuid.Parse(challengeID); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID format")
    }

    var req StartChallengeRequest
    if err := c.Bind().Body(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }
    if req.PlayerID == "" {
        return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
    }

    output, err := h.startChallengeUC.Execute(appDuel.StartChallengeInput{
        PlayerID:    req.PlayerID,
        ChallengeID: challengeID,
    })
    if err != nil {
        return mapDuelError(err)
    }

    return c.JSON(fiber.Map{"data": output})
}
```

**Step 3: Добавить Swagger модели в `swagger_models.go`**

```go
// StartChallengeRequest - request to start a challenge
type StartChallengeRequest struct {
    PlayerID string `json:"playerId" validate:"required"`
}

// StartChallengeResponse - game started
type StartChallengeResponse struct {
    Data struct {
        GameID string `json:"gameId"`
    } `json:"data"`
}
```

**Step 4: Зарегистрировать route в `routes.go`**

Найти блок duel routes (строка ~658-671) и добавить:
```go
duel.Post("/challenge/:challengeId/start", duelHandler.StartChallenge)
```

Также инициализировать `startChallengeUC` в routes.go и передать в `NewDuelHandler`. По аналогии с `acceptByLinkCodeUC` — нужен `questionRepo`.

**Step 5: Generate swagger + compile check**
```bash
cd backend && make swagger 2>&1 | tail -5
cd backend && go build ./... 2>&1
```
Expected: no errors

**Step 6: Commit**
```bash
git add backend/internal/infrastructure/http/handlers/duel_handlers.go
git add backend/internal/infrastructure/http/handlers/swagger_models.go
git add backend/internal/infrastructure/http/routes/routes.go
git commit -m "feat(pvp-duel): add POST /duel/challenge/:id/start endpoint"
```

---

## Task 7: Frontend — регенерация типов

**Files:**
- Run: `tma/` — генерация

**Step 1: Regenerate**
```bash
cd tma && pnpm run generate:all
```
Expected: no errors, обновлены файлы в `tma/src/api/generated/`

**Step 2: Проверить новые типы**

```bash
grep -r "challengeId\|StartChallenge\|accepted_waiting" tma/src/api/generated/ | head -10
```

**Step 3: Commit**
```bash
git add tma/src/api/generated/
git commit -m "chore(tma): regenerate API types for challenge flow redesign"
```

---

## Task 8: Frontend — usePvPDuel: добавить startChallenge + outgoingChallenges

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts`

**Контекст:** Сейчас polling отслеживает только `hasActiveDuel`. Нужно также отслеживать `outgoingChallenges` со статусом `accepted_waiting_inviter`. Добавить `startChallenge(challengeId)` функцию.

**Step 1: Добавить импорт нового хука**

Найти блок импортов generated хуков и добавить:
```ts
import { usePostDuelChallengeChallengeidStart } from '@/api/generated/hooks/duelController/usePostDuelChallengeChallengeidStart'
```
> Точное имя хука проверить в `tma/src/api/generated/hooks/duelController/` после регенерации.

**Step 2: Добавить `outgoingChallenges` computed**

После объявления query-хуков добавить:
```ts
const outgoingReadyChallenges = computed(() =>
  (statusData.value?.data?.outgoingChallenges ?? []).filter(
    (c) => c.status === 'accepted_waiting_inviter'
  )
)

const outgoingPendingChallenges = computed(() =>
  (statusData.value?.data?.outgoingChallenges ?? []).filter(
    (c) => c.status === 'pending'
  )
)
```

**Step 3: Добавить `startChallenge` функцию**

```ts
const { mutateAsync: startChallengeMutation } = usePostDuelChallengeChallengeidStart()

const startChallenge = async (challengeId: string) => {
  if (!playerId) return
  const response = await startChallengeMutation({
    params: { challengeId },
    data: { playerId },
  })
  if (response.data?.gameId) {
    router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
  }
}
```

**Step 4: Обновить polling**

Обновить `startOutgoingPoll` чтобы также проверял `outgoingReadyChallenges` (опционально — они и так видны через status, polling уже есть).

**Step 5: Экспортировать новые значения**

В `return` composable добавить:
```ts
outgoingChallenges: statusData.value?.data?.outgoingChallenges ?? [],
outgoingReadyChallenges,
outgoingPendingChallenges,
startChallenge,
```

**Step 6: Commit**
```bash
git add tma/src/composables/usePvPDuel.ts
git commit -m "feat(pvp-duel): add startChallenge + outgoingReadyChallenges to usePvPDuel"
```

---

## Task 9: Frontend — DuelLobbyView: модал подтверждения + карточки исходящих

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Контекст:** Сейчас при переходе по ссылке `handleAcceptByLinkCode` вызывается автоматически без подтверждения. Нужно показать модал. Также добавить карточки исходящих вызовов.

**Step 1: Добавить стейт модала**

В `<script setup>`:
```ts
const showConfirmModal = ref(false)
const pendingLinkCode = ref<string | null>(null)
const challengerInfo = ref<{ username: string } | null>(null)
```

**Step 2: Изменить логику deep link**

Найти в `onMounted` блок:
```ts
// Check for deep link challenge
const challengeCode = route.query.challenge as string
if (challengeCode) {
  deepLinkChallenge.value = challengeCode
  // Auto-accept the challenge        ← УБРАТЬ авто-принятие
  await handleAcceptByLinkCode(challengeCode)
}
```

Заменить на:
```ts
const challengeCode = route.query.challenge as string
if (challengeCode) {
  pendingLinkCode.value = challengeCode
  showConfirmModal.value = true   // показываем модал
}
```

**Step 3: Добавить функции модала**

```ts
const handleConfirmChallenge = async () => {
  if (!pendingLinkCode.value) return
  showConfirmModal.value = false
  await handleAcceptByLinkCode(pendingLinkCode.value)
  pendingLinkCode.value = null
}

const handleDeclineChallenge = () => {
  showConfirmModal.value = false
  pendingLinkCode.value = null
  router.replace({ name: 'duel-lobby' })
}
```

**Step 4: Импортировать и использовать startChallenge**

```ts
const {
  // ... существующие
  outgoingReadyChallenges,
  outgoingPendingChallenges,
  startChallenge,
} = usePvPDuel(playerId.value)
```

**Step 5: Добавить в template — модал подтверждения**

После блока `<!-- Deep Link Error -->` добавить:
```html
<!-- Confirmation Modal -->
<UModal v-model:open="showConfirmModal" :dismissible="false">
  <template #content>
    <div class="p-6 text-center">
      <div class="text-4xl mb-4">⚔️</div>
      <h3 class="text-xl font-bold mb-2">{{ t('duel.incomingChallenge') }}</h3>
      <p class="text-gray-600 dark:text-gray-400 mb-6">
        {{ t('duel.wantsToFight') }}
      </p>
      <div class="space-y-3">
        <UButton block size="lg" color="primary" :loading="isAcceptingChallenge" @click="handleConfirmChallenge">
          {{ t('duel.acceptChallenge') }}
        </UButton>
        <UButton block size="lg" color="gray" variant="ghost" @click="handleDeclineChallenge">
          {{ t('duel.decline') }}
        </UButton>
      </div>
    </div>
  </template>
</UModal>
```

**Step 6: Добавить карточки исходящих вызовов в Play tab**

Найти в template блок `<!-- Invite Friend -->` и добавить перед ним:

```html
<!-- Outgoing Challenges -->
<div v-if="outgoingReadyChallenges.length > 0 || outgoingPendingChallenges.length > 0" class="space-y-2">
  <h3 class="text-sm font-semibold text-gray-600 dark:text-gray-400">
    {{ t('duel.outgoingChallenges') }}
  </h3>

  <!-- Ready to start -->
  <UCard
    v-for="challenge in outgoingReadyChallenges"
    :key="challenge.id"
    class="border-green-200 dark:border-green-800"
  >
    <div class="flex items-center gap-2 mb-3">
      <span class="text-green-500">✅</span>
      <p class="font-medium">
        {{ challenge.inviteeName || t('duel.friend') }} {{ t('duel.isReady') }}
      </p>
    </div>
    <UButton color="green" block @click="() => startChallenge(challenge.id!)">
      {{ t('duel.startDuel') }}
    </UButton>
  </UCard>

  <!-- Waiting for response -->
  <UCard v-for="challenge in outgoingPendingChallenges" :key="challenge.id">
    <div class="flex items-center gap-2">
      <UIcon name="i-heroicons-paper-airplane" class="size-5 text-blue-500" />
      <div class="flex-1">
        <p class="font-medium">{{ t('duel.waitingForResponse') }}</p>
        <p class="text-xs text-gray-500">
          {{ t('duel.linkExpiresIn', { time: formatExpiry(challenge.expiresAt) }) }}
        </p>
      </div>
    </div>
  </UCard>
</div>
```

Добавить helper `formatExpiry`:
```ts
const formatExpiry = (expiresAt: number) => {
  const diff = expiresAt - Math.floor(Date.now() / 1000)
  if (diff <= 0) return t('duel.expired')
  const hours = Math.floor(diff / 3600)
  const minutes = Math.floor((diff % 3600) / 60)
  return hours > 0 ? `${hours}ч ${minutes}мин` : `${minutes}мин`
}
```

**Step 7: Добавить i18n ключи** в `tma/src/i18n/` (ru.json и en.json):
```json
"duel.incomingChallenge": "Тебя вызывают на дуэль!",
"duel.wantsToFight": "Хочешь принять вызов?",
"duel.acceptChallenge": "Принять вызов",
"duel.outgoingChallenges": "Мои вызовы",
"duel.isReady": "готов к дуэли!",
"duel.startDuel": "Начать дуэль →",
"duel.waitingForResponse": "Ожидаем ответа...",
"duel.linkExpiresIn": "Ссылка истекает через {time}",
"duel.expired": "Истекла",
"duel.friend": "Друг"
```

**Step 8: Lint + type-check**
```bash
cd tma && pnpm lint && pnpm run type-check
```
Expected: no errors

**Step 9: Commit**
```bash
git add tma/src/views/Duel/DuelLobbyView.vue tma/src/i18n/
git commit -m "feat(pvp-duel): add confirmation modal + outgoing challenge cards in lobby"
```

---

## Task 10: Frontend — DuelCard: бейдж для готовых вызовов

**Files:**
- Modify: `tma/src/components/Duel/DuelCard.vue`

**Контекст:** Сейчас бейдж показывает только `pendingChallenges.length`. Нужно добавить счётчик `outgoingReadyChallenges` — когда друг готов, это важно показать.

**Step 1: Обновить composable вызов**

```ts
const {
  // ... существующие
  outgoingReadyChallenges,
  initialize,
} = usePvPDuel(props.playerId)
```

**Step 2: Обновить `totalAlerts` computed**

```ts
const totalAlerts = computed(
  () => pendingChallenges.value.length + outgoingReadyChallenges.value.length
)
```

**Step 3: Обновить template**

Заменить `pendingChallenges.length > 0` на `totalAlerts > 0`:
```html
<!-- header badge -->
<UBadge v-if="totalAlerts > 0" color="orange" variant="soft" size="sm">
  {{ totalAlerts }}
</UBadge>

<!-- button disabled condition -->
:disabled="!canPlay && !hasActiveDuel && totalAlerts === 0"
```

Обновить `buttonText`:
```ts
const buttonText = computed(() => {
  if (hasActiveDuel.value) return t('duel.continueDuel')
  if (outgoingReadyChallenges.value.length > 0) return t('duel.friendReady')  // новый ключ
  if (pendingChallenges.value.length > 0) return t('duel.challengesCount', { count: pendingChallenges.value.length })
  return t('duel.findOpponent')
})
```

Добавить i18n ключ: `"duel.friendReady": "Друг готов к дуэли! →"`

**Step 4: Lint + type-check**
```bash
cd tma && pnpm lint && pnpm run type-check
```

**Step 5: Commit**
```bash
git add tma/src/components/Duel/DuelCard.vue tma/src/i18n/
git commit -m "feat(pvp-duel): update DuelCard badge to include outgoing ready challenges"
```

---

## Task 11: Финальная проверка

**Step 1: Backend tests**
```bash
cd backend && go test ./... 2>&1 | grep -E "PASS|FAIL|ok"
```
Expected: all PASS

**Step 2: Frontend build**
```bash
cd tma && pnpm build 2>&1 | tail -10
```
Expected: no errors

**Step 3: Запустить dev окружение и проверить вручную**
```bash
# Terminal 1
cd backend && docker compose -f docker-compose.dev.yml up

# Terminal 2
cd tma && pnpm dev
```

Проверить сценарии:
- [ ] Создать ссылку → скопировать → открыть в другом браузере/инкогнито → показывается модал
- [ ] Принять вызов → в первом браузере появляется карточка "✅ Друг готов!"
- [ ] Нажать "Начать дуэль" → оба переходят в игру
- [ ] Бейдж на главном экране обновляется
- [ ] Карточка "Ожидаем ответа..." отображается для pending вызовов

**Step 4: Push**
```bash
git push origin pvp-duel
```
