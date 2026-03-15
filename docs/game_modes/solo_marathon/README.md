# Solo Marathon Documentation

> **Аудит реализации: 2026-03-15**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 9 | 3 | 3 | Нет weekly leaderboard, нет shop, coins не списываются |
| 02_gameplay.md | 10 | 6 | 4 | Freeze +5 вместо +10, feedback 1.8с вместо 3с |
| 03_rules.md | 15 | 4 | 8 | Leaderboard sort по streak, нет античита, нет weekly |
| 04_rewards.md | 3 | 2 | 11 | Система наград почти полностью отсутствует |
| 05_api.md | 2 | 7 | 1 | Response shapes расходятся, нет thin-client лейблов |
| 06_domain.md | 10 | 8 | 3 | Нет Redis, сервисы встроены в value objects |
| 07_edge_cases.md | 8 | 6 | 3 | Abandon обновляет personal best (баг) |
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
- [ ] Weekly leaderboard (Redis, Mon-Sun UTC)
- [ ] Reward distribution (coins, badges, packs)
- [ ] Coin deduction for continues (TODO stub)
- [ ] Personal best +500 coins bonus
- [ ] Milestone rewards (tracked but no rewards)
- [ ] Bonus shop / packs
- [ ] Anti-cheat (timeTaken validation, flags)
- [ ] Difficulty transition toasts
- [ ] Onboarding overlay
- [ ] Network disconnect overlay
- [ ] Share card
- [ ] Fix: Freeze frontend +5s → +10s
- [ ] Fix: Abandon should NOT update personal best
- [ ] Fix: Leaderboard sort by score DESC (not streak)
- [ ] Fix: streakCount lost on DB reconstruction
