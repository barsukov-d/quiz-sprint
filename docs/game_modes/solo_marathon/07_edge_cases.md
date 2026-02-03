# Solo Marathon - Edge Cases & Error Handling

## Lives System Edge Cases

### Shield activated but answer correct
**Behavior:**
- Shield NOT consumed
- Shield deactivates after question (does NOT carry to next question)
- Player must activate again manually for next question

**Implementation:**
```go
if isCorrect {
    // Don't consume shield, just deactivate
    shieldActive = false
    return result
}

if shieldActive {
    consumeBonus(BonusShield)
    shieldActive = false
    // Lives unchanged
}
```

### Last life + Shield active
**Scenario:** Player has ❤️, shield active, wrong answer.

**Behavior:**
- Shield consumed
- Lives remain at 1 (❤️)
- Game continues

### No Shield, wrong answer at 1 life
**Behavior:**
- Instant game over
- Continue offer shown immediately
- Cannot answer more questions until continue/quit

---

## Bonus Combinations

### Multiple bonuses same question

**Allowed:**
```
✅ Freeze + 50/50
✅ Freeze + Shield
✅ Shield + 50/50
```

**Result:** Both consumed independently.

**Not allowed:**
```
❌ Skip + anything (Skip ends question immediately)
❌ Shield twice (only one active at a time)
```

### 50/50 then Skip
**Scenario:** Used 50/50, then decides to Skip.

**Behavior:**
- Both bonuses consumed
- Question skipped (2 answers already removed)
- No score change

### Shield active, used Skip
**Scenario:** Shield active, player skips question.

**Behavior:**
- Shield NOT consumed (no wrong answer)
- Shield deactivates (does NOT carry to next question)
- Skip consumed
- Player must re-activate Shield on next question if needed

---

## Continue Edge Cases

### Continue at exact 0 lives
**Behavior:**
- Lives reset to 1 (NOT +1)
- Same question reloaded
- Timer resets to full (15s/12s/10s/8s)

**Example:**
```
Before: ❤️ → wrong answer → 🖤 (game over)
Continue: 🖤 → ❤️ (1 life)
```

### Multiple continues same game
**Cost escalation:**
```
1st: 200 coins
2nd: 400 coins
3rd: 600 coins
4th: 800 coins (no ad option)
```

**Unlimited continues allowed** (if player has coins).

### Continue with insufficient coins
**Response:**
```json
{
  "error": "INSUFFICIENT_COINS",
  "required": 400,
  "current": 150,
  "action": {
    "type": "navigate",
    "route": "/shop"
  }
}
```

Frontend shows:
```
Недостаточно монет!
Нужно: 400 💰
У тебя: 150 💰

[ Пополнить в магазине ]
```

### Rewarded Ad failed to load
**Behavior:**
- Show "Ad unavailable" message
- Offer coins option only
- No free continue

### Continue after quit
**Not allowed.**
- Once player quits (closes game over screen), cannot continue
- Must start fresh game

---

## Disconnect & Network Issues

### Disconnect during question
**State saved:**
- Lives, score, bonuses after each answer
- Current question index
- Active bonuses state (Shield on/off)

**Resume:**
- Player returns → Same question
- Timer resets to full (integrity - no exploit)
- Can continue playing

### Network timeout on answer submit
**Server behavior:**
- If no answer received within 30s after timer: Auto-submit empty
- Counts as wrong answer (NO special treatment)

**Client behavior:**
- Retry submission on reconnect
- If already submitted (idempotency check): Use existing result

### Multiple answer submissions (network retry)
**Protection:**
```go
if session.IsQuestionAnswered(questionID) {
    return ErrQuestionAlreadyAnswered
}
```

HTTP: `409 Conflict`

---

## Leaderboard Edge Cases

### Tied scores
**Example:**
```
Player A: 87 correct, 87 total
Player B: 87 correct, 90 total
Player C: 87 correct, 90 total, completed earlier
```

**Ranking:**
```
#1: Player A (more efficient: 87/87)
#2: Player C (87/90, completed earlier)
#3: Player B (87/90, completed later)
```

**Redis score formula (no timestamp component):**
```
redisScore = correctAnswers * 1000000 - totalQuestions
```

**When Redis scores are equal** (same correct + same total), tiebreak by `completedAt ASC` (earlier = better) — resolved at application level when querying.

### Player plays multiple times same week
**Only best score counts.**

**Leaderboard shows:**
- Best game per player
- If 2nd game better: Replaces 1st in leaderboard

### Week boundary
**Scenario:** Player starts at Sunday 23:58, finishes Monday 00:02.

**Behavior:**
- Game belongs to WEEK OF COMPLETION (`completedAt`)
- Goes to next week's leaderboard

### Rank calculation delay
**Real-time:**
- Score added to Redis instantly on completion
- Rank calculated on-demand (ZREVRANK)

