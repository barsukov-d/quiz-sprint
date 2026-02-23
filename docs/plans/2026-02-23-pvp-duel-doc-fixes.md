# PvP Duel — Fix Documentation Contradictions

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Устранить 9 внутренних противоречий в документации `docs/game_modes/pvp_duel/`, найденных при аудите.

**Architecture:** Только правки Markdown-файлов. Никакого кода не трогаем. Каждый таск — одно противоречие, один коммит.

**Tech Stack:** Markdown, `docs/game_modes/pvp_duel/`

---

## Контекст: найденные противоречия

Источник — результаты аудита от 2026-02-23. Все противоречия задокументированы там же.
Файлы: `01_concept.md`, `02_gameplay.md`, `03_rules.md`, `04_rewards.md`, `05_api.md`, `06_domain.md`, `07_edge_cases.md`, `README.md`.

---

### Task 1: Исправить ошибку в таблице MMR (03_rules.md)

**Приоритет:** 🔴 Критический (математическая ошибка)

**Проблема:**
`03_rules.md` строка с "1500 vs 2000" показывает `+28/-28`, но формула даёт `+30/-30`:
```
expected = 1 / (1 + 10^(500/400)) = 1 / 18.78 = 0.0532
winnerDelta = int(32 * (1 - 0.0532)) = int(30.3) = 30
loserDelta  = int(32 * (0 - 0.9468)) = int(-30.3) = -30
```

**Files:**
- Modify: `docs/game_modes/pvp_duel/03_rules.md`

**Step 1: Исправить таблицу**

Найти строку:
```
| 1500 | 2000 | +28 | -28 |
| 2000 | 1500 | +10 | -10 |
```

Заменить на:
```
| 1500 | 2000 | +30 | -30 |
| 2000 | 1500 | +10 | -10 |
```

Пояснение: строка `2000 vs 1500` (сильный победил слабого) остаётся `+10/-10` — это minimum clamp, не ошибка.

**Step 2: Проверить остальные строки таблицы**

Проверить вручную по формуле:
- 1500 vs 1500 → expected=0.5, delta = 32*(1-0.5)=16 ✅ таблица +16
- 1500 vs 1700 → expected=0.240, delta = 32*0.760=24.3→24 ✅ таблица +24
- 1700 vs 1500 → expected=0.760, delta = 32*0.240=7.7→7 → clamp → +10 ✅

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/03_rules.md
git commit -m "docs(pvp-duel): fix MMR table calculation for 1500 vs 2000 (+28→+30)"
```

---

### Task 2: Унифицировать формулу MMR (03_rules.md vs 06_domain.md)

**Приоритет:** 🔴 Критический (разные алгоритмы округления)

**Проблема:**
- `03_rules.md:18-19` — усечение `int()` (truncation)
- `06_domain.md:341-342` — округление `int(math.Round())`

**Решение:** Принять `math.Round()` как эталон (более корректно для финансово-значимых вычислений). Обновить `03_rules.md`.

**Files:**
- Modify: `docs/game_modes/pvp_duel/03_rules.md`

**Step 1: Обновить функцию CalculateMMRChange в 03_rules.md**

Найти:
```go
winnerDelta = int(float64(K) * (actualWinner - expectedWinner))
loserDelta = int(float64(K) * (actualLoser - expectedLoser))
```

Заменить на:
```go
winnerDelta = int(math.Round(float64(K) * (actualWinner - expectedWinner)))
loserDelta = int(math.Round(float64(K) * (actualLoser - expectedLoser)))
```

**Step 2: Убедиться что импорт math указан в псевдокоде**

Если в коде-примере нет `import "math"` — добавить комментарий перед функцией:
```go
// Requires: import "math"
```

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/03_rules.md
git commit -m "docs(pvp-duel): align MMR rounding to math.Round() (consistent with domain model)"
```

---

### Task 3: Исправить билеты Daily Challenge (04_rewards.md vs 06_domain.md)

**Приоритет:** 🔴 Критический (разные числа)

**Проблема:**

| Сундук | `04_rewards.md` | `06_domain.md` |
|--------|----------------|----------------|
| Silver (5-7 correct) | 2 🎟️ | 2-3 🎟️ |
| Golden (8-10 correct) | 3 🎟️ | 4-5 🎟️ |

**Решение:** `04_rewards.md` — основной документ по наградам, его значения точнее (фиксированные числа лучше диапазонов для Daily Challenge). Обновить `06_domain.md`.

**Files:**
- Modify: `docs/game_modes/pvp_duel/06_domain.md`

**Step 1: Найти таблицу в 06_domain.md**

```
| 5-7 correct (Silver Chest) | 2-3 🎟️ |
| 8-10 correct (Golden Chest) | 4-5 🎟️ |
```

**Step 2: Заменить на значения из 04_rewards.md**

