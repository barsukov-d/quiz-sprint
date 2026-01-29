# Daily Challenge - API Specification

## Architecture Note: Thin Client Pattern

**Backend owns ALL game state and logic. Frontend is pure rendering layer.**

API responses include:
- ✅ All data needed for UI rendering
- ✅ Localized labels/text where needed
- ✅ Pre-calculated values (scores, bonuses, etc.)
- ✅ UI hints (canRetry, hasAd, etc.)

Frontend does NOT:
- ❌ Calculate scores locally
- ❌ Validate answers client-side
- ❌ Determine chest types
- ❌ Track game state beyond API responses

---

## Authentication
All endpoints require Telegram auth:
```
Authorization: tma <base64_encoded_init_data>
```

## Endpoints

### 1. Start Daily Challenge

```http
POST /api/v1/daily/start
Content-Type: application/json
```

**Request:**
```json
{
  "date": "2026-01-28"  // Optional, defaults to today UTC
}
```

**Response 201 Created:**
```json
{
  "data": {
    "gameId": "dg_abc123xyz",
    "dailyQuizId": "dq_20260128",
    "date": "2026-01-28",
    "questions": [
      {
        "id": "q_001",
        "text": "В каком году основали Москву?",
        "answers": [
          {"id": "a_001", "text": "1147 год", "position": 0},
          {"id": "a_002", "text": "1240 год", "position": 1},
          {"id": "a_003", "text": "988 год", "position": 2},
          {"id": "a_004", "text": "1380 год", "position": 3}
        ],
        "timeLimit": 15,
        "difficulty": "medium"
      }
      // ... 9 more questions
    ],
    "currentStreak": 5,
    "streakBonus": 1.25,
    "startedAt": 1706428800
  }
}
```

**Errors:**
- `409 Conflict` - Already played today
- `404 Not Found` - No daily quiz available (system error)

---

### 2. Submit Answer

```http
POST /api/v1/daily/:gameId/answer
Content-Type: application/json
```

**Request:**
```json
{
  "questionId": "q_001",
  "answerId": "a_001",
  "timeTaken": 8.5
}
```

**Response 200 OK:**
```json
{
  "data": {
    "questionIndex": 0,
    "remainingQuestions": 9,
    "isGameCompleted": false
  }
}
```

**On Last Question (auto-completion):**
```json
{
  "data": {
    "questionIndex": 9,
    "remainingQuestions": 0,
    "isGameCompleted": true,
    "results": {
      "finalScore": 920,
      "baseScore": 800,
      "streakBonus": 1.25,
      "correctAnswers": 8,
      "totalQuestions": 10,
      "chestType": "golden",
      "chestIcon": "🏆",
      "chestLabel": "Золотой сундук",
      "rank": 847,
      "totalPlayers": 12847,
      "rankLabel": "#847 из 12,847 игроков",
      "canShare": true,
      "shareText": "Я занял #847 место в Daily Challenge!"
    }
  }
}
```

**Note:** Backend provides ALL display data. Frontend just renders.

**Errors:**
- `400 Bad Request` - Invalid time, already answered
- `404 Not Found` - Game or question not found
- `409 Conflict` - Game already completed

---

### 3. Get Daily Status

```http
GET /api/v1/daily/status?playerId=user_123&date=2026-01-28
```

**Response 200 OK (Not played):**
```json
{
  "data": {
    "date": "2026-01-28",
    "hasPlayed": false,
    "currentStreak": 5,
    "nextStreakMilestone": 7,
    "canPlayNow": true
  }
}
```

**Response 200 OK (Already played):**
```json
{
  "data": {
    "date": "2026-01-28",
    "hasPlayed": true,
    "gameId": "dg_abc123",
    "results": {
      "finalScore": 920,
      "correctAnswers": 8,
      "rank": 847,
      "rankLabel": "#847",
      "chestType": "golden",
      "chestIcon": "🏆",
      "chestLabel": "Золотой сундук"
    },
    "currentStreak": 6,
    "streakLabel": "🔥 6 дней подряд",
    "nextMilestone": {
      "days": 7,
      "bonus": "+25%",
      "daysUntil": 1
    },
    "canRetry": true,
    "retryLabel": "Попробовать ещё раз",
    "retryCost": {
      "coins": 100,
      "coinsLabel": "100 💰",
      "hasAd": true,
      "adLabel": "Смотреть рекламу"
    }
  }
}
```

