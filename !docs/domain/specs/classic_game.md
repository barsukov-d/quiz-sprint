# Specification: Classic Game Core Gameplay Mechanics
**Context:** ClassicGame
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to play solo quizzes, turning knowledge testing into an exciting score pursuit, where speed and accuracy are rewarded with visual and emotional triumph.

## 2. Ubiquitous Language
- **Classic Game:** Solo quiz mode focused on score maximization through streak combinations and answer speed.
- **Streak:** Chain of consecutive correct answers that builds momentum and increases Multiplier.
- **Multiplier:** Score coefficient that amplifies rewards based on Streak level (x1.0 → x1.5 → x2.0).
- **Speed Window:** First 5-8 seconds of a question when maximum Time Bonus is available.

## 3. Business Rules & Logic
1. **Scoring Formula:** `FinalScore = (BasePoints + TimeBonus) * Multiplier`.
2. **Multiplier Levels:**
   - Streak 0-2: x1.0
   - Streak 3-5: x1.5
   - Streak 6+: x2.0
3. **Streak Reset:** Any incorrect answer or timeout immediately resets Multiplier to x1.0 and Streak to 0.
4. **Time Bonus:** Decreases linearly from maximum to zero over the question time limit.

## 4. Scenarios (User Flows)
- **Scenario: Entering Flow State**
    - **Given:** Player has Streak = 2.
    - **When:** Player gives correct answer to 3rd question within Speed Window.
    - **Then:** Streak becomes 3, Multiplier increases to x1.5, and UI activates "On Fire" effects.

- **Scenario: Dramatic Failure**
    - **Given:** Player is in "Godlike" state (Streak = 7, Multiplier = x2.0).
    - **When:** Player selects incorrect answer.
    - **Then:** Multiplier drops to x1.0, Streak resets to 0, and fire effects disappear.
