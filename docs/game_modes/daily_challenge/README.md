# Daily Challenge Documentation

> **Аудит реализации: 2026-03-15**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 6 | 2 | 3 | Streak recovery, Premium не реализованы |
| 02_gameplay.md | 6 | 6 | 2 | Нет ABANDONED статуса, нет анимации сундука |
| 03_rules.md | 7 | 6 | 4 | Нет валидации timeTaken, нет античита |
| 04_rewards.md | 5 | 1 | 2 | Нет inventory системы, нет Premium |
| 05_api.md | 2 | 7 | 0 | Response shapes расходятся, нет thin-client лейблов |
| 06_domain.md | 8 | 7 | 3 | DailyQuiz структура другая, нет Redis |
| 07_edge_cases.md | 5 | 3 | 4 | Нет серверной валидации времени |
| 08_frontend_integration.md | 4 | 4 | 2 | Thin client нарушения, нет ChestOpening |

## Quick Navigation

- **Backend domain**: `backend/internal/domain/daily_challenge/`
- **Backend app**: `backend/internal/application/daily_challenge/`
- **Backend handlers**: `backend/internal/infrastructure/http/handlers/daily_challenge_handlers.go`
- **Frontend views**: `tma/src/views/DailyChallenge/`
- **Frontend composable**: `tma/src/composables/useDailyChallenge.ts`

## Implementation Checklist

- [x] Domain model (DailyQuiz, DailyGame)
- [x] Basic gameplay flow (10 вопросов, 15с, feedback)
- [x] Streak system (5 тиров, immutable value object)
- [x] Leaderboard (PostgreSQL, не Redis как в доках)
- [x] Chest rewards (ChestRewardCalculator, 3 типа, вероятности)
- [x] Second attempt (use case есть, но оплата заглушена)
- [x] Frontend views (Play, Results, Review)
- [ ] Streak recovery
- [ ] Premium подписка
- [ ] User inventory (rewards хранятся, но не в inventory)
- [ ] Chest opening анимация
- [ ] Anti-cheat (timeTaken validation, server time check)
- [ ] ABANDONED статус + 24h timeout
- [ ] Thin-client лейблы от бэкенда
- [ ] Push notifications
