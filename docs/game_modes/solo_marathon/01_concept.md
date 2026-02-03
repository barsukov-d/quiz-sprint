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
| Questions | Endless (until game over) |
| Starting lives | 3 |
| Time per question | 15s → 8s (adaptive) |
| Wrong answer penalty | -1 life |
| Game over | 0 lives |
| Score | Correct answers count |

## Core Loop
```
Daily Challenge → Earn Bonuses → Use in Marathon → Set Record → Compete Weekly
```

## Unique Features

### 1. Lives System
- Start with 3 ❤️❤️❤️
- Wrong answer = -1 life
- 0 lives = game over
- **NO life regeneration** (except continue)

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

### 4. Continue Mechanic (Monetization)
At game over:
- **1 continue:** 200 coins OR Rewarded Ad → lives reset to 1
- Multiple continues possible (escalating cost)

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
