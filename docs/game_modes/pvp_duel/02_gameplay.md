# PvP Duel - Gameplay Flow

## Entry Point
Home → "Дуэль" → Shows:
- Current league + MMR (e.g., "🥇 Gold III — 1,650 MMR")
- PvP tickets available: 🎟️×3
- Win/Loss record this season
- **Two main buttons:** "Случайный соперник" / "Вызвать друга"
- Friends online indicator

---

## Game Flow

### 1. Pre-Game Screen
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

### 1b. Friend Challenge Flow
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

### 2. Matchmaking Screen (Random)
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

**Queue expansion:**
- 0-10s: ±50 MMR
- 10-20s: ±100 MMR
- 20-30s: ±200 MMR
- 30-45s: ±300 MMR
- 45-60s: ±500 MMR
- 60s+: Offer bot game

---

### 3. Opponent Found Screen
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

### 4. Question Screen (In-Duel)
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

**Real-time opponent status:**
- Show when opponent answers (⏳ → ✅/❌)
- Do NOT show which answer they selected
- Create tension without revealing strategy

---

### 5. Answer Feedback (Brief)
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

### 6. Game Result Screen
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
- **Rematch button** (if opponent accepts)
- **Share button** (generates victory card)

---

### 6b. Share Victory Card
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

**Share link behavior:**
- Clicking link → Opens TMA directly with `?startapp=duel_xxx` parameter
- TMA extracts `startParam`, authenticates user, navigates to duel lobby
- Deep link handler auto-accepts challenge via `POST /duel/challenge/accept-by-code`
- New user → Registers first, then challenge is auto-accepted
- Existing user → Direct to duel lobby

---

### 7. Defeat Screen
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
│  💡 Совет: Быстрее отвечай на       │
│  вопросы по науке — твоё слабое     │
│  место.                             │
│                                     │
│  [  РЕВАНШ  ]  [  В МЕНЮ  ]         │
└─────────────────────────────────────┘
```

---

### 8. Tiebreaker Result
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

---

## Rematch Flow

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

## Rank Promotion/Demotion

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

**Protection zone:** No demotion at division floor for first 3 games at new rank.

---

## State Management (Backend)

**Duel states:**
```
MATCHMAKING → COUNTDOWN → IN_PROGRESS → COMPLETED
      ↓
   CANCELLED (player quit queue)
```

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

### Disconnect Handling
**During matchmaking:**
- Disconnect → Auto-cancel queue
- Ticket refunded

**During duel:**
- 10s grace period to reconnect
- If reconnected → Resume (timer continues)
- If timeout → Auto-lose current question (no answer = wrong)
- 3 consecutive timeouts → Forfeit game, opponent wins

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

## Bot Game (Fallback)

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

**Bot behavior:**
- Answers with realistic timing (3-8s)
- Accuracy based on player's league (harder in higher ranks)
- Clearly labeled as "🤖 Bot"
- NO MMR change for bot games
- **Ticket handling:** Ticket is never consumed for bot games (player keeps their ticket)
