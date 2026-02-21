# Marathon Endless Mode — Design

**Date:** 2026-02-21
**Status:** Approved
**Goal:** Remove artificial time-gate barriers so players can stay in Marathon as long as they want.

---

## Problem

Current Marathon flow creates forced stops:
1. Player runs out of 5 energy (lives) → Game Over
2. Options: pay coins for Continue OR wait 4 hours for energy regen
3. Result: players leave instead of playing more

---

## Solution: Energy + Instant Restart

### Core Changes

| Before | After |
|--------|-------|
| "Жизни" (❤️) | "Энергия" (⚡) — UI rename only, `MaxLives=5` unchanged |
| 0 lives → pay/wait | 0 energy → Continue (pay) OR instant new run (free) |
| 4-hour regen between runs | No regen gate — play as many runs as you want |
| Single run session | Session model: multiple runs, best score to leaderboard |

### Energy Mechanics (within a run)

| Event | Energy change |
|-------|---------------|
| Run start | ⚡⚡⚡⚡⚡ (5/5) |
| Wrong answer | −1 ⚡ |
| Wrong answer + shield active | 0 (shield absorbs) |
| 5 correct in a row (Marathon Momentum) | +1 ⚡ (capped at max) |
| Continue (coins / rewarded ad) | reset to 1 ⚡ |

### At Zero Energy — Two Paths

```
┌──────────────────────────────┐
│  ⚡ Энергия кончилась!        │
│  Забег: 47 правильных        │
│                              │
│  [ 200 💰 ] или [ 📺 ]       │  ← Continue (unchanged)
│                              │
│         — или —              │
│                              │
│  [ ▶ Новый забег ]           │  ← Free instant restart
└──────────────────────────────┘
```

"Новый забег" → results screen → new run starts with 5 ⚡, no cost, no wait.

---

## Session Model

A **session** = continuous time in Marathon without exiting. May contain multiple runs.

Session state is in-memory on the client (not persisted). Resets on app exit.

**Between-run screen:**
```
┌──────────────────────────────┐
│  🏁 Забег завершён           │
│                              │
│  ✅ 47 правильных            │
│  🔥 Лучшая серия: 12         │
│                              │
│  Эта сессия:                 │
│  Забег #2 | Лучший: 47       │
│                              │
│  [ ▶ Новый забег  ]          │
│  [ 📊 Лидерборд   ]          │
│  [ 🚪 Выйти       ]          │
└──────────────────────────────┘
```

**Leaderboard:** Best single run per week (unchanged from current model).

---

## Engagement Hooks

### 1. Cross-run Hot Streak 🔥

Correct answer streak does NOT reset between runs — only resets on a wrong answer.

This creates a powerful retention hook: player with streak 23 doesn't want to lose it.

**Implementation:** streak is stored in session state on frontend (not backend).
When starting a new run, the backend receives `sessionStreak` as a hint for display purposes only.
Score calculation is unaffected — streak is a UI/engagement feature.

### 2. Best Streak Record

Separate personal record: longest streak this week. Shown alongside best score.

```
Рекорд: 87 ✅  |  Серия: 23 🔥
```

### 3. Motivational Prompts (between runs)

Non-blocking text shown on the between-run screen:

```
"До рекорда 12 ответов. Ещё один забег?"
"Серия 15 — это твой лучший старт!"
"#127 на этой неделе. До топ-100: 8 ответов"
```

### 4. Daily Missions

```
📋 Задачи на сегодня:
  ☐ Ответить правильно 10 раз подряд  → +50 💰
  ☐ Дойти до вопроса 30              → +30 💰
  ☐ Сыграть 3 забега                 → +20 💰
```

Reset daily. Give a reason to return each day.

---

## What Doesn't Change

- Continue mechanic (pay coins / watch ad → +1 energy, stay in current run)
- Marathon Momentum (5 correct in a row → +1 energy) — already implemented
- Leaderboard: best run per week
- Adaptive difficulty within a run
- Bonus inventory (Shield, 50/50, Skip, Freeze)
- All API contracts

---

## Changes Required

### Backend

| Change | Scope |
|--------|-------|
| Remove time-based life regen gate (4-hour timer) | `value_objects.go` — remove `RegenerateLives()` call on session resume |
| "Energy" terminology in API responses | Swagger annotations + DTO field comments |
| New run can start immediately after previous game completed/abandoned | No code change needed — `StartMarathon` already allows new game after completed |

### Frontend

| Change | Scope |
|--------|-------|
| Rename "жизни/❤️" → "энергия/⚡" | All Marathon UI strings |
| Add "Новый забег" button to Game Over screen | `MarathonGameOver.vue` |
| Session state: track run count, session best | Composable `useMarathonSession` |
| Between-run screen | New view or modal |
| Cross-run streak display | `useMarathon` composable |
| Motivational prompts | Static data, shown on between-run screen |

### Documentation

| Change | Scope |
|--------|-------|
| Update `01_concept.md` | Remove "NO life regeneration", add session model |
| Update `02_gameplay.md` | New between-run screen wireframe, streak UX |
| Update `03_rules.md` | Energy table, cross-run streak rule |

---

## Out of Scope

- Daily missions (Phase 2 — separate feature)
- Best streak leaderboard (Phase 2)
- Monetization changes
- Cross-device session sync
