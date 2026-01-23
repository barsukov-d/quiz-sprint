# Specification: Head-to-Head Duel - Fighting Game Mode
**Context:** DuelGame
**Status:** Draft

## 1. Business Goal (User Story)
> As a player, I want to challenge other players to real-time 1v1 quiz battles, where each correct answer deals damage to my opponent, creating the intensity of a fighting game combined with intellectual competition.

## 2. Value Proposition
Head-to-Head Duel transforms quizzes into **"Intellectual Combat"**. Its appeal is built on:
1. **Direct Competition:** Face real opponents in real-time, not just leaderboards.
2. **Fighting Game Feel:** HP bars, combo attacks, critical hits, and special moves.
3. **Psychological Pressure:** Seeing opponent's HP drop (or yours) creates tension and excitement.
4. **Quick Matches:** 3-5 minute battles, perfect for mobile gaming.

## 3. Terminology (Ubiquitous Language)
- **Duel:** 1v1 real-time quiz battle between two players.
- **HP (Health Points):** Each player starts with 100 HP. First to 0 HP loses.
- **Combo:** Chain of consecutive correct answers that amplifies damage.
- **Damage:** HP reduction dealt to opponent on correct answer.
- **Critical Hit:** Bonus damage for answering within the first 3 seconds.
- **Special Move:** Powerful attack unlocked at Combo 3, 5, or 7.
- **Block/Defense:** Damage reduction when opponent answers slower than you.
- **KO (Knockout):** Victory achieved when opponent's HP reaches 0.

## 4. Business Rules and Invariants

### 4.1. Combat Mechanics

#### Base Damage System
- **Correct Answer:** Deal 15 HP damage to opponent.
- **Incorrect Answer:**
  - Take 10 HP self-damage (reflect damage).
  - Reset Combo to 0.
- **Timeout:** Take 5 HP self-damage, reset Combo.

#### Combo System (Amplified Damage)
Consecutive correct answers multiply your damage:
- **Combo 0-1:** x1.0 (15 HP)
- **Combo 2-3:** x1.3 (19 HP) - "Warming Up" 🔥
- **Combo 4-5:** x1.6 (24 HP) - "On Fire" 🔥🔥
- **Combo 6+:** x2.0 (30 HP) - "Unstoppable" 🔥🔥🔥

#### Critical Hit (Speed Bonus)
- If you answer within **first 3 seconds (Speed Window)**, deal **+50% damage**.
- Example: Combo 4 + Critical = 24 * 1.5 = **36 HP damage**.
- Visual: Screen flash, "CRITICAL!" text, sound effect.

