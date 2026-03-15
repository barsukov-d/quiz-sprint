# PvP Duel - Concept

> **Статус реализации (аудит 2026-03-15)**
> ✅ Реализовано: 8 | ⚠️ Расходится: 5 | ❌ Не реализовано: 5

## What?
Competitive 1v1 ranked mode where players battle in real-time quiz duels using identical questions. **Designed for maximum virality and friend engagement.**

## Why?
**Primary goal:** Pure skill competition — no bonuses, no luck, only knowledge and speed.
**Secondary:** Invite friends, climb together, share victories, build community.

## For Whom?
- Competitive players: climb ranks, prove skill
- Social players: challenge friends, compare ranks, share victories
- Influencers: build audience through shareable moments

## Key Mechanics

| Parameter | Value | Status |
|-----------|-------|--------|
| Players | 2 (1v1) | ✅ |
| Questions | 7 (identical for both) | ✅ |
| Time per question | 10 seconds | ✅ |
| Entry cost | 1 PvP ticket | ❌ Ticket system not implemented — no consumption or refund logic |
| Bonuses/hints | **FORBIDDEN** | ✅ |
| Win condition | Most correct answers | ⚠️ Code uses POINTS (100 + SpeedBonus up to 50), not simple correct count |
| Tiebreaker | Total time spent | ⚠️ Time is baked into speed bonus points; if points tied, code returns nil (draw) — no explicit time tiebreaker |

## Core Loop
```
Daily Challenge → Earn PvP Tickets → Enter Duel → Win/Lose → MMR Change → Climb Leagues
```

## Unique Features

### 1. Pure Skill Mode
- **NO bonuses** (Shield, 50/50, Skip, Freeze — all disabled)
- **NO luck factor** (same questions, same order)
- **Identical conditions** for both players

### 2. ELO/MMR Rating System <!-- ⚠️ K factor differs: code uses K=32 for new players (<30 games), K=16 for regular; doc says K=32 always. Min ±10 implemented. -->
Skill-based matchmaking:
- Win against stronger opponent → more MMR (up to ~+30)
- Win against weaker opponent → less MMR (minimum +10)
- Lose to stronger opponent → less MMR lost (minimum -10)
- Lose to weaker opponent → more MMR lost (up to ~-30)
- Exact values: ELO formula, K=32, min ±10. See `03_rules.md`.

> ⚠️ **Расхождение:** Код использует K=32 для игроков с <30 играми и K=16 для остальных. Документ специфицирует K=32 для всех. Min ±10 реализован корректно.

### 3. League System <!-- ✅ Fully implemented with correct MMR ranges and 6 leagues × 4 divisions -->

| League | Icon | MMR Range | Division Count |
|--------|------|-----------|----------------|
| Bronze | 🥉 | 0-999 | 4 (IV → I) |
| Silver | 🥈 | 1000-1499 | 4 |
| Gold | 🥇 | 1500-1999 | 4 |
| Platinum | 💍 | 2000-2499 | 4 |
| Diamond | 💎 | 2500-2999 | 4 |
| Legend | 👑 | 3000+ | 1 |

### 4. Seasonal Structure <!-- ✅ SeasonReset method with correct soft-reset formula -->
- **Season duration:** 1 month
- **Soft reset:** MMR compressed toward 1000 at season start
- **Seasonal rewards:** Exclusive cosmetics based on peak rank <!-- ❌ No reward distribution implemented -->

> ❌ **Не реализовано:** Seasonal rewards — распределение наград не реализовано.

## Entry Requirement

**PvP Ticket** (earned from Daily Challenge): <!-- ❌ Not implemented — tickets are not consumed on duel entry or refunded on cancel -->
- Free-to-play players: Limited tickets per day
- No tickets → Cannot enter duel
- Tickets purchasable in shop <!-- ❌ No shop system -->

> ❌ **Не реализовано:** Тикет-система — тикеты не списываются при входе и не возвращаются при отмене.

## Matchmaking

### Queue Time Expectations <!-- ⚠️ Code has 4 tiers (5s/10s/15s/15s+), not 5 tiers. After 15s matches anyone regardless of MMR. -->

| MMR Difference | Max Wait | Статус |
|----------------|----------|--------|
| ±50 | 10s | ⚠️ |
| ±100 | 20s | ⚠️ |
| ±200 | 30s | ⚠️ |
| ±300 | 45s | ⚠️ |
| ±500 | 60s | ⚠️ |

> ⚠️ **Расхождение:** Код реализует 4 тира (5с/10с/15с/15с+), после 15с матчит любого. Документ специфицирует 5 тиров с расширением до 60с.

### Fallback
If no opponent found in 60s → Offer bot game (clearly labeled as 🤖 Bot). <!-- ❌ Not implemented -->

> ❌ **Не реализовано:** Bot game fallback после 60с.

