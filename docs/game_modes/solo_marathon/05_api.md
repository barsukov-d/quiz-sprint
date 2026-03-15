# Solo Marathon - API Specification

> **Статус реализации (аудит 2026-03-15)**
> ✅ Реализовано: 2 | ⚠️ Расходится: 7 | ❌ Не реализовано: 1
>
> - ✅ Domain events (все существуют + дополнительные)
> - ✅ Error codes (все маппируются в HTTP)
> - ⚠️ POST /marathon/start — endpoint есть, но `selectedBonuses` в request не используется (бонусы грузятся автоматически); response не содержит `milestone`, `onboarding`, `startedAt`
> - ⚠️ POST /marathon/:gameId/bonus — endpoint есть, но нет `statusMessage`; 50/50 возвращает `hiddenAnswerIds` вместо `remainingAnswers`; skip возвращает `nextQuestion` ✅
> - ⚠️ POST /marathon/:gameId/answer — endpoint есть, но `shieldActive` не принимается в request; response не содержит `feedbackMessage`, `correctAnswerText`, `explanation`; `gameOverData` не содержит `weeklyRank`
> - ⚠️ POST /marathon/:gameId/continue — endpoint есть, но `remainingCoins` отсутствует в response
> - ⚠️ GET /marathon/status — endpoint есть, но не содержит `weeklyBest`, `weeklyRank`, `canStart`
> - ⚠️ GET /marathon/leaderboard — endpoint есть, но нет `weekId`, `endsAt`, `totalPlayers`; `playerRank` всегда nil
> - ⚠️ Errors — маппируются как plain text, не как структурированный `{error: {code, message}}`
> - ❌ POST /marathon/:gameId/complete — не реализован; вместо него код использует DELETE /:gameId (abandon)

## Architecture Note: Thin Client Pattern

Backend owns ALL game state. Frontend renders server responses.

API returns:
- Complete game state (lives, score, bonuses)
- UI-ready labels and hints
- Next action options

---

## Authentication
```
Authorization: tma <base64_encoded_init_data>
```

---

## Endpoints

### 1. Start Marathon

> ⚠️ Endpoint существует. `selectedBonuses` в request игнорируется (бонусы грузятся из инвентаря автоматически). Response не содержит `milestone`, `onboarding`, `startedAt`.

```http
POST /api/v1/marathon/start
```

**Request:**
```json
{
  "selectedBonuses": {
    "shield": 2,
    "fiftyFifty": 1,
    "skip": 0,
    "freeze": 3
  }
}
```

**Response 201:**
```json
{
  "data": {
    "gameId": "mg_abc123",
    "status": "in_progress",
    "lives": 3,
    "livesLabel": "❤️❤️❤️",
    "score": 0,
    "scoreLabel": "✅ 0",
    "personalBest": 87,
    "currentQuestion": {
      "questionNumber": 1,
      "questionId": "q_001",
      "text": "В каком году основали Москву?",
      "answers": [
        {"id": "a_001", "text": "1147 год", "position": 0},
        {"id": "a_002", "text": "1240 год", "position": 1},
        {"id": "a_003", "text": "988 год", "position": 2},
        {"id": "a_004", "text": "1380 год", "position": 3}
      ],
      "timeLimit": 15,
      "difficulty": "easy",
      "difficultyChanged": false,
      "difficultyMessage": null
    },
    "milestone": {
      "next": 25,
      "current": 0,
      "remaining": 25,
      "label": "Следующая цель: 25 ✅ (ещё 25)"
    },
    "onboarding": null,
    "bonusInventory": {
      "shield": 2,
      "fiftyFifty": 1,
      "skip": 0,
      "freeze": 3
    },
    "startedAt": 1706428800
  }
}
```

---

### 2. Use Bonus

> ⚠️ Endpoint существует. Нет `statusMessage` в response. 50/50 возвращает `hiddenAnswerIds` вместо `remainingAnswers`. Skip возвращает `nextQuestion` ✅.

```http
POST /api/v1/marathon/:gameId/bonus
```

**Request (Shield):**
```json
{
  "bonusType": "shield",
  "questionId": "q_001"
}
```

**Response 200:**
```json
{
  "data": {
    "bonusType": "shield",
    "bonusActive": true,
    "statusMessage": "🛡️ Щит активирован",
    "bonusInventory": {
      "shield": 1,
      "fiftyFifty": 1,
      "skip": 0,
      "freeze": 3
    }
  }
}
```

**Request (50/50):**
```json
{
  "bonusType": "fifty_fifty",
  "questionId": "q_001"
}
```

