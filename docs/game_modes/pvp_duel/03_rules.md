# PvP Duel - Business Rules

## ELO/MMR System

### Rating Calculation
```go
func CalculateMMRChange(winnerMMR, loserMMR int, result GameResult) (winnerDelta, loserDelta int) {
    K := 32  // K-factor (sensitivity)

    // Expected score (probability of winning)
    expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserMMR-winnerMMR)/400))
    expectedLoser := 1.0 - expectedWinner

    // Actual score
    actualWinner := 1.0  // Win = 1
    actualLoser := 0.0   // Lose = 0

    winnerDelta = int(float64(K) * (actualWinner - expectedWinner))
    loserDelta = int(float64(K) * (actualLoser - expectedLoser))

    // Minimum change
    if winnerDelta < 10 {
        winnerDelta = 10
    }
    if loserDelta > -10 {
        loserDelta = -10
    }

    return winnerDelta, loserDelta
}
```

### MMR Examples

| Winner MMR | Loser MMR | Winner Δ | Loser Δ |
|------------|-----------|----------|---------|
| 1500 | 1500 | +16 | -16 |
| 1500 | 1700 | +24 | -24 |
| 1700 | 1500 | +10 | -10 |
| 1500 | 2000 | +28 | -28 |
| 2000 | 1500 | +10 | -10 |

### Initial MMR
New players start at **1000 MMR** (Silver IV).

---

## League System

### League Thresholds

**Note:** Bronze divisions have 250 MMR range (vs 125 for other leagues) to provide more progression room for new players learning the game.

| League | Division | MMR Min | MMR Max | Range |
|--------|----------|---------|---------|-------|
| 🥉 Bronze IV | 0 | 0 | 249 | 250 |
| 🥉 Bronze III | 1 | 250 | 499 | 250 |
| 🥉 Bronze II | 2 | 500 | 749 | 250 |
| 🥉 Bronze I | 3 | 750 | 999 | 250 |
| 🥈 Silver IV | 4 | 1000 | 1124 | 125 |
| 🥈 Silver III | 5 | 1125 | 1249 |
| 🥈 Silver II | 6 | 1250 | 1374 |
| 🥈 Silver I | 7 | 1375 | 1499 |
| 🥇 Gold IV | 8 | 1500 | 1624 |
| 🥇 Gold III | 9 | 1625 | 1749 |
| 🥇 Gold II | 10 | 1750 | 1874 |
| 🥇 Gold I | 11 | 1875 | 1999 |
| 💍 Platinum IV | 12 | 2000 | 2124 |
| 💍 Platinum III | 13 | 2125 | 2249 |
| 💍 Platinum II | 14 | 2250 | 2374 |
| 💍 Platinum I | 15 | 2375 | 2499 |
| 💎 Diamond IV | 16 | 2500 | 2624 |
| 💎 Diamond III | 17 | 2625 | 2749 |
| 💎 Diamond II | 18 | 2750 | 2874 |
| 💎 Diamond I | 19 | 2875 | 2999 |
| 👑 Legend | 20 | 3000 | ∞ |

### Promotion/Demotion

**Promotion:**
```go
func CheckPromotion(oldMMR, newMMR int) bool {
    oldLeague := GetLeagueByMMR(oldMMR)
    newLeague := GetLeagueByMMR(newMMR)
    return newLeague > oldLeague
}
```

**Demotion Protection:**
- First 3 games at new rank: Cannot demote below division floor
- After 3 games: Normal demotion rules apply

```go
func CanDemote(gamesAtCurrentRank int) bool {
    return gamesAtCurrentRank > 3
}
```

---

## Win Conditions

### Primary: Correct Answers
```go
func DetermineWinner(player1Score, player2Score int) int {
    if player1Score > player2Score {
        return 1  // Player 1 wins
    }
    if player2Score > player1Score {
        return 2  // Player 2 wins
    }
    return 0  // Tie - check tiebreaker
}
```

### Tiebreaker 1: Total Time
```go
func DetermineWinnerByTime(player1TotalTime, player2TotalTime int64) int {
    if player1TotalTime < player2TotalTime {
        return 1  // Player 1 wins (faster)
    }
    if player2TotalTime < player1TotalTime {
        return 2
    }
    return 0  // Still tied - extremely rare
}
```

### Tiebreaker 2: First Correct Answer
```go
func DetermineWinnerByFirstCorrect(answers1, answers2 []Answer) int {
    for i := 0; i < len(answers1); i++ {
        if answers1[i].IsCorrect && !answers2[i].IsCorrect {
            return 1
        }
        if answers2[i].IsCorrect && !answers1[i].IsCorrect {
            return 2
        }
    }
    return 0  // Complete tie (both wrong all questions - impossibly rare)
}
```

---

## Matchmaking

### Queue Algorithm
```go
func FindOpponent(player Player) (*DuelGame, error) {
    startTime := time.Now()

    for {
        elapsed := time.Since(startTime)
        mmrRange := CalculateMMRRange(elapsed)

        opponent := FindPlayerInRange(
            player.MMR - mmrRange,
            player.MMR + mmrRange,
            excludeID: player.ID,
        )

        if opponent != nil {
            return CreateDuelGame(player, opponent), nil
        }

        if elapsed > 60*time.Second {
            return nil, ErrNoOpponentFound  // Offer bot
        }

        time.Sleep(1 * time.Second)
    }
}

func CalculateMMRRange(elapsed time.Duration) int {
    switch {
    case elapsed < 10*time.Second:
        return 50
    case elapsed < 20*time.Second:
        return 100
    case elapsed < 30*time.Second:
        return 200
    case elapsed < 45*time.Second:
        return 300
    default:
        return 500
    }
}
```