## Victory Conditions

### Primary: Correct Answers <!-- ⚠️ Code uses point-based system (100 + speed bonus), not raw correct count -->
```
Player A: 5/7 correct
Player B: 4/7 correct
Winner: Player A
```

### Tiebreaker: Total Time <!-- ⚠️ If points are equal, code returns nil (draw); no explicit time tiebreaker -->
```
Player A: 5/7 correct, 42.5s total
Player B: 5/7 correct, 38.2s total
Winner: Player B (faster)
```

### Second Tiebreaker: First Correct
If total time equal (extremely rare):
```
Winner: First player to answer correctly on any question
```

## Monetization

| Feature | Cost | Effect | Статус |
|---------|------|--------|--------|
| PvP Ticket | Daily Challenge reward | 1 duel entry | ❌ Не реализовано |
| Ticket Pack (5) | 300 coins | 5 duel entries | ❌ Нет магазина |
| Ticket Pack (15) | 750 coins | 15 entries (17% discount) | ❌ Нет магазина |
| Premium cosmetics | Seasonal rewards / Shop | Visual only, no gameplay effect | ❌ Нет магазина |

## Virality & Social Features

### 1. Friend Challenge (Direct Invite) <!-- ✅ Both direct and link challenges implemented -->
```
Challenge flow:
Player A → "Вызвать друга" → Share link → Friend opens → Instant duel
```

**No queue waiting** for friend games — instant start.

### 2. Referral System <!-- ✅ Domain model exists (referral.go), repository exists -->
| Milestone | Reward for Inviter | Reward for Invitee |
|-----------|-------------------|-------------------|
| Friend registers | 3 🎟️ + 100 coins | 3 🎟️ + 100 coins |
| Friend plays 5 duels | 5 🎟️ + 300 coins | 200 coins |
| Friend reaches Silver | 10 🎟️ + 500 coins + 🏷️ "Гуру" badge | 300 coins |
| Friend reaches Gold | 20 🎟️ + 1,000 coins + Exclusive avatar | 500 coins |
| Friend reaches Platinum | 50 🎟️ + 3,000 coins + Legendary title | 1,000 coins |

**Referral link:** `https://t.me/quiz_sprint_dev_bot?startapp=ref_USER123`

### 3. Friends Leaderboard <!-- ⚠️ Leaderboard exists but no separate friends tab -->
Separate tab showing only friends:
```
┌─────────────────────────────────┐
│  👥 ДРУЗЬЯ (8 игроков)          │
│                                 │
│  #1 🥇 @ProGamer     Gold I     │
│  #2 🥈 @BestQuizzer  Gold III   │
│  #3 🥉 Ты            Gold III   │  ← You
│  #4    @NewbieFriend Silver II  │
│                                 │
│  [ Пригласить ещё друзей ]      │
└─────────────────────────────────┘
```

> ⚠️ **Расхождение:** Лидерборд реализован, но отдельной вкладки "Друзья" нет.

### 4. Shareable Victory Cards <!-- ❌ Not implemented — no image generation -->
After each win, generate shareable image:
```
┌─────────────────────────────────┐
│  ⚔️ ПОБЕДА В ДУЭЛИ!             │
│                                 │
│  PlayerName 🏆                  │
│  5 : 3                          │
│  🥇 Gold III                    │
│                                 │
│  "Попробуй победить меня!"      │
│  [QR код / ссылка]              │
└─────────────────────────────────┘
```

**Share targets:** Telegram, Instagram Stories, Twitter.

> ❌ **Не реализовано:** Генерация Victory Card — нет image generation.

### 5. Revenge System <!-- ❌ Not implemented — no revenge notifications -->
After losing to a friend:
```
💔 @Friend победил тебя!

Хочешь отомстить?
[ РЕВАНШ ] → Sends challenge notification
```

Creates engagement loop: lose → revenge → rematch → repeat.

> ❌ **Не реализовано:** Revenge System — уведомления реванша не реализованы.

### 6. Spectator Mode (Future)
Friends can watch live duels. Deferred to Phase 4+.

---

## Viral Loop

```
                    ┌──────────────────┐
                    │   Player wins    │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Share victory   │
                    │  card to Telegram│
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Friend sees,    │
                    │  clicks link     │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Friend installs,│
                    │  challenges back │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Duel happens    │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Both share      │◄────┐
                    │  results         │     │
                    └────────┬─────────┘     │
                             │               │
                             └───────────────┘
```

---

## Success Metrics

- Avg duels per player/day: 3-5
- **Friend duel ratio: >30%** (vs random)
- **Referral conversion: >15%**
- **Share rate: >25%** of wins shared
- Queue time <30s: >85%
- Rematch rate: >40%
- Seasonal climb engagement: >60% players ranked
- Ticket purchase conversion: >10%