**No delay**, but heavy load may cause slight lag (<1s).

---

## Bonus Inventory Edge Cases

### Bonus depleted mid-game
**Scenario:** Started with 🛡️×1, used it, no more shields.

**UI:**
- Shield button grayed out
- Tooltip: "У тебя нет щитов"

**In-game offer (optional monetization):**
```
┌───────────────────────────┐
│ Бонусы закончились!       │
│ Купить экстренный набор:  │
│ 🛡️×1  ❄️×3               │
│ [ 150 💰 ] или [ 📺 ]     │
└───────────────────────────┘
```

### Starting game with 0 bonuses
**Allowed.**
- Player can play "hardcore mode"
- Harder to achieve high score

### Bonus usage history
**Tracked for analytics:**
```sql
INSERT INTO marathon_bonus_usage (
    game_id, player_id, bonus_type, question_id, used_at
)
```

Used for:
- Balance analysis (which bonuses most valuable?)
- Fraud detection (suspicious patterns)

---

## Personal Best Edge Cases

### New record exactly equals old record
**Not considered new:**
```
Old: 87
Current: 87
→ No bonus, no notification
```

### New record after using continues
**Still counts as new record:**
```
Score: 100 (with 2 continues)
Old best: 90
→ New record! +500 coins
```

**UI shows:**
```
🎉 НОВЫЙ РЕКОРД! 🎉
100 правильных ответов
(Продолжений: 2)
+500 монет
```

### Multiple new records same week
**Each milestone once:**
```
Game 1: Score 30 → Milestone 25 ✓ (+100 coins)
Game 2: Score 60 → Milestone 50 ✓ (+250 coins)
Game 3: Score 70 → No new milestone
Game 4: Score 120 → Milestones 100 ✓ (+500 coins)
```

---

## Quit & Abandon

### Quit mid-game
**Warning:**
```
Точно выйти?
Прогресс будет потерян!
[ Выйти ]  [ Продолжить игру ]
```

**If confirmed:**
- Status: `ABANDONED` (intentional quit, NOT `COMPLETED`)
- Score saved to history (but NOT personal best)
- NOT in leaderboard (incomplete run)

### Abandon due to timeout
**Trigger:** No activity for 30+ minutes.

**Behavior:**
- Auto-quit
- Status: `ABANDONED`
- No leaderboard entry

### Resume abandoned game
**Not allowed.**
- Start fresh game only

---

## Scoring Edge Cases

### Skipped questions count?
**No.**
```
Answered: 47
Skipped: 3
Total shown: 50
Score: 47 (only correct)
```

**Leaderboard:**
```
score = 47
totalQuestions = 50
efficiency = 47/50 = 94%
```

### Wrong answer then correct on retry (after continue)
**Both counted:**
```
Question 25: Wrong (at 0 lives → continue)
Question 25 retry: Correct
Total questions: Still counts as 2 attempts at Q25
Score: +1 for correct retry
```

---

## Security & Anti-Cheat

### Impossible score (too high too fast)
**Threshold:** >200 correct answers.

**Action:**
- Flag for manual review
- Temporarily hide from leaderboard
- Investigate timing patterns

### Suspicious timing patterns
**Red flags:**
- All answers exactly 1.0 second
- All answers correct with <0.5s
- Perfect score (300+) with no bonuses

**Action:**
- Automated flag
- Manual review required for leaderboard

### Multiple devices same account
**Allowed** (legitimate use case).

**Restriction:**
- Only 1 active game at a time
- Starting new game abandons previous

---

## API Error Responses

### Standard format
```json
{
  "error": {
    "code": "INSUFFICIENT_BONUSES",
    "message": "У вас недостаточно бонусов",
    "details": {
      "bonusType": "shield",
      "required": 1,
      "current": 0
    },
    "action": {
      "type": "show_offer",
      "offerId": "emergency_bonus_pack"
    }
  }
}
```

### Error codes
```
GAME_NOT_FOUND
GAME_NOT_ACTIVE
GAME_OVER
ACTIVE_GAME_EXISTS
INSUFFICIENT_BONUSES
INSUFFICIENT_COINS
INVALID_TIME_TAKEN
QUESTION_NOT_FOUND
ANSWER_NOT_FOUND
QUESTION_ALREADY_ANSWERED
AD_UNAVAILABLE
```

---

## Monitoring & Alerts

**Key metrics:**
- Game completion rate (% reached game over vs quit)
- Avg score per game
- Continue conversion rate
- Bonus usage rate (% games using bonuses)
- Top 100 score threshold (difficulty indicator)

**Alerts:**
- Avg score drops >20% → Questions too hard
- Continue conversion <15% → Price too high
- Bonus usage <50% → Players hoarding (need incentive)
- Leaderboard fraud detected → Manual review
