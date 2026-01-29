# Solo Marathon - Rewards System

## Weekly Leaderboard Rewards

### Week Schedule
- **Start:** Monday 00:00 UTC
- **End:** Sunday 23:59 UTC
- **Distribution:** Monday 00:05 UTC (after reset)

### Reward Tiers

| Rank | Coins | PvP Tickets | Bonus Pack | Badge |
|------|-------|-------------|------------|-------|
| **1-3** | 5,000 | 50 | Premium Pack | 🥇 Legend |
| **4-10** | 3,000 | 30 | Premium Pack | 🥈 Champion |
| **11-25** | 2,000 | 20 | Standard Pack | 🥉 Master |
| **26-50** | 1,000 | 10 | Standard Pack | 🏅 Expert |
| **51-100** | 500 | 5 | Basic Pack | 🎖️ Challenger |
| **101+** | 0 | 0 | - | - |

### Bonus Packs

**Premium Pack:**
- 🛡️ Shield: 5
- 🔀 50/50: 5
- ⏭️ Skip: 5
- ❄️ Freeze: 10

**Standard Pack:**
- 🛡️ Shield: 3
- 🔀 50/50: 3
- ⏭️ Skip: 3
- ❄️ Freeze: 6

**Basic Pack:**
- 🛡️ Shield: 2
- 🔀 50/50: 2
- ⏭️ Skip: 2
- ❄️ Freeze: 4

### Badge Display
Badges shown:
- In leaderboard next to name
- On profile
- In game results screen

**Format:** "🥇 Legend (Week 42, 2026)"

---

## All-Time Rewards

**NO material rewards** (Hall of Fame is prestige only).

**Benefits:**
- Permanent "Legend" title
- Special profile frame
- Featured in main menu ("Current Champion: @username")

---

## Personal Best Bonus

First time reaching new personal record:
```
New record → +500 coins (one-time)
```

**Milestones:**
| Score Reached | Bonus | Badge |
|---------------|-------|-------|
| 25 | 100 coins | 🌟 Beginner |
| 50 | 250 coins | ⭐ Intermediate |
| 100 | 500 coins | ✨ Advanced |
| 200 | 1,000 coins | 💫 Expert |
| 500 | 5,000 coins | 🌠 Master |

---

## Continue Economics

### Cost Progression
```
Continue #1: 200 coins OR Rewarded Ad
Continue #2: 400 coins OR Rewarded Ad
Continue #3: 600 coins OR Rewarded Ad
Continue #4+: 800 coins (no ad option)
```

### Expected Value Analysis

**Average player (40 questions/run):**
- Without continue: 40 score, ~0 rewards
- With 1 continue (+10 questions): 50 score, top 100 chance

**ROI for competitive player:**
```
Cost: 200 coins
Potential: Top 100 → 500+ coins reward
Break-even: Need to reach top 100
```

---

## Monetization: Bonus Packs

### Shop Offers

**Marathon Starter Pack:**
- 🛡️ Shield: 3
- ❄️ Freeze: 5
- Cost: **300 coins**

**Marathon Pro Pack:**
- 🛡️ Shield: 5
- 🔀 50/50: 5
- ⏭️ Skip: 5
- ❄️ Freeze: 10
- Cost: **800 coins** (vs 1,000 if bought separately)

**Emergency Pack (in-game offer):**
Shown when all bonuses depleted during run:
```
┌─────────────────────────────────┐
│  У тебя закончились бонусы!     │
│  Получи экстренный набор:       │
│                                 │
│  🛡️×1  ❄️×3                    │
│                                 │
│  [ 150 💰 ]  или  [ 📺 ]        │
└─────────────────────────────────┘
```

---

## Reward Distribution Flow

### Weekly Rewards

**Monday 00:05 UTC:**
1. Backend job queries Redis:
```
ZREVRANGE marathon:leaderboard:weekly:{prev_week} 0 99 WITHSCORES
```

