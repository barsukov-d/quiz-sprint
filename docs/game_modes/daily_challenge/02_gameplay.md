# Daily Challenge - Gameplay Flow

> **Статус реализации (аудит 2026-03-15)**
> ✅ Реализовано: 6 | ⚠️ Расходится: 6 | ❌ Не реализовано: 2

## Changes

| Date | Change | Reason |
|------|--------|--------|
| 2026-01-31 | Question screen: `NO feedback` → `instant feedback after each answer` | Player must see correct/incorrect right away |
| 2026-01-31 | Added section `2b. Feedback State` with wireframe | New UI state for answer feedback |
| 2026-01-31 | Removed section `6. Review Mistakes` | Redundant — feedback is now shown inline during gameplay |
| 2026-01-31 | Pre-game screen text: `Результаты покажутся в конце` → `Мгновенная обратная связь` | Reflects new feedback behavior |
| 2026-02-01 | Question screen: compact header (1 line), question text is primary focus | Question was buried below progress/timer |
| 2026-02-01 | Feedback: all non-selected wrong answers → muted/dimmed | Only selected + correct should stand out |
| 2026-02-01 | Removed permanent "Select your answer" alert | Redundant hint, wastes space |
| 2026-02-01 | Answer buttons: Tailwind-only, no custom CSS | Consistency with project style |
| 2026-02-01 | Card (Completed): score first, no progress bar, compact streak | Score is what user wants to see |
| 2026-02-01 | Card (Completed): removed "Next: milestone" progress | Secondary info, moved to results page |
| 2026-02-01 | Card (Not Played): simpler layout, no redundant badges | Reduced visual noise |

## Entry Point
Home screen → "Daily Challenge" button → Shows:
- Today's date
- Streak counter (🔥 5 days)
- "Already played" or "Start" button

## Flow Steps

### 1a. Card — Not Played (Home Screen)

> ⚠️ **Расходится:** данные доступны через `/status` API, но точное соответствие UI карточки не верифицировано

```
┌─────────────────────────────────────┐
│  📅 Today's Challenge               │
│                                     │
│  10 questions • 15s each            │
│  🔥 1 day streak                    │
│                                     │
│  [     Start Challenge     ]        │
│                                     │
│  ⏱ 14:43:47        👥 1 player     │  ← small, gray
└─────────────────────────────────────┘
```

**Layout:**
- Title: "Today's Challenge" (no redundant "Daily Challenge - Available" + badge)
- Challenge info: bullets, one line
- Streak: inline, emoji + count
- Action: primary button
- Meta: reset timer + players (small, bottom)

### 1b. Card — Completed (Home Screen)

> ⚠️ **Расходится:** данные доступны через `/status` API, но точное соответствие UI карточки не верифицировано

```
┌─────────────────────────────────────┐
│  📅 Today's Challenge          ✓    │
│                                     │
│         178 points                  │  ← primary focus
│         🔥 1 day streak             │
│                                     │
│  [     View Results     ]           │
│                                     │
│  ⏱ 14:43:47        👥 1 player     │
└─────────────────────────────────────┘
```

**Layout:**
- Title: "Today's Challenge" + checkmark icon (no "Completed" badge duplication)
- **Score first** — large, center-aligned (what user wants to see)
- Streak: below score, one line
- Action: gray button "View Results"
- Meta: same as not-played
- **No progress bar** (game done, 100% is obvious)
- **No "Next: milestone"** (moved to results page)

### 1c. Card — In Progress (Home Screen)

> ⚠️ **Расходится:** данные доступны через `/status` API, но точное соответствие UI карточки не верифицировано

```
┌─────────────────────────────────────┐
│  📅 Today's Challenge        🕐     │
│                                     │
│  Question 5/10                      │
│  ━━━━━━━━━━━━━━━━━━━━━━ 50%        │
│                                     │
│  [      Continue      ]             │
│                                     │
│  ⏱ 14:43:47        👥 1 player     │
└─────────────────────────────────────┘
```

**Layout:**
- Title + clock icon (not "In Progress" text badge)
- Progress: question count + thin bar
- Action: primary button "Continue"
- Meta: same

### 2. Question Screen

> ⚠️ **Расходится:** счётчик и заголовок в одной строке, но таймер реализован как отдельная градиентная полоса снизу, а не inline UProgress как в спеке

**Layout priority (top → bottom):** compact header → question text (primary) → answers

```
┌─────────────────────────────────────┐
│  3/10   ━━━━━━━━━━━━━━━━━━   00:12  │  ← single line: counter + progress + timer
│                                     │
│                                     │
│  В каком году основали Москву?      │  ← large text, primary focus
│                                     │
│                                     │
│  ┌─ A ── 1147 год ────────────────┐ │
│  ├─ B ── 1240 год ────────────────┤ │  ← full-width answer buttons
│  ├─ C ── 988 год ─────────────────┤ │
│  └─ D ── 1380 год ────────────────┘ │
└─────────────────────────────────────┘
```

