# Specification: Classic Game Core Gameplay Mechanics
**Context:** ClassicGame
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to play solo quizzes, turning knowledge testing into an exciting pursuit of records, where speed and accuracy are rewarded with visual and emotional triumph.

## 2. Ubiquitous Language
- **Classic Game:** Solo quiz mode focused on score optimization and personal records.
- **Streak:** Chain of consecutive correct answers that builds momentum and increases Multiplier.
- **Multiplier:** Score coefficient that amplifies rewards based on Streak level (x1.0 → x1.5 → x2.0).
- **Speed Window:** First 5-8 seconds of a question when maximum Time Bonus is available.
- **Personal Best:** Player's highest score for a specific quiz (used for Ghost Battle).
- **Ghost Battle:** Real-time comparison of current score vs Personal Best at same question point.

## 3. Business Rules & Logic
1. **Scoring Formula:** `FinalScore = (BasePoints + TimeBonus) * Multiplier`.
2. **Multiplier Levels:**
   - 0-2 Streak: x1.0
   - 3-5 Streak: x1.5
   - 6+ Streak: x2.0
3. **Streak Reset:** Any incorrect answer or timeout immediately resets Multiplier to x1.0 and Streak to 0.
4. **Time Bonus:** Decreases linearly from maximum to zero over the question time limit.
5. **Ghost Comparison:** If `PersonalBest` exists, calculate `GhostComparison` = `CurrentScore` - `BestScoreAtThisPoint`.
6. **Record Update:** Only successful games (reaching `PassingScore`) can update `PersonalBest`.

## 4. Manifest Updates (Intent)
- **New Fields (Aggregate ClassicGame):**
    - `CurrentStreak`: int (Current consecutive correct answers)
    - `MaxStreak`: int (Highest Streak achieved in current session)
    - `CurrentMultiplier`: float64 (Current score multiplier: 1.0, 1.5, or 2.0)
    - `GhostComparison`: int (Score difference vs Personal Best at same question index)
- **New Methods:**
    - `SubmitAnswer`: Handles scoring logic, multiplier updates, and streak management.
- **New Events:**
    - `ClassicGameFinished`: Published with final score and stats.

## 5. Scenarios (User Flows)
- **Scenario: Entering Flow State**
    - **Given:** Player has Streak = 2.
    - **When:** Player gives correct answer to 3rd question within Speed Window.
    - **Then:** Streak becomes 3, Multiplier jumps to x1.5, and UI activates "On Fire" effects.

- **Scenario: Ghost Battle (Beating Personal Best)**
    - **Given:** Player has Personal Best of 5000 points for this quiz.
    - **When:** At Question 5, player has 2800 points (Personal Best was 2500 at this point).
    - **Then:** Ghost indicator shows `+300`, indicating player is ahead of their record pace.

- **Scenario: Dramatic Error**
    - **Given:** Player is in "Godlike" state (Streak = 7, Multiplier = x2.0).
    - **When:** Player selects incorrect answer.
    - **Then:** Multiplier drops to x1.0, Streak resets to 0, and fire effects disappear.