2. For each player in top 100:
```sql
INSERT INTO reward_claims (
    player_id,
    reward_type,
    coins,
    pvp_tickets,
    bonuses,
    week_id
)
```

3. Update user inventory:
```sql
UPDATE user_inventory SET
    coins = coins + reward.coins,
    pvp_tickets = pvp_tickets + reward.tickets,
    bonus_shield = bonus_shield + reward.bonuses.shield,
    ...
```

4. Send notification:
```
🏆 Награда за неделю #42!
Твоё место: #15
Получено: 2,000💰, 20🎟️, Standard Pack
```

---

## Claiming Rewards (Frontend)

### Notification Banner
```
┌─────────────────────────────────┐
│  🎁 У тебя новая награда!       │
│  [ Открыть ]                    │
└─────────────────────────────────┘
```

### Reward Screen
```
┌─────────────────────────────────┐
│  🏆 НАГРАДА ЗА НЕДЕЛЮ #42       │
│                                 │
│  Твоё место: #15 из 5,847       │
│                                 │
│  Получено:                      │
│  💰 +2,000 монет                │
│  🎟️ +20 PvP билетов             │
│                                 │
│  Бонусы:                        │
│  🛡️ +3  🔀 +3  ⏭️ +3  ❄️ +6    │
│                                 │
│  [ Забрать ]                    │
└─────────────────────────────────┘
```

Animation: Coins/tickets fly to inventory counters.

---

## Database Schema

### Weekly Leaderboard Archive
```sql
CREATE TABLE marathon_weekly_leaderboard (
    week_id VARCHAR(20),        -- "2026-W42"
    player_id VARCHAR(36),
    score INT,
    total_questions INT,
    continue_count INT,
    rank INT,
    completed_at TIMESTAMP,

    PRIMARY KEY (week_id, player_id),
    INDEX idx_week_rank (week_id, rank)
);
```

### Reward Claims
```sql
CREATE TABLE reward_claims (
    id VARCHAR(36) PRIMARY KEY,
    player_id VARCHAR(36),
    reward_type VARCHAR(50),    -- 'weekly_marathon', 'personal_best'
    coins INT,
    pvp_tickets INT,
    bonuses JSONB,
    week_id VARCHAR(20),
    claimed_at TIMESTAMP,

    INDEX idx_player_unclaimed (player_id, claimed_at)
);
```

---

## Edge Cases

### Player deleted account
- Rewards NOT distributed
- Rank skipped (e.g., if #3 deleted, #4 becomes #3)

### Tied scores
Tiebreaker: total_questions ASC → completedAt ASC
Both get SAME rank rewards.

**Example:**
```
Rank 10: Player A (87/87)
Rank 10: Player B (87/87, completed earlier)
Rank 12: Player C (86/86)

Both A and B get rank 10 rewards.
```

### Multiple games in same week
Only **best score** counts for leaderboard.

### Playing after Sunday 23:59 UTC
Score goes to **next week** (based on completedAt).

---

## Fraud Prevention

### Bot Detection
- Suspiciously high scores (>500) → Manual review
- Identical question timings → Flag account
- Excessive continues (>20/game) → Flag

### Reward Validation
```go
func ValidateRewardClaim(playerID, weekID) error {
    // Check not already claimed
    if isAlreadyClaimed(playerID, weekID) {
        return ErrRewardAlreadyClaimed
    }

    // Verify rank in leaderboard
    actualRank := getWeeklyRank(playerID, weekID)
    if actualRank > 100 || actualRank == 0 {
        return ErrNotEligible
    }

    return nil
}
```

---

## Metrics to Track

- **Weekly participation rate:** Active players who played Marathon this week
- **Top 100 competition:** Avg score of #100 player (difficulty indicator)
- **Continue conversion:** % of players who continue at game over
- **Bonus usage:** Avg bonuses used per game
- **Reward claim rate:** % of eligible players who claimed rewards
