# Specification: Daily Quiz (Daily Mode)
**Context:** ClassicMode / QuizCatalog
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to complete a special quiz every day to maintain a visit streak and earn bonus points.

## 2. Key Mechanics (Essentials)

### 2.1. Daily Streak
- **Daily Streak**: Counter of consecutive days when user completed the daily quiz.
- **Reset**: If user missed a calendar day (UTC), the streak resets to 0.

### 2.2. Daily Bonus
- For the first completion of the daily quiz in the current day, a fixed multiplier **x1.5** is applied to the final score.

### 2.3. Shared Content
- All users receive the same quiz within a calendar day (UTC).

## 3. Data Model
- **DailyStreak** (int): Current streak of days.
- **LastDailyCompletedAt** (Timestamp): Time of last daily quiz completion.

## 4. Main Scenario
1. User enters Daily Quiz.
2. After completion, system checks: "Has less than 48 hours passed since last time AND has a new calendar day started?".
3. If yes — `DailyStreak` increases by 1.
4. If more than 48 hours passed — `DailyStreak` becomes 1.
5. Multiplier x1.5 is applied to final score.
