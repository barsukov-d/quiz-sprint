# PvP Duel - API Specification

## Architecture Note: Real-Time + REST

- **REST API:** Game setup, results, history
- **WebSocket:** Live duel sync (questions, answers, opponent status)

---

## Authentication
```
Authorization: tma <base64_encoded_init_data>
```

---

## REST Endpoints

### 1. Get Duel Status

```http
GET /api/v1/duel/status
```

**Response 200:**
```json
{
  "data": {
    "hasActiveDuel": false,
    "player": {
      "id": "user_123",
      "mmr": 1650,
      "league": "gold",
      "division": 3,
      "leagueLabel": "🥇 Gold III",
      "seasonWins": 23,
      "seasonLosses": 18,
      "winRate": "56%"
    },
    "tickets": 3,
    "friendsOnline": [
      {
        "id": "user_456",
        "username": "ProGamer",
        "avatar": "https://...",
        "league": "gold",
        "division": 1,
        "leagueLabel": "🥇 Gold I",
        "status": "online",
        "lastSeen": null
      },
      {
        "id": "user_789",
        "username": "BestQuiz",
        "league": "silver",
        "division": 2,
        "leagueLabel": "🥈 Silver II",
        "status": "recent",
        "lastSeen": 1706428500
      }
    ],
    "pendingChallenges": [
      {
        "challengeId": "ch_abc123",
        "from": {
          "id": "user_999",
          "username": "Challenger",
          "leagueLabel": "💍 Platinum IV"
        },
        "expiresAt": 1706428900,
        "expiresIn": 45
      }
    ]
  }
}
```

---

### 2. Start Matchmaking (Random)

```http
POST /api/v1/duel/queue/join
```

**Request:**
```json
{}
```

**Response 201:**
```json
{
  "data": {
    "queueId": "q_abc123",
    "status": "searching",
    "estimatedWait": 15,
    "mmrRange": {
      "min": 1600,
      "max": 1700
    }
  }
}
```

**WebSocket message when game found:**
```json
{
  "type": "game_found",
  "data": {
    "gameId": "g_xyz789",
    "opponent": {
      "id": "user_456",
      "username": "ProGamer",
      "avatar": "https://...",
      "mmr": 1720,
      "leagueLabel": "🥇 Gold II"
    },
    "startsIn": 3
  }
}
```

---

### 3. Cancel Matchmaking

```http
DELETE /api/v1/duel/queue/leave
```

**Response 200:**
```json
{
  "data": {
    "success": true,
    "ticketRefunded": true,
    "tickets": 3
  }
}
```

---

### 4. Challenge Friend

```http
POST /api/v1/duel/challenge
```

**Request:**
```json
{
  "friendId": "user_456"
}
```

**Response 201:**
```json
{
  "data": {
    "challengeId": "ch_abc123",
    "status": "pending",
    "expiresAt": 1706429000,
    "expiresIn": 60,
    "friend": {
      "id": "user_456",
      "username": "ProGamer",
      "status": "online"
    },
    "ticketConsumed": true,
    "ticketsRemaining": 2
  }
}
```

**WebSocket to friend:**
```json
{
  "type": "challenge_received",
  "data": {
    "challengeId": "ch_abc123",
    "from": {
      "id": "user_123",
      "username": "Challenger",
      "leagueLabel": "🥇 Gold III"
    },
    "expiresIn": 60
  }
}
```

---

### 5. Respond to Challenge

```http
POST /api/v1/duel/challenge/:challengeId/respond
```

**Request (Accept):**
```json
{
  "action": "accept"
}
```

**Response 200 (Game starts):**
```json
{
  "data": {
    "gameId": "g_xyz789",
    "status": "starting",
    "ticketConsumed": true,
    "ticketsRemaining": 2,
    "opponent": {
      "id": "user_123",
      "username": "Challenger",
      "leagueLabel": "🥇 Gold III"
    },
    "startsIn": 3
  }
}
```

**Request (Decline):**
```json
{
  "action": "decline"
}
```

