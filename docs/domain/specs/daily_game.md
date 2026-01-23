# Specification: Daily Marathon - Survival Challenge
**Context:** DailyMarathon
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to test my endurance in a daily survival challenge with progressive difficulty, where every mistake costs me dearly and careful strategy determines how far I can go.

## 2. Value Proposition
Daily Marathon is a **"High-Stakes Survival Test"**. Its appeal is built on:
1. **Scarcity:** One attempt per day creates urgency and emotional investment.
2. **Progressive Challenge:** Difficulty escalates, pushing players to their limits.
3. **Strategic Depth:** Power-ups add tactical layer beyond pure knowledge.
4. **Community Competition:** Daily leaderboard fosters FOMO and rivalry.

## 3. Terminology (Ubiquitous Language)
- **Daily Marathon:** Time-limited survival mode with progressive stages, playable once per 24 hours.
- **Lives:** Health system represented as 3 discrete chances (❤️ ❤️ ❤️).
- **Permadeath:** Immediate session end when Lives reach zero.
- **Stage:** Sequential difficulty level (Easy → Medium → Hard → Elite).
- **Power-up:** Consumable item that provides tactical advantage (50/50, Time Extension, Skip).
- **Streak:** Chain of consecutive correct answers (triggers Combo Bonus power-up at 5 Streak).

## 4. Business Rules and Invariants

### 4.1. Survival Mechanics

#### Lives System
- **Starting Lives:** Every player begins with exactly 3 lives.
- **Life Loss Conditions:**
  - Incorrect answer: -1 life
  - Question timeout: -1 life
- **Permadeath Rule:** Session ends immediately when Lives = 0. No revival, no second chances.
- **Visual Feedback:** Heart icons disappear with animation, screen shake on last life lost.

#### Daily Restriction
- **One Attempt Rule:** Player can start Daily Marathon only once per 24-hour period.
- **Reset Time:** Midnight UTC (configurable per timezone).
- **Enforcement:** Server validates `LastAttemptTimestamp` before allowing new session.

### 4.2. Progressive Difficulty

Stages increase in challenge through three vectors: question count, time limit, and topic complexity.

| Stage | Difficulty | Questions | Time Limit | Topics |
|-------|-----------|-----------|------------|--------|
| 1 | Easy | 5 | 30s | Basic general knowledge |
| 2 | Medium | 7 | 20s | Intermediate topics |
| 3 | Hard | 10 | 15s | Advanced, specialized |
| 4+ | Elite | 12+ | 12s | Expert-level, niche |

**Advancement Rule:** Must complete current stage with Lives ≥ 1 to unlock next stage.

### 4.3. Power-up System

#### Earning Power-ups
1. **Stage Completion Reward:** One random power-up upon finishing any stage.
2. **Streak Bonus:** Earn power-up for achieving 5 consecutive correct answers within a stage.

#### Power-up Types
- **50/50:** Removes two incorrect answer options, leaving one correct and one incorrect.
- **Time Extension:** Adds +20 seconds to current question timer.
- **Skip:** Skip current question without answering (no points, no life lost).

#### Inventory Rules
- **Session-Specific:** Power-ups do NOT carry over to next day's marathon.
- **No Limit:** Can accumulate multiple power-ups, use strategically.
- **One Use:** Each power-up is consumed after single use.

### 4.4. Scoring System

**Question Score:** `BasePoints + SpeedBonus + StreakMultiplier`

- **Base Points:** Fixed per difficulty (Easy: 100, Medium: 200, Hard: 300, Elite: 500).
- **Speed Bonus:** Decreases linearly from max to zero over question time limit.
- **Streak Multiplier:** Same as Classic Game (0-2: x1.0, 3-5: x1.5, 6+: x2.0).

**Stage Bonuses:**
- **Stage Completion Bonus:** Large point reward (500 × Stage Number).
- **Survival Bonus:** +200 points per remaining life at stage end.

**Final Score:** Sum of all question scores + stage bonuses.

## 5. Data Model Changes