```
| 5-7 correct (Silver Chest) | 2 🎟️ |
| 8-10 correct (Golden Chest) | 3 🎟️ |
```

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/06_domain.md
git commit -m "docs(pvp-duel): align Daily Challenge ticket rewards with 04_rewards.md (Silver:2, Gold:3)"
```

---

### Task 4: Исправить grace period при дисконнекте (05_api.md)

**Приоритет:** 🔴 Критический (30s в API spec vs 10s в логике)

**Проблема:**
- `05_api.md:720` — WebSocket `opponent_disconnected` → `"reconnectIn": 30`
- `02_gameplay.md:429` — "10s grace period"
- `07_edge_cases.md:143` — код явно: `time.Sleep(10 * time.Second)`

**Решение:** 10 секунд — значение из кода и gameplay doc. Исправить API spec.

**Files:**
- Modify: `docs/game_modes/pvp_duel/05_api.md`

**Step 1: Найти WebSocket сообщение opponent_disconnected**

```json
{
  "type": "opponent_disconnected",
  "data": {
    "playerId": "user_456",
    "reconnectIn": 30
  }
}
```

**Step 2: Исправить значение**

```json
{
  "type": "opponent_disconnected",
  "data": {
    "playerId": "user_456",
    "reconnectIn": 10
  }
}
```

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/05_api.md
git commit -m "docs(pvp-duel): fix reconnectIn to 10s (was 30s, inconsistent with game rules)"
```

---

### Task 5: Убрать pointsEarned из WebSocket spec (05_api.md)

**Приоритет:** 🔴 Критический (несуществующая механика)

**Проблема:**
`05_api.md` WebSocket сообщение `answer_result` содержит `"pointsEarned": 100`.
Нигде в документации нет системы очков. Score в PvP Duel = количество правильных ответов (1 за каждый). Значение 100 не объяснено.

**Files:**
- Modify: `docs/game_modes/pvp_duel/05_api.md`

**Step 1: Найти answer_result сообщение**

```json
{
  "type": "answer_result",
  "data": {
    "playerId": "user_123",
    "questionId": "q_003",
    "isCorrect": true,
    "correctAnswer": "a_002",
    "pointsEarned": 100,
    "timeTaken": 4200,
    "player1Score": 3,
    "player2Score": 2
  }
}
```

**Step 2: Удалить поле pointsEarned**

```json
{
  "type": "answer_result",
  "data": {
    "playerId": "user_123",
    "questionId": "q_003",
    "isCorrect": true,
    "correctAnswer": "a_002",
    "timeTaken": 4200,
    "player1Score": 3,
    "player2Score": 2
  }
}
```

`player1Score` и `player2Score` — это и есть счёт (количество правильных ответов), они остаются.

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/05_api.md
git commit -m "docs(pvp-duel): remove pointsEarned from answer_result WS message (no points system in PvP)"
```

---

### Task 6: Исправить момент списания билета (03_rules.md + 05_api.md)

**Приоритет:** 🔴 Критический (две несовместимые модели)

**Проблема:**
- `03_rules.md:259` — `StartDuel()` списывает билет при **старте игры**
- `05_api.md` — API показывает списание при **входе в очередь** (`ticketConsumed: true` в Challenge ответе, `ticketRefunded: true` при отмене очереди)

**Решение:** Принять модель "билет списывается при входе в очередь / отправке вызова" как правильную. Это соответствует API spec и UX (пользователь видит немедленную реакцию). Обновить `03_rules.md`.

**Files:**
- Modify: `docs/game_modes/pvp_duel/03_rules.md`

**Step 1: Заменить функцию StartDuel на правильную модель**

Найти блок:
```go
func StartDuel(player1, player2 *Player) error {
    // Consume tickets
    if err := player1.ConsumeTicket(); err != nil {
        return ErrInsufficientTickets
    }
    if err := player2.ConsumeTicket(); err != nil {
        player1.RefundTicket()  // Rollback
        return ErrInsufficientTickets
    }

    return nil
}
```

Заменить на:
```go
// Tickets are consumed BEFORE game start:
// - Random queue: consumed at JoinQueue()
// - Friend challenge (challenger): consumed at CreateChallenge()
// - Friend challenge (challengee): consumed at AcceptChallenge()
// StartDuel() assumes tickets already consumed and validated.
func StartDuel(player1, player2 *Player) error {
    // Tickets already consumed. Validate game can start.
    return nil
}
```

**Step 2: Убедиться, что таблица Ticket Refund корректна**

Проверить таблицу (lines 276-283) — она остаётся без изменений, так как уже описывает правильное поведение рефанда.

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/03_rules.md
git commit -m "docs(pvp-duel): clarify ticket consumed at queue join/challenge, not at game start"
```

---

### Task 7: Исправить противоречие "Наставник" (04_rewards.md)

**Приоритет:** 🟡 Средний (одно название — два условия)

**Проблема:**
- `04_rewards.md` таблица Titles: "Наставник" → 5 referred friends
- `04_rewards.md` таблица Milestones: "Наставник" badge → first friend reaches Silver

Два разных триггера для одного названия, разный тип (badge vs title).

**Решение:** Разделить — milestone badge и title за рефералы это разные предметы с разными названиями.

**Files:**
- Modify: `docs/game_modes/pvp_duel/04_rewards.md`

**Step 1: Переименовать milestone badge в таблице Referral Milestones**

Найти:
```
| **Reaches Silver** | 10 🎟️ + 500 coins + 🏷️ "Наставник" badge | 300 coins |
```

