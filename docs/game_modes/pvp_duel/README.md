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

- **Backend**: `backend/internal/domain/pvp_duel/` (TBD)
- **Frontend**: `tma/src/views/DuelView.vue` (TBD)
- **API**: `/api/v1/duel/*`
- **WebSocket**: `/ws/duel`

---

## Key Features

### Core Gameplay
- [x] 1v1 real-time duels
- [x] 10 identical questions
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
- [ ] Domain model (DuelMatch, PlayerRating)
- [ ] Matchmaking queue (Redis)
- [ ] WebSocket server for real-time sync
- [ ] Question distribution (same for both)
- [ ] Answer submission + scoring
- [ ] MMR calculation
- [ ] Basic result screen

### Phase 2: Friend Features
- [ ] Direct friend challenge
- [ ] Challenge link generation
- [ ] Push notifications (Telegram)
- [ ] Friends online status
- [ ] Rematch flow

### Phase 3: Ranking & Seasons
- [ ] League system (6 leagues × 4 divisions)
- [ ] Promotion/demotion logic
- [ ] Seasonal leaderboard
- [ ] Season reset job
- [ ] Seasonal reward distribution

### Phase 4: Virality
- [ ] Referral tracking
- [ ] Referral milestone rewards
- [ ] Victory card generation (image)
- [ ] Share to Telegram/Stories
- [ ] Friends leaderboard
- [ ] Revenge notifications

### Phase 5: Frontend
- [ ] Duel lobby view
- [ ] Matchmaking screen
- [ ] Friend challenge UI
- [ ] Live duel screen (WebSocket)
- [ ] Result screen with sharing
- [ ] Leaderboard views

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
