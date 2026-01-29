# Backend Implementation TODO

## ✅ Status Update (2026-01-26 20:30)

**Great Progress!** Daily Challenge endpoints are now **WORKING** 🎉

**Test Results:**
- ✅ `/api/v1/daily-challenge/status` - Returns data correctly
- ✅ `/api/v1/daily-challenge/streak` - Returns data correctly
- ❌ `/api/v1/marathon/status` - Returns 500 error
- ❌ `/api/v1/marathon/personal-bests` - Returns 500 error

**Next Steps:**
1. ✅ Run migrations on dev server (likely already done for Daily Challenge)
2. ⚠️ Run Marathon migrations: `008_create_marathon_tables.sql` and `009_create_daily_challenge_tables.sql`
3. 🧪 Test all Daily Challenge flow end-to-end
4. 🔧 Fix Marathon 500 errors (likely just missing table migrations)

---

## 🚨 Required Backend Endpoints

The frontend UI is complete and ready. Below is the detailed status of each endpoint:

### Daily Challenge Endpoints

#### 1. `GET /api/v1/daily-challenge/status`
**Query Params:**
- `playerId` (string, required)
- `date` (string, optional) - Format: YYYY-MM-DD

**Response:**
```typescript
{
  data: {
    hasPlayed: boolean
    game: DailyGameDTO | null
    timeToExpire: number  // seconds until midnight UTC
    totalPlayers: number
  }
}
```

**Current Status:** ✅ **WORKING**

---

#### 2. `GET /api/v1/daily-challenge/streak`
**Query Params:**
- `playerId` (string, required)

**Response:**
```typescript
{
  data: {
    streak: {
      currentStreak: number
      longestStreak: number
      lastPlayedAt: number  // Unix timestamp
      brokenYesterday: boolean
    }
  }
}
```

**Current Status:** ✅ **WORKING**

---

#### 3. `POST /api/v1/daily-challenge/start`
**Body:**
```typescript
{
  playerId: string
  date?: string  // YYYY-MM-DD, defaults to today
}
```

**Response:**
```typescript
{
  data: {
    game: DailyGameDTO
    firstQuestion: QuestionDTO
    timeLimit: number  // seconds per question (15)
  }
}
```

**Current Status:** ❌ 500 Internal Server Error

---

#### 4. `POST /api/v1/daily-challenge/{gameId}/answer`
**Body:**
```typescript
{
  questionId: string
  answerId: string
  playerId: string
  timeTaken: number  // seconds
}
```

**Response:**
```typescript
{
  data: {
    questionIndex: number
    isGameCompleted: boolean
    nextQuestion?: QuestionDTO
    gameResults?: GameResultsDTO
  }
}
```

**Current Status:** 🔨 Not tested yet

---

#### 5. `GET /api/v1/daily-challenge/leaderboard`
**Query Params:**
- `date` (string, optional) - YYYY-MM-DD

**Response:**
```typescript
{
  data: {
    leaderboard: LeaderboardEntryDTO[]
  }
}
```

**Current Status:** 🔨 Not tested yet

---

### Marathon Endpoints

#### 6. `GET /api/v1/marathon/status`
**Query Params:**
- `playerId` (string, required)

**Response:**
```typescript
{
  data: {
    hasActiveGame: boolean
    game: MarathonGameDTO | null
    lives: number
    maxLives: number
    timeToLifeRestore: number  // seconds
  }
}
```

**Current Status:** ❌ 500 Internal Server Error

---

#### 7. `GET /api/v1/marathon/personal-bests`
**Query Params:**
- `playerId` (string, required)

**Response:**
```typescript
{
  data: {
    personalBests: MarathonPersonalBestDTO[]
  }
}
```

**Current Status:** ❌ 500 Internal Server Error

---

#### 8. `POST /api/v1/marathon/start`
**Body:**
```typescript
{
  playerId: string
  categoryId: string
}
```

**Response:**
```typescript
{
  data: {
    game: MarathonGameDTO
    firstQuestion: QuestionDTO
    timeLimit: number
    hints: {
      fiftyFifty: number
      extraTime: number
      skip: number
      hint: number
    }
  }
}
```

