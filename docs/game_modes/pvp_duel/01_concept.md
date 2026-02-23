# PvP Duel - Concept

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

| Parameter | Value |
|-----------|-------|
| Players | 2 (1v1) |
| Questions | 7 (identical for both) |
| Time per question | 10 seconds |
| Entry cost | 1 PvP ticket |
| Bonuses/hints | **FORBIDDEN** |
| Win condition | Most correct answers |
| Tiebreaker | Total time spent |

## Core Loop
```
Daily Challenge → Earn PvP Tickets → Enter Duel → Win/Lose → MMR Change → Climb Leagues
```

## Unique Features

### 1. Pure Skill Mode
- **NO bonuses** (Shield, 50/50, Skip, Freeze — all disabled)
- **NO luck factor** (same questions, same order)
- **Identical conditions** for both players

### 2. ELO/MMR Rating System
Skill-based matchmaking:
- Win against stronger opponent → more MMR (up to ~+30)
- Win against weaker opponent → less MMR (minimum +10)
- Lose to stronger opponent → less MMR lost (minimum -10)
- Lose to weaker opponent → more MMR lost (up to ~-30)
- Exact values: ELO formula, K=32, min ±10. See `03_rules.md`.

### 3. League System

| League | Icon | MMR Range | Division Count |
|--------|------|-----------|----------------|
| Bronze | 🥉 | 0-999 | 4 (IV → I) |
| Silver | 🥈 | 1000-1499 | 4 |
| Gold | 🥇 | 1500-1999 | 4 |
| Platinum | 💍 | 2000-2499 | 4 |
| Diamond | 💎 | 2500-2999 | 4 |
| Legend | 👑 | 3000+ | 1 |

### 4. Seasonal Structure
- **Season duration:** 1 month
- **Soft reset:** MMR compressed toward 1000 at season start
- **Seasonal rewards:** Exclusive cosmetics based on peak rank

## Entry Requirement

**PvP Ticket** (earned from Daily Challenge):
- Free-to-play players: Limited tickets per day
- No tickets → Cannot enter duel
- Tickets purchasable in shop

## Matchmaking

### Queue Time Expectations
| MMR Difference | Max Wait |
|----------------|----------|
| ±50 | 10s |
| ±100 | 20s |
| ±200 | 30s |
| ±300 | 45s |
| ±500 | 60s |

### Fallback
If no opponent found in 60s → Offer bot game (clearly labeled as 🤖 Bot).

## Victory Conditions

### Primary: Correct Answers
```
Player A: 5/7 correct
Player B: 4/7 correct
Winner: Player A
```

### Tiebreaker: Total Time
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

| Feature | Cost | Effect |
|---------|------|--------|
| PvP Ticket | Daily Challenge reward | 1 duel entry |
| Ticket Pack (5) | 300 coins | 5 duel entries |
| Ticket Pack (15) | 750 coins | 15 entries (17% discount) |
| Premium cosmetics | Seasonal rewards / Shop | Visual only, no gameplay effect |

## Virality & Social Features

### 1. Friend Challenge (Direct Invite)
```
Challenge flow:
Player A → "Вызвать друга" → Share link → Friend opens → Instant duel
```

**No queue waiting** for friend games — instant start.

### 2. Referral System
| Milestone | Reward for Inviter | Reward for Invitee |
|-----------|-------------------|-------------------|
| Friend registers | 3 🎟️ + 100 coins | 3 🎟️ + 100 coins |
| Friend plays 5 duels | 5 🎟️ + 300 coins | 200 coins |
| Friend reaches Silver | 10 🎟️ + 500 coins + 🏷️ "Гуру" badge | 300 coins |
| Friend reaches Gold | 20 🎟️ + 1,000 coins + Exclusive avatar | 500 coins |
| Friend reaches Platinum | 50 🎟️ + 3,000 coins + Legendary title | 1,000 coins |

**Referral link:** `https://t.me/quiz_sprint_dev_bot?startapp=ref_USER123`

### 3. Friends Leaderboard
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

### 4. Shareable Victory Cards
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

### 5. Revenge System
After losing to a friend:
```
💔 @Friend победил тебя!

Хочешь отомстить?
[ РЕВАНШ ] → Sends challenge notification
```

Creates engagement loop: lose → revenge → rematch → repeat.

### 6. Weekly Friend Tournaments
**Every Sunday:** Mini-tournament among friends who played each other.
- Auto-generated bracket from week's duels
- Winner gets "Champion of Friends" badge
- Shareable results

### 7. "Bring a Friend" Events
Monthly events:
```
🎉 НЕДЕЛЯ ДРУЖБЫ

Сыграй 10 дуэлей с друзьями →
Получи эксклюзивную рамку аватара!

Прогресс: 4/10
[ Пригласить друга ]
```

### 8. Spectator Mode (Future)
Friends can watch live duels:
- Real-time question + answers
- React with emojis
- Share ongoing game

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
