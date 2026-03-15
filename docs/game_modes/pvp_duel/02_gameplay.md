# PvP Duel - Gameplay Flow

> **Статус реализации (аудит 2026-03-15)**
> ✅ Реализовано: 13 | ⚠️ Расходится: 6 | ❌ Не реализовано: 7

## Entry Point
Home → "Дуэль" → Shows: <!-- ⚠️ DuelLobbyView exists but tickets not enforced -->
- Current league + MMR (e.g., "🥇 Gold III — 1,650 MMR")
- PvP tickets available: 🎟️×3 <!-- ❌ Ticket count shown but not enforced on entry -->
- Win/Loss record this season
- **Two main buttons:** "Случайный соперник" / "Вызвать друга" <!-- ✅ -->
- Friends online indicator <!-- ✅ Redis OnlineTracker -->

> ⚠️ **Расхождение:** DuelLobbyView существует, но тикеты не списываются при входе в дуэль.

---

## Invite Link Flow (Friend Challenge по ссылке)

### Технический механизм <!-- ✅ Well implemented, matching doc flow -->

```
POST /api/v1/duel/challenge/link → { challengeLink: "t.me/bot?startapp=duel_abc123" }
Telegram share sheet → получатель кликает

TMA открывается: SDK launchParams.startParam = "duel_abc123"
useAuth.ts → сохраняет startParam в памяти

App.vue (onMounted):
  registerUser() → Welcome bonus +3 🎟️ (новый пользователь)
  consumeStartParam() → "duel_abc123"
  handleDeepLink() → router.push('/duel?challenge=duel_abc123')

DuelLobbyView (onMounted):
  route.query.challenge = "duel_abc123"
  → Показывает модал подтверждения (НЕ принимает автоматически)

Invitee нажимает "Принять вызов":
  POST /duel/challenge/accept-by-code { linkCode: "duel_abc123" }
  → { challengeId: "ch_xyz", status: "accepted_waiting_inviter" }
  → Telegram Bot API отправляет сообщение инвайтеру: "Vasya принял твой вызов!"
  → UI показывает "Ждём инвайтера..."

Inviter видит карточку "✅ Vasya готов к дуэли!" (GET /duel/status polling):
  Нажимает "Начать дуэль →"
  POST /duel/challenge/:challengeId/start { playerId: "user_456" }
  → { gameId: "g_xyz" }
  → Оба переходят в DuelPlay
```

### UX Flow — Invitee (принимающий) <!-- ✅ -->

| Шаг | Экран | Действие |
|-----|-------|---------|
| 1 | Telegram чат | Кликает invite-ссылку |
| 2 | TMA splash | Загрузка + Telegram Auth |
| 3 | (фон) | `POST /user/register` → +3 🎟️ welcome bonus |
| 4 | DuelLobby | **Модал подтверждения** "Тебя вызывают на дуэль!" |
| 5 | (фон) | `POST /duel/challenge/accept-by-code` → статус `accepted_waiting_inviter` |
| 6 | DuelLobby | "Ждём инвайтера..." |
| 7 | DuelPlay | Инвайтер нажал "Начать" → игра |

### UX Flow — Inviter (создавший ссылку) <!-- ✅ -->

| Шаг | Экран | Действие |
|-----|-------|---------|
| 1 | DuelLobby | Карточка "✈ Ожидаем ответа..." (pending) |
| 2 | (TG) | Получает уведомление: "Vasya принял твой вызов!" |
| 3 | DuelLobby | Карточка меняется: "✅ Vasya готов к дуэли!" |
| 4 | (тап) | Нажимает "Начать дуэль →" |
| 5 | (фон) | `POST /duel/challenge/:id/start` → `{ gameId }` |
| 6 | DuelPlay | 3...2...1 → игра |

### Состояния модала подтверждения (Invitee) <!-- ✅ DuelLobbyView handles challenge param -->

```
┌───────────────────────────────────┐
│  ⚔️  Тебя вызывают на дуэль!      │
│                                   │
│  Хочешь принять вызов?            │
│                                   │
│  [ Принять вызов ]                │
│  [ Отклонить     ]                │
└───────────────────────────────────┘

┌───────────────────────────────────┐
│  ✗  Ссылка устарела               │  ← 409 CHALLENGE_EXPIRED
│     Попроси друга прислать новую  │
│                          [Закрыть]│
└───────────────────────────────────┘
```

### Карточки исходящих вызовов (Inviter, в DuelLobby) <!-- ✅ -->