Заменить на:
```
| **Reaches Silver** | 10 🎟️ + 500 coins + 🏷️ "Гуру" badge | 300 coins |
```

**Step 2: Обновить таблицу Titles**

Найти:
```
| "Наставник" | 5 referred friends |
```

Оставить без изменений (title за 5 рефералов остаётся "Наставник").

**Step 3: Обновить ссылку в 01_concept.md**

В `01_concept.md` таблица Referral System строка "Friend reaches Silver":

Найти:
```
| Friend reaches Silver | 10 🎟️ + 500 coins + 🏷️ "Наставник" | 300 coins |
```

Заменить на:
```
| Friend reaches Silver | 10 🎟️ + 500 coins + 🏷️ "Гуру" badge | 300 coins |
```

**Step 4: Обновить ссылку в 05_api.md (referral claim response)**

В `05_api.md` найти:
```json
"badge": "Наставник"
```

Заменить на:
```json
"badge": "Гуру"
```

**Step 5: Commit**

```bash
git add docs/game_modes/pvp_duel/04_rewards.md docs/game_modes/pvp_duel/01_concept.md docs/game_modes/pvp_duel/05_api.md
git commit -m "docs(pvp-duel): rename Silver milestone badge to 'Гуру' to avoid conflict with 'Наставник' title"
```

---

### Task 8: Исправить пути домена в README.md (quick_duel → pvp_duel)

**Приоритет:** 🟡 Средний (устаревшие пути)

**Проблема:**
`README.md:22-23` указывает пути `quick_duel/`, но bounded context называется `pvp_duel` (`06_domain.md:4`).

**Files:**
- Modify: `docs/game_modes/pvp_duel/README.md`

**Step 1: Обновить пути в секции Quick Navigation**

Найти:
```
- **Domain**: `backend/internal/domain/quick_duel/`
- **Application**: `backend/internal/application/quick_duel/`
```

Заменить на:
```
- **Domain**: `backend/internal/domain/pvp_duel/`
- **Application**: `backend/internal/application/pvp_duel/`
```

**Step 2: Убрать "(TBD)" у WebSocket**

Найти:
```
- **WebSocket**: `/ws/duel` (TBD)
```

Заменить на:
```
- **WebSocket**: `/ws/duel/:gameId` (spec in 05_api.md)
```

**Step 3: Commit**

```bash
git add docs/game_modes/pvp_duel/README.md
git commit -m "docs(pvp-duel): fix domain paths quick_duel→pvp_duel, mark WebSocket as specced"
```

---

### Task 9: Уточнить диапазоны MMR в концепте (01_concept.md)

**Приоритет:** 🟡 Средний (приблизительные числа не совпадают с формулой)

**Проблема:**
`01_concept.md:41-44` описывает MMR изменения как "+25-40" для победы над более сильным, но формула при разнице 200 MMR даёт +24 (не входит в диапазон).

**Решение:** Заменить конкретные диапазоны на формульное описание — так концепт останется актуальным при любых изменениях K-фактора.

**Files:**
- Modify: `docs/game_modes/pvp_duel/01_concept.md`

**Step 1: Найти и заменить описание MMR в секции ELO/MMR Rating System**

Найти:
```
- Win against stronger opponent → +25-40 MMR
- Win against weaker opponent → +10-15 MMR
- Lose to stronger opponent → -10-15 MMR
- Lose to weaker opponent → -25-40 MMR
```

Заменить на:
```
- Win against stronger opponent → more MMR (up to ~+30)
- Win against weaker opponent → less MMR (minimum +10)
- Lose to stronger opponent → less MMR lost (minimum -10)
- Lose to weaker opponent → more MMR lost (up to ~-30)
- Exact values: ELO formula, K=32, min ±10. See `03_rules.md`.
```

**Step 2: Commit**

```bash
git add docs/game_modes/pvp_duel/01_concept.md
git commit -m "docs(pvp-duel): replace approximate MMR ranges with formula reference in concept"
```

---

## Порядок выполнения

Выполнять строго по порядку задач — каждая независима, но нумерация отражает приоритет. Коммит после каждого таска.

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6 → Task 7 → Task 8 → Task 9
```

После выполнения всех задач — запушить ветку:

```bash
git push origin pvp-duel
```

---

## Быстрая проверка после выполнения

Пройтись по каждому противоречию из аудита и убедиться что исправлено:

- [ ] Таблица MMR: 1500 vs 2000 = ±30 (не ±28)
- [ ] Формула: `math.Round()` в обоих файлах
- [ ] Daily Challenge tickets: Silver=2, Golden=3 в обоих файлах
- [ ] `reconnectIn: 10` в 05_api.md
- [ ] Нет `pointsEarned` в `answer_result` WS message
- [ ] `StartDuel()` не списывает билеты (билет списывается при входе в очередь)
- [ ] "Наставник" = title за 5 рефералов; "Гуру" = badge за Silver milestone
- [ ] Пути: `pvp_duel/` (не `quick_duel/`)
- [ ] MMR в концепте — ссылка на формулу, не hardcoded диапазоны