**Note:** Backend includes UI labels to simplify frontend rendering.

---

### 4. Get Leaderboard

```http
GET /api/v1/daily/leaderboard?date=2026-01-28&limit=10&playerId=user_123
```

**Response 200 OK:**
```json
{
  "data": {
    "date": "2026-01-28",
    "totalPlayers": 12847,
    "entries": [
      {
        "rank": 1,
        "playerId": "user_456",
        "username": "ProGamer",
        "score": 1750,
        "correctAnswers": 10,
        "completedAt": 1706428920
      }
      // ... top 10
    ],
    "playerRank": {
      "rank": 847,
      "playerId": "user_123",
      "username": "You",
      "score": 920,
      "correctAnswers": 8,
      "completedAt": 1706429800
    }
  }
}
```

**Query Params:**
- `date` - YYYY-MM-DD (default: today)
- `limit` - 1-100 (default: 10)
- `playerId` - Include player's rank (optional)
- `type` - `global`/`friends`/`country` (default: global)

---

### 5. Get Player Streak

```http
GET /api/v1/daily/streak?playerId=user_123
```

**Response 200 OK:**
```json
{
  "data": {
    "currentStreak": 5,
    "lastPlayedDate": "2026-01-28",
    "longestStreak": 14,
    "streakBonus": 1.25,
    "nextMilestone": {
      "days": 7,
      "bonus": 1.25
    }
  }
}
```

---

### 6. Open Chest (Get Rewards)

```http
POST /api/v1/daily/:gameId/chest/open
```

**Response 200 OK:**
```json
{
  "data": {
    "chestType": "golden",
    "rewards": {
      "coins": 420,
      "pvpTickets": 5,
      "bonuses": [
        {"type": "shield", "quantity": 1},
        {"type": "freeze", "quantity": 1}
      ]
    },
    "streakBonus": 1.25,
    "premiumApplied": false
  }
}
```

**Notes:**
- Idempotent (multiple calls return same rewards)
- Rewards already in DB (chest opening = UI only)

---

### 7. Retry Challenge (Second Attempt)

```http
POST /api/v1/daily/:gameId/retry
Content-Type: application/json
```

**Request:**
```json
{
  "paymentMethod": "coins"  // or "ad"
}
```

**Response 201 Created:**
```json
{
  "data": {
    "newGameId": "dg_xyz789",
    "coinsDeducted": 100,
    "remainingCoins": 450
  }
}
```

**Errors:**
- `400 Bad Request` - Insufficient coins
- `409 Conflict` - Retry limit reached (non-premium)

---

## Domain Events (Server-side)

Published to event bus on completion:

```go
type DailyGameStartedEvent struct {
    GameID        string
    PlayerID      string
    DailyQuizID   string
    Date          string
    CurrentStreak int
    Timestamp     int64
}

type DailyGameCompletedEvent struct {
    GameID         string
    PlayerID       string
    FinalScore     int
    CorrectAnswers int
    ChestType      string
    Streak         int
    Rank           *int
    Timestamp      int64
}

type ChestOpenedEvent struct {
    PlayerID   string
    GameID     string
    ChestType  string
    Rewards    ChestContents
    Timestamp  int64
}

type StreakMilestoneReachedEvent struct {
    GameID       string
    PlayerID     string
    Streak       int
    BonusPercent int
    Timestamp    int64
}
```

## Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_TIME_TAKEN` | Time not in 0-15 range |
| 400 | `GAME_NOT_ACTIVE` | Game not in progress |
| 400 | `INSUFFICIENT_COINS` | Not enough coins for retry |
| 404 | `GAME_NOT_FOUND` | Game doesn't exist |
| 404 | `QUIZ_NOT_FOUND` | Daily quiz missing |
| 409 | `ALREADY_PLAYED_TODAY` | Free attempt used |
| 409 | `GAME_COMPLETED` | Cannot answer completed game |