**Response 200:**
```json
{
  "data": {
    "success": true,
    "message": "Вызов отклонён"
  }
}
```

---

### 6. Generate Challenge Link

```http
POST /api/v1/duel/challenge/link
```

**Response 201:**
```json
{
  "data": {
    "challengeLink": "https://t.me/quiz_sprint_dev_bot?startapp=duel_abc123",
    "expiresAt": 1706515200,
    "expiresIn": 86400,
    "shareText": "⚔️ Вызываю тебя на дуэль в Quiz Sprint!\nПокажи кто здесь умнее! 🧠\n\nhttps://t.me/quiz_sprint_dev_bot?startapp=duel_abc123"
  }
}
```

---

### 6a. Accept Challenge by Link Code

Used by TMA when user opens deep link with `?startapp=duel_xxx` parameter.

```http
POST /api/v1/duel/challenge/accept-by-code
```

**Request:**
```json
{
  "playerId": "user_123",
  "linkCode": "duel_abc12345"
}
```

**Response 200:**
```json
{
  "data": {
    "success": true,
    "gameId": "g_xyz789",
    "ticketConsumed": true,
    "startsIn": 3,
    "challengerId": "user_456"
  }
}
```

**Error 404:** Challenge not found
**Error 409:** Challenge expired or already accepted

---

### 7. Get Game Result

```http
GET /api/v1/duel/game/:gameId
```

**Response 200:**
```json
{
  "data": {
    "gameId": "g_xyz789",
    "status": "completed",
    "winner": "user_123",
    "isFriendGame": true,
    "players": {
      "player1": {
        "id": "user_123",
        "username": "You",
        "score": 5,
        "totalTime": 42500,
        "mmrBefore": 1650,
        "mmrAfter": 1678,
        "mmrChange": 28
      },
      "player2": {
        "id": "user_456",
        "username": "ProGamer",
        "score": 4,
        "totalTime": 38200,
        "mmrBefore": 1720,
        "mmrAfter": 1692,
        "mmrChange": -28
      }
    },
    "questions": [
      {
        "questionId": "q_001",
        "text": "Какой элемент имеет символ Au?",
        "correctAnswerId": "a_002",
        "player1Answer": {
          "answerId": "a_002",
          "isCorrect": true,
          "timeTaken": 4200
        },
        "player2Answer": {
          "answerId": "a_001",
          "isCorrect": false,
          "timeTaken": 3800
        }
      }
      // ... 7 questions
    ],
    "winReason": "score",
    "completedAt": 1706429100,
    "share": {
      "text": "⚔️ ПОБЕДА В ДУЭЛИ!\n@YourName 🏆\n5 : 4\n🥇 Gold III\n\nПопробуй победить меня!\nhttps://t.me/quiz_sprint_dev_bot?startapp=duel_abc123",
      "imageUrl": "https://api.quiz-sprint.online/share/m_xyz789.png"
    },
    "rematch": {
      "available": true,
      "opponentWaiting": false,
      "ticketCost": 1
    }
  }
}
```

---

### 8. Request Rematch

```http
POST /api/v1/duel/game/:gameId/rematch
```

**Response 201:**
```json
{
  "data": {
    "rematchId": "rm_abc123",
    "status": "waiting_opponent",
    "expiresIn": 15,
    "ticketConsumed": true
  }
}
```

**WebSocket to opponent:**
```json
{
  "type": "rematch_request",
  "data": {
    "rematchId": "rm_abc123",
    "from": {
      "id": "user_123",
      "username": "Challenger"
    },
    "expiresIn": 15
  }
}
```

---

### 8b. Surrender

```http
POST /api/v1/duel/game/:gameId/surrender
```

**Availability:** Only after question 3 (to prevent rage-quits on first question).

**Response 200:**
```json
{
  "data": {
    "gameId": "g_xyz789",
    "result": "forfeit",
    "winner": "user_456",
    "mmrChange": -25,
    "newMmr": 1625,
    "message": "Вы сдались. Победа присуждена сопернику."
  }
}
```