**Header:** single row, 3 elements inline:
- Left: `3/10` (question counter, text only)
- Center: thin `UProgress` bar
- Right: `00:12` timer (mono font, color changes: green → orange → red)

**Question:** `text-xl` / `text-2xl`, no card wrapper, just text with vertical padding.

**Behavior:**
- Timer counts down from 15
- Answer locks after selection (no change)
- Auto-submit at 0:00 (counts as wrong) — ⚠️ **Расходится:** авто-сабмит отправляет ПЕРВЫЙ вариант ответа вместо пустого — первый вариант может оказаться правильным
- **Instant feedback** after each answer (see 2b)

### 2b. Feedback State (after answer selected)

> ✅ **Реализовано:** все 4 состояния кнопок, блокировка кнопок, остановка таймера, авто-переход

Example: user selected C (wrong), correct is A:
```
┌─────────────────────────────────────┐
│  3/10   ━━━━━━━━━━━━━━━━━━   00:08  │
│                                     │
│  В каком году основали Москву?      │
│                                     │
│  ┌─ A ── 1147 год ─── ✓ ─────────┐ │ ← green bg + border
│  ├─ B ── 1240 год ───────────────┤ │ ← muted (opacity-40)
│  ├─ C ── 988 год ──── ✗ ─────────┤ │ ← red bg + border (user's pick)
│  └─ D ── 1380 год ───────────────┘ │ ← muted (opacity-40)
└─────────────────────────────────────┘
```

**Feedback rules (4 states per button):** ✅

| Condition | Style | Icon |
|-----------|-------|------|
| Correct answer | `bg-green`, `border-green`, full opacity | `✓` checkmark |
| Selected + wrong | `bg-red`, `border-red`, full opacity | `✗` cross |
| Not selected + not correct | `opacity-40`, no border change | none |
| Selected + correct | `bg-green`, `border-green`, full opacity | `✓` checkmark |

**Timing:** ✅
- All answer buttons **disabled** during feedback ✅
- Timer **stops** during feedback ✅
- Auto-transition to next question after **1.5s** ✅
- Backend `submitAnswer` returns `{ isCorrect, correctAnswerId }` — frontend renders feedback from this

### 3. Progress Indicator
- `3/10` shown inline left in header
- Thin progress bar between counter and timer
- Timer always visible, color-coded (green > 5s, orange <= 5s, red = 0)

### 4. Completion Screen

> ⚠️ **Расходится:** экран результатов показывает счёт, ранг и тип сундука, но отдельная кнопка "Открыть сундук" и её анимация отсутствуют

```
┌─────────────────────────────────────┐
│  📅 РЕЗУЛЬТАТЫ ДНЯ                  │
│  24 января 2026                     │
│                                     │
│  Твой результат: 8/10 ✓             │
│  Счёт: 920 очков                    │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  🏆 Твоя позиция: #847      │    │
│  │  из 12,847 игроков          │    │
│  └─────────────────────────────┘    │
│                                     │
│  Твоя награда:                      │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  🏆 ЗОЛОТОЙ СУНДУК          │    │
│  └─────────────────────────────┘    │
│                                     │
│  [  ✨ ОТКРЫТЬ СУНДУК ✨   ]         │
│                                     │
│  [ Поделиться ]  [ Лидерборд ]      │
└─────────────────────────────────────┘
```

### 5. Chest Opening

> ❌ **Не реализовано:** компонент `ChestOpening.vue` отсутствует, анимация открытия сундука не реализована

Animation → Shows rewards:
```
┌─────────────────────────────────────┐
│         🏆 ЗОЛОТОЙ СУНДУК           │
│                                     │
│  💰 +420 монет                      │
│  🎟️ +5 PvP билетов                 │
│  🛡️ +1 Щит (Marathon)              │
│  ❄️ +1 Заморозка (Marathon)        │
│                                     │
│  Бонус серии: +25%                  │
│                                     │
│  [        ЗАБРАТЬ        ]          │
└─────────────────────────────────────┘
```

## Navigation

- Question screen: NO back button (can't change answers) ✅
- Can quit mid-game → Shows "abandon" warning → Game saved as incomplete — ❌ **Не реализовано:** диалог предупреждения об abandon отсутствует, route guard не установлен

## States

> ⚠️ **Расходится:** статус `ABANDONED` отсутствует в коде — только `in_progress` и `completed`

```
NOT_STARTED → IN_PROGRESS → COMPLETED
              ↓
           ABANDONED (24h timeout)
```

## Edge Cases

**Disconnect during game:**
- State saved after each answer
- Can resume on reconnect
- Timer continues (may lose time)

**Started at 23:58, finished at 00:02:**
- Belongs to START date
- Streak updated for START date

**Second attempt:**
- Only after completion
- Button: "Попробовать ещё раз (100💰 / 📺)"
- Creates NEW game, old result kept for leaderboard
