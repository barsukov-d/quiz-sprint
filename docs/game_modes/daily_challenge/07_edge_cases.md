# Daily Challenge - Edge Cases & Error Handling

## Timing Edge Cases

### Started at 23:58, finished at 00:02
**Behavior:**
- Game belongs to START date (23:58 date)
- Streak updated for START date
- Result appears in START date leaderboard

**Implementation:**
```go
game.date = currentDate  // Fixed at start
```

### Server time vs Client time
**Problem:** Client can fake time.

**Solution:**
- Server timestamps all events
- Client `timeTaken` validated: `0 < t ≤ 15`
- Server checks: `abs(clientTime - serverTime) < 2s`
- Violation → `ErrInvalidTimeTaken`

### Timeout at 15 seconds
**Behavior:**
- Timer reaches 0:00
- Answer auto-submitted as empty
- Counts as wrong answer (0 points)
- Next question appears automatically

### Game abandoned (24h timeout)
**Trigger:** Player started but didn't finish within 24h.

**Behavior:**
```
if now - startedAt > 24h && status == "in_progress":
    status = "abandoned"
    streak = 0  // Broken
```

**Recovery:** Cannot resume, must start new attempt (if has coins/ad).

---

## Connection Issues

### Disconnect during game
**State saved:**
- Each answer submission → DB update
- Progress persists

**Resume:**
- Player returns → `GET /api/v1/daily/status`
- Response includes `gameId` if incomplete
- Frontend resumes from `currentQuestionIndex`

**Timer:**
- NOT paused (server continues counting)
- May lose time during disconnect
- If timeout → auto-submit empty

### Network delay on answer submission
**Problem:** Answer sent at 14s, arrives at 16s.

**Solution:**
- Server checks `answeredAt` timestamp (client-sent)
- Validates: `answeredAt - startedAt < 15s * questionIndex`
- If valid: Accept with client's `timeTaken`
- If invalid: Reject with `ErrInvalidTimeTaken`

### Duplicate answer submission
**Scenario:** Double-tap or network retry.

**Solution:**
- Check `session.IsQuestionAnswered(questionID)`
- If already answered: Return `ErrQuestionAlreadyAnswered`
- HTTP: `409 Conflict`

---

## Leaderboard Edge Cases

### Exact same score
**Tiebreaker:** `completedAt ASC` (earlier = better).

**Redis score:**
```
leaderboardScore = finalScore * 1000000 + (maxTimestamp - completedAt)
```

### Player not in top 100
**Response:**
```json
{
  "entries": [/* top 100 */],
  "playerRank": {
    "rank": 5847,
    "playerId": "user_123",
    "score": 450
  }
}
```

Separate ZREVRANK call for player's rank.

### Leaderboard updates
**Timing:**
- Real-time: Added to Redis on game completion
- Rank: Calculated on-demand (ZREVRANK)
- No delay, no batch processing

### Historical leaderboards
**Retention:** 7 days in Redis.

**Archival:**
```
After 7 days:
- Copy to PostgreSQL (daily_leaderboard_archives)
- Remove from Redis
- Query historical via SQL
```

---

## Streak Edge Cases

### Played today, then retry
**Behavior:**
- First completion updates streak
- Retry does NOT update streak again
- Streak timestamp = first completion

### Played yesterday at 23:59, today at 00:01
**Dates:** 2026-01-27 and 2026-01-28 (consecutive).

**Streak:** Incremented ✓

### Played yesterday, skip today, play tomorrow
**Dates:** 2026-01-27, skip 2026-01-28, play 2026-01-29.

**Streak:** Reset to 0 (gap detected).

### Streak recovery
**Available:** Within 24h of break.

**Example:**
- Last played: 2026-01-26
- Today: 2026-01-28 (1 day gap)
- Can recover: Only if `now < 2026-01-28 00:00 + 24h`

**After recovery:**
- Streak restored to previous value
- `lastPlayedDate` updated to 2026-01-27 (fill gap)

### Streak milestone notification
**Trigger:** Exactly at 3, 7, 14, 30, 100 days.

**Event:** `StreakMilestoneReachedEvent`

**UI:** Toast notification with bonus info.

---

## Second Attempt Edge Cases

### Retry before opening chest
**Allowed:** Yes.

**Behavior:**
- First game completed (chest not opened)
- Retry creates NEW game
- Both games independent
- Can open both chests

### Retry with better score
**Leaderboard:** Shows best score.