#### Special Moves (Combo Finishers)
At specific combo levels, you can trigger a special move (optional, replaces normal attack):
- **Combo 3:** "Power Strike" - Deal 40 HP (instead of 19 HP), but reset Combo to 0.
- **Combo 5:** "Fury Combo" - Deal 50 HP + stun opponent for 2 seconds (they can't answer next question immediately).
- **Combo 7:** "Ultimate Attack" - Deal 70 HP, reset Combo to 0.

**Rule:** Special Move is triggered automatically if you answer correctly at these combo levels. Player sees dramatic animation.

#### Defense Mechanism
- If **both players answer correctly**, compare answer times:
  - Faster player deals **full damage**.
  - Slower player deals only **50% damage** (opponent "blocked").
- Visual: Shield icon appears on faster player.

### 4.2. Match Flow

1. **Matchmaking:** Player enters queue, matched with opponent of similar rating.
2. **Pre-Battle:** 3-second countdown, players see each other's avatars and HP bars.
3. **Battle:** Both players answer the same 10 questions simultaneously.
4. **Real-time Updates:** After each question, both see damage dealt/taken and HP bars update.
5. **Victory Conditions:**
   - **KO Victory:** Opponent's HP reaches 0 before all questions answered.
   - **Decision Victory:** After 10 questions, player with more HP wins.
   - **Draw:** Equal HP (rare) - both get participation rewards.

### 4.3. Scoring and Rewards

- **Winner:** +50 Rating, +500 coins, +3 trophies 🏆.
- **Loser:** -20 Rating, +100 coins (participation).
- **Perfect Victory (No HP lost):** +100 bonus coins, "Flawless Victory" badge.
- **Comeback Victory (Win from <20 HP):** "Clutch Master" achievement.

## 5. Data Model Changes

### Aggregate `DuelGame`
- `DuelID` (UUID): Unique match identifier.
- `Player1`, `Player2` (UserID): Participants.
- `Player1HP`, `Player2HP` (int): Current health (0-100).
- `Player1Combo`, `Player2Combo` (int): Current combo counters.
- `CurrentQuestionIndex` (int): Which question (0-9).
- `Player1Answers`, `Player2Answers` ([]AnswerSubmission): Timestamped answers.
- `Status` (enum): `WaitingForPlayers`, `InProgress`, `Completed`.
- `Winner` (UserID): Set when match ends.
- `EndReason` (enum): `KO`, `Decision`, `Forfeit`, `Draw`.

### Entity `AnswerSubmission`
- `QuestionID` (UUID)
- `SelectedAnswerID` (UUID)
- `IsCorrect` (bool)
- `ResponseTime` (int): Milliseconds taken.
- `DamageDealt` (int): HP damage to opponent.
- `IsCritical` (bool): Was it a critical hit?
- `ComboAtTime` (int): Combo level when answer submitted.

## 6. Scenarios (User Flows)

### Scenario 1: Critical Hit Combo
- **Given:** Player has Combo = 2, both players answer Question 3 correctly.
- **When:** Player answers in 2.5 seconds (critical), opponent in 5 seconds.
- **Then:**
  - Player's Combo becomes 3.
  - Player deals: 19 (base) * 1.5 (critical) = **28 HP**.
  - Opponent deals: 19 * 0.5 (blocked) = **9 HP**.
  - Visual: Player sees "CRITICAL HIT! 28 DMG", opponent's HP bar drops dramatically.
  - UI: Player's side shows fire effects, "Warming Up 🔥" status.

### Scenario 2: Special Move - Ultimate Attack
- **Given:** Player has Combo = 6, opponent has 80 HP.
- **When:** Player answers Question 7 correctly.
- **Then:**
  - Combo reaches 7, triggering "Ultimate Attack" automatically.
  - Screen zooms in on player's avatar, special animation plays.
  - Opponent takes **70 HP damage** (80 → 10 HP).
  - Player's Combo resets to 0.
  - Text overlay: "ULTIMATE ATTACK!"

### Scenario 3: Dramatic Comeback
- **Given:** Player has 15 HP, opponent has 60 HP, Question 8 starting.
- **When:**
  - Player answers Q8 correctly in 2s (critical, Combo 3).
  - Opponent makes mistake.
- **Then:**
  - Player deals: 19 * 1.5 = 28 HP (60 → 32 HP).
  - Opponent takes 10 self-damage (32 → 22 HP).
  - Gap closes: 15 HP vs 22 HP.
  - Player sees "Comeback Potential!" motivational text.

### Scenario 4: Perfect KO Victory
- **Given:** Player has 100 HP (no damage taken), opponent has 25 HP, Question 6.
- **When:** Player answers correctly (Combo 5), opponent makes mistake.
- **Then:**
  - Player deals 24 HP (25 → 1 HP).
  - Opponent takes 10 self-damage (1 → 0 HP).
  - **KO! Match ends immediately.**
  - Player sees "FLAWLESS VICTORY!" screen.
  - Both players see match summary, winner gets bonus rewards.

### Scenario 5: Mutual Correct Answer, Defense Trigger
- **Given:** Player answers in 4s, opponent in 3.8s, both correct.
- **When:** Damage calculation occurs.
- **Then:**
  - Opponent (faster) deals full damage: 15 HP.
  - Player (slower) deals reduced: 15 * 0.5 = 7 HP.
  - Both see shield icon on opponent (defender).
  - Text: "Opponent blocked your attack!"

## 7. UI/UX Design

### Battle Screen Layout
```
┌─────────────────────────────────────────┐
│  [🧑 You: 75 HP ████████░░]            │
│  Combo: 4 🔥🔥 "On Fire"                │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  Question: What is 2+2?         │   │
│  │  Timer: ⏱️ 8s                    │   │
│  ├─────────────────────────────────┤   │
│  │  A) 3    [tap]                  │   │
│  │  B) 4    [tap]                  │   │
│  │  C) 5    [tap]                  │   │
│  │  D) 22   [tap]                  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  [👤 Opponent: 50 HP █████░░░░░]       │
│  Combo: 2 🔥 "Warming Up"               │
└─────────────────────────────────────────┘
```

### Damage Animation
- **Hit:** Opponent's HP bar shakes and decreases with red flash.
- **Critical:** Screen border flashes gold, "CRITICAL!" text.
- **Special Move:** Full-screen animation (e.g., fireball, lightning).
- **Combo Milestone:** Character portrait glows, fire particles.

### Sound Design
- **Correct Answer:** Attack sound (punch, slash).
- **Critical Hit:** Power-up chime + whoosh.
- **Combo Build:** Rising pitch sounds (1→2→3).
- **Special Move:** Dramatic orchestral hit.
- **Wrong Answer:** Damage grunt, HP bar alert sound.
- **KO:** Knockout bell, victory fanfare.

## 8. Technical Requirements

### Real-time Sync (WebSocket)
- Both players must see live HP updates.
- Answer submissions processed immediately.
- Latency must be <200ms for fair fights.

### Anti-cheat
- Server validates all answers (client cannot fake correct answers).
- Timestamps verified server-side (no time manipulation).
- Disconnect = forfeit (loser penalty).

### Matchmaking Algorithm
- Pair players by **Rating ±150 points**.
- Max queue time: 30s, then match with anyone.
- No rematches with same opponent within 1 hour.

## 9. Future Enhancements (Post-MVP)

### Character Classes
Players choose a "Fighter" with unique passive:
- **Tank:** Start with 120 HP, deal -20% damage.
- **Speedster:** +1s Speed Window, -10 HP starting health.
- **Berserker:** Deal +30% damage, take +20% self-damage on mistakes.

### Power-ups (Pickups)
Random power-ups appear mid-match:
- **Shield:** Block next incoming attack.
- **Rage Mode:** Double damage for next 2 questions.
- **Heal:** Restore 20 HP.

### Ranked Seasons
- Monthly seasons with tiers (Bronze → Silver → Gold → Diamond).
- Top 100 players get exclusive badges.
- Season rewards: Avatars, titles, coins.

## 10. Success Metrics
- **Engagement:** 40% of daily active users try Duel mode within first week.
- **Retention:** 60% of players who try Duel play 3+ matches.
- **Session Length:** Average 4 matches per session (12-20 minutes).
- **Viral Growth:** 20% of players invite friends for duels (social feature).

---

**Status:** Ready for prototyping. Requires WebSocket infrastructure and matchmaking service.
