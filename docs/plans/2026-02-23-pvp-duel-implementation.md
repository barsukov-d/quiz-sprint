# PvP Duel — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Довести PvP Duel до рабочего end-to-end состояния, где два игрока могут провести полный матч в реальном времени с корректной логикой ответов, сохранением результатов и экраном итогов.

**Architecture:** Domain layer полностью готов (game aggregate, ELO, challenge, referral). Application layer ~80%. Основные пробелы: WebSocket hub (30%), Answer validation (placeholder), Postgres repo (40%), Frontend results view (0%). Реализуем слой за слоем, начиная с критического пути "сыграть дуэль до конца".

**Tech Stack:** Go + Fiber + Redis + PostgreSQL (backend), Vue 3 + TypeScript + WebSocket (frontend). Пакет домена: `quick_duel` (не `pvp_duel` — расхождение с доками, не трогаем).

---

## Приоритет задач

```
PHASE A: Core playable duel     ← КРИТИЧЕСКИЙ ПУТЬ
PHASE B: Frontend game flow     ← КРИТИЧЕСКИЙ ПУТЬ
PHASE C: Ticket system          ← нужен для корректной экономики
PHASE D: Missing API endpoints  ← нужен для результатов и сдачи
PHASE E: Friends & Social       ← деферд, но незаглушенные TODOs
PHASE F: Seasons & Rewards      ← деферд
PHASE G: Push notifications     ← деферд
```

---

## PHASE A: Core Playable Duel

### Task 1: PostgreSQL — добавить FindActiveByPlayer и Paginate

**Проблема:** `GetDuelStatus` не может найти активную игру игрока, история матчей не работает.

**Files:**
- Modify: `backend/internal/infrastructure/persistence/postgres/duel_game_repository.go`

**Step 1: Найди метод-заглушку `FindActiveByPlayer`**

```bash
grep -n "FindActiveByPlayer\|TODO\|not implemented" backend/internal/infrastructure/persistence/postgres/duel_game_repository.go
```

**Step 2: Реализуй FindActiveByPlayer**

```go
func (r *DuelGameRepository) FindActiveByPlayer(ctx context.Context, playerID domain.PlayerID) (*domain.DuelGame, error) {
    query := `
        SELECT id, status, player1_id, player2_id, questions_data, answers_data,
               player1_score, player2_score, player1_mmr_before, player2_mmr_before,
               started_at, completed_at, is_friend_game, challenge_id
        FROM duel_matches
        WHERE (player1_id = $1 OR player2_id = $1)
          AND status IN ('waiting_start', 'in_progress')
        ORDER BY started_at DESC
        LIMIT 1
    `
    row := r.db.QueryRowContext(ctx, query, string(playerID))
    return r.scanGame(row)
}
```

**Step 3: Реализуй FindByPlayerPaginated**

```go
func (r *DuelGameRepository) FindByPlayerPaginated(
    ctx context.Context,
    playerID domain.PlayerID,
    limit, offset int,
    filter string, // "all", "wins", "losses", "friends"
) ([]*domain.DuelGame, int, error) {
    whereClauses := []string{"(player1_id = $1 OR player2_id = $1)", "status = 'finished'"}
    args := []interface{}{string(playerID)}
    argN := 2

    switch filter {
    case "wins":
        whereClauses = append(whereClauses, fmt.Sprintf("winner_id = $%d", argN))
        args = append(args, string(playerID))
        argN++
    case "losses":
        whereClauses = append(whereClauses, fmt.Sprintf("winner_id != $%d AND winner_id IS NOT NULL", argN))
        args = append(args, string(playerID))
        argN++
    case "friends":
        whereClauses = append(whereClauses, "is_friend_game = true")
    }

    countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM duel_matches WHERE %s`, strings.Join(whereClauses, " AND "))
    // ... full paginated query
}
```

**Step 4: Запусти тесты репозитория**

```bash
cd backend && go test ./internal/infrastructure/persistence/postgres/... -v -run TestDuel
```

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/persistence/postgres/duel_game_repository.go
git commit -m "feat(pvp-duel): implement FindActiveByPlayer and paginated history in repo"
```