```
┌──────────────────────────────────────┐
│  ✅  @vasya готов к дуэли!            │  ← accepted_waiting_inviter
│                                      │
│  [ Начать дуэль →               ]    │
└──────────────────────────────────────┘

┌──────────────────────────────────────┐
│  ✈ Ожидаем ответа...                 │  ← pending
│  Ссылка истекает через: 23ч 45мин   │
└──────────────────────────────────────┘
```

### Edge Cases <!-- ✅ All validated in domain -->

| Ситуация | HTTP | UI |
|----------|------|----|
| Ссылка истекла | 409 | "Ссылка устарела. Попроси новую" |
| Ссылка уже использована | 409 | "Вызов уже принят другим игроком" |
| Inviter уже в игре | 409 | "Твой друг сейчас в игре" |
| Inviter отменил ссылку | 404 | "Вызов отменён. Хочешь создать новый?" |
| Себе отправил ссылку | 400 | "Нельзя вызвать самого себя" |
| Уже в очереди | 409 | "Отмени поиск и попробуй снова" |

---

## Game Flow

### 1. Pre-Game Screen <!-- ⚠️ DuelLobbyView exists but tickets not enforced -->
```
┌─────────────────────────────────────┐
│  ⚔️ РЕЙТИНГОВАЯ ДУЭЛЬ        🎟️ × 3│
│                                     │
│  Твой ранг: 🥇 Gold III             │
│  MMR: 1,650                         │
│  Сезон 4: 23W / 18L (56%)           │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  [  СЛУЧАЙНЫЙ СОПЕРНИК  ]   │    │
│  │  Стоимость: 1 🎟️           │    │
│  └─────────────────────────────┘    │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  [    ВЫЗВАТЬ ДРУГА     ]   │    │
│  │  Стоимость: 1 🎟️ (обоим)   │    │
│  └─────────────────────────────┘    │
│                                     │
│  👥 Друзья онлайн: 3                │
│  @ProGamer 🟢  @BestQuiz 🟢         │
│                                     │
│  [ Лидерборд ]  [ История ]         │
└─────────────────────────────────────┘
```

---

### 1b. Friend Challenge Flow <!-- ✅ -->
```
┌─────────────────────────────────────┐
│  👥 ВЫЗВАТЬ ДРУГА                   │
│                                     │
│  Друзья онлайн:                     │
│  ┌─────────────────────────────┐    │
│  │ 🟢 @ProGamer    Gold I       │    │
│  │    [ ВЫЗВАТЬ ]              │    │
│  ├─────────────────────────────┤    │
│  │ 🟢 @BestQuiz    Silver II   │    │
│  │    [ ВЫЗВАТЬ ]              │    │
│  ├─────────────────────────────┤    │
│  │ 🟡 @NewFriend   Bronze III  │    │
│  │    был 5 мин назад          │    │
│  │    [ ОТПРАВИТЬ ВЫЗОВ ]      │    │
│  └─────────────────────────────┘    │
│                                     │
│  ─────── или ───────                │
│                                     │
│  [ 📤 Поделиться ссылкой ]          │
│  Пригласи любого по ссылке!         │
│                                     │
└─────────────────────────────────────┘
```

**Friend Challenge Types:**

| Type | Description | Wait Time |
|------|-------------|-----------|
| 🟢 Online | Friend in app now | Instant accept/decline |
| 🟡 Recent | Was online <30min | Push notification, 5min wait |
| 📤 Link | Share to any chat | 24h valid, one-time use |

**Challenge Notification (for recipient):**
```
┌─────────────────────────────────────┐
│  ⚔️ @PlayerName вызывает тебя!      │
│                                     │
│  🥇 Gold III vs 🥈 Silver I (ты)    │
│                                     │
│  [ ПРИНЯТЬ ]    [ ОТКЛОНИТЬ ]       │
│                                     │
│  Осталось: 58 сек                   │
└─────────────────────────────────────┘
```

**If friend not in app → Telegram push:**
```
⚔️ Quiz Sprint: @PlayerName вызывает тебя на дуэль!
Покажи кто здесь умнее! 🧠
[Принять вызов]
```

---

### 2. Matchmaking Screen (Random) <!-- ✅ In DuelLobbyView -->
```
┌─────────────────────────────────────┐
│                                     │
│        🔍 ПОИСК СОПЕРНИКА...        │
│                                     │
│        ⏱️ 00:12                      │
│                                     │
│  Ищем игрока с близким MMR          │
│  Диапазон: 1,550 - 1,750            │
│                                     │
│  [ Отменить поиск ]                 │
│                                     │
└─────────────────────────────────────┘
```

