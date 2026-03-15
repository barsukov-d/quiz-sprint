# PvP Duel Documentation

> **Аудит реализации: 2026-03-15 | Обновлено: 2026-03-15 (Phase 8: bot fallback + same-opponent prevention)**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 14 | 1 | 3 | Нет emotes, cosmetics, ticket shop |
| 02_gameplay.md | 18 | 2 | 4 | Нет emotes, victory cards, push notifications |
| 03_rules.md | 11 | 4 | 7 | Scoring+K-factor исправлены, surrender добавлен |
| 04_rewards.md | 6 | 1 | 14 | Referral claiming + seasonal distribution добавлены |
| 05_api.md | 27 | 1 | 0 | Surrender + referral endpoints добавлены |
| 06_domain.md | 22 | 11 | 1 | Surrender, BotFallback, CalculateDrawRating добавлены |
| 07_edge_cases.md | 13 | 11 | 3 | Tiebreaker, bot game, same-opponent исправлены |

## Quick Navigation

- **Backend domain**: `backend/internal/domain/quick_duel/`
- **Backend app**: `backend/internal/application/quick_duel/`
- **Backend handlers**: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- **WebSocket**: `backend/internal/infrastructure/http/handlers/duel_websocket_handler.go`
- **Lobby WebSocket**: `backend/internal/infrastructure/http/handlers/duel_lobby_hub.go`
- **Redis queue**: `backend/internal/infrastructure/persistence/redis/matchmaking_queue.go`
- **Frontend views**: `tma/src/views/Duel/`
- **Frontend composables**: `tma/src/composables/usePvPDuel.ts`, `useDuelWebSocket.ts`, `useLobbyWebSocket.ts`

## Implementation Checklist

### Phase 1: Core Duel (MVP)
- [x] Domain model (DuelGame, DuelPlayer, EloRating, RoundAnswer)
- [x] PlayerRating (ELO, leagues, divisions, demotion protection)
- [x] Matchmaking queue (Redis sorted set)
- [x] WebSocket server for real-time sync
- [x] Question distribution (same for both)
- [x] Answer submission + scoring (points: 100 base + speed bonus)
- [x] Anti-cheat: min answer time 200ms
- [x] Game result screen (DuelResultsView.vue)

### Phase 2: Friend Features
- [x] Direct friend challenge (1h expiry)
- [x] Challenge link generation (24h expiry)
- [x] Challenge AcceptWaiting flow (invitee accepts → inviter starts)
- [x] Friends online status (Redis OnlineTracker)
- [x] Rematch flow (RequestRematchUseCase)
- [ ] Push notifications (Telegram Bot API)
- [ ] Revenge notifications

### Phase 3: Ranking & Seasons
- [x] League system (6 leagues × 4 divisions)
- [x] Promotion/demotion logic (3-game protection)
- [x] Season reset (soft reset formula: `1000 + (mmr-1000)*0.5`, min 500)
- [x] Peak MMR tracking
- [x] Seasonal reward distribution (DistributeSeasonalRewardsUseCase, credits by peak league)
- [ ] Seasonal leaderboard with percentile

### Phase 4: Virality & Rewards
- [x] Referral tracking (5 milestones, referral.go)
- [x] Referral reward claiming (GET /referrals + POST /referrals/:friendId/claim, 5-level milestone rewards)
- [ ] Victory card image generation
- [ ] Share to Telegram/Stories
- [ ] Daily/weekly missions
- [ ] Cosmetics (frames, titles, emotes)
- [ ] Ticket shop (coin/real money packs)

### Phase 5: Missing Features
- [x] Ticket consumption/refund system (real pvp_tickets via InventoryService)
- [x] Time-based tiebreaker (by totalTimeMs, then playerID — deterministic, no more nil/draw)
- [x] Server-side time validation (reject negative, clamp >10500ms to 10000ms)
- [x] ELO draw fix (symmetric CalculateDrawRating for equal scores)
- [x] Swagger: missing fields added (ExpiredChallenges, RankChange, Questions, RematchExpiresIn)
- [x] Bot games fallback (60s queue timeout)
- [x] Same-opponent prevention in matchmaking (Redis duel:recent:{p}:{o} EX 300, bypass after 30s)
- [x] Surrender endpoint (POST /duel/game/:gameId/surrender, requires 3+ answers)
- [x] Structured error codes (AppError with errorCode field, 30+ codes)
- [ ] Anti-cheat: pattern detection, penalties

### Bugs / Fixes Needed
- [x] Fix: Scoring doc updated — 100 base + speed bonus (50/25/10/0), not correct count
- [x] Fix: Matchmaking ranges — updated to 5-tier spec (10/20/30/45/45+ seconds)
- [x] Fix: K-factor doc updated — K=32 for <30 games, K=16 for 30+ games
- [x] Fix: Package name — docs updated to `quick_duel` (note added to 06_domain.md)
