# PvP Duel - API Specification

## Architecture Note: Real-Time + REST

- **REST API:** Match setup, results, history
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

**WebSocket message when match found:**
```json
{
  "type": "match_found",
  "data": {
    "matchId": "m_xyz789",
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

**Response 200 (Match starts):**
```json
{
  "data": {
    "matchId": "m_xyz789",
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
    "challengeLink": "t.me/quiz_sprint_dev_bot?start=duel_abc123",
    "expiresAt": 1706515200,
    "expiresIn": 86400,
    "shareText": "⚔️ Вызываю тебя на дуэль в Quiz Sprint!\nПокажи кто здесь умнее! 🧠\n\nt.me/quiz_sprint_dev_bot?start=duel_abc123"
  }
}
```

---

### 7. Get Match Result

```http
GET /api/v1/duel/match/:matchId
```

**Response 200:**
```json
{
  "data": {
    "matchId": "m_xyz789",
    "status": "completed",
    "winner": "user_123",
    "isFriendMatch": true,
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
      "text": "⚔️ ПОБЕДА В ДУЭЛИ!\n@YourName 🏆\n5 : 4\n🥇 Gold III\n\nПопробуй победить меня!\nt.me/quiz_sprint_dev_bot?start=duel_abc123",
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
POST /api/v1/duel/match/:matchId/rematch
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

### 9. Get Match History

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
    "matches": [
      {
        "matchId": "m_xyz789",
        "opponent": {
          "id": "user_456",
          "username": "ProGamer",
          "leagueLabel": "🥇 Gold II"
        },
        "result": "win",
        "score": "5:4",
        "mmrChange": 28,
        "isFriendMatch": true,
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
    "referralLink": "t.me/quiz_sprint_dev_bot?start=ref_user123",
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
          "badge": "Наставник"
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
      "badge": "Наставник"
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
wss://quiz-sprint-tma.online/ws/duel?token=<auth_token>
```

### Message Types (Server → Client)

#### match_found
```json
{
  "type": "match_found",
  "data": {
    "matchId": "m_xyz789",
    "opponent": {...},
    "startsIn": 3
  }
}
```

#### question
```json
{
  "type": "question",
  "data": {
    "questionNumber": 3,
    "totalQuestions": 7,
    "questionId": "q_003",
    "text": "Какой элемент имеет символ Au?",
    "answers": [
      {"id": "a_001", "text": "Серебро", "position": 0},
      {"id": "a_002", "text": "Золото", "position": 1},
      {"id": "a_003", "text": "Медь", "position": 2},
      {"id": "a_004", "text": "Алюминий", "position": 3}
    ],
    "timeLimit": 10000,
    "serverTime": 1706429050000
  }
}
```

#### opponent_answered
```json
{
  "type": "opponent_answered",
  "data": {
    "questionNumber": 3,
    "opponentAnswered": true,
    "yourStatus": "answering"
  }
}
```

#### answer_result
```json
{
  "type": "answer_result",
  "data": {
    "questionNumber": 3,
    "yourAnswer": {
      "answerId": "a_002",
      "isCorrect": true,
      "timeTaken": 4200
    },
    "opponentAnswer": {
      "isCorrect": false,
      "timeTaken": 3800
    },
    "correctAnswerId": "a_002",
    "scores": {
      "you": 3,
      "opponent": 2
    },
    "nextQuestionIn": 1500
  }
}
```

#### match_complete
```json
{
  "type": "match_complete",
  "data": {
    "result": "win",
    "finalScore": {
      "you": 5,
      "opponent": 4
    },
    "mmrChange": 28,
    "newMmr": 1678,
    "rankChange": null,
    "matchId": "m_xyz789"
  }
}
```

#### challenge_received
```json
{
  "type": "challenge_received",
  "data": {
    "challengeId": "ch_abc123",
    "from": {...},
    "expiresIn": 60
  }
}
```

#### opponent_disconnected
```json
{
  "type": "opponent_disconnected",
  "data": {
    "gracePeriod": 10,
    "message": "Соперник отключился. Ожидание..."
  }
}
```

### Message Types (Client → Server)

#### submit_answer
```json
{
  "type": "submit_answer",
  "data": {
    "matchId": "m_xyz789",
    "questionId": "q_003",
    "answerId": "a_002",
    "clientTime": 4200
  }
}
```

#### send_emote
```json
{
  "type": "send_emote",
  "data": {
    "matchId": "m_xyz789",
    "emote": "fire"
  }
}
```

---

## Error Codes

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `INSUFFICIENT_TICKETS` | Not enough tickets |
| 400 | `ALREADY_IN_QUEUE` | Already searching for match |
| 400 | `ALREADY_IN_MATCH` | Currently in active duel |
| 400 | `INVALID_CHALLENGE` | Challenge expired or invalid |
| 404 | `MATCH_NOT_FOUND` | Match doesn't exist |
| 404 | `FRIEND_NOT_FOUND` | Friend ID invalid |
| 409 | `CHALLENGE_EXPIRED` | Challenge timed out |
| 409 | `FRIEND_BUSY` | Friend already in queue/match |
| 429 | `RATE_LIMIT` | Too many requests |

---

## Domain Events

```go
type DuelMatchStartedEvent struct {
    MatchID      string
    Player1ID    string
    Player2ID    string
    IsFriendMatch bool
    Timestamp    int64
}

type DuelAnswerSubmittedEvent struct {
    MatchID      string
    PlayerID     string
    QuestionID   string
    AnswerID     string
    IsCorrect    bool
    TimeTaken    int64
    Timestamp    int64
}

type DuelMatchCompletedEvent struct {
    MatchID       string
    WinnerID      string
    LoserID       string
    WinnerScore   int
    LoserScore    int
    WinnerMMRDelta int
    LoserMMRDelta  int
    WinReason     string  // "score", "time", "forfeit"
    IsFriendMatch bool
    Timestamp     int64
}

type DuelChallengeCreatedEvent struct {
    ChallengeID  string
    ChallengerID string
    ChallengedID string
    ExpiresAt    int64
    Timestamp    int64
}

type ReferralMilestoneReachedEvent struct {
    InviterID    string
    InviteeID    string
    Milestone    string  // "registered", "played_5", "reached_silver"
    Timestamp    int64
}
```
