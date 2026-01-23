# Specification: Classic Mode Core Gameplay Mechanics
**Context:** ClassicMode
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to play solo quizzes, turning knowledge testing into an exciting pursuit of records, where speed and accuracy are rewarded with visual and emotional triumph.

## 2. Value Proposition
Classic Mode is not just a test, it's a **"Personal Challenge"**. Its appeal is built on three pillars:
1. **Flow:** Quick sessions, instant feedback, and increasing tempo.
2. **Mastery:** Ability to beat your previous result ("Ghost Battle").
3. **Juice:** Visual confirmation of success through combo bonuses and dynamic effects.

## 3. Terminology (Ubiquitous Language)
- **Classic Game:** Main game loop focused on a single player.
- **Streak:** Chain of correct answers that turns a regular game into a "hot streak".
- **Multiplier:** Excitement coefficient that increases rewards for risk and speed.
- **Speed Window:** First seconds of a question when maximum bonus can be obtained.

## 4. Business Rules and Invariants

### 4.1. Game Juice Mechanics
1. **Building Momentum:** When Streak >= 3, UI transitions to "Heat" state. When Streak >= 6, it enters "Fire" state. This is accompanied by animation and sound changes.
2. **Streak Reset:** Any error or timeout immediately resets Multiplier and visual effects, creating emotional risk.

### 4.2. Scoring System (Scoring 2.0)
Final score for a question is calculated by the formula:
`Score = (BasePoints + TimeBonus) * Multiplier`

1. **BasePoints:** Fixed value per question.
2. **TimeBonus:** Linearly decreases from maximum to zero over the question time limit.
3. **Multiplier (Streak-based):**
   - 0-2 correct: x1.0
   - 3-5 correct: x1.5 ("On Fire")
   - 6+ correct: x2.0 ("Godlike")

### 4.3. Lifecycle and Progress
1. **Battle with Record:** If user has a `PersonalBest` for this quiz, during gameplay a "Ghost" indicator is displayed (comparison of current score with best result at the same quiz point).
2. **Completion:** Game is considered successful if `PassingScore` is reached. Only successful games update `PersonalBest`.

## 5. Data Model Changes
- **Aggregate `ClassicGame`**:
  - `CurrentStreak` (int): Current streak counter.
  - `MaxStreak` (int): Best streak for current session.
  - `CurrentMultiplier` (float): Current multiplier based on streak.
  - `GhostComparison` (int): Score difference relative to Personal Best.

## 6. Scenarios (User Flows)

### Scenario: Entering "Flow" State
- **Given:** Player has Streak = 2.
- **When:** Player gives correct answer to 3rd question within Speed Window.
- **Then:**
  - Streak becomes 3.
  - Multiplier jumps to x1.5.
  - UI activates "On Fire" effects.
  - Player receives significantly more points than for previous question.

### Scenario: Battle with Yourself (Ghost Run)
- **Given:** Player has a record of 5000 points.
- **When:** On 5th question player has 2800 points (record was 2500 at this point).
- **Then:** "Ghost" indicator shows `+300` and highlights green, motivating player to maintain pace.

### Scenario: Dramatic Error
- **Given:** Player is in "Godlike" state (Streak = 7, Multiplier = x2.0).
- **When:** Player selects incorrect answer.
- **Then:**
  - Screen briefly shakes (Shake effect).
  - Multiplier drops to x1.0.
  - Fire effects disappear.
  - Player sees correct answer for learning.