**Response 200:**
```json
{
  "data": {
    "bonusType": "fifty_fifty",
    "remainingAnswers": ["a_001", "a_003"],
    "statusMessage": "🔀 50/50 использован",
    "bonusInventory": {
      "shield": 2,
      "fiftyFifty": 0,
      "skip": 0,
      "freeze": 3
    }
  }
}
```

**Request (Skip):**
```json
{
  "bonusType": "skip",
  "questionId": "q_001"
}
```

**Response 200:**
```json
{
  "data": {
    "bonusType": "skip",
    "nextQuestion": { /* question data */ },
    "statusMessage": "⏭️ Вопрос пропущен",
    "bonusInventory": { /* updated */ }
  }
}
```

---

### 3. Submit Answer

> ⚠️ Endpoint существует. `shieldActive` не принимается в request. Response не содержит `feedbackMessage`, `correctAnswerText`, `explanation`. `gameOverData` не содержит `weeklyRank`.

```http
POST /api/v1/marathon/:gameId/answer
```

**Request:**
```json
{
  "questionId": "q_001",
  "answerId": "a_001",
  "timeTaken": 8.5,
  "shieldActive": false
}
```

**Response 200 (Correct):**
```json
{
  "data": {
    "isCorrect": true,
    "correctAnswerId": "a_001",
    "feedbackMessage": "✅ Правильно!",
    "lives": 3,
    "livesLabel": "❤️❤️❤️",
    "score": 1,
    "scoreLabel": "✅ 1",
    "isGameOver": false,
    "nextQuestion": { /* next question data */ }
  }
}
```

**Response 200 (Wrong - with lives):**
```json
{
  "data": {
    "isCorrect": false,
    "correctAnswerId": "a_001",
    "correctAnswerText": "1147 год",
    "feedbackMessage": "❌ Неправильно",
    "explanation": "Москва основана в 1147 году Юрием Долгоруким.",
    "lives": 2,
    "livesLabel": "❤️❤️🖤",
    "livesLost": 1,
    "score": 1,
    "scoreLabel": "✅ 1",
    "isGameOver": false,
    "nextQuestion": { /* next question */ }
  }
}
```

**Response 200 (Wrong - game over):**
```json
{
  "data": {
    "isCorrect": false,
    "correctAnswerId": "a_001",
    "feedbackMessage": "❌ Неправильно",
    "lives": 0,
    "livesLabel": "🖤🖤🖤",
    "score": 23,
    "scoreLabel": "✅ 23",
    "isGameOver": true,
    "gameOverData": {
      "finalScore": 23,
      "totalQuestions": 24,
      "personalBest": 87,
      "isNewRecord": false,
      "weeklyRank": 342,
      "continueOffer": {
        "available": true,
        "costCoins": 200,
        "hasAd": true,
        "continueCount": 0,
        "message": "Хочешь продолжить? Получи ещё одну жизнь!"
      }
    }
  }
}
```

---

### 4. Continue Game

> ⚠️ Endpoint существует. `remainingCoins` отсутствует в response.

```http
POST /api/v1/marathon/:gameId/continue
```

**Request:**
```json
{
  "paymentMethod": "coins"  // or "ad"
}
```

**Response 200:**
```json
{
  "data": {
    "success": true,
    "lives": 1,
    "livesLabel": "❤️",
    "continueCount": 1,
    "coinsDeducted": 200,
    "remainingCoins": 1250,
    "status": "in_progress",
    "nextContinueCost": 400,
    "currentQuestion": { /* same question */ }
  }
}
```

**Error 400 (Insufficient coins):**
```json
{
  "error": {
    "code": "INSUFFICIENT_COINS",
    "message": "Недостаточно монет",
    "required": 200,
    "current": 50,
    "action": {
      "type": "navigate",
      "route": "/shop"
    }
  }
}
```

---

### 5. Get Marathon Status

> ⚠️ Endpoint существует. Response не содержит `weeklyBest`, `weeklyRank`, `canStart`.

```http
GET /api/v1/marathon/status?playerId=user_123
```

**Response 200 (No active game):**
```json
{
  "data": {
    "hasActiveGame": false,
    "personalBest": 87,
    "weeklyBest": 47,
    "weeklyRank": 342,
    "weeklyRankLabel": "#342 из 5,847",
    "bonusInventory": {
      "shield": 2,
      "fiftyFifty": 1,
      "skip": 0,
      "freeze": 3
    },
    "canStart": true
  }
}
```

**Response 200 (Active game):**
```json
{
  "data": {
    "hasActiveGame": true,
    "gameId": "mg_abc123",
    "currentScore": 23,
    "lives": 2,
    "canResume": true
  }
}
```

