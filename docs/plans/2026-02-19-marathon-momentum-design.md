# Marathon Momentum — Design Doc

**Date:** 2026-02-19
**Branch:** marathon-update
**Status:** Approved

## Problem

3 lives → game over in 5 questions on bad streak. Too harsh for 5-10 min session target.

## Solution: Marathon Momentum

Increase starting lives + add streak-based regeneration.

---

## Mechanics

### Lives System

| Parameter | Before | After |
|-----------|--------|-------|
| Starting lives | 3 | **5** |
| Max lives | 3 | **5 (fixed)** |
| Life regen | none | **+1 every 5 correct in a row** |
| Max cap progression | none | **none** |

### Streak Rules

```
streak_count: resets on any wrong answer (including Shield-protected)
streak_count++ on correct answer
if streak_count % 5 == 0 && lives < max_lives:
    lives += 1
    emit LifeRestored
```

**Shield interaction:** Shield prevents life loss but wrong answer still resets streak.
Rationale: bonus protects life, not correctness.

---

## Scoring

No changes. `score = correctAnswers`. Streak is survival mechanic only.

End-of-game stats (UI only, not in ranking):
- Best streak: `bestStreak`
- Lives restored count: `livesRestored`

Leaderboard tiebreaker unchanged: `correctAnswers DESC`, `totalQuestions ASC`, `completedAt ASC`.

---

## Domain Changes

### MarathonSession — new fields

```go
type MarathonSession struct {
    // existing fields unchanged...
    Lives         int // 5 (was 3)
    MaxLives      int // 5 (fixed)
    StreakCount   int // current streak, resets on wrong answer
    BestStreak    int // peak streak this session
    LivesRestored int // total times life was restored
}
```

### Constants

```go
const (
    MarathonStartLives    = 5  // was 3
    MarathonMaxLives      = 5  // fixed cap
    MarathonStreakForRegen = 5  // correct streak threshold for +1 life
)
```

---

## API Changes

### SubmitAnswer response — additional fields

```json
{
  "isCorrect": true,
  "lives": 4,
  "maxLives": 5,
  "streakCount": 3,
  "lifeRestored": false
}
```

`lifeRestored: true` → frontend shows ❤️+1 animation.

### No DB schema changes

`streakCount`, `bestStreak`, `livesRestored` stored in session (memory/Redis).
`bestStreak` and `livesRestored` written to session result on game over.

---

## Frontend Changes

1. Lives bar: render 5 hearts (was 3)
2. Streak counter: show current streak (e.g. `🔥 3`)
3. Life restored animation: pulse ❤️+1 when `lifeRestored: true`
4. End screen: show `bestStreak` and `livesRestored`

---

## Edge Cases

| Scenario | Behaviour |
|----------|-----------|
| Regen at max lives | Skip regen, streak continues counting |
| Shield on wrong answer | Life saved, streak resets to 0 |
| Skip bonus | No streak change (question skipped, not answered) |
| Continue after game over | Lives reset to 1 (unchanged), streak resets to 0 |
