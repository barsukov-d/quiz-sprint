# Specification: Classic Game Core Gameplay Mechanics
**Context:** ClassicGame
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to play solo quizzes, turning knowledge testing into an exciting pursuit of records, where speed and accuracy are rewarded with visual and emotional triumph.

## 2. Value Proposition
Classic Game is not just a test, it's a **"Personal Challenge"**. Its appeal is built on three pillars:
1. **Flow:** Quick sessions, instant feedback, and increasing tempo.
2. **Mastery:** Ability to beat your previous result ("Ghost Battle").
3. **Juice:** Visual confirmation of success through combo bonuses and dynamic effects.

## 3. Terminology (Ubiquitous Language)
> See [UBIQUITOUS_LANGUAGE.md](../UBIQUITOUS_LANGUAGE.md) for full domain glossary.

- **Classic Game:** Solo quiz mode focused on score optimization and personal records.
- **Streak:** Chain of consecutive correct answers that builds momentum and increases Multiplier.
- **Multiplier:** Score coefficient that amplifies rewards based on Streak level (x1.0 → x1.5 → x2.0).
- **Speed Window:** First 5-8 seconds of a question when maximum Time Bonus is available.
- **Personal Best:** Player's highest score for a specific quiz (used for Ghost Battle).
- **Ghost Battle:** Real-time comparison of current score vs Personal Best at same question point.

## 4. Business Rules and Invariants

### 4.1. Game Juice Mechanics
1. **Building Momentum:** When Streak >= 3, UI transitions to "Heat" state. When Streak >= 6, it enters "Fire" state. This is accompanied by animation and sound changes.
2. **Streak Reset:** Any error or timeout immediately resets Multiplier and visual effects, creating emotional risk.

### 4.2. Scoring System
Final score for a question is calculated by the formula:
`FinalScore = (BasePoints + TimeBonus) * Multiplier`

1. **Base Points:** Fixed value per question (varies by difficulty).
2. **Time Bonus:** Decreases linearly from maximum to zero over the question time limit.
3. **Multiplier (Streak-based):**
   - 0-2 Streak: x1.0 (Normal)
   - 3-5 Streak: x1.5 ("On Fire" 🔥)
   - 6+ Streak: x2.0 ("Godlike" 🔥🔥)

### 4.3. Lifecycle and Progress
1. **Battle with Record:** If user has a `PersonalBest` for this quiz, during gameplay a "Ghost" indicator is displayed (comparison of current score with best result at the same quiz point).
2. **Completion:** Game is considered successful if `PassingScore` is reached. Only successful games update `PersonalBest`.

## 5. Data Model Changes
- **Aggregate `ClassicGame`** (extends base Session):
  - `CurrentStreak` (int): Current consecutive correct answers.
  - `MaxStreak` (int): Highest Streak achieved in current session.
  - `CurrentMultiplier` (float): Current score multiplier based on Streak (1.0, 1.5, or 2.0).
  - `GhostComparison` (int): Score difference vs Personal Best at same question index.

## 6. Scenarios (User Flows)

### Scenario: Entering "Flow" State
- **Given:** Player has Streak = 2.
- **When:** Player gives correct answer to 3rd question within Speed Window.
- **Then:**
  - Streak becomes 3.
  - Multiplier jumps to x1.5.
  - UI activates "On Fire" effects.
  - Player receives significantly more points than for previous question.

### Scenario: Ghost Battle (Beating Personal Best)
- **Given:** Player has Personal Best of 5000 points for this quiz.
- **When:** At Question 5, player has 2800 points (Personal Best was 2500 at this point).
- **Then:** Ghost indicator shows `+300` in green, showing player is ahead of their record pace.

### Scenario: Dramatic Error
- **Given:** Player is in "Godlike" state (Streak = 7, Multiplier = x2.0).
- **When:** Player selects incorrect answer.
- **Then:**
  - Screen briefly shakes (Shake effect).
  - Multiplier drops to x1.0.
  - Fire effects disappear.
  - Player sees correct answer for learning.