---

### Task 2: Redis — сохранять ответы раундов вместо in-memory map

**Проблема:** `SubmitDuelAnswer` хранит ответы раундов в `sync.Map` в памяти процесса. При перезапуске или нескольких инстансах — данные теряются.

**Files:**
- Modify: `backend/internal/application/quick_duel/submit_duel_answer.go`
- Modify: `backend/internal/infrastructure/persistence/redis/matchmaking_queue.go` (или создать `duel_game_cache.go`)

**Step 1: Найди in-memory хранилище ответов**

```bash
grep -n "sync.Map\|roundAnswers\|in-memory\|TODO.*redis\|TODO.*Redis" backend/internal/application/quick_duel/submit_duel_answer.go
```

**Step 2: Создай Redis-хелпер для раундов**

```go
// backend/internal/infrastructure/persistence/redis/duel_round_cache.go

type DuelRoundCache struct {
    client *redis.Client
}

func (c *DuelRoundCache) SetAnswer(ctx context.Context, gameID, playerID string, round int, answer RoundAnswer) error {
    key := fmt.Sprintf("duel:round:%s:%d", gameID, round)
    field := playerID
    data, _ := json.Marshal(answer)
    return c.client.HSet(ctx, key, field, data).Err()
}

func (c *DuelRoundCache) GetBothAnswers(ctx context.Context, gameID string, round int) (p1, p2 *RoundAnswer, err error) {
    key := fmt.Sprintf("duel:round:%s:%d", gameID, round)
    vals, err := c.client.HGetAll(ctx, key).Result()
    // parse and return
}

// TTL = 24h (game results only needed during active game + brief review)
func (c *DuelRoundCache) SetTTL(ctx context.Context, gameID string) {
    for i := 0; i < 7; i++ {
        key := fmt.Sprintf("duel:round:%s:%d", gameID, i)
        c.client.Expire(ctx, key, 24*time.Hour)
    }
}
```

**Step 3: Замени sync.Map на Redis в SubmitDuelAnswer use case**

```go
// Было:
r.roundAnswers.Store(key, answer)

// Стало:
if err := r.roundCache.SetAnswer(ctx, gameID, playerID, round, answer); err != nil {
    return nil, fmt.Errorf("cache answer: %w", err)
}
```

**Step 4: Запусти тесты**

```bash
cd backend && go test ./internal/application/quick_duel/... -v -run TestSubmitAnswer
```

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/submit_duel_answer.go \
        backend/internal/infrastructure/persistence/redis/duel_round_cache.go
git commit -m "feat(pvp-duel): persist round answers to Redis instead of memory"
```

---

### Task 3: Исправить валидацию ответов

**Проблема:** `SubmitDuelAnswer` не проверяет правильность ответа — `isCorrect` всегда `true` (placeholder).

**Files:**
- Modify: `backend/internal/application/quick_duel/submit_duel_answer.go`

**Step 1: Найди placeholder**

```bash
grep -n "isCorrect\|always.*correct\|TODO.*valid\|placeholder" backend/internal/application/quick_duel/submit_duel_answer.go
```

**Step 2: Получи правильный ответ из вопроса**

```go
// В use case уже должен быть доступ к DuelGame (через repo).
// DuelGame.Questions[roundIndex].CorrectAnswerID → сравни с AnswerID из запроса.

game, err := uc.gameRepo.FindByID(ctx, domain.GameID(input.GameID))
if err != nil {
    return nil, err
}

question := game.GetCurrentQuestion()
isCorrect := question.CorrectAnswerID == domain.AnswerID(input.AnswerID)
```

**Step 3: Напиши тест**

```go
func TestSubmitDuelAnswer_WrongAnswer(t *testing.T) {
    // setup game with known questions
    result, err := uc.Execute(ctx, SubmitDuelAnswerInput{
        GameID:   "game_1",
        PlayerID: "player_1",
        AnswerID: "wrong_answer_id",
    })
    require.NoError(t, err)
    assert.False(t, result.IsCorrect)
    assert.Equal(t, 0, result.Points) // no points for wrong answer
}
```

**Step 4: Запусти тесты**

```bash
cd backend && go test ./internal/application/quick_duel/... -v -run TestSubmitDuelAnswer
```

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/submit_duel_answer.go
git commit -m "fix(pvp-duel): validate answer correctness against question data"
```

