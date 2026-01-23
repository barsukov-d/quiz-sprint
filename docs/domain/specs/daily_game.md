# Daily Survival Marathon

The Daily Survival Marathon is a special competitive mode available once per day. It combines the challenge of survival with increasing difficulty across multiple quiz stages.

## Core Mechanics

### 1. The Survival Rule
- **Lives**: Every player starts with **3 lives** (❤️ ❤️ ❤️).
- **Losing Lives**: A life is lost for every incorrect answer or when the question timer expires.
- **Permadeath**: The session ends immediately when lives reach zero.
- **Daily Limit**: Players can attempt the Marathon only once per day.

### 2. Progressive Difficulty
The marathon consists of sequential stages (quizzes). To advance, a player must complete the current stage with at least one life remaining.
- **Stage 1 (Easy)**: 5 questions, 30s timer, basic topics.
- **Stage 2 (Medium)**: 7 questions, 20s timer, intermediate topics.
- **Stage 3 (Hard)**: 10 questions, 15s timer, advanced topics.
- **Stage 4+ (Elite)**: Increasing question count and even shorter timers.

### 3. Power-ups (Bonuses)
Players earn bonuses to help them survive longer.
- **Earning**: 
  - One random bonus is awarded upon successful completion of each stage.
  - A "Combo Bonus" is awarded for 5 consecutive correct answers.
- **Types**:
  - **50/50**: Removes two incorrect answer options.
  - **Time Extension**: Adds +20 seconds to the current question timer.
  - **Skip**: Skip the current question without losing a life (no points awarded).

### 4. Scoring
- Points are awarded for correct answers.
- **Speed Bonus**: Extra points for fast responses.
- **Stage Completion Bonus**: Large point boost for finishing a whole quiz stage.
- **Survival Bonus**: Points awarded for each remaining life at the end of a stage.

## User Flow
1. **Entry**: User opens the "Daily Marathon" from the main menu.
2. **Start**: User sees the "Stage 1" intro and their 3 lives.
3. **Gameplay**: User answers questions, using power-ups if they have them.
4. **Transition**: After Stage 1, a "Stage Clear" screen shows the reward (e.g., "+1 50/50") and the "Start Stage 2" button.
5. **Game Over**: When lives hit 0, a summary screen shows total points, stages cleared, and their rank for the day.

## Technical Requirements
- **State Persistence**: The server must track `lives`, `current_stage`, and `inventory` (bonuses) within the session.
- **Daily Reset**: A new marathon sequence is generated every 24 hours.
- **Fair Play**: Power-ups are session-specific and do not carry over to the next day.
