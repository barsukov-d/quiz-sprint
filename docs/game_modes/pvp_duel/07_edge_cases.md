# PvP Duel - Edge Cases & Error Handling

## Matchmaking Edge Cases

### Both players find each other simultaneously
**Behavior:**
- Server assigns game atomically
- First transaction wins
- Other player stays in queue

```go
// Use Redis SETNX for atomic game creation
if !redis.SetNX("game:pending:"+player1+":"+player2, gameID) {
    // Already matched by other instance
    return
}
```

### Player disconnects during queue
**Behavior:**
- Auto-remove from queue after 10s heartbeat timeout
- Ticket NOT consumed (never started)

### No opponent found in 60 seconds
**Behavior:**
- Offer bot game
- If declined → Return ticket
- If accepted → Play vs bot, no MMR change

### Same player matched twice in a row
**Prevention:**
```go
func canBeMatched(player1, player2 PlayerID) bool {
    lastGame := getLastGame(player1)
    if lastGame != nil && lastGame.OpponentID == player2 {
        return false  // Prevent immediate rematch
    }
    return true
}
```

---

## Friend Challenge Edge Cases

### Challenge sent to offline friend
**Behavior:**
- Push notification sent via Telegram
- 5 minute expiry (longer than online)
- If no response → Challenge expires, ticket refunded

### Friend opens app but challenge already expired
**Behavior:**
- Show "Вызов истёк" message
- Offer to challenge back

### Both friends challenge each other simultaneously
**Behavior:**
- First challenge wins (by timestamp)
- Second challenge auto-converts to "accept"
- Game starts immediately

### Challenge link used by wrong person
**Behavior:**
- Only original link creator can start the game
- Other person sees: "Ссылка уже использована"

### Challenge link used after 24h expiry
**Response:**
```json
{
  "error": {
    "code": "CHALLENGE_EXPIRED",
    "message": "Срок действия вызова истёк",
    "action": {
      "type": "show_create_new",
      "message": "Попроси друга отправить новый вызов"
    }
  }
}
```

---

## During Duel Edge Cases

### Player answers exactly at 0 seconds
**Behavior:**
- Accept if server receives before timeout
- Network latency tolerance: 500ms
- If received after timeout → Treat as no answer (wrong)

### Both players answer at exact same time
**Behavior:**
- Both answers recorded with server timestamp
- If truly identical (to millisecond) → Use smaller playerID as tiebreaker
- Extremely rare scenario

### Player submits answer twice
**Behavior:**
```go
if isAlreadyAnswered(gameID, questionID, playerID) {
    return ErrQuestionAlreadyAnswered
}
```
- First answer counts
- Subsequent attempts rejected

### Player tries to answer wrong question
**Behavior:**
```go
if game.CurrentQuestionIndex != questionIndex {
    return ErrInvalidQuestion
}
```

### One player answers, other disconnects
**Behavior:**
1. Start 10s grace period
2. If reconnected → Continue game
3. If timeout → Current question = wrong for disconnected player
4. 3 consecutive timeouts → Forfeit game

---

## Disconnect Handling

### Disconnect during countdown (3-2-1)
**Behavior:**
- 5s grace period
- If reconnected → Continue countdown
- If timeout → Game cancelled, both tickets refunded

### Disconnect mid-question
**Behavior:**
- Opponent sees "Соперник переподключается..."
- Timer continues for disconnected player
- If reconnected → Resume (remaining time)
- If timeout → Wrong answer, move to next question

```go
func handleDisconnect(gameID, playerID string) {
    go func() {
        time.Sleep(10 * time.Second)
        if !isReconnected(playerID) {
            submitEmptyAnswer(gameID, playerID)
            incrementMissedQuestions(gameID, playerID)

            if getMissedQuestions(gameID, playerID) >= 3 {
                forfeitGame(gameID, playerID)
            }
        }
    }()
}
```

### Both players disconnect
**Behavior:**
- Game paused
- 30s to reconnect for either
- If neither returns → Game cancelled, no MMR change, tickets refunded

### Reconnect after game completed
**Behavior:**
- Show final results
- MMR already applied
- Can view game history

---

## Score & Tiebreaker Edge Cases

### Tied score (5:5)
**Tiebreaker:** Total time
```go
func determineWinner(game *DuelGame) PlayerID {
    if game.Player1.Score > game.Player2.Score {
        return game.Player1.ID
    }
    if game.Player2.Score > game.Player1.Score {
        return game.Player2.ID
    }

    // Tied score - check time
    if game.Player1.TotalTime < game.Player2.TotalTime {
        return game.Player1.ID
    }
    if game.Player2.TotalTime < game.Player1.TotalTime {
        return game.Player2.ID
    }

    // Extremely rare: same score, same time
    // Use first correct answer
    return firstCorrectPlayer(game)
}
```

### Both players get 0 correct
**Behavior:**
- Tiebreaker by time (who "failed faster")
- If same time → First to answer any question
- Winner still gets MMR (even 0:0)

### Perfect tie (same score, same time)
**Probability:** <0.001%
**Tiebreaker:** First correct answer on any question
**If still tied:** Lower playerID wins (arbitrary but deterministic)

---

## MMR Edge Cases

### New player vs Legend
**Behavior:**
- Matchmaking shouldn't allow (60s timeout)
- If forced (bot decline + no other players):
  - New player wins → +28-32 MMR (big boost)
  - Legend wins → +10 MMR (minimum)
  - New player loses → -10 MMR (protected)

### Player at 0 MMR loses
**Behavior:**
- MMR cannot go negative
- Stay at 0
- Still in Bronze IV

