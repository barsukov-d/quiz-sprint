# Daily Challenge - Rewards System

## Chest Types & Contents

### 🪵 Wooden Chest (0-4 correct)

| Resource | Amount | Notes |
|----------|--------|-------|
| Coins | 50-100 | Random in range |
| PvP Tickets | 1 | Fixed |
| Marathon Bonuses | 50% chance 1 bonus | Random type |

**Bonus Selection:**
- 50% chance: 1 random bonus (avg 0.5 per chest)
- 50% chance: No bonus
- Equal probability per type (25% each)

**With Streak Bonus:**
```
3d streak → 55-110 coins
7d streak → 62-125 coins
30d streak → 75-150 coins
```

### 🥈 Silver Chest (5-7 correct)

| Resource | Amount | Notes |
|----------|--------|-------|
| Coins | 150-250 | Random |
| PvP Tickets | 2-3 | Random |
| Marathon Bonuses | 1-2 bonuses | 1 guaranteed |

**Bonus Selection:**
- 100%: 1 random bonus (guaranteed)
- 30% chance: +1 extra bonus (2 total, different types)
- Avg 1.3 bonuses per chest

**With Streak:**
```
3d → 165-275 coins
7d → 187-312 coins
30d → 225-375 coins
```

### 🏆 Golden Chest (8-10 correct)

| Resource | Amount | Notes |
|----------|--------|-------|
| Coins | 300-500 | Random |
| PvP Tickets | 4-5 | Random |
| Marathon Bonuses | 2-3 bonuses | 2 guaranteed |

**Bonus Selection:**
- 100%: 2 random bonuses (guaranteed, different types)
- 40% chance: +1 extra bonus (3 total, all different types)
- Avg 2.4 bonuses per chest

**With Streak:**
```
3d → 330-550 coins
7d → 375-625 coins
30d → 450-750 coins
```

## Marathon Bonuses (4 types)

| Bonus | Effect in Marathon | Rarity Weight |
|-------|-------------------|---------------|
| 🛡️ Shield | 1 free mistake (no life loss) | 25% |
| 🔀 50/50 | Remove 2 wrong answers | 25% |
| ⏭️ Skip | Skip question without penalty | 25% |
| ❄️ Freeze | +10 seconds to timer | 25% |

**Notes:**
- Stored in user inventory
- Can stack (e.g., 3 shields)
- No expiration
- Used in Solo Marathon only (not in Daily)

## PvP Tickets

**Purpose:** Entry cost for PvP Duel and Party Mode.

**Storage:**
- User inventory (database)
- No expiration
- Can accumulate (no limit)

**Usage:**
- PvP Duel: 1 ticket per match
- Party Mode: 1 ticket per game

## Coins

**Purpose:** Universal currency.

**Usage:**
- Second attempt: 100 coins
- Streak recovery: 50 coins
- Marathon continue: 200 coins (TBD)
- Shop purchases (cosmetics, bonuses)

## Reward Calculation Algorithm

```go
func CalculateRewards(chestType ChestType, streak int) ChestContents {
    baseRewards := getBaseRewards(chestType)

    // Apply streak multiplier to coins
    multiplier := getStreakMultiplier(streak)
    coins := int(float64(baseRewards.Coins) * multiplier)

    // Bonuses selection
    bonuses := selectBonuses(chestType)

    return ChestContents{
        Coins: coins,
        PvPTickets: baseRewards.PvPTickets,
        Bonuses: bonuses,
    }
}

func getBaseRewards(chestType ChestType) BaseRewards {
    switch chestType {
    case Wooden:
        return BaseRewards{
            Coins: random(50, 100),
            PvPTickets: 1,
        }
    case Silver:
        return BaseRewards{
            Coins: random(150, 250),
            PvPTickets: random(2, 3),
        }
    case Golden:
        return BaseRewards{
            Coins: random(300, 500),
            PvPTickets: random(4, 5),
        }
    }
}

func selectBonuses(chestType ChestType) []BonusType {
    switch chestType {
    case Wooden:
        // 50% chance for 1 bonus
        if random() < 0.5 {
            return [randomBonus()]
        }
        return []
    case Silver:
        // 100% × 1 bonus + 30% chance for 2nd
        b1 := randomBonus()
        if random() < 0.3 {
            return [b1, randomDifferentBonus(b1)]
        }
        return [b1]
    case Golden:
        // 100% × 2 bonuses + 40% chance for 3rd
        b1 := randomBonus()
        b2 := randomDifferentBonus(b1)
        if random() < 0.4 {
            return [b1, b2, randomDifferentBonus(b1, b2)]
        }
        return [b1, b2]
    }
}
```

## Premium Subscription Effects

**Chest Upgrade:**
```
Original → Upgraded
─────────────────────
Wooden  → Silver
Silver  → Golden
Golden  → Golden (no change)
```

**Additional Coins Bonus (Golden with Premium):**
```
baseCoins = 300-500
premiumBonus = baseCoins * 0.5
totalCoins = baseCoins + premiumBonus = 450-750
```

Applied AFTER streak multiplier:
```
finalCoins = (baseCoins + premiumBonus) * streakMultiplier
```

## Reward Distribution Timeline

1. Game completed → Chest type determined
2. Results screen → Shows chest type
3. Player taps "Open" → Animation plays (3 sec)
4. Animation completes → Rewards revealed
5. Player taps "Collect" → **Resources added to inventory**
6. Database updated (atomic transaction)

## Database Schema

```sql
-- User inventory
CREATE TABLE user_inventory (
    user_id VARCHAR(36) PRIMARY KEY,
    coins INTEGER DEFAULT 0,
    pvp_tickets INTEGER DEFAULT 0,
    bonus_shield INTEGER DEFAULT 0,
    bonus_fifty_fifty INTEGER DEFAULT 0,
    bonus_skip INTEGER DEFAULT 0,
    bonus_freeze INTEGER DEFAULT 0,
    updated_at TIMESTAMP
);

-- Reward history
CREATE TABLE reward_history (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    source VARCHAR(50),  -- 'daily_challenge'
    chest_type VARCHAR(20),
    coins INTEGER,
    pvp_tickets INTEGER,
    bonuses JSONB,
    created_at TIMESTAMP
);
```

## Edge Cases

**Chest not opened:**
- Rewards still granted on completion
- Chest animation = just UI feedback

**Disconnect during chest opening:**
- Rewards already in DB (transaction on completion)
- Re-opening shows same rewards (idempotent)

**Inventory overflow:**
- No limits currently
- TBD: Max caps per resource (future balance)
