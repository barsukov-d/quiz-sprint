# PvP Duel Documentation

> **Аудит реализации: 2026-03-15**

## Статус документации vs кода

| Файл | ✅ | ⚠️ | ❌ | Главные расхождения |
|------|----|----|----|--------------------|
| 01_concept.md | 8 | 5 | 5 | Scoring по points а не count, matchmaking тиры другие |
| 02_gameplay.md | 13 | 6 | 7 | Нет emotes, surrender, bot games, victory cards |
| 03_rules.md | 8 | 6 | 9 | K-factor двойной, нет тикетов, нет time tiebreaker |
| 04_rewards.md | 3 | 2 | 16 | Система наград почти полностью отсутствует |
| 05_api.md | 24 | 2 | 3 | WebSocket полностью реализован, нет surrender/referrals |
| 06_domain.md | 18 | 15 | 2 | Пакет quick_duel (не pvp_duel), сервисы в use cases |
| 07_edge_cases.md | 8 | 14 | 5 | Draw вместо time tiebreaker, нет bot/surrender |

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
- [x] Season reset (soft reset formula)
- [x] Peak MMR tracking
- [ ] Seasonal reward distribution
- [ ] Seasonal leaderboard with percentile

### Phase 4: Virality & Rewards
- [x] Referral tracking (5 milestones, referral.go)
- [ ] Referral reward claiming
- [ ] Victory card image generation
- [ ] Share to Telegram/Stories
- [ ] Daily/weekly missions
- [ ] Cosmetics (frames, titles, emotes)
- [ ] Ticket shop (coin/real money packs)

### Phase 5: Missing Features
- [ ] Ticket consumption/refund system
- [ ] Time-based tiebreaker (currently draw on equal points)
- [ ] Bot games fallback (60s queue timeout)
- [ ] Surrender endpoint (after Q3)
- [ ] Structured error codes (JSON format)
- [ ] Server-side time validation (500ms tolerance)
- [ ] Anti-cheat: pattern detection, penalties
- [ ] Same-opponent prevention in matchmaking

### Bugs / Fixes Needed
- [ ] Fix: Scoring doc says "correct count", code uses points+speed bonus — align doc or code
- [ ] Fix: Matchmaking ranges don't match doc (code: 5/10/15s, doc: 10/20/30/45/60s)
- [ ] Fix: K-factor doc says K=32 always, code has K=32/<30 games, K=16 after
- [ ] Fix: Package name `quick_duel` vs doc `pvp_duel` — choose one