---

### 6. Get Leaderboard

> ⚠️ Endpoint существует. Response не содержит `weekId`, `endsAt`, `totalPlayers`. `playerRank` всегда nil.

```http
GET /api/v1/marathon/leaderboard?type=weekly&limit=10&playerId=user_123
```

**Query Params:**
- `type`: `weekly` | `alltime` | `friends`
- `limit`: 1-100 (default 10)
- `playerId`: Include player's rank (optional)

**Response 200:**
```json
{
  "data": {
    "type": "weekly",
    "weekId": "2026-W42",
    "endsAt": "2026-01-26T23:59:59Z",
    "endsInLabel": "Осталось 3 дня",
    "totalPlayers": 5847,
    "entries": [
      {
        "rank": 1,
        "playerId": "user_456",
        "username": "ProGamer",
        "score": 187,
        "totalQuestions": 187,
        "efficiency": "100%",
        "continueCount": 0,
        "badge": "🥇",
        "completedAt": 1706428920
      }
      // ... top 10
    ],
    "playerRank": {
      "rank": 342,
      "playerId": "user_123",
      "username": "You",
      "score": 47,
      "badge": null,
      "rewardTier": null
    }
  }
}
```

---

### 7. Complete Game (Quit)

> ❌ Endpoint не реализован. В коде используется `DELETE /:gameId` (abandon). Отдельного complete endpoint нет.

```http
POST /api/v1/marathon/:gameId/complete
```

**Request:**
```json
{
  "reason": "quit"  // or "game_over" (auto-called)
}
```

**Response 200:**
```json
{
  "data": {
    "finalScore": 23,
    "totalQuestions": 24,
    "personalBest": 87,
    "isNewRecord": false,
    "newRecordBonus": 0,
    "weeklyRank": 342,
    "weeklyRankLabel": "#342",
    "bonusesUsed": {
      "shield": 2,
      "fiftyFifty": 0,
      "skip": 0,
      "freeze": 3
    },
    "continueCount": 1,
    "summary": "Правильных: 23 из 24 вопросов",
    "personalBestProgress": {
      "current": 23,
      "best": 87,
      "percent": 26,
      "label": "23/87 рекорда"
    },
    "topRankGap": {
      "targetRank": 100,
      "gap": 12,
      "label": "До топ-100 не хватает 12 ответов!"
    },
    "share": {
      "text": "🏃 Мой марафон в Quiz Sprint!\n✅ 23 правильных ответа\n🏆 #342 на этой неделе\nПопробуй побить мой рекорд!",
      "url": "https://quiz-sprint-tma.online/marathon"
    }
  }
}
```

---

## Domain Events

> ✅ Все события реализованы плюс дополнительные: `LifeLostEvent`, `DifficultyIncreasedEvent`.

```go
type MarathonGameStartedEvent struct {
    GameID        string
    PlayerID      string
    PersonalBest  int
    BonusInventory map[string]int
    Timestamp     int64
}

type MarathonQuestionAnsweredEvent struct {
    GameID        string
    PlayerID      string
    QuestionID    string
    AnswerID      string
    IsCorrect     bool
    LivesRemaining int
    CurrentScore  int
    Timestamp     int64
}

type BonusUsedEvent struct {
    GameID     string
    PlayerID   string
    BonusType  string
    QuestionID string
    Timestamp  int64
}

type MarathonGameOverEvent struct {
    GameID          string
    PlayerID        string
    FinalScore      int
    TotalQuestions  int
    PersonalBest    int
    IsNewRecord     bool
    ContinueCount   int
    Timestamp       int64
}

type ContinueUsedEvent struct {
    GameID        string
    PlayerID      string
    ContinueCount int
    PaymentMethod string  // "coins" | "ad"
    Cost          int
    Timestamp     int64
}
```

---

## Error Codes

> ⚠️ Коды маппируются в HTTP-статусы, но возвращаются как plain text, не как структурированный `{"error": {"code": ..., "message": ...}}`.

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `INVALID_TIME_TAKEN` | Time not in valid range |
| 400 | `GAME_NOT_ACTIVE` | Game not in progress |
| 400 | `INSUFFICIENT_COINS` | Not enough coins |
| 400 | `INSUFFICIENT_BONUSES` | Bonus not available |
| 404 | `GAME_NOT_FOUND` | Game doesn't exist |
| 409 | `GAME_ALREADY_OVER` | Cannot answer after game over |
| 409 | `ACTIVE_GAME_EXISTS` | Finish current game first |