---

### Task 4: Завершить WebSocket Hub — полный message routing

**Проблема:** `DuelWebSocketHub` — 30% реализован. Есть структура подключений, но нет полного роутинга сообщений по документации `05_api.md`.

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/duel_websocket_handler.go`

**Ожидаемые message types (Server → Client):**

| Type | Когда отправлять |
|------|-----------------|
| `connected` | При подключении к хабу |
| `game_ready` | Оба игрока подключены |
| `new_question` | Начало каждого раунда (serverTime в payload) |
| `answer_result` | Когда любой игрок ответил |
| `round_complete` | Оба ответили, переход к следующему |
| `round_timeout` | Время вышло |
| `game_complete` | Игра завершена, MMR изменился |
| `opponent_disconnected` | Соперник отключился (с grace period) |

**Step 1: Посмотри текущее состояние хаба**

```bash
cat backend/internal/infrastructure/http/handlers/duel_websocket_handler.go
```

**Step 2: Реализуй метод `handleMessage` с роутингом**

```go
func (h *DuelWebSocketHub) handleMessage(gameID, playerID string, msg ClientMessage) {
    switch msg.Type {
    case "player_ready":
        h.handlePlayerReady(gameID, playerID)
    case "submit_answer":
        h.handleSubmitAnswer(gameID, playerID, msg.Data)
    case "ping":
        h.sendToPlayer(playerID, ServerMessage{Type: "pong"})
    default:
        log.Printf("unknown message type: %s", msg.Type)
    }
}
```

**Step 3: Реализуй broadcast new_question**

```go
func (h *DuelWebSocketHub) broadcastNewQuestion(gameID string, round int, question DuelQuestionDTO) {
    msg := ServerMessage{
        Type: "new_question",
        Data: map[string]interface{}{
            "roundNum":     round,
            "totalRounds":  7,
            "question":     question,
            "serverTime":   time.Now().UnixMilli(),
        },
    }
    h.broadcastToGame(gameID, msg)
}
```

**Step 4: Реализуй answer_result (отправить обоим игрокам)**

```go
func (h *DuelWebSocketHub) broadcastAnswerResult(gameID string, result AnswerResultDTO) {
    msg := ServerMessage{
        Type: "answer_result",
        Data: result,
    }
    // Отправить ОБОИМ игрокам — противник видит результат сразу
    h.broadcastToGame(gameID, msg)
}
```

**Step 5: Реализуй disconnect grace period**

```go
func (h *DuelWebSocketHub) handleDisconnect(gameID, playerID string) {
    h.broadcastToGame(gameID, ServerMessage{
        Type: "opponent_disconnected",
        Data: map[string]interface{}{
            "playerId":    playerID,
            "reconnectIn": 10,
        },
    })

    // Grace period — 10 секунд
    time.AfterFunc(10*time.Second, func() {
        if !h.isConnected(gameID, playerID) {
            h.submitEmptyAnswer(gameID, playerID)
        }
    })
}
```

**Step 6: Проверь WebSocket вручную через wscat**

```bash
# Установи wscat если нет:
npm install -g wscat

# Подключись:
wscat -c "ws://localhost:3000/ws/duel/g_test?playerId=user_1" \
  -H "Authorization: tma <base64_init_data>"

