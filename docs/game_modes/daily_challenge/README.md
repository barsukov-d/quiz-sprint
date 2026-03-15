# Daily Challenge Documentation

> **Аудит реализации: 2026-03-15 | Обновлено: 2026-03-15 (Phase 8: leaderboard friends/country filtering)**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 9 | 1 | 1 | Нет анимации сундука |
| 02_gameplay.md | 10 | 4 | 0 | ABANDONED добавлен, лейблы добавлены |
| 03_rules.md | 13 | 4 | 0 | Anti-cheat добавлен (suspicious flag) |
| 04_rewards.md | 7 | 1 | 0 | Inventory работает, Premium stub |
| 05_api.md | 7 | 2 | 0 | Основные поля добавлены (rankLabel, chestLabel, shareText) |
| 06_domain.md | 10 | 5 | 3 | DailyQuiz структура другая, нет Redis |
| 07_edge_cases.md | 8 | 3 | 1 | Серверная валидация времени добавлена |
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
- [x] Second attempt (use case + реальное списание монет)
- [x] Frontend views (Play, Results, Review)
- [x] Streak recovery (use case + интеграция с inventoryService, 50 coins)
- [x] Premium подписка (interface + NoopPremiumService stub, nil-guarded)
- [x] User inventory (InventoryService: coins, tickets, bonuses — полностью работает)
- [x] Anti-cheat (suspicious score flag: avg time < 1s/question)
- [x] ABANDONED статус + CleanupAbandonedGamesUseCase (24h timeout)
- [x] Thin-client лейблы от бэкенда (RankLabel, ChestLabel, ShareText, CanRetry, RetryCost, CanPlayNow)
- [x] Ad verification interface (NoopAdVerificationService stub)
- [x] Swagger: RecoverStreak аннотации, ErrAlreadyAnswered → HTTP 409
- [x] Фильтрация лидерборда (friends/country) — FindTopByDateAndFriends + FindTopByDateAndCountry
- [ ] Chest opening анимация (frontend)
- [ ] Push notifications
- [ ] Redis-лидерборд (вместо PostgreSQL)
