# Solo Marathon - Concept

## What?
Endless PvE mode where player answers questions until losing 3 lives.

## Why?
**Primary goal:** Use bonuses earned from Daily Challenge strategically to set records.
**Secondary:** Long-term progression, skill demonstration, weekly competition.

## For Whom?
- Hardcore players: push for records
- Strategic players: optimize bonus usage
- Casual players: relaxed endless mode (no stress of PvP)

## Key Mechanics

| Parameter | Value |
|-----------|-------|
| Questions | Endless (until energy runs out) |
| Starting energy | 5 ⚡ |
| Time per question | 15s → 8s (adaptive) |
| Wrong answer penalty | −1 ⚡ |
| Energy regen | +1 ⚡ every 5 correct in a row (Marathon Momentum) |
| Run over | 0 ⚡ → Continue (coins/ad) OR instant new run |
| Score | Correct answers count (best run per week) |

## Core Loop
```
Daily Challenge → Earn Bonuses → Use in Marathon → Set Record → Compete Weekly
```

## Unique Features

### 1. Energy System
- Start with 5 ⚡⚡⚡⚡⚡
- Wrong answer = −1 ⚡
- 5 correct in a row = +1 ⚡ (Marathon Momentum)
- 0 ⚡ = run over → instant free restart OR pay to continue
- **NO waiting** between runs

### 2. Strategic Bonuses (from Daily Challenge)

4 types, limited quantity:

| Bonus | Icon | Effect | Use Case |
|-------|------|--------|----------|
| Shield | 🛡️ | 1 free mistake (no life loss) | Uncertain answer |
| 50/50 | 🔀 | Remove 2 wrong answers | 50/50 guess |
| Skip | ⏭️ | Skip question (no penalty) | Unknown topic |
| Freeze | ❄️ | +10 seconds to timer | Late-game time pressure |

### 3. Adaptive Difficulty
Questions get harder over time:
- Timer decreases (15s → 8s)
- Topics become narrower
- Questions become more complex

### 4. Continue Mechanic (optional monetization)
At 0 energy:
- **Continue:** 200 coins OR Rewarded Ad → energy resets to 1 ⚡ (resume same run)
- **New run:** Free → 5 ⚡ fresh start (best run per week goes to leaderboard)

## Leaderboards

| Type | Period | Top Rewards |
|------|--------|-------------|
| Weekly | Monday-Sunday UTC | Top 100: coins, bonuses |
| All-Time | Forever | Hall of Fame only |
| Friends | Current week | Social comparison |

Weekly resets → Fresh competition every week.

## Monetization

| Feature | Cost | Effect |
|---------|------|--------|
| Continue (1st) | 200 coins / Ad | Lives reset to 1 |
| Continue (2nd) | 400 coins / Ad | Lives reset to 1 |
| Continue (3rd) | 600 coins / Ad | Lives reset to 1 |
| Continue (4th+) | 800+ coins (no Ad) | Lives reset to 1 |
| Bonus pack | 500 coins | 3 Shields, 5 Freezes |

## Success Metrics

- Avg questions per run: 30-50
- Continue conversion: >25%
- Weekly top 100 participation: >5% active users
- Bonus usage rate: >80% (not hoarding)