# Отправь:
{"type":"player_ready"}
# Ожидай: {"type":"game_ready",...}
```

**Step 7: Commit**

```bash
git add backend/internal/infrastructure/http/handlers/duel_websocket_handler.go
git commit -m "feat(pvp-duel): complete WebSocket hub message routing"
```

---

## PHASE B: Frontend Game Flow

### Task 5: Завершить DuelPlayView.vue

**Проблема:** Play view на 50% — нет обработки результатов раунда, feedback экрана, UI отключения соперника.

**Files:**
- Modify: `tma/src/views/Duel/DuelPlayView.vue`
- Modify: `tma/src/composables/useDuelWebSocket.ts`

**Step 1: Посмотри текущее состояние**

```bash
cat tma/src/views/Duel/DuelPlayView.vue
cat tma/src/composables/useDuelWebSocket.ts
```

**Step 2: Завершить обработку message types в useDuelWebSocket.ts**

```typescript
// Каждый WS message type должен обновлять реактивный state:
function handleMessage(msg: WebSocketMessage) {
  switch (msg.type) {
    case 'new_question':
      currentQuestion.value = msg.data.question
      currentRound.value = msg.data.roundNum
      questionStartTime.value = msg.data.serverTime
      opponentAnswered.value = false
      countdownSeconds.value = 10
      startCountdown()
      break

    case 'answer_result':
      if (msg.data.playerId !== myPlayerId) {
        opponentAnswered.value = true
        opponentLastResult.value = { isCorrect: msg.data.isCorrect }
      }
      break

    case 'round_complete':
      lastRoundResult.value = {
        myScore: msg.data.player1Score, // или player2Score в зависимости от playerID
        opponentScore: msg.data.player2Score,
        nextIn: msg.data.nextRoundIn,
      }
      break

    case 'game_complete':
      isFinished.value = true
      gameResult.value = msg.data
      break

    case 'opponent_disconnected':
      opponentReconnecting.value = true
      opponentReconnectCountdown.value = msg.data.reconnectIn
      break
  }
}
```

**Step 3: Добавить Answer Feedback (1.5s overlay)**

```vue
<!-- После выбора ответа — brief feedback overlay -->
<Transition name="fade">
  <div v-if="showFeedback" class="answer-feedback">
    <span v-if="lastAnswerCorrect">✅ ПРАВИЛЬНО!</span>
    <span v-else>❌ НЕВЕРНО</span>
    <div class="times">
      <span>Твоё время: {{ myAnswerTime }}с</span>
      <span v-if="opponentAnswered">{{ opponent.username }}: {{ opponentAnswerTime }}с</span>
    </div>
  </div>
</Transition>
```

**Step 4: Добавить Emote button (max 3/игра)**

```vue
<div class="emote-bar" v-if="emotesLeft > 0">
  <button
    v-for="emote in unlockedEmotes"
    :key="emote"
    @click="sendEmote(emote)"
    :disabled="emotesLeft === 0"
  >{{ emote }}</button>
  <span class="emotes-left">{{ emotesLeft }}/3</span>
</div>
```

**Step 5: Disconnect overlay**

```vue
<div v-if="opponentReconnecting" class="disconnect-overlay">
  🔄 {{ opponent.username }} переподключается...
  {{ opponentReconnectCountdown }}с