**Current Status:** 🔨 Not tested yet

---

#### 9. `POST /api/v1/marathon/{gameId}/answer`
**Body:**
```typescript
{
  questionId: string
  answerId: string
  playerId: string
  timeTaken: number
}
```

**Response:**
```typescript
{
  data: {
    isCorrect: boolean
    correctAnswerId: string
    pointsEarned: number
    currentStreak: number
    lives: number
    isGameOver: boolean
    nextQuestion?: QuestionDTO
    gameStats?: {
      score: number
      questionCount: number
      accuracy: number
    }
  }
}
```

**Current Status:** 🔨 Not tested yet

---

#### 10. `POST /api/v1/marathon/{gameId}/hint`
**Body:**
```typescript
{
  hintType: 'fifty_fifty' | 'extra_time' | 'skip' | 'hint'
  playerId: string
}
```

**Response:**
```typescript
{
  data: {
    hintsRemaining: {
      fiftyFifty: number
      extraTime: number
      skip: number
      hint: number
    }
    eliminatedAnswers?: string[]  // for fifty_fifty
    timeAdded?: number  // for extra_time
    nextQuestion?: QuestionDTO  // for skip
    hintText?: string  // for hint
  }
}
```

**Current Status:** 🔨 Not tested yet

---

#### 11. `DELETE /api/v1/marathon/{gameId}`
**Query Params:**
- `playerId` (string, required)

**Response:**
```typescript
{
  message: 'Game abandoned successfully'
}
```

**Current Status:** 🔨 Not tested yet

---

#### 12. `GET /api/v1/marathon/leaderboard`
**Query Params:**
- `categoryId` (string, required)

**Response:**
```typescript
{
  data: {
    leaderboard: LeaderboardEntryDTO[]
  }
}
```

**Current Status:** 🔨 Not tested yet

---

## 📝 Implementation Notes

### DTOs Reference

All DTO types are defined in:
- `backend/internal/infrastructure/http/handlers/swagger_models.go`

Key DTOs:
- `DailyGameDTO` - Daily challenge game state
- `GameResultsDTO` - Final results with leaderboard
- `ReviewAnswerDTO` - Answer review data
- `StreakDTO` - Player streak info
- `MarathonGameDTO` - Marathon game state
- `MarathonPersonalBestDTO` - Personal record per category
- `QuestionDTO` - Question with answers
- `AnswerDTO` - Single answer option
- `LeaderboardEntryDTO` - Leaderboard entry

### Business Logic

See backend domain documentation:
- `backend/internal/domain/solo_marathon/` - Marathon game logic
- `backend/internal/domain/daily_challenge/` - Daily challenge logic (to be created)

### Testing

Once endpoints are implemented, test the complete flows:

**Daily Challenge Flow:**
1. Home → Check status (idle)
2. Start game → Get first question
3. Answer 10 questions
4. View results with leaderboard
5. Review all answers
6. Next day → New challenge available

**Marathon Flow:**
1. Home → Check status (3 lives)
2. Select category
3. Start game → Get first question
4. Answer until wrong (lose life) or give up
5. Use hints during gameplay
6. Game over → View stats
7. View leaderboard

---

## ✅ Frontend Status

**Fully Implemented:**
- ✅ All composables (useDailyChallenge, useMarathon, useGameTimer, useStreaks)
- ✅ All shared components (GameTimer, QuestionCard, AnswerButton)
- ✅ Daily Challenge complete flow (Play, Results, Review)
- ✅ Home screen cards (DailyChallengeCard, MarathonCard)
- ✅ Router integration
- ✅ TypeScript types generated from Swagger
- ✅ Vue Query hooks for all endpoints

**Waiting on Backend:**
- ⏳ All API endpoints listed above
- ⏳ Marathon views (will be created once endpoints are ready)

---

## 🔧 Development Workflow

1. Backend implements endpoint
2. Update Swagger docs in Go handlers
3. Run `pnpm run generate:all` from `tma/` directory
4. Frontend types auto-update
5. Test in browser at `https://dev.quiz-sprint-tma.online`

---

**Last Updated:** 2026-01-26