### Aggregate `DailyMarathon`
- `SessionID` (UUID): Unique session identifier.
- `UserID` (UUID): Player.
- `CurrentStage` (int): Current stage number (1-based).
- `Lives` (int): Remaining lives (0-3).
- `Streak` (int): Current correct answer chain.
- `Inventory` ([]PowerUp): Available power-ups (`{Type: "50/50", Quantity: 2}`).
- `TotalScore` (int): Cumulative score across all stages.
- `LastAttemptTimestamp` (int64): Unix timestamp of last marathon start.
- `Status` (enum): `InProgress`, `Completed`, `Failed`.
- `StagesCleared` (int): Highest stage completed.

### Value Object `PowerUp`
- `Type` (enum): `FiftyFifty`, `TimeExtension`, `Skip`.
- `Quantity` (int): Number of uses available.

## 6. Scenarios (User Flows)

### Scenario 1: Perfect Stage Clear
- **Given:** Player is on Stage 1 with 3 lives, has answered 4/5 questions correctly.
- **When:** Player answers final question correctly within Speed Window.
- **Then:**
  - Streak reaches 5 → Earns Combo Bonus power-up (50/50).
  - Stage 1 complete → Receives Stage Completion Bonus (500 pts) + Survival Bonus (600 pts for 3 lives).
  - Random power-up awarded (e.g., Time Extension).
  - "Stage Clear" screen displays rewards, button to start Stage 2.
  - Lives remain at 3 for next stage.

### Scenario 2: Strategic Power-up Use
- **Given:** Player is on Stage 3 (Hard), Question 8/10, 1 life remaining, has 50/50 and Skip in inventory.
- **When:** Player faces extremely difficult question they don't know.
- **Then:**
  - Player activates **50/50** → Two wrong answers disappear, 50/50 chance remains.
  - If still unsure, player uses **Skip** → Moves to Question 9 without losing life (0 points for Q8).
  - Player survives to continue marathon.

### Scenario 3: Dramatic Permadeath
- **Given:** Player is on Stage 4 (Elite), Question 5/12, 1 life left, Timer: 3 seconds remaining.
- **When:** Player hesitates and timer expires.
- **Then:**
  - Screen flashes red, heart icon shatters.
  - Lives: 1 → 0 (Permadeath triggered).
  - Session immediately ends.
  - "Game Over" screen shows:
    - Final Score: 8,500 points
    - Stages Cleared: 3
    - Daily Rank: #47 out of 1,203 players
  - Message: "See you tomorrow! Marathon resets in 16h 23m"

### Scenario 4: Comeback After Mistake
- **Given:** Player is on Stage 2, Question 3/7, 2 lives, Streak: 2.
- **When:** Player makes mistake on difficult question.
- **Then:**
  - Lives: 2 → 1 (screen shake, heart disappears).
  - Streak resets to 0.
  - Correct answer displayed for learning.
  - Player continues with heightened tension, plays more carefully.
  - On next question, uses Time Extension power-up to avoid another mistake.

## 7. Technical Requirements

### State Persistence
- Server must track: `Lives`, `CurrentStage`, `Inventory`, `Streak`, `Score`.
- All updates validated server-side (client cannot fake power-ups or lives).

### Daily Reset Logic
- Cron job runs at midnight UTC.
- Generates new marathon sequence (randomized questions per stage).
- Resets all players' `LastAttemptTimestamp` eligibility.

### Fair Play
- Power-up usage logged with timestamps.
- Answer validation server-side (no client manipulation).
- Disconnect during active question = timeout penalty (-1 life).

### Leaderboard
- Real-time updates throughout the day.
- Displays: Rank, Username, Score, Stages Cleared.
- Resets daily, archived for historical stats.

## 8. Success Metrics
- **Engagement:** 30% of daily active users attempt Daily Marathon within first week.
- **Completion Rate:** 40% of players clear at least Stage 2.
- **Retention:** 50% of players who try Marathon return next day.
- **Average Depth:** Players reach Stage 2.5 on average (median).
- **Social Sharing:** 15% of players share their rank on Telegram.

---

**Status:** Ready for implementation. Requires daily reset job, power-up inventory system, and leaderboard infrastructure.
