# Solo Marathon - Gameplay Flow

## Entry Point
Home → "Марафон" → Shows:
- Personal best (e.g., "Рекорд: 87")
- Current lives: ❤️❤️❤️
- Available bonuses: 🛡️×2 🔀×1 ❄️×3
- Weekly rank (if played this week)
- Coin balance (for continue budget planning)

---

## Game Flow

### 1. Pre-Start Screen
```
┌─────────────────────────────────────┐
│  🏃 МАРАФОН                  💰 1,250│
│                                     │
│  Твой рекорд: 87 правильных         │
│  Эта неделя: #342                   │
│                                     │
│  Доступные бонусы:                  │
│  🛡️ × 2   🔀 × 1   ❄️ × 3          │
│                                     │
│  Правила:                           │
│  • 3 жизни, ошибка = -1 жизнь       │
│  • Сложность растёт со временем     │
│  • Используй бонусы стратегически   │
│  💡 Continue: от 200💰               │
│                                     │
│  [      НАЧАТЬ МАРАФОН      ]       │
│                                     │
│  [ Лидерборд ]  [ Магазин бонусов ] │
└─────────────────────────────────────┘
```

---

### 2. Question Screen (In-Game)
```
┌─────────────────────────────────────┐
│  🏃 Марафон                         │
│  ❤️❤️❤️    ✅ 23    Вопрос 24  ⏱️ 11 │
│─────────────────────────────────────│
│                                     │
│  В каком году был основан Рим?      │
│  (по легенде)                       │
│                                     │
│  [ A. 753 г. до н.э.      ]         │
│  [ B. 509 г. до н.э.      ]         │
│  [ C. 27 г. до н.э.       ]         │
│  [ D. 476 г. н.э.         ]         │
│                                     │
│  ── Следующая цель: 25 ✅ (ещё 2) ──│
│                                     │
│─────────────────────────────────────│
│  Бонусы:                            │
│  [ 🛡️×2 ] [ 🔀×1 ] [ ⏭️×0 ] [ ❄️×3 ] │
└─────────────────────────────────────┘
```

**UI Elements:**
- **Score:** `✅ 23` — count of correct answers (primary metric)
- **Question number:** `Вопрос 24` — current question index
- **Lives:** Visual hearts (❤️ = active, 🖤 = lost)
- **Timer:** Countdown (color changes: green → yellow → red)
- **Milestone:** Progress toward next milestone (25, 50, 100, 200, 500)
- **Bonuses:** Active buttons (grayed if 0 quantity, long-press for tooltip)

**Timer behavior:**
- Questions 1-10: 15 seconds
- Questions 11-25: 12 seconds
- Questions 26-50: 10 seconds
- Questions 51+: 8 seconds

**Difficulty transition notification:**
When timer limit changes (e.g., question 11, 26, 51), show brief toast:
```
⚡ Сложность растёт! Время: 12 сек
```
Duration: 2 seconds, non-blocking.

**Difficulty indicators:**
- Question number visible (e.g., "Вопрос 47")
- No explicit "Easy/Hard" label

---

### 3. Bonus Usage

#### Using Shield 🛡️
```
Player taps Shield BEFORE answering
→ Visual indicator: "🛡️ Активен" above question
→ If answer wrong: Shield consumed, NO life lost
→ If answer correct: Shield NOT consumed, deactivates after question
→ Shield does NOT carry to next question (must re-activate manually)
```

#### Using 50/50 🔀
```
Player taps 50/50
→ 2 wrong answers fade out instantly
→ 2 answers remain
→ Bonus consumed (regardless of correctness)
```

#### Using Skip ⏭️
```
Player taps Skip
→ Question skipped immediately
→ Next question appears
→ NO score increment (doesn't count as wrong)
→ NO life lost
```

#### Using Freeze ❄️
```
Player taps Freeze
→ Timer +10 seconds instantly
→ Visual effect: ❄️ animation
→ Can use multiple freezes on same question
```

---

### 4. Answer Feedback (Immediate)

**Correct Answer:**
```
┌─────────────────────────────────────┐
│          ✅ ПРАВИЛЬНО!              │
│                                     │
│  Счёт: ✅ 24                        │
│                                     │
│  [ Далее ]                          │
└─────────────────────────────────────┘
```
Duration: 1.5 seconds → Auto-advance

**Wrong Answer (with lives left):**
```
┌─────────────────────────────────────┐
│          ❌ НЕПРАВИЛЬНО              │
│                                     │
│  Правильный ответ: A. 753 г. до н.э│
│  -1 жизнь: ❤️❤️🖤                   │
│  Счёт: ✅ 23                        │
│                                     │
│  [ Продолжить ]                     │
└─────────────────────────────────────┘
```
Duration: 3 seconds (read answer) → Continue

**Wrong Answer (last life) → See section 5**

---

### 5. Game Over Screen
```
┌─────────────────────────────────────┐
│          💀 ИГРА ОКОНЧЕНА            │
│                                     │
│  Твой результат: 47 правильных      │
│  [████████░░░░░░] 47/87 рекорда     │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  Хочешь продолжить?         │    │
│  │  Получи ещё одну жизнь!     │    │
│  │                             │    │
│  │  [ 200 💰 ] или [ 📺 ]      │    │
│  │  💰 Баланс: 1,250           │    │
│  └─────────────────────────────┘    │
│                                     │
│  [ Закончить забег ]                │
└─────────────────────────────────────┘
```

**Continue options:**
1. Pay coins (show current balance)
2. Watch rewarded ad
3. Decline → Go to results

**Progress bar:** Visual comparison to personal best — motivates continue.