**Response 400 (too early):**
```json
{
  "error": {
    "code": "SURRENDER_TOO_EARLY",
    "message": "Сдаться можно только после 3-го вопроса"
  }
}
```

**WebSocket to opponent:**
```json
{
  "type": "opponent_surrendered",
  "data": {
    "gameId": "g_xyz789",
    "message": "Соперник сдался. Победа!"
  }
}
```

---

### 9. Get Game History

```http
GET /api/v1/duel/history?limit=20&offset=0&filter=friends
```

**Query params:**
- `limit`: 1-50 (default 20)
- `offset`: pagination
- `filter`: `all` | `friends` | `wins` | `losses`

**Response 200:**
```json
{
  "data": {
    "games": [
      {
        "gameId": "g_xyz789",
        "opponent": {
          "id": "user_456",
          "username": "ProGamer",
          "leagueLabel": "🥇 Gold II"
        },
        "result": "win",
        "score": "5:4",
        "mmrChange": 28,
        "isFriendGame": true,
        "completedAt": 1706429100
      }
    ],
    "total": 41,
    "hasMore": true
  }
}
```

---

### 10. Get Leaderboards

```http
GET /api/v1/duel/leaderboard?type=seasonal&limit=10
```

**Query params:**
- `type`: `seasonal` | `friends` | `referrals`
- `limit`: 1-100

**Response 200:**
```json
{
  "data": {
    "type": "seasonal",
    "seasonId": "2026-S04",
    "endsAt": 1709251199,
    "endsInLabel": "Осталось 12 дней",
    "entries": [
      {
        "rank": 1,
        "playerId": "user_top",
        "username": "ChampionX",
        "mmr": 3250,
        "leagueLabel": "👑 Legend",
        "wins": 187,
        "losses": 42
      }
    ],
    "playerRank": {
      "rank": 342,
      "percentile": "top 5%"
    }
  }
}
```

---

### 11. Get Referral Stats

```http
GET /api/v1/duel/referrals
```

**Response 200:**
```json
{
  "data": {
    "referralLink": "https://t.me/quiz_sprint_dev_bot?startapp=ref_user123",
    "totalReferrals": 8,
    "activeReferrals": 5,
    "pendingRewards": [
      {
        "friendId": "user_new",
        "friendUsername": "NewPlayer",
        "milestone": "reached_silver",
        "reward": {
          "tickets": 10,
          "coins": 500,
          "badge": "Гуру"
        },
        "claimable": true
      }
    ],
    "referralLeaderboardRank": 47,
    "monthlyReferrals": 3,
    "referrals": [
      {
        "friendId": "user_ref1",
        "username": "Friend1",
        "registeredAt": 1706400000,
        "currentLeague": "silver",
        "milestonesCompleted": ["registered", "played_5_duels", "reached_silver"],
        "rewardsClaimed": true
      }
    ]
  }
}
```

---

### 12. Claim Referral Reward

```http
POST /api/v1/duel/referrals/:friendId/claim
```

**Request:**
```json
{
  "milestone": "reached_silver"
}
```

**Response 200:**
```json
{
  "data": {
    "success": true,
    "rewards": {
      "tickets": 10,
      "coins": 500,
      "badge": "Гуру"
    },
    "newTicketBalance": 15,
    "newCoinBalance": 2500
  }
}
```

---

## WebSocket Protocol

### Connection
```
wss://quiz-sprint-tma.online/ws/duel/:gameId?playerId=<player_id>
```

### Message Types (Server → Client)

#### connected
```json
{
  "type": "connected",
  "data": {
    "gameId": "g_xyz789",
    "playerId": "user_123"
  }
}
```

#### game_ready
When both players are connected:
```json
{
  "type": "game_ready",
  "data": {
    "gameId": "g_xyz789",
    "player1Id": "user_123",
    "player2Id": "user_456",
    "startsIn": 3,
    "totalRounds": 7
  }
}
```

