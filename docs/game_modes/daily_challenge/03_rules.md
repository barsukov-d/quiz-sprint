# Daily Challenge - Business Rules

## Score Calculation

### Per Question
```
basePoints = 100
timeBonus = max(0, (15 - timeTaken) * 5)
questionScore = isCorrect ? (basePoints + timeBonus) : 0

Max per question: 175 (answered instantly)
Min per question: 0 (wrong or timeout)
```

### Total Score
```
baseScore = sum(questionScores)  // 0 to 1750
streakMultiplier = getStreakBonus(streak)
finalScore = floor(baseScore * streakMultiplier)
```

### Streak Multiplier Table

| Streak | Multiplier | Formula |
|--------|------------|---------|
| 0-2 days | 1.0 | No bonus |
| 3-6 days | 1.1 | +10% |
| 7-13 days | 1.25 | +25% |
| 14-29 days | 1.4 | +40% |
| 30+ days | 1.5 | +50% |

**Note:** Milestones trigger event at 3, 7, 14, 30, 100 days.

## Chest Type Determination

```go
func GetChestType(correctAnswers int) ChestType {
    if correctAnswers >= 8 {
        return ChestTypeGolden    // 8-10 correct
    }
    if correctAnswers >= 5 {
        return ChestTypeSilver    // 5-7 correct
    }
    return ChestTypeWooden        // 0-4 correct
}
```

## Streak Rules

### Update Logic
```
if currentDate == lastPlayedDate + 1 day:
    streak++
else if currentDate == lastPlayedDate:
    // Already played today, no change
else:
    streak = 0  // Broken
```

**Important:**
- Streak updates ONLY on game completion
- Starting game doesn't count (must finish)
- Timezone: All dates in UTC

### Streak Recovery
- Available: Within 24h after break
- Cost: 50 coins OR Rewarded Ad
- Effect: Restores previous streak

## Validation Rules

### Time Taken
```
0 < timeTaken ≤ 15 seconds
```
Violations:
- `timeTaken ≤ 0` → `ErrInvalidTimeTaken`
- `timeTaken > 15` → `ErrInvalidTimeTaken`
- Server validates: `abs(clientTime - serverTime) < 2s` (anti-cheat)

### Answer Submission
- Each question answered exactly once
- Cannot change answer after submission
- Violation: `ErrQuestionAlreadyAnswered`

### Daily Limit
```
maxFreeAttempts = 1 per UTC day
```

Check:
```sql
SELECT COUNT(*) FROM daily_games
WHERE player_id = ? AND date = ? AND status != 'abandoned'
```

Violations:
- Already played: `ErrAlreadyPlayedToday`
- Can retry: 100 coins or Premium subscription

### Game State Transitions

```
NOT_STARTED --[start()]--> IN_PROGRESS
IN_PROGRESS --[answer all]--> COMPLETED
IN_PROGRESS --[24h timeout]--> ABANDONED
```

Invalid transitions throw `ErrInvalidGameStatus`.

## Leaderboard Rules

### Ranking
Primary sort: `finalScore DESC`
Tiebreaker: `completedAt ASC` (earlier = better)

### Formula
```
leaderboardScore = finalScore * 1000000 + (maxTimestamp - completedAt)
```
Stored in Redis Sorted Set (higher = better rank).

### Visibility
- **Global:** All players worldwide
- **Friends:** Telegram contacts (TBD: implementation)
- **Country:** By user profile country (TBD: detection method)

### Update Timing
- Real-time: Added to Redis on game completion
- Rank calculated: On leaderboard request
- Historical: Kept 7 days, then archived

## Second Attempt Rules

**Availability:**
- After completing first attempt
- Same day only

**Cost:**
- 100 coins (deducted upfront)
- OR Rewarded Ad

**Effect:**
- Creates NEW `DailyGame` with same `date`
- Both attempts saved
- **Best score** counts for leaderboard
- Streak: Uses FIRST completion (not affected by retry)

**Limits:**
- Free users: 1 retry per day (total 2 attempts)
- Premium users: Unlimited retries

## Premium Subscription Benefits

**Chest Upgrade:**
```
wooden → silver
silver → golden
golden → golden + 50% bonus coins
```

Applied automatically at chest opening.

**Other:**
- Unlimited second attempts
- +50% coins from chest
- Exclusive cosmetics (TBD)

## Anti-Cheat

**Server validates:**
1. Time taken (realistic range)
2. Answer sequence (no skipping questions)
3. Completion time (min 10 seconds total)
4. Request signatures (Telegram auth)

**Penalties:**
- Suspicious activity → Game invalidated
- Repeated violations → Account flag (TBD: ban system)
