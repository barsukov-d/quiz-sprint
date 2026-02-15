# PvP Duel Documentation

## Completeness Status

| File | Status | Notes |
|------|--------|-------|
| 01_concept.md | ✅ 90% | Core concept + virality focus |
| 02_gameplay.md | ✅ 85% | Full flow with friend challenges |
| 03_rules.md | ✅ 90% | ELO/MMR, leagues, matchmaking |
| 04_rewards.md | ✅ 90% | Seasons, referrals, friend rewards |
| 05_api.md | ✅ 85% | REST + WebSocket spec |
| 06_domain.md | ✅ 85% | Aggregates, services, DB schema |
| 07_edge_cases.md | ✅ 85% | Disconnect, ties, fraud |

**Overall: 87% complete**

---

## Quick Navigation

- **Domain**: `backend/internal/domain/quick_duel/`
- **Application**: `backend/internal/application/quick_duel/`
- **Handlers**: `backend/internal/infrastructure/http/handlers/duel_handlers.go`
- **Repositories**: `backend/internal/infrastructure/persistence/postgres/*_repository.go`
- **Redis Queue**: `backend/internal/infrastructure/persistence/redis/matchmaking_queue.go`
- **Migration**: `backend/migrations/013_create_duel_tables.sql`
- **Frontend**: `tma/src/views/Duel/` (DuelLobbyView, DuelPlayView, DuelResultsView)
- **Components**: `tma/src/components/Duel/DuelCard.vue`
- **Composables**: `tma/src/composables/usePvPDuel.ts`, `tma/src/composables/useDuelWebSocket.ts`
- **API**: `/api/v1/duel/*`
- **WebSocket**: `/ws/duel` (TBD)

---

## Key Features

### Core Gameplay
- [x] 1v1 real-time duels
- [x] 7 identical questions
- [x] 10 seconds per question
- [x] No bonuses (pure skill)
- [x] Score + time tiebreaker

### Ranking System
- [x] ELO/MMR calculation
- [x] 6 leagues (Bronze → Legend)
- [x] 4 divisions per league
- [x] Seasonal resets (monthly)
- [x] Demotion protection (3 games)

### Social & Virality
- [x] Friend challenge (direct)
- [x] Challenge links (shareable)
- [x] Friends leaderboard
- [x] Referral rewards (5 milestones)
- [x] Victory card sharing
- [x] Revenge system
- [x] Rematch flow

### Rewards
- [x] Seasonal rewards by peak rank
- [x] Referral milestone rewards
- [x] Friend duel weekly bonuses
- [x] Daily/weekly missions
- [x] Cosmetics (frames, titles, emotes)

---

## Implementation Checklist

### Phase 1: Core Duel (MVP)
- [x] Domain model (DuelGame, PlayerRating, EloRating, League, Division)
- [x] Matchmaking queue (Redis sorted sets)
- [x] WebSocket server for real-time sync (DuelWebSocketHub)
- [x] Question distribution (same for both via StartMatchUseCase)
- [x] Answer submission + scoring (SubmitDuelAnswerUseCase)
- [x] MMR/ELO calculation (K-factor 32, min ±10)
- [x] Basic result screen (frontend) - DuelResultsView.vue

### Phase 2: Friend Features
- [x] Direct friend challenge (60s expiry)
- [x] Challenge link generation (24h expiry)
- [ ] Push notifications (Telegram)
- [x] Friends online status (Redis OnlineTracker)
- [x] Rematch flow (RequestRematchUseCase)

### Phase 3: Ranking & Seasons
- [x] League system (6 leagues × 4 divisions)
- [x] Promotion/demotion logic (with 3-game protection)
- [x] Seasonal leaderboard
- [x] Season reset job (soft reset formula)
- [ ] Seasonal reward distribution

### Phase 4: Virality
- [x] Referral tracking (5 milestones)
- [x] Referral milestone rewards
- [ ] Victory card generation (image)
- [ ] Share to Telegram/Stories
- [x] Friends leaderboard
- [ ] Revenge notifications

### Phase 5: Frontend
- [x] Duel lobby view (DuelLobbyView.vue)
- [x] Matchmaking screen (in lobby)
- [x] Friend challenge UI (in lobby)
- [x] Live duel screen (DuelPlayView.vue + WebSocket)
- [x] Result screen with sharing (DuelResultsView.vue)
- [x] Leaderboard views (in lobby tabs)

---

## Dependencies

### From Other Contexts
- `user` - Player accounts, inventory
- `quiz` - Questions pool (medium difficulty)
- `daily_challenge` - Source of PvP tickets
- `notifications` - Push notifications

### External Services
- Redis - Matchmaking queue, online status
- WebSocket - Real-time duel sync
- Telegram Bot API - Push notifications

---

## Metrics to Track

### Engagement
- Duels per player per day
- Queue time distribution
- Match completion rate
- Rematch rate

### Social
- Friend duel ratio (vs random)
- Referral conversion rate
- Share rate (% of wins shared)
- Challenge acceptance rate

### Ranking
- MMR distribution by league
- Promotion/demotion rates
- Season participation rate

### Monetization
- Ticket purchase conversion
- Avg tickets used per day
- Revenue per user (tickets)

---

## Open Questions

1. **Bot intelligence level?** - Should bots be beatable or challenging?
2. **Spectator mode priority?** - Feature for watching friend duels
3. **Tournament mode?** - 8/16 player brackets
4. **Voice chat?** - Real-time voice during duel with friends
5. **Ranked decay?** - Lose MMR for inactivity?