#### new_question
```json
{
  "type": "new_question",
  "data": {
    "roundNum": 3,
    "totalRounds": 7,
    "question": {
      "id": "q_003",
      "text": "Какой элемент имеет символ Au?",
      "answers": [
        {"id": "a_001", "text": "Серебро"},
        {"id": "a_002", "text": "Золото"},
        {"id": "a_003", "text": "Медь"},
        {"id": "a_004", "text": "Алюминий"}
      ],
      "timeLimit": 10
    },
    "serverTime": 1706429050000
  }
}
```

#### answer_result
Sent to both players when someone answers:
```json
{
  "type": "answer_result",
  "data": {
    "playerId": "user_123",
    "questionId": "q_003",
    "isCorrect": true,
    "correctAnswer": "a_002",
    "timeTaken": 4200,
    "player1Score": 3,
    "player2Score": 2
  }
}
```

#### round_complete
When both players have answered:
```json
{
  "type": "round_complete",
  "data": {
    "roundNum": 3,
    "player1Score": 3,
    "player2Score": 2,
    "nextRoundIn": 2
  }
}
```

#### round_timeout
When time runs out:
```json
{
  "type": "round_timeout",
  "data": {
    "roundNum": 3
  }
}
```

#### game_complete
```json
{
  "type": "game_complete",
  "data": {
    "winnerId": "user_123",
    "player1Score": 5,
    "player2Score": 4,
    "player1MMRChange": 28,
    "player2MMRChange": -28,
    "player1NewMMR": 1678,
    "player2NewMMR": 1692
  }
}
```

#### opponent_disconnected
```json
{
  "type": "opponent_disconnected",
  "data": {
    "playerId": "user_456",
    "reconnectIn": 10
  }
}
```

#### error
```json
{
  "type": "error",
  "error": "Error message"
}
```

### Message Types (Client → Server)

#### submit_answer
```json
{
  "type": "submit_answer",
  "data": {
    "playerId": "user_123",
    "gameId": "g_xyz789",
    "questionId": "q_003",
    "answerId": "a_002",
    "timeTaken": 4200
  }
}
```

#### player_ready
Signal ready for next round:
```json
{
  "type": "player_ready"
}
```

#### ping
Heartbeat:
```json
{
  "type": "ping"
}
```

---

## Error Codes

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `INSUFFICIENT_TICKETS` | Not enough tickets |
| 400 | `ALREADY_IN_QUEUE` | Already searching for match |
| 400 | `ALREADY_IN_GAME` | Currently in active duel |
| 400 | `INVALID_CHALLENGE` | Challenge expired or invalid |
| 404 | `GAME_NOT_FOUND` | Game doesn't exist |
| 404 | `FRIEND_NOT_FOUND` | Friend ID invalid |
| 409 | `CHALLENGE_EXPIRED` | Challenge timed out |
| 409 | `FRIEND_BUSY` | Friend already in queue/match |
| 429 | `RATE_LIMIT` | Too many requests |

---

## Domain Events

```go
type DuelGameStartedEvent struct {
    GameID       string
    Player1ID    string
    Player2ID    string
    IsFriendGame bool
    Timestamp    int64
}

type PlayerAnsweredEvent struct {
    GameID       string
    PlayerID     string
    QuestionID   string
    AnswerID     string
    IsCorrect    bool
    TimeTaken    int64
    Timestamp    int64
}

type DuelGameFinishedEvent struct {
    GameID         string
    WinnerID       string
    LoserID        string
    WinnerScore    int
    LoserScore     int
    WinnerMMRDelta int
    LoserMMRDelta  int
    WinReason      string  // "score", "time", "forfeit"
    IsFriendGame   bool
    Timestamp      int64
}

type ChallengeCreatedEvent struct {
    ChallengeID  string
    ChallengerID string
    ChallengedID string
    ExpiresAt    int64
    Timestamp    int64
}

type ReferralMilestoneEvent struct {
    InviterID    string
    InviteeID    string
    Milestone    string  // "registered", "played_5", "reached_silver"
    Timestamp    int64
}
```