</div>
```

**Step 6: Запусти lint и type-check**

```bash
cd tma && pnpm lint && pnpm run type-check
```

**Step 7: Commit**

```bash
git add tma/src/views/Duel/DuelPlayView.vue tma/src/composables/useDuelWebSocket.ts
git commit -m "feat(pvp-duel): complete play view - round feedback, emotes, disconnect UI"
```

---

### Task 6: Создать DuelResultsView.vue

**Проблема:** `DuelResultsView.vue` не существует. Игра завершается, но некуда перейти.

**Files:**
- Create: `tma/src/views/Duel/DuelResultsView.vue`
- Modify: `tma/src/router/index.ts` (добавить route)

**Step 1: Проверь роутер**

```bash
grep -n "Duel\|duel" tma/src/router/index.ts
```

**Step 2: Добавь route для результатов**

```typescript
// В router/index.ts:
{
  path: '/duel/results/:gameId',
  name: 'DuelResults',
  component: () => import('@/views/Duel/DuelResultsView.vue'),
  props: true,
}
```

**Step 3: Создай компонент по wireframe из 02_gameplay.md**

```vue
<!-- DuelResultsView.vue -->
<template>
  <div class="duel-results">
    <!-- Win/Lose header -->
    <div class="result-header" :class="didWin ? 'win' : 'lose'">
      <h1>{{ didWin ? '🏆 ПОБЕДА!' : '💔 ПОРАЖЕНИЕ' }}</h1>
    </div>

    <!-- Score summary -->
    <div class="score-card">
      <div class="player">
        <Avatar :src="me.avatar" />
        <span>{{ me.username }}</span>
        <span class="score">{{ me.score }}</span>
      </div>
      <div class="vs">:</div>
      <div class="player">
        <span class="score">{{ opponent.score }}</span>
        <span>{{ opponent.username }}</span>
        <Avatar :src="opponent.avatar" />
      </div>
    </div>

    <!-- Time comparison -->
    <div class="times">
      <span>{{ formatTime(me.totalTime) }}</span>
      <span v-if="winReason === 'time'">⚡ победа по времени</span>
      <span>{{ formatTime(opponent.totalTime) }}</span>
    </div>

    <!-- Per-question breakdown -->
    <div class="questions-breakdown">
      <div v-for="(q, i) in questions" :key="i" class="question-row">
        <span>Q{{ i + 1 }}:</span>
        <span :class="q.myAnswer.isCorrect ? 'correct' : 'wrong'">
          {{ q.myAnswer.isCorrect ? '✅' : '❌' }} {{ formatTime(q.myAnswer.timeTaken) }}
        </span>
        <span :class="q.opponentAnswer.isCorrect ? 'correct' : 'wrong'">
          {{ q.opponentAnswer.isCorrect ? '✅' : '❌' }} {{ formatTime(q.opponentAnswer.timeTaken) }}
        </span>
      </div>
    </div>

    <!-- MMR change -->
    <div class="mmr-change">
      <span>MMR: {{ mmrBefore }} → {{ mmrAfter }}</span>
      <span :class="mmrDelta > 0 ? 'positive' : 'negative'">
        {{ mmrDelta > 0 ? '+' : '' }}{{ mmrDelta }}
      </span>
    </div>

    <!-- Actions -->
    <div class="actions">
      <button @click="requestRematch" :disabled="!rematchAvailable">РЕВАНШ</button>
      <button @click="goToLobby">В МЕНЮ</button>
    </div>

    <!-- Share (win only) -->
    <button v-if="didWin" @click="shareVictory" class="share-btn">
      📤 Поделиться
    </button>
  </div>
</template>
```

**Step 4: Используй GET /api/v1/duel/game/:gameId для данных**

```typescript
// usePvPDuel.ts или локально в компоненте:
const { data: gameResult } = useGetDuelGame(props.gameId)
```

**Step 5: Подключи навигацию из DuelPlayView**

```typescript
// В useDuelWebSocket.ts при game_complete:
router.push({ name: 'DuelResults', params: { gameId: msg.data.gameId } })
```

**Step 6: Запусти lint и type-check**

```bash
cd tma && pnpm lint && pnpm run type-check
```

**Step 7: Commit**

```bash
git add tma/src/views/Duel/DuelResultsView.vue tma/src/router/index.ts
git commit -m "feat(pvp-duel): create results view with per-question breakdown and share"
```

---

## PHASE C: Ticket System

### Task 7: Интеграция реального баланса билетов

**Проблема:** Все use cases возвращают захардкоженные 10 билетов. `JoinQueue` не списывает билет, `LeaveQueue` не возвращает.

**Files:**
- Modify: `backend/internal/application/quick_duel/get_duel_status.go`
- Modify: `backend/internal/application/quick_duel/join_queue.go`
- Modify: `backend/internal/application/quick_duel/leave_queue.go`
- Modify: `backend/internal/application/quick_duel/send_challenge.go`
- Modify: `backend/internal/application/quick_duel/respond_challenge.go`

**Step 1: Найди интерфейс TicketService**

```bash
grep -rn "TicketService\|ConsumeTicket\|RefundTicket" backend/internal/
```

**Step 2: Убедись что TicketService реализован**

```bash
grep -rn "func.*ConsumeTicket\|func.*RefundTicket" backend/internal/
```

**Step 3: Подключи TicketService в JoinQueue**

```go
// В JoinQueueUseCase.Execute():