**Implementation:**
```sql
SELECT MAX(final_score) FROM daily_games
WHERE player_id = ? AND date = ?
```

### Retry with worse score
**Leaderboard:** Still shows best (first attempt).

**Streak:** Uses first completion (no change).

### Non-premium retry limit
**Free users:** 1 retry (total 2 attempts).

**Check:**
```sql
SELECT COUNT(*) FROM daily_games
WHERE player_id = ? AND date = ?
```

If count ≥ 2 and not premium → `ErrRetryLimitReached`

### Insufficient coins
**Response:**
```json
{
  "error": "insufficient_coins",
  "required": 100,
  "current": 45
}
```

HTTP: `400 Bad Request`

---

## Reward Edge Cases

### Chest not opened
**Behavior:**
- Rewards ALREADY granted on completion (DB transaction)
- Chest opening = UI animation only
- Idempotent: Multiple opens show same rewards

### Disconnect during chest animation
**Resume:**
- Rewards already in DB
- Animation replays (same rewards)
- No duplication

### Premium upgrade at exact 8 correct
**Without premium:** Golden chest (8-10 correct).
**With premium:** Still Golden (no higher tier).

**Bonus:** +50% coins applied.

**Example:**
```
Base: 400 coins
Premium: 400 * 1.5 = 600 coins
```

### Bonus selection randomness
**Seed:** Use `gameID + timestamp` for determinism.

**Goal:** Same game → same bonuses on replay (for testing).

---

## Database Edge Cases

### Transaction failure during completion
**Scenario:** Game completed, but DB write fails.

**Solution:**
- All writes in single transaction
- If fails: Game remains `in_progress`
- Player can continue answering
- Auto-complete triggers again

### Concurrent game starts
**Scenario:** User taps "Start" twice rapidly.

**Solution:**
```sql
UNIQUE (player_id, date)
```

First insert succeeds, second fails with `ErrAlreadyPlayedToday`.

### Redis leaderboard desync
**Problem:** Redis updated, but DB write fails.

**Solution:**
- DB write FIRST (source of truth)
- Then Redis update
- If Redis fails: Background job repairs from DB

**Repair query:**
```sql
SELECT player_id, final_score, completed_at
FROM daily_games
WHERE date = ? AND status = 'completed'
```

---

## Security & Anti-Cheat

### Impossible time values
**Examples:**
- `timeTaken = -5` → Reject
- `timeTaken = 0.001` → Suspicious (flag)
- `timeTaken > 15` → Reject

**Action:**
```go
if timeTaken <= 0 || timeTaken > 15 {
    return ErrInvalidTimeTaken
}
if timeTaken < 0.5 {
    logSuspiciousActivity(playerID, "impossible_time")
}
```

### Answering questions out of order
**Protection:** Track `currentQuestionIndex`.

**Validation:**
```go
if questionID != session.CurrentQuestion().ID() {
    return ErrInvalidQuestion
}
```

### Total game time too fast
**Check:**
```go
totalTime = completedAt - startedAt
minTime = 10 * 0.5  // 5 seconds minimum (0.5s per question)

if totalTime < minTime {
    invalidateGame()
    logSuspiciousActivity(playerID, "game_too_fast")
}
```

### Multiple games from same IP
**Monitoring:** Track active games per IP.

**Threshold:** >10 concurrent games → Rate limit.

### Modified client time
**Protection:** Server timestamps used for all calculations.

**Client `timeTaken`:** Only for scoring bonus (validated).

---

## API Error Responses

### Standard format:
```json
{
  "error": {
    "code": "ALREADY_PLAYED_TODAY",
    "message": "You have already played today",
    "details": {
      "date": "2026-01-28",
      "gameId": "dg_abc123"
    }
  }
}
```

### Error codes:
```
INVALID_TIME_TAKEN
GAME_NOT_FOUND
GAME_NOT_ACTIVE
GAME_COMPLETED
ALREADY_PLAYED_TODAY
INSUFFICIENT_COINS
RETRY_LIMIT_REACHED
QUIZ_NOT_FOUND
QUESTION_NOT_FOUND
INVALID_QUESTION
QUESTION_ALREADY_ANSWERED
```

---

## Monitoring & Alerts

**Key metrics:**
- Games abandoned rate (target: <5%)
- Average completion time
- Suspicious activity flags
- API error rates
- Leaderboard desync incidents

**Alerts:**
- Abandoned rate >10% → Investigate UX
- Suspicious activity spike → Potential cheating
- Redis-DB desync detected → Run repair job
