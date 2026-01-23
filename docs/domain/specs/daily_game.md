# Specification: Daily Marathon - Survival Challenge
**Context:** DailyMarathon
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to test my endurance in a daily survival challenge with progressive difficulty, where every mistake costs me dearly and careful strategy determines how far I can go.

## 2. Ubiquitous Language
- **Daily Marathon:** Time-limited survival mode with progressive stages, playable once per 24 hours.
- **Lives:** Health system represented as 3 discrete chances (❤️ ❤️ ❤️).
- **Permadeath:** Immediate session end when Lives reach zero.
- **Stage:** Sequential difficulty level (Easy → Medium → Hard → Elite).
- **Power-up:** Consumable item that provides tactical advantage (50/50, Time Extension, Skip).

## 3. Business Rules & Logic
1. **Starting State:** Every player begins with exactly 3 lives.
2. **Life Loss:** Incorrect answers or timeouts reduce Lives by 1.
3. **Permadeath:** Session ends immediately when Lives reach 0. No revivals allowed.
4. **Daily Limit:** Only one attempt per 24-hour period (resets at Midnight UTC).
5. **Advancement:** Must complete current stage with Lives ≥ 1 to unlock the next stage.
6. **Earning Power-ups:** Awarded for stage completion (random) or achieving a 5-question Streak.
7. **Power-up Usage:** Consumable items (50/50, Time Extension, Skip) are session-specific and do not carry over.
8. **Scoring Formula:** `FinalScore = Sum(QuestionScores) + StageBonuses`. 
    - **Stage Completion Bonus:** `500 * StageNumber`.
    - **Survival Bonus:** `200 * RemainingLives` at stage end.

## 4. Manifest Updates (Intent)
- **New Fields (Aggregate DailyMarathon):**
    - `SessionID`: uuid.UUID
    - `CurrentStage`: int (1-based stage index)
    - `Lives`: int (0-3)
    - `Streak`: int (Current correct answer chain)
    - `Inventory`: []PowerUp (List of earned power-ups)
    - `TotalScore`: int (Cumulative score)
    - `LastAttemptTimestamp`: int64 (Unix timestamp)
    - `Status`: string (InProgress, Completed, Failed)
- **New Value Object (PowerUp):**
    - `Type`: string (FiftyFifty, TimeExtension, Skip)
    - `Quantity`: int
- **New Methods:**
    - `UsePowerUp`: Validates and applies power-up effects.
    - `CompleteStage`: Calculates bonuses and transitions to next stage.

## 5. Scenarios (User Flows)
- **Scenario: Perfect Stage Clear**
    - **Given:** Player is on Stage 1 with 3 lives, has answered 4/5 questions correctly.
    - **When:** Player answers final question correctly within Speed Window.
    - **Then:** Streak reaches 5, player earns a 50/50 power-up, receives Stage Completion Bonus, and proceeds to Stage 2.

- **Scenario: Strategic Power-up Use**
    - **Given:** Player is on Stage 3, has 1 life remaining and a Skip power-up.
    - **When:** Player faces an extremely difficult question and uses Skip.
    - **Then:** Player moves to the next question without losing a life and without earning points for the skipped question.

- **Scenario: Dramatic Permadeath**
    - **Given:** Player is on Stage 4 with 1 life left.
    - **When:** Player's timer expires on a question.
    - **Then:** Lives drop to 0, the session ends immediately, and the "Game Over" screen is displayed.