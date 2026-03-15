# Solo Marathon Documentation

> **Аудит реализации: 2026-03-15 | Обновлено: 2026-03-15 (Phase 8: freeze fix verified + difficulty toasts)**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 12 | 2 | 1 | Нет shop |
| 02_gameplay.md | 13 | 4 | 3 | feedback 1.8с вместо 3с |
| 03_rules.md | 21 | 3 | 3 | Anti-cheat добавлен, weekly rewards use case готов |
| 04_rewards.md | 9 | 2 | 5 | Milestone rewards добавлены, weekly use case готов |
| 05_api.md | 5 | 4 | 1 | Complete endpoint добавлен, correctAnswerText добавлен |
| 06_domain.md | 13 | 5 | 3 | Streak персистируется, нет Redis |
| 07_edge_cases.md | 12 | 4 | 1 | Abandon NOT updating PB (исправлено) |
| 08_frontend_integration.md | 6 | 8 | 5 | Компоненты inline, нет отдельных виджетов |

## Quick Navigation

- **Backend domain**: `backend/internal/domain/solo_marathon/`
- **Backend app**: `backend/internal/application/marathon/`
- **Backend handlers**: `backend/internal/infrastructure/http/handlers/marathon_handlers.go`
- **Frontend views**: `tma/src/views/Marathon/`
- **Frontend composable**: `tma/src/composables/useMarathon.ts`

## Implementation Checklist

- [x] Domain model (MarathonGameV2, LivesSystem, BonusInventory)
- [x] Lives system (5 жизней, game over, regen каждые 5 streak)
- [x] Bonus mechanics (Shield, 50/50, Skip, Freeze)
- [x] Adaptive difficulty (timer 15→8s, вопросы сложнее)
- [x] Continue flow (формула cost, lives=1, ad flag)
- [x] All-time leaderboard (PostgreSQL)
- [x] Frontend views (Category, Play, GameOver)
- [x] Bonus UI controls (inline в PlayView)
- [x] Personal best tracking
- [x] POST /marathon/:gameId/complete endpoint (отдельный от abandon)
- [x] Coin deduction for continues (реально работает через InventoryService)
- [x] Milestone rewards (25→100, 50→250, 100→500, 200→1000, 500→5000 coins)
- [x] Weekly reward distribution use case (DistributeWeeklyMarathonRewardsUseCase, 5 tier'ов)
- [x] Anti-cheat (suspicious flag для score > 200)
- [x] marathon_bonus_usage table (migration 024, аналитика бонусов)
- [x] API: CorrectAnswerText в answer output, CanStart в status
- [x] Fix: Abandon should NOT update personal best (dead code removed)
- [x] Fix: streakCount/bestStreak/livesRestored persisted (migration 023)
- [x] Fix: CurrentStreak()/MaxStreak() getters fixed (returned score → streak)
- [ ] Weekly leaderboard (Redis sorted set, needs infra implementation)
- [ ] WeeklyRewardDistributionRepository (PostgreSQL impl needed)
- [ ] Milestone deduplication (prevent re-crediting same threshold)
- [ ] Bonus shop / packs
- [x] Difficulty transition toasts (frontend) — toast shown on difficulty level up
- [x] Fix: Freeze frontend +5s → +10s (confirmed correct, matches backend FREEZE_BONUS_SECONDS=10)
- [ ] Onboarding overlay (frontend)
- [ ] Network disconnect overlay (frontend)
- [ ] Share card (frontend)
