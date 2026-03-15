# PvP Duel Documentation

> **Аудит реализации: 2026-03-15 | Обновлено: 2026-03-15 (Phase 8: bot fallback + same-opponent prevention)**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 12 | 3 | 3 | Нет bot games, emotes, cosmetics |
| 02_gameplay.md | 16 | 4 | 6 | Нет emotes, surrender, bot games, victory cards |
| 03_rules.md | 14 | 4 | 5 | Matchmaking исправлен, tiebreaker добавлен |
| 04_rewards.md | 9 | 2 | 10 | Referral rewards добавлены, seasonal добавлен |
| 05_api.md | 26 | 2 | 1 | Referral endpoints добавлены, нет surrender |
| 06_domain.md | 22 | 11 | 2 | Пакет quick_duel (не pvp_duel), сервисы в use cases |
| 07_edge_cases.md | 12 | 10 | 5 | Tiebreaker по времени, нет bot/surrender |

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
- [ ] Surrender endpoint (after Q3)
- [ ] Structured error codes (JSON format)
- [ ] Anti-cheat: pattern detection, penalties

### Bugs / Fixes Needed
- [ ] Fix: Scoring doc says "correct count", code uses points+speed bonus — align doc or code
- [x] Fix: Matchmaking ranges — updated to 5-tier spec (10/20/30/45/45+ seconds)
- [ ] Fix: K-factor doc says K=32 always, code has K=32/<30 games, K=16 after
- [ ] Fix: Package name `quick_duel` vs doc `pvp_duel` — choose one
