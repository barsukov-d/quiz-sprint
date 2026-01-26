# Specification: Leaderboard System
**Context:** Leaderboard
**Status:** Approved

## 1. Business Goal (User Story)
> As a player, I want to see my ranking compared to other players on each quiz, so I can compete and be motivated to improve my scores.

## 2. Value Proposition
Leaderboards transform solo quiz-taking into a **social competitive experience**:
1. **Competition:** Players see where they stand among peers
2. **Motivation:** Desire to reach top positions drives replays
3. **Recognition:** Top 3 podium provides status and achievement
4. **Progress:** Players track improvement through rank changes

## 3. Terminology (Ubiquitous Language)
- **Leaderboard:** Ranked list of all scores for a specific quiz
- **Entry:** A single player's score submission to a leaderboard
- **Rank:** Player's position in the leaderboard (1 = first place)
- **Podium:** Top 3 positions with special recognition (🥇🥈🥉)
- **Score:** Final points achieved in a completed game
- **Tiebreaker:** When scores are equal, earlier completion time ranks higher

## 4. Business Rules and Invariants

### 4.1. Ranking Logic
1. **One Leaderboard per Quiz:** Each quiz maintains its own independent leaderboard
2. **Score-based Ranking:** Entries sorted by score (highest first)
3. **Timestamp Tiebreaker:** Equal scores ranked by completion time (earlier = better)
4. **Immutable Entries:** Once submitted, scores cannot be changed
5. **Multiple Entries Allowed:** Players can submit multiple scores (all appear in leaderboard)

### 4.2. Visibility Rules
1. **Public Leaderboard:** All player scores are visible to everyone
2. **Username Display:** Shows Telegram username for each entry
3. **Current User Highlight:** User's own entries always highlighted
4. **Podium Emphasis:** Top 3 displayed with special styling (medals)
5. **Rank Always Shown:** Every entry displays its current rank position

### 4.3. Performance Rules
1. **Pagination:** Load leaderboard in chunks (50 entries at a time)
2. **Context Window:** Show entries around user's position for relevance
3. **Real-time Updates:** Broadcast rank changes via WebSocket when new scores arrive
4. **Denormalized Data:** Cache usernames in entries (no join queries needed)

## 5. Data Model Changes

### Aggregate: Leaderboard
- **QuizID:** uuid.UUID (which quiz this leaderboard is for)
- **Entries:** []LeaderboardEntry (sorted by score DESC, timestamp ASC)
- **TotalEntries:** int (total number of scores submitted)
- **LastUpdatedAt:** int64 (Unix timestamp of last update)

**Methods:**
- `AddEntry(entry)` - Add new score and recalculate ranks
- `GetTopN(n)` - Retrieve top N entries
- `GetPlayerRank(playerID)` - Find player's position
- `GetEntriesAroundPlayer(playerID, range)` - Get context around player

### Entity: LeaderboardEntry
- **ID:** uuid.UUID
- **QuizID:** uuid.UUID
- **PlayerID:** uuid.UUID
- **PlayerUsername:** string (denormalized for display)
- **Score:** int (final score from completed game)
- **Rank:** int (calculated position, 1-indexed)
- **CompletedAt:** int64 (Unix timestamp)
- **MaxStreak:** int (best streak achieved, for additional context)

**Methods:**
- `UpdateRank(newRank)` - Set calculated position
- `IsPodium()` - Returns true if rank <= 3

### Domain Events
- **LeaderboardEntryCreated:** New score submitted
  - EntryID, QuizID, PlayerID, Score, Rank, CreatedAt
- **LeaderboardUpdated:** Ranks recalculated (batch)
  - QuizID, AffectedPlayerIDs, UpdatedAt
- **TopScoreAchieved:** Player reached #1
  - QuizID, PlayerID, Score, PreviousTopScore, AchievedAt
- **PodiumPositionAchieved:** Player entered top 3
  - QuizID, PlayerID, Rank, Score, AchievedAt

## 6. Scenarios (User Flows)

### Scenario: First Player Completes Quiz
- **Given:** Leaderboard is empty for quiz "World Capitals"
- **When:** Alice completes quiz with score 850
- **Then:**
  - LeaderboardEntry created with Rank = 1
  - Alice sees "🥇 You're #1!" message
  - TopScoreAchieved event published
  - PodiumPositionAchieved event published

### Scenario: New High Score
- **Given:** Current #1 is Bob with 920 points
- **When:** Alice submits 950 points
- **Then:**
  - Alice's entry created with Rank = 1
  - Bob's rank updates to 2
  - Alice sees "🥇 You took #1!" celebration
  - TopScoreAchieved event published
  - LeaderboardUpdated event broadcasts to all viewers

### Scenario: Tied Scores, Timestamp Breaks Tie
- **Given:** Charlie has 800 points (submitted at 10:00 AM)
- **When:** Dave submits 800 points at 10:05 AM
- **Then:**
  - Charlie remains at higher rank (earlier submission)
  - Dave ranked one position below Charlie
  - Both see same score but different ranks

### Scenario: Player Outside Top 50 Views Leaderboard
- **Given:** Eve is ranked #156 out of 500 players
- **When:** Eve opens leaderboard screen
- **Then:**
  - Top 50 entries displayed at top
  - Sticky row at bottom shows: "Your position: #156"
  - Eve can tap sticky row to scroll to her context (entries #146-166)

### Scenario: Real-time Rank Drop
- **Given:** Frank is viewing leaderboard, currently #7
- **When:** Two new players submit higher scores
- **Then:**
  - WebSocket pushes LeaderboardUpdated event
  - Frank's row animates from #7 to #9
  - Smooth transition animation shows rank change

### Scenario: Multiple Entries from Same Player
- **Given:** Alice previously scored 800 (Rank #15)
- **When:** Alice plays again and scores 950 (Rank #1)
- **Then:**
  - Both entries exist in leaderboard
  - Entry at #1 shows Alice with 950
  - Entry at #15 (now #16 after rank shift) shows Alice with 800
  - Leaderboard highlights both of Alice's entries

## 7. Integration Points

### From Classic Mode
- **Event:** ClassicGameFinished
- **Action:** Create LeaderboardEntry with final score

### From Daily Mode
- **Event:** DailyQuizCompleted
- **Action:** Create LeaderboardEntry with bonus-applied score

### To Identity Context
- **Query:** Fetch username and avatar for display
- **Cached:** Username stored in LeaderboardEntry for performance

### To Notification Context (Future)
- **Event:** PodiumPositionAchieved
- **Action:** Send push notification "You reached top 3!"

## 8. Edge Cases

### Empty Leaderboard
- **Display:** "No rankings yet. Be the first to complete this quiz!"
- **No errors:** System handles gracefully

### Single Player
- **Display:** Player always #1 until someone else plays
- **Podium:** Still shows 🥇 even if only entry

### Mass Score Submission
- **Scenario:** 100 players complete quiz simultaneously
- **Handling:** Queue score submissions, process sequentially
- **Real-time:** Batch rank updates, broadcast once every 2 seconds

### Player Deleted/Banned
- **Entry remains:** Scores stay in leaderboard (historical data)
- **Username:** Shows "[Deleted User]" if player removed
- **Note:** Discuss retention policy separately