**Queue expansion:** <!-- ⚠️ Code has 4 tiers (5s/10s/15s/15s+), not 5 tiers; after 15s matches anyone -->
- 0-10s: ±50 MMR
- 10-20s: ±100 MMR
- 20-30s: ±200 MMR
- 30-45s: ±300 MMR
- 45-60s: ±500 MMR
- 60s+: Offer bot game <!-- ❌ Bot game fallback not implemented -->

> ⚠️ **Расхождение:** Код реализует 4 тира (5с/10с/15с/15с+), после 15с матчит любого. Документ специфицирует 5 тиров с постепенным расширением до ±500 MMR.
> ❌ **Не реализовано:** Bot game fallback по истечении 60с.

---

### 3. Opponent Found Screen <!-- ✅ WebSocket game_ready with startsIn -->
```
┌─────────────────────────────────────┐
│                                     │
│        ⚔️ СОПЕРНИК НАЙДЕН!          │
│                                     │
│  ┌─────────┐     ┌─────────┐        │
│  │  [AVA]  │ VS  │  [AVA]  │        │
│  │   Ты    │     │ ProQuiz │        │
│  │ Gold III│     │ Gold II │        │
│  │  1,650  │     │  1,720  │        │
│  └─────────┘     └─────────┘        │
│                                     │
│        Дуэль начнётся через 3...    │
│                                     │
└─────────────────────────────────────┘
```

**Countdown:** 3 seconds → Start duel.

---

### 4. Question Screen (In-Duel) <!-- ✅ DuelPlayView.vue with WebSocket -->
```
┌─────────────────────────────────────┐
│  ⚔️ Дуэль                    ⏱️ 7   │
│  Вопрос 3/7                         │
│─────────────────────────────────────│
│                                     │
│  Какой элемент имеет символ Au?     │
│                                     │
│  [ A. Серебро          ]            │
│  [ B. Золото           ]  ← selected│
│  [ C. Медь             ]            │
│  [ D. Алюминий         ]            │
│                                     │
│─────────────────────────────────────│
│  Ты: ✅✅⬜⬜⬜⬜⬜                   │
│  ProQuiz: ✅⬜⬜⬜⬜⬜⬜              │
│─────────────────────────────────────│
└─────────────────────────────────────┘
```

**UI Elements:**
- **Timer:** 10s countdown (color: green → yellow → red)
- **Question number:** 3/7
- **Score indicators:** Visual progress for both players
  - ✅ = correct
  - ❌ = wrong
  - ⬜ = not answered yet
  - ⏳ = opponent still answering current question

**Real-time opponent status:** <!-- ✅ Via WebSocket answer_result -->
- Show when opponent answers (⏳ → ✅/❌)
- Do NOT show which answer they selected
- Create tension without revealing strategy

**Emotes (in-game reactions):** <!-- ❌ Not implemented -->
- Emote button in top corner → pick from unlocked set
- Max 1 emote per question, max 3 per game
- Appears briefly (~1.5s) on opponent's screen
- Default: 👋. More unlocked via achievements (see `04_rewards.md`)

> ❌ **Не реализовано:** Emotes — реакции в процессе игры не реализованы.

---

### 5. Answer Feedback (Brief) <!-- ✅ -->
```
┌─────────────────────────────────────┐
│                                     │
│          ✅ ПРАВИЛЬНО!              │
│                                     │
│  Твоё время: 4.2 сек                │
│                                     │
│  ProQuiz: ✅ (3.8 сек)              │
│                                     │
└─────────────────────────────────────┘
```

**Duration:** 1.5 seconds → Auto-advance to next question.

**Show:**
- Your result (correct/wrong)
- Opponent's result (✅/❌)
- Both times (after both answered)
- Correct answer if you were wrong

---

### 6. Game Result Screen <!-- ✅ DuelResultsView.vue -->
```
┌─────────────────────────────────────┐
│                                     │
│          🏆 ПОБЕДА!                  │
│                                     │
│  ┌─────────────────────────────┐    │
│  │   Ты        5 : 4    ProQuiz│    │
│  │  42.5s              38.2s   │    │
│  └─────────────────────────────┘    │
│                                     │
│  Детали:                            │
│  Q1: ✅ 2.1s    ✅ 3.0s             │
│  Q2: ✅ 5.2s    ❌ 4.1s             │
│  Q3: ✅ 4.2s    ✅ 3.8s             │
│  Q4: ❌ 6.0s    ✅ 5.5s             │
│  Q5: ✅ 8.1s    ✅ 7.2s             │
│  Q6: ✅ 9.0s    ❌ 8.0s             │
│  Q7: ❌ 8.0s    ❌ 7.6s             │
│                                     │
│  MMR: 1,650 → 1,678 (+28)           │
│  Ранг: 🥇 Gold III                  │
│                                     │
│  [  РЕВАНШ  ]  [  В МЕНЮ  ]         │
│                                     │
│  [ Поделиться ]                     │
└─────────────────────────────────────┘
```

