# Daily Challenge Documentation

> **Аудит реализации: 2026-03-15 | Обновлено: 2026-03-15 (Phase 9: streak recovery, ABANDONED, anti-cheat, thin-client labels, leaderboard filtering, ErrAlreadyAnswered→409, retry coins)**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 7 | 2 | 2 | Premium stub, нет анимации сундука |
| 02_gameplay.md | 8 | 4 | 2 | ABANDONED ✅, лейблы ✅, нет ChestOpening анимации |
| 03_rules.md | 10 | 5 | 2 | Recovery ✅, retry coins ✅, anti-cheat flag ✅; timeTaken валидация ❌ |
| 04_rewards.md | 7 | 1 | 0 | Inventory работает, Premium stub |
| 05_api.md | 7 | 2 | 0 | rankLabel/chestLabel/shareText ✅, leaderboard filtering ✅, RecoverStreak ✅ |
| 06_domain.md | 10 | 6 | 2 | ABANDONED ✅, DailyQuiz структура другая, нет Redis |
| 07_edge_cases.md | 7 | 3 | 2 | ABANDONED ✅, ErrAlreadyAnswered 409 ✅, SuspiciousScore ✅ |
| 08_frontend_integration.md | 5 | 3 | 2 | rankLabel/chestLabel от бэкенда ✅, нет ChestOpening |

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
- [x] Second attempt (use case + реальное списание монет)
- [x] Frontend views (Play, Results, Review)
- [x] Streak recovery (use case + интеграция с inventoryService, 50 coins)
- [x] Premium подписка (interface + NoopPremiumService stub, nil-guarded)
- [x] User inventory (InventoryService: coins, tickets, bonuses — полностью работает)
- [x] Anti-cheat (suspicious score flag: avg time < 1s/question)
- [x] ABANDONED статус + CleanupAbandonedGamesUseCase (24h timeout)
- [x] Thin-client лейблы от бэкенда (RankLabel, ChestLabel, ShareText, CanRetry, RetryCost, CanPlayNow)
- [x] Ad verification interface (NoopAdVerificationService wired in routes.go)
- [x] Swagger: RecoverStreak аннотации, ErrAlreadyAnswered → HTTP 409
- [x] Фильтрация лидерборда (friends/country) — FindTopByDateAndFriends + FindTopByDateAndCountry
- [ ] Chest opening анимация (frontend — `ChestOpening.vue` отсутствует)
- [ ] Push notifications
- [ ] Redis-лидерборд (вместо PostgreSQL)
- [ ] timeTaken диапазон валидация (0-15s) на сервере
