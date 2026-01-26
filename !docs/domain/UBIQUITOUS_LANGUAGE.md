# Ubiquitous Language - Quiz Sprint Domain Glossary

**Purpose:** Unified terminology across all game modes for consistent communication between teams.

---

## Game Modes

- **Classic Game:** Solo quiz with score optimization and personal records
- **Daily Marathon:** Survival mode with progressive difficulty, one attempt per day
- **Duel Game:** Real-time 1v1 quiz battle with fighting game mechanics

## Core Gameplay

- **Session:** Single gameplay instance from start to finish (`SessionID`)
- **Question:** Quiz question with multiple answer choices
- **Answer Submission:** Player's timestamped response to a question
- **Stage:** (Daily only) Sequential difficulty level in marathon progression
- **Timeout:** When time limit expires without answer submission

## Streak & Combo

- **Streak:** Chain of consecutive correct answers (Classic, Daily Marathon)
  - Builds momentum, increases Multiplier
  - Reset on incorrect answer or timeout
  - Visual states: Normal → "On Fire" (3+) → "Godlike" (6+)

- **Combo:** Chain of consecutive correct answers (Duel only - combat context)
  - Amplifies damage dealt to opponent
  - Reset on incorrect answer or timeout
  - Levels: x1.0 → x1.3 → x1.6 → x2.0

## Scoring

- **Base Points:** Fixed score value per question (varies by difficulty)
- **Time Bonus:** Additional points for fast answers (decreases linearly)
- **Multiplier:** Score coefficient based on Streak/Combo level
  - Formula (Classic): `FinalScore = (BasePoints + TimeBonus) * Multiplier`
- **Speed Window:** First N seconds when maximum time bonus available
  - Classic: 5-8 seconds (configurable)
  - Duel: 3 seconds (triggers Critical Hit)

## Health Systems

- **Lives:** (Daily Marathon) Discrete health system with 3 lives
  - Lost on incorrect answer or timeout
  - **Permadeath:** Session ends when Lives = 0

- **HP (Health Points):** (Duel) Continuous health (0-100 per player)
  - Reduced by opponent attacks or self-damage
  - Victory when opponent HP = 0 or higher HP after 10 questions

## Power-ups & Bonuses

- **Power-up:** Consumable item providing tactical advantage
  - 50/50: Removes two incorrect options
  - Time Extension: +20 seconds to timer
  - Skip: Skip question without losing a life

- **Bonus:** Automatic reward or score boost
  - Stage Completion Bonus, Survival Bonus, Streak Bonus

## Combat (Duel Only)

- **Damage:** HP reduction dealt to opponent on correct answer (base: 15 HP)
- **Critical Hit:** +50% damage for answering within Speed Window (3s)
- **Special Move:** Auto-triggered powerful attack at Combo milestones
  - Combo 3: Power Strike (40 HP)
  - Combo 5: Fury Combo (50 HP + stun)
  - Combo 7: Ultimate Attack (70 HP)
- **Defense (Block):** Damage reduction when opponent answers faster
  - Faster player: 100% damage
  - Slower player: 50% damage
- **KO (Knockout):** Victory when opponent HP = 0

## Progress & Records

- **Personal Best:** Player's highest score for specific quiz (Classic)
- **Ghost Battle:** Real-time score comparison vs Personal Best
- **Passing Score:** Minimum score required for successful completion
- **Rating:** (Duel) Player's competitive rank for matchmaking

## Session States

- **NotStarted:** Session created, gameplay hasn't begun
- **InProgress:** Player actively answering questions
- **Completed:** All questions answered or win condition met
- **Abandoned:** Player disconnected/quit before completion
- **Failed:** (Daily) Lives = 0, (Duel) HP = 0

---

**Key Distinctions:**
- **Streak** (solo modes) vs **Combo** (Duel fighting context)
- **Lives** (discrete: 3→2→1→0) vs **HP** (continuous: 100→0)
- **Power-up** (player activates) vs **Bonus** (automatic reward)