If continued:
- Lives reset to 1 (❤️) — NOT "+1", always exactly 1
- Resume from same question
- Next continue costs more (400, 600, 800, ...)

---

### 6. Results Screen (Final)
```
┌─────────────────────────────────────┐
│  🏁 ФИНАЛЬНЫЙ РЕЗУЛЬТАТ             │
│                                     │
│  Правильных ответов: 47             │
│  Использовано бонусов:              │
│    • 🛡️ Щит: 2                     │
│    • ❄️ Заморозка: 3                │
│                                     │
│  Твой рекорд: 87                    │
│  [████████░░░░░░] 47/87 (54%)       │
│                                     │
│  Позиция на этой неделе: #127       │
│  До топ-100 не хватает 12 ответов!  │
│                                     │
│  [  ИГРАТЬ ЕЩЁ РАЗ  ]               │
│                                     │
│  [ Лидерборд ]  [ Поделиться ]      │
└─────────────────────────────────────┘
```

**If new personal record:**
```
🎉 НОВЫЙ РЕКОРД! 🎉
+500 монет за достижение
```

**Share card format** (for "Поделиться" button):
```
🏃 Мой марафон в Quiz Sprint!
✅ 47 правильных ответов
🏆 #127 на этой неделе
Попробуй побить мой рекорд!
[link to app]
```

---

## State Management (Backend)

**Game states:**
```
IN_PROGRESS → GAME_OVER → COMPLETED (declined continue / final game over)
     │             │
     │             └─→ IN_PROGRESS (continue used)
     │
     └─→ ABANDONED (player quit mid-game or timeout 30min)
```

**Statuses:**
- `in_progress` — actively playing
- `game_over` — 0 lives, continue offer shown
- `completed` — game ended normally (after game over, no more continues)
- `abandoned` — player quit mid-game or inactivity timeout

**State stored on backend:**
- Current question index
- Lives remaining (0-3)
- Correct answers count
- Bonus inventory state
- Continue count
- All answers history

**Frontend only tracks:**
- UI animations
- Timer visual
- Selected answer ID (before submit)

---

## Adaptive Difficulty Details

### Timer Progression
```
Questions 1-10:   15s
Questions 11-25:  12s
Questions 26-50:  10s
Questions 51-100: 8s
Questions 101+:   8s (no lower)
```

### Question Selection Algorithm
Backend selects questions:
1. Questions 1-10: `difficulty = 'easy' OR 'medium'` (80% easy, 20% medium)
2. Questions 11-30: `difficulty = 'medium'` (100%)
3. Questions 31-50: `difficulty = 'medium' OR 'hard'` (70% medium, 30% hard)
4. Questions 51+: `difficulty = 'hard'` (100%)

### Category Narrowing (Optional Enhancement)
- Early: Broad categories (Geography, History)
- Late: Narrow categories (Byzantine History, Molecular Biology)

---

## Bonus Tooltips (Long-Press)

Each bonus button shows tooltip on long-press:

| Bonus | Tooltip text |
|-------|-------------|
| 🛡️ Shield | "Защита от ошибки. Не тратится при правильном ответе" |
| 🔀 50/50 | "Убирает 2 неправильных ответа" |
| ⏭️ Skip | "Пропустить вопрос без потери жизни" |
| ❄️ Freeze | "+10 секунд к таймеру" |

**First 3 games:** Show labels under icons by default (no long-press needed).
After 3 games → icons only + long-press tooltips.

---

## Bonus Strategy Tips (In-Game Hints)

**Shown on loading screen / first-time tutorial:**

```
💡 Совет: Используй Щит 🛡️ только когда у тебя 1 жизнь!

💡 Совет: 50/50 🔀 лучше работает на вопросах с числами.

💡 Совет: Заморозка ❄️ критична после 50-го вопроса!

💡 Совет: Пропуск ⏭️ не портит твою серию правильных ответов.
```

---

## First-Time Onboarding

**Trigger:** Player's first Marathon game ever (`gamesPlayed == 0`).

**Flow:**
1. Question 1 — normal question, no hints
2. Question 2 — highlight Shield button: "Нажми 🛡️ чтобы защититься от ошибки"
3. Question 3 — highlight 50/50 button: "Нажми 🔀 чтобы убрать 2 неправильных ответа"
4. Question 4+ — normal gameplay, no more hints

**Implementation:**
- Backend returns `isOnboarding: true` + `onboardingStep: 2` in question response
- Frontend shows animated highlight on specified bonus button
- Hints are non-blocking (player can ignore and play normally)

---

## Pause / Quit

**Pause NOT allowed** (integrity of run).

**Quit:**
- Shows warning: "Прогресс будет потерян!"
- If confirmed: Game state saved as `ABANDONED`
- Cannot resume (fresh start only)

---

## Network Issues

**Disconnect during game:**
- State saved after each answer
- Can resume on reconnect (same question)
- Timer paused (server-side tracking)

**Disconnect UI overlay:**
```
┌─────────────────────────────────────┐
│                                     │
│        🔄 Переподключение...         │
│                                     │
│    Твой прогресс сохранён.          │
│    Таймер на паузе.                 │
│                                     │
└─────────────────────────────────────┘
```
Overlay blocks interaction until reconnect. On reconnect → resume seamlessly.

**Timeout:**
- If no answer for 30+ seconds: Auto-submit empty → Wrong answer

---

## Edge Cases

**Used Shield but answered correctly:**
- Shield NOT consumed (saved for later)

**Multiple bonuses on same question:**
- Can use: Freeze + 50/50 (both consumed)
- Can use: Freeze + Shield (Shield only if wrong)

**Skip after using 50/50:**
- Both bonuses consumed (50/50 already applied)

**Continue with bonuses still available:**
- Bonuses persist (not reset)
