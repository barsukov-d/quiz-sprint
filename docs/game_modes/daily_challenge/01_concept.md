# Daily Challenge - Concept

> **Статус реализации (аудит 2026-03-15)**
> ✅ Реализовано: 6 | ⚠️ Расходится: 2 | ❌ Не реализовано: 3

## Changes

| Date | Change | Reason |
|------|--------|--------|
| 2026-01-31 | Feedback: `after completion only` → `instant (after each answer)` | UX: player sees correct/incorrect immediately after answering |

## What?
Central daily event where all players answer same 10 questions. <!-- ✅ -->

## Why?
**Primary goal:** Earn resources (PvP tickets, coins, bonuses) for other game modes.
**Secondary:** Global competition, streak building. <!-- ✅ Реализовано через leaderboard -->

## For Whom?
- Core players: daily resource farming
- Casual players: quick 3-5 min session
- Competitive players: leaderboard rankings

## Key Mechanics

| Parameter | Value | Status |
|-----------|-------|--------|
| Questions | 10 | <!-- ✅ --> |
| Time per question | 15 sec | <!-- ✅ --> |
| Attempts (free) | 1/day | <!-- ✅ --> |
| Reset | 00:00 UTC daily | <!-- ✅ --> |
| Feedback | Instant (after each answer) | <!-- ✅ --> |

## Core Loop
```
Play Daily → Earn Chest → Get Resources → Use in PvP/Marathon → Return Tomorrow
```

## Main Reward: Daily Chest

3 types based on performance: <!-- ✅ -->

| Correct Answers | Chest Type | Contains |
|-----------------|------------|----------|
| 0-4 / 10 | 🪵 Wooden | Few coins, 1 PvP ticket |
| 5-7 / 10 | 🥈 Silver | More coins, 2-3 tickets, bonus chance |
| 8-10 / 10 | 🏆 Golden | Many coins, 4-5 tickets, guaranteed bonuses |

## Streak System
Consecutive days played → multiplier to chest rewards. <!-- ⚠️ расходится: код содержит 5 градаций (добавлена 14d=+40%), документ описывает только 3. Полная таблица в 03_rules.md -->

| Days | Bonus |
|------|-------|
| 3 | +10% |
| 7 | +25% |
| 30+ | +50% |

## Monetization

| Feature | Cost | Effect | Status |
|---------|------|--------|--------|
| Second attempt | 100 coins / Ad | Better chest chance | <!-- ⚠️ расходится: use case существует, но монеты НЕ списываются (TODO stub), реклама не верифицирована (TODO stub) --> |
| Streak recovery | 50 coins / Ad | Don't lose streak | <!-- ❌ не реализовано --> |
| Premium subscription | Monthly | Auto-upgrade chest +1 tier | <!-- ❌ не реализовано (premium захардкожен как false) --> |

## Success Metrics

- Daily active users: > 60%
- Chest open rate: > 98%
- Second attempt conversion: > 15%
- 7-day streak retention: > 40%