**Victory screen elements:**
- Final score (5:4)
- Total time comparison
- Per-question breakdown
- MMR change (+/-)
- New rank if promoted/demoted
- **Rematch button** (if opponent accepts) <!-- ✅ RequestRematchUseCase exists -->
- **Share button** (generates victory card) <!-- ❌ No image generation -->

---

### 6b. Share Victory Card <!-- ❌ Not implemented — no image generation -->
```
┌─────────────────────────────────────┐
│  📤 ПОДЕЛИТЬСЯ ПОБЕДОЙ              │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  ⚔️ ПОБЕДА В ДУЭЛИ!         │    │
│  │                             │    │
│  │      @YourName 🏆           │    │
│  │        5 : 4                │    │
│  │      🥇 Gold III            │    │
│  │                             │    │
│  │  "Кто следующий?" 😎        │    │
│  │                             │    │
│  │  ▶️ Сыграй со мной:         │    │
│  │  t.me/quiz_sprint_dev_bot?  │    │
│  │  startapp=duel_abc123       │    │
│  └─────────────────────────────┘    │
│                                     │
│  Отправить в:                       │
│  [ Telegram ] [ Stories ] [ Copy ]  │
│                                     │
└─────────────────────────────────────┘
```

> ❌ **Не реализовано:** Share Victory Card — генерация изображения не реализована.

**Share link behavior:**
- Clicking link → Opens TMA directly with `?startapp=duel_xxx` parameter
- TMA extracts `startParam`, authenticates user, navigates to duel lobby
- Deep link handler показывает **модал подтверждения** (не авто-принимает)
- Пользователь подтверждает → `POST /duel/challenge/accept-by-code` → `accepted_waiting_inviter`
- New user → Registers first, then sees confirmation modal
- Existing user → Direct to duel lobby with modal

---

### 7. Defeat Screen <!-- ✅ -->
```
┌─────────────────────────────────────┐
│                                     │
│          💔 ПОРАЖЕНИЕ                │
│                                     │
│  ┌─────────────────────────────┐    │
│  │   Ты        3 : 5    ProQuiz│    │
│  │  45.2s              42.1s   │    │
│  └─────────────────────────────┘    │
│                                     │
│  MMR: 1,650 → 1,625 (-25)           │
│  Ранг: 🥇 Gold III                  │
│                                     │
│  [  РЕВАНШ  ]  [  В МЕНЮ  ]         │
└─────────────────────────────────────┘
```

---

### 8. Tiebreaker Result <!-- ⚠️ Code returns nil (draw) when points tied; no explicit time tiebreaker -->
```
┌─────────────────────────────────────┐
│                                     │
│       ⚖️ НИЧЬЯ ПО ОЧКАМ!            │
│                                     │
│  ┌─────────────────────────────┐    │
│  │   Ты        5 : 5    ProQuiz│    │
│  │  38.2s   ⚡        42.5s   │    │
│  └─────────────────────────────┘    │
│                                     │
│  Победа по времени: ТЫ! 🏆          │
│                                     │
│  MMR: 1,650 → 1,672 (+22)           │
│                                     │
└─────────────────────────────────────┘
```

> ⚠️ **Расхождение:** Когда очки равны, код возвращает nil (ничья), а не применяет тайбрейкер по времени. Документ специфицирует победу более быстрого игрока.

---

## Rematch Flow <!-- ✅ RequestRematchUseCase exists, 15s window -->

### Rematch Request
```
┌─────────────────────────────────────┐
│                                     │
│  ProQuiz хочет реванш!              │
│                                     │
│  Стоимость: 1 🎟️                   │
│  У тебя: 🎟️ × 2                    │
│                                     │
│  [  ПРИНЯТЬ  ]  [  ОТКЛОНИТЬ  ]     │
│                                     │
│  Ожидание: 15 сек...                │
│                                     │
└─────────────────────────────────────┘
```

**Rules:**
- Both players must have tickets
- 15 second acceptance window
- Decline → Return to menu
- Same matchmaking (no search needed)