// Проверь баланс
balance, err := uc.ticketService.GetTicketBalance(ctx, input.PlayerID)
if err != nil {
    return nil, err
}
if balance < 1 {
    return nil, domain.ErrInsufficientTickets
}

// Спиши билет
if err := uc.ticketService.ConsumeTicket(ctx, input.PlayerID, "pvp_queue"); err != nil {
    return nil, err
}
```

**Step 4: Возврат билета в LeaveQueue**

```go
// В LeaveQueueUseCase.Execute():
if err := uc.ticketService.RefundTicket(ctx, input.PlayerID, "queue_cancelled"); err != nil {
    log.Printf("warn: failed to refund ticket for player %s: %v", input.PlayerID, err)
    // не фейлим запрос, просто логируем
}
```

**Step 5: Напиши тест на недостаток билетов**

```go
func TestJoinQueue_InsufficientTickets(t *testing.T) {
    mockTickets.On("GetTicketBalance", mock.Anything, playerID).Return(0, nil)

    _, err := uc.Execute(ctx, JoinQueueInput{PlayerID: playerID})

    assert.ErrorIs(t, err, domain.ErrInsufficientTickets)
    mockTickets.AssertNotCalled(t, "ConsumeTicket")
}
```

**Step 6: Запусти тесты**

```bash
cd backend && go test ./internal/application/quick_duel/... -v
```

**Step 7: Commit**

```bash
git add backend/internal/application/quick_duel/join_queue.go \
        backend/internal/application/quick_duel/leave_queue.go \
        backend/internal/application/quick_duel/send_challenge.go \
        backend/internal/application/quick_duel/respond_challenge.go
git commit -m "feat(pvp-duel): integrate real ticket balance in queue and challenge use cases"
```

---

## PHASE D: Missing API Endpoints

### Task 8: Добавить GET /api/v1/duel/game/:gameId

**Проблема:** Эндпоинт задокументирован в `05_api.md` как основной источник данных для экрана результатов. В хэндлерах отсутствует.

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- Create: `backend/internal/application/quick_duel/get_game_result.go`

**Step 1: Создай use case GetGameResult**

```go
// get_game_result.go
type GetGameResultInput struct {
    GameID   string
    PlayerID string // для определения "who is me"
}

type GetGameResultOutput struct {
    GameID      string
    Status      string
    Winner      string
    IsFriendGame bool
    Players     GamePlayersDTO
    Questions   []QuestionResultDTO
    WinReason   string
    CompletedAt int64
    Share       *ShareDTO     // только если winner == playerID
    Rematch     RematchDTO
}
```

**Step 2: Добавь handler**

```go
// @Summary Get game result
// @Tags duel
// @Produce json
// @Param gameId path string true "Game ID"
// @Success 200 {object} GetGameResultResponse
// @Router /duel/game/{gameId} [get]
func (h *DuelHandler) GetGameResult(c *fiber.Ctx) error {
    playerID := c.Locals("userID").(string)
    gameID := c.Params("gameId")

    result, err := h.getGameResultUC.Execute(c.Context(), application.GetGameResultInput{
        GameID:   gameID,
        PlayerID: playerID,
    })
    if err != nil {
        return h.handleError(c, err)
    }

    return c.JSON(fiber.Map{"data": result})
}
```

**Step 3: Зарегистрируй route**

```go
// В router setup:
duel.Get("/game/:gameId", duelHandler.GetGameResult)
```

**Step 4: Regenerate Swagger + TypeScript**

```bash
cd backend && make swagger
cd ../tma && pnpm run generate:all
```

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/get_game_result.go \
        backend/internal/infrastructure/http/handlers/duel_handlers.go
git commit -m "feat(pvp-duel): add GET /duel/game/:gameId endpoint for results screen"
```