### Matchmaking Constraints
- Cannot be matched with same opponent twice in a row
- Cannot be matched with blocked players
- Bot games: No MMR change, ticket refunded

---

## Question Selection

### Duel Question Pool
```go
func SelectDuelQuestions(player1Category, player2Category string) []Question {
    // Select from balanced pool - no advantage to either player
    questions := SelectRandomQuestions(
        count: 7,
        difficulty: "medium",  // All medium difficulty
        excludeRecent: []QuestionID{...},  // Not seen by either player recently
    )

    return ShuffleAnswers(questions)
}
```

**Rules:**
- 7 questions per duel
- All medium difficulty (balanced)
- Neither player has seen questions in last 50 games
- Answer order randomized (same for both players)

---

## Time Validation

### Client Time vs Server Time
```go
func ValidateAnswerTime(clientTime, serverStartTime, serverSubmitTime int64) (int64, error) {
    serverTime := serverSubmitTime - serverStartTime

    // Allow 500ms network latency tolerance
    if abs(clientTime - serverTime) > 500 {
        return serverTime, nil  // Use server time, flag suspicious
    }

    // Time must be positive and <= 10 seconds
    if clientTime < 0 || clientTime > 10000 {
        return 0, ErrInvalidTime
    }

    return clientTime, nil
}
```

### Timeout Handling
```go
const QuestionTimeout = 10 * time.Second

func HandleTimeout(playerID string) {
    // No answer within 10s = wrong answer
    submitAnswer(playerID, nil, 10000)  // time = 10s, answer = nil
}
```

---

## Ticket System

### Ticket Consumption
```go
func StartDuel(player1, player2 *Player) error {
    // Consume tickets
    if err := player1.ConsumeTicket(); err != nil {
        return ErrInsufficientTickets
    }
    if err := player2.ConsumeTicket(); err != nil {
        player1.RefundTicket()  // Rollback
        return ErrInsufficientTickets
    }

    return nil
}
```

### Ticket Refund Scenarios

| Scenario | Refund |
|----------|--------|
| Queue cancelled by player | ✅ Yes |
| Queue timeout (no opponent) | ✅ Yes |
| Game completed normally | ❌ No |
| Opponent disconnected (forfeit) | ❌ No (win awarded) |
| Server error | ✅ Yes |
| Bot game accepted | ✅ Yes |

---

## Forfeit Rules

### Disconnect Forfeit
```go
func HandleDisconnect(gameID, playerID string) {
    // Start 10s grace period
    time.AfterFunc(10*time.Second, func() {
        if !IsReconnected(playerID) {
            IncrementMissedQuestions(gameID, playerID)

            if GetMissedQuestions(gameID, playerID) >= 3 {
                ForfeitGame(gameID, playerID)
            }
        }
    })
}

func ForfeitGame(gameID, loserID string) {
    game := GetDuelGame(gameID)
    winner := GetOpponent(game, loserID)

    // Award win to opponent
    CompleteGame(game, winner.ID, "forfeit")

    // Normal MMR calculation applies
    ApplyMMRChange(winner.ID, loser.ID)
}
```

### Voluntary Forfeit
- "Surrender" button available after question 3
- Immediate loss, normal MMR penalty
- Opponent gets full win MMR

---

## Anti-Cheat

### Impossible Timing
```go
func ValidateAnswer(playerID string, answer Answer) error {
    // Minimum human reaction time: 0.5 seconds
    if answer.TimeTaken < 500 {
        FlagSuspicious(playerID, "too_fast", answer.TimeTaken)
    }

    // Perfect score with all answers < 2 seconds
    if IsPerfectGame(playerID) && AllAnswersUnder(playerID, 2000) {
        FlagSuspicious(playerID, "inhuman_perfect")
    }

    return nil
}
```

### Pattern Detection
- Same answer timing patterns across games
- 100% accuracy over many games
- Coordination with specific opponents (win trading)

### Penalties

| Violation | Penalty |
|-----------|---------|
| First offense | Warning, games reviewed |
| Second offense | 24h ranked ban |
| Third offense | Season ban, rank reset |
| Confirmed cheating | Permanent ban |

---

## Game History

### Data Retention
```sql
-- Keep full game details for 90 days
-- Keep summary (W/L, MMR change) forever

CREATE TABLE duel_games (
    id VARCHAR(36) PRIMARY KEY,
    player1_id VARCHAR(36),
    player2_id VARCHAR(36),
    winner_id VARCHAR(36),
    player1_score INT,
    player2_score INT,
    player1_total_time INT,
    player2_total_time INT,
    player1_mmr_before INT,
    player1_mmr_after INT,
    player2_mmr_before INT,
    player2_mmr_after INT,
    win_reason VARCHAR(20),  -- 'score', 'time', 'forfeit'
    completed_at TIMESTAMP,

    INDEX idx_player1 (player1_id, completed_at),
    INDEX idx_player2 (player2_id, completed_at)
);
```

---

## Seasonal Reset

### Soft Reset Formula
```go
func CalculateSeasonalReset(currentMMR int) int {
    // Compress toward 1000
    baseline := 1000
    compression := 0.5  // 50% compression

    newMMR := baseline + int(float64(currentMMR - baseline) * compression)

    // Minimum 500, no cap
    if newMMR < 500 {
        return 500
    }

    return newMMR
}
```

**Examples:**
| Current MMR | After Reset |
|-------------|-------------|
| 500 | 750 |
| 1000 | 1000 |
| 1500 | 1250 |
| 2000 | 1500 |
| 3000 | 2000 |

### Reset Timing
- Season ends: Last Sunday of month, 23:59 UTC
- New season starts: First Monday of month, 00:00 UTC
- Rewards distributed during reset window (~5 minutes)