---

## Rank Promotion/Demotion <!-- ⚠️ Events exist but frontend promotion/demotion animations not verified -->

### Promotion Animation
```
┌─────────────────────────────────────┐
│                                     │
│       ⬆️ ПОВЫШЕНИЕ РАНГА!           │
│                                     │
│    🥈 Silver I  →  🥇 Gold IV       │
│                                     │
│       ✨ Поздравляем! ✨             │
│                                     │
│  Новые награды сезона доступны!     │
│                                     │
└─────────────────────────────────────┘
```

### Demotion Warning
```
┌─────────────────────────────────────┐
│                                     │
│       ⚠️ ВНИМАНИЕ                    │
│                                     │
│  Ещё 1 поражение и ты упадёшь в     │
│  Silver I!                          │
│                                     │
│  Текущий MMR: 1,510                 │
│  Граница Gold IV: 1,500             │
│                                     │
└─────────────────────────────────────┘
```

**Protection zone:** No demotion at division floor for first 3 games at new rank. <!-- ✅ DemotionProtection=3 in player_rating.go -->

> ⚠️ **Расхождение:** События повышения/понижения ранга существуют в бэкенде, но анимации на фронтенде не верифицированы.

---

## State Management (Backend) <!-- ⚠️ Code uses waiting_start/in_progress/finished/abandoned — no COUNTDOWN state -->

**Duel states:**
```
MATCHMAKING → COUNTDOWN → IN_PROGRESS → COMPLETED
      ↓
   CANCELLED (player quit queue)
```

> ⚠️ **Расхождение:** Код использует состояния `waiting_start / in_progress / finished / abandoned`. Состояние `COUNTDOWN` отсутствует — код переходит напрямую из матчмейкинга в `in_progress`.

**Player states in duel:**
```
WAITING_QUESTION → ANSWERING → ANSWERED → WAITING_OPPONENT
```

**State stored on backend:**
- Current question index
- Both players' answers and times
- Running scores
- MMR before/after
- Game timestamp

**Frontend only tracks:**
- UI animations
- Timer visual
- Selected answer (before submit)

---

## Network Considerations

### Synchronization
- Backend controls question timing
- Both clients receive question at same server timestamp
- Answer submission includes client-side time measurement
- Server validates time against server clock (anti-cheat)

### Disconnect Handling <!-- ⚠️ HandlePlayerDisconnect exists, tracks status, but no 10s timer or 3-timeout forfeit in domain -->
**During matchmaking:**
- Disconnect → Auto-cancel queue
- Ticket refunded

**During duel:**
- 10s grace period to reconnect <!-- ⚠️ HandlePlayerDisconnect exists but 10s grace timer not implemented -->
- If reconnected → Resume (timer continues)
- If timeout → Auto-lose current question (no answer = wrong)
- 3 consecutive timeouts → Forfeit game, opponent wins <!-- ⚠️ 3-timeout forfeit logic not implemented in domain -->

> ⚠️ **Расхождение:** `HandlePlayerDisconnect` существует и отслеживает статус соединения, но 10-секундный таймер переподключения и логика форфейта после 3 таймаутов в домене не реализованы.

**Reconnect UI:**
```
┌─────────────────────────────────────┐
│                                     │
│     🔄 ПЕРЕПОДКЛЮЧЕНИЕ...           │
│                                     │
│  Дуэль продолжается!                │
│  Осталось: 7 сек                    │
│                                     │
└─────────────────────────────────────┘
```

---

## Bot Game (Fallback) <!-- ❌ Not implemented -->

If queue timeout (60s):
```
┌─────────────────────────────────────┐
│                                     │
│  Не удалось найти соперника         │
│                                     │
│  Хочешь сыграть с ботом?            │
│  (MMR не изменится)                 │
│                                     │
│  [  ДА  ]  [  ВЕРНУТЬСЯ  ]          │
│                                     │
└─────────────────────────────────────┘
```

> ❌ **Не реализовано:** Bot game fallback — бот-игры не реализованы.

**Bot behavior:**
- Answers with realistic timing (3-8s)
- Accuracy based on player's league (harder in higher ranks)
- Clearly labeled as "🤖 Bot"
- NO MMR change for bot games
- **Ticket handling:** Ticket is never consumed for bot games (player keeps their ticket)

---

## Additional Missing Features

> ❌ **Не реализовано:** Surrender button — кнопка сдаться после Q3 (нет endpoint'а).
> ❌ **Не реализовано:** Emotes — реакции в процессе игры.