---

### Task 9: Добавить POST /duel/game/:gameId/surrender

**Проблема:** Задокументировано в `05_api.md`, отсутствует в коде. Нужно для voluntary forfeit после Q3.

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- Create: `backend/internal/application/quick_duel/surrender_game.go`

**Step 1: Найди метод Forfeit в domain**

```bash
grep -n "Forfeit\|Surrender\|surrender" backend/internal/domain/quick_duel/duel_game_aggregate.go
```

**Step 2: Создай use case SurrenderGame**

```go
type SurrenderGameInput struct {
    GameID   string
    PlayerID string
}

func (uc *SurrenderGameUseCase) Execute(ctx context.Context, input SurrenderGameInput) error {
    game, err := uc.gameRepo.FindByID(ctx, domain.GameID(input.GameID))
    if err != nil {
        return err
    }

    // Surrender доступен только после Q3 (roundIndex >= 2 = после 3го вопроса)
    if game.CurrentRound() < 3 {
        return domain.ErrSurrenderTooEarly
    }

    return game.Forfeit(domain.PlayerID(input.PlayerID))
}
```

**Step 3: Добавь handler**

```go
// @Summary Surrender (forfeit) active duel
// @Tags duel
// @Produce json
// @Param gameId path string true "Game ID"
// @Success 200 {object} SurrenderResponse
// @Failure 400 {object} ErrorResponse "Too early to surrender"
// @Router /duel/game/{gameId}/surrender [post]
func (h *DuelHandler) Surrender(c *fiber.Ctx) error {
    // ...
}
```

**Step 4: Запусти тесты**

```bash
cd backend && go test ./internal/application/quick_duel/... -v -run TestSurrender
```

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/surrender_game.go \
        backend/internal/infrastructure/http/handlers/duel_handlers.go
git commit -m "feat(pvp-duel): add surrender endpoint (available after Q3)"
```

---

## PHASE E: Friends & Social

### Task 10: Реализовать Friends Online статус

**Проблема:** `GetDuelStatus` всегда возвращает `friendsOnline: []`. Нужна интеграция с Telegram friend list.

**Files:**
- Modify: `backend/internal/application/quick_duel/get_duel_status.go`
- Read: `backend/internal/infrastructure/persistence/redis/` — онлайн статус уже должен быть в Redis (ключ `duel:online:{playerId}`)

**Step 1: Проверь есть ли онлайн трекер**

```bash
grep -rn "duel:online\|OnlineTracker\|SetOnline\|online.*tracker" backend/internal/
```

**Step 2: Если трекер есть — подключи в GetDuelStatus**

```go
// В GetDuelStatusUseCase:
// 1. Получи список Telegram-друзей игрока (если есть friends service)
//    или всех пользователей с которыми играл (из истории матчей)
// 2. Для каждого — проверь Redis ключ duel:online:{playerId}
// 3. Добавь статус: "online", "recent" (был < 30 мин), "offline"

onlineFriends, err := uc.onlineTracker.GetOnlineFriends(ctx, playerID)
```

**Step 3: Если трекера нет — создай минимальный**

```go
// online_tracker.go
func (t *OnlineTracker) SetOnline(ctx context.Context, playerID string) error {
    return t.redis.SetEx(ctx, "duel:online:"+playerID, "1", 60*time.Second).Err()
}

func (t *OnlineTracker) IsOnline(ctx context.Context, playerID string) (bool, error) {
    val, err := t.redis.Exists(ctx, "duel:online:"+playerID).Result()
    return val > 0, err
}
```

**Step 4: Вызывай SetOnline при каждом WebSocket heartbeat (ping)**

```go
// В WS handler при получении ping:
case "ping":
    uc.onlineTracker.SetOnline(ctx, playerID)
    sendPong(playerID)