### Player at exactly rank boundary
**Example:** 1500 MMR (Gold IV floor)
```go
func checkDemotion(rating *PlayerRating, newMMR int) bool {
    // Protection for first 3 games at new rank
    if rating.GamesAtRank <= 3 {
        return false
    }

    oldDivision := getLeagueAndDivision(rating.MMR)
    newDivision := getLeagueAndDivision(newMMR)

    return newDivision < oldDivision
}
```

### Season reset edge case
**Player at 3500 MMR:**
```
newMMR = 1000 + (3500 - 1000) * 0.5 = 2250
```
Placed in Platinum II, not Legend

---

## Referral Edge Cases

### Friend registers but never plays
**Behavior:**
- Inviter gets "registered" reward only
- No further rewards until friend plays

### Friend reaches milestone, inviter deleted account
**Behavior:**
- Rewards not distributed
- Friend still gets their rewards

### Self-referral attempt (same device)
**Detection:**
```go
func validateReferral(inviter, invitee *Player) error {
    if inviter.DeviceFingerprint == invitee.DeviceFingerprint {
        return ErrSelfReferral
    }
    if inviter.IP == invitee.IP && invitee.RegisteredAt-inviter.RegisteredAt < 86400 {
        flagForReview(inviter.ID, "suspicious_referral")
    }
    return nil
}
```

### Referral link used after invitee already registered
**Behavior:**
- No referral created
- Message: "Этот игрок уже зарегистрирован"

### Inviter and invitee duel immediately after registration
**Behavior:**
- Allowed (friend games are encouraged)
- But: No MMR for brand new accounts (placement protection)
- Win trading detection: 50/50 win rate over 20+ games → flag

---

## Rematch Edge Cases

### Rematch requested but opponent left
**Behavior:**
- 15s timeout
- Request expires
- Ticket refunded if not accepted

### Both request rematch simultaneously
**Behavior:**
- Both requests treated as "accept"
- Game starts immediately
- Both tickets consumed

### Rematch with insufficient tickets
**Behavior:**
- Cannot send rematch request
- UI: "Недостаточно билетов"
- Offer to buy tickets

### Opponent accepts rematch but disconnects before start
**Behavior:**
- Ticket consumed (rematch was accepted)
- Game cancelled
- No MMR change

---

## Season Edge Cases

### Game starts Sunday 23:59, ends Monday 00:01
**Behavior:**
- Game counts for NEW season
- Based on `completedAt` timestamp

### Player banned mid-season
**Behavior:**
- Removed from leaderboards
- No seasonal rewards
- MMR frozen

### Multiple peak ranks in season
**Only highest peak counts:**
```go
func updatePeakRank(rating *PlayerRating) {
    if rating.MMR > rating.PeakMMR {
        rating.PeakMMR = rating.MMR
        rating.PeakLeague = rating.League
        rating.PeakDivision = rating.Division
    }
}
```

### Season reward claim after expiry
**Rewards never expire:**
- Can claim anytime
- But must claim before account deletion

---

## Network & Security Edge Cases

### Client time manipulation
**Server validation:**
```go
func validateClientTime(clientTime, serverTime int64) int64 {
    diff := abs(clientTime - serverTime)

    if diff > 500 {  // >500ms discrepancy
        flagSuspicious(playerID, "time_manipulation")
        return serverTime  // Use server time
    }

    return clientTime
}
```

### Answer replay attack
**Prevention:**
```go
// Each answer has unique nonce
type AnswerSubmission struct {
    GameID     string
    QuestionID string
    AnswerID   string
    Nonce      string
    Signature  string
}

func validateSubmission(sub *AnswerSubmission) error {
    if isNonceUsed(sub.Nonce) {
        return ErrReplayAttack
    }
    markNonceUsed(sub.Nonce)
    return nil
}
```

### WebSocket connection hijacking
**Mitigation:**
- Token-based authentication
- Token expires after game
- New token for each game

---

## Bot Game Edge Cases

### Player accepts bot, then finds real player
**Not possible:**
- Bot game starts immediately
- Cannot be in queue during game

### Bot difficulty by league
```go
func getBotAccuracy(playerLeague League) float64 {
    accuracy := map[League]float64{
        LeagueBronze:   0.40,
        LeagueSilver:   0.50,
        LeagueGold:     0.60,
        LeaguePlatinum: 0.70,
        LeagueDiamond:  0.80,
        LeagueLegend:   0.85,
    }
    return accuracy[playerLeague]
}
```

### Player rage-quits bot game
**Behavior:**
- Game ends, no penalty
- No MMR change (it's vs bot)
- Ticket was refunded at bot game start

---

## API Error Responses

### Standard error format
```json
{
  "error": {
    "code": "CHALLENGE_EXPIRED",
    "message": "Вызов истёк",
    "details": {
      "challengeId": "ch_abc123",
      "expiredAt": 1706429000
    },
    "action": {
      "type": "dismiss",
      "nextStep": "Return to lobby"
    }
  }
}
```

### Error codes
```
INSUFFICIENT_TICKETS
ALREADY_IN_QUEUE
ALREADY_IN_MATCH
NO_OPPONENT_FOUND
CHALLENGE_EXPIRED
FRIEND_BUSY
FRIEND_OFFLINE
MATCH_NOT_FOUND
QUESTION_ALREADY_ANSWERED
MATCH_ALREADY_COMPLETED
INVALID_TIME
RATE_LIMITED
BANNED_FROM_RANKED
SEASON_ENDED
```

---

## Monitoring & Alerts

### Key metrics
- Game completion rate
- Avg queue time
- Disconnect rate during game
- MMR distribution by league
- Referral conversion rate

### Alerts
- Queue time >45s for >10% of players → Expand MMR range
- Disconnect rate >5% → Check server health
- Win rate >70% for any player over 50 games → Review for cheating
- Referral fraud pattern → Manual review