```

**Step 5: Commit**

```bash
git commit -m "feat(pvp-duel): implement friends online status via Redis TTL"
```

---

## PHASE F: Seasons & Rewards

### Task 11: Seasonal reward distribution

**Проблема:** README checklist: `[ ] Seasonal reward distribution`. Season reset job есть, но награды не раздаются.

**Files:**
- Найди: `backend/internal/application/quick_duel/` или `backend/internal/infrastructure/jobs/`

**Step 1: Найди season reset job**

```bash
find backend/ -name "*season*" -o -name "*job*" | grep -v test
grep -rn "SeasonReset\|season.*reset\|cron" backend/internal/
```

**Step 2: Добавь distribute rewards step в reset job**

```go
// В season reset:
// 1. Получи топ игроков по peak_mmr из player_seasons
// 2. Для каждого — рассчитай награду по таблице из 04_rewards.md
// 3. Добавь монеты и билеты через wallet service
// 4. Добавь косметику через cosmetics service
// 5. Отправь уведомление (если есть Telegram notification service)
// 6. Пометь rewards_distributed = true в player_seasons

func (j *SeasonResetJob) distributeRewards(ctx context.Context, seasonID string) error {
    players, err := j.seasonRepo.GetPlayersForRewards(ctx, seasonID)
    for _, p := range players {
        reward := calculateSeasonReward(p.PeakLeague)
        j.wallet.AddCoins(ctx, p.PlayerID, reward.Coins, "season_reward")
        j.ticketService.AddTickets(ctx, p.PlayerID, reward.Tickets, "seasonal")
    }
}
```

**Step 3: Commit**

```bash
git commit -m "feat(pvp-duel): implement seasonal reward distribution in reset job"
```

---

## PHASE G: Push Notifications

### Task 12: Telegram push для challenge notifications

**Проблема:** README checklist: `[ ] Push notifications (Telegram)`. При вызове офлайн друга — уведомление не отправляется.

**Files:**
- Найди: `backend/internal/infrastructure/notifications/` или Telegram bot integration
- Modify: `backend/internal/application/quick_duel/send_challenge.go`

**Step 1: Найди существующий notification service**

```bash
find backend/ -name "*notif*" -o -name "*telegram*" -o -name "*bot*"
grep -rn "NotificationService\|SendMessage\|bot.*token" backend/internal/
```

**Step 2: Подключи в SendChallenge use case**

```go
// В SendChallengeUseCase после создания challenge:
if targetUser.IsOnline {
    // Уже отправляем WS сообщение challenge_received
} else {
    // Отправляем Telegram push
    uc.notifier.SendChallenge(ctx, NotifyChallengeInput{
        ToPlayerID: input.FriendID,
        FromUsername: challenger.Username,
        ExpiresIn: 300, // 5 мин для оффлайн друга
        DeepLink: "https://t.me/quiz_sprint_dev_bot?startapp=duel_" + challenge.LinkCode,
    })
}
```

**Step 3: Commit**

```bash
git commit -m "feat(pvp-duel): send Telegram push notification for friend challenges"
```

---

## Итоговый порядок выполнения

```
Phase A: Task 1 → Task 2 → Task 3 → Task 4   (Core backend)
Phase B: Task 5 → Task 6                       (Frontend)
Phase C: Task 7                                 (Tickets)
Phase D: Task 8 → Task 9                       (API gaps)
Phase E: Task 10                               (Friends)
Phase F: Task 11                               (Seasons)
Phase G: Task 12                               (Notifications)
```

**Минимально рабочая игра (MVP):** Tasks 1, 3, 4, 5, 6 + Task 8.
Остальные можно итерировать после запуска.

---

## Известные технические долги (не блокируют)

| Проблема | Файл | Приоритет |
|----------|------|-----------|
| Domain package называется `quick_duel`, docs говорят `pvp_duel` | весь backend | Низкий — rename когда будет время |
| Friends leaderboard в GetLeaderboard возвращает пустой список | get_leaderboard.go | Средний — нужен friends service |
| Victory card image generation не реализована | — | Низкий — Phase 2+ |
| Revenge notifications | — | Низкий — Phase 2+ |
