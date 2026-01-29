# Daily Challenge - Frontend Integration

## Thin Client Architecture

**Critical:** Frontend has ZERO game logic. All state and calculations on backend.

---

## Component Responsibilities

### DailyChallengeView.vue

**Only does:**
- Fetch status: `GET /api/v1/daily/status`
- Render based on `hasPlayed` flag
- Show "Start" or "Results" screen

**Does NOT:**
- Calculate if player can play
- Determine streak locally
- Validate game state

---

### QuestionCard.vue

**Only does:**
- Display question text from server
- Render 4 answer buttons
- Track timer (visual only)
- Send answer: `POST /api/v1/daily/:gameId/answer`

**Does NOT:**
- Validate if answer is correct
- Calculate score
- Enforce time limit (server does this)
- Prevent answer changes (server rejects)

**Timer behavior:**
```typescript
// ✅ Visual countdown only
const timeLeft = ref(15)
const interval = setInterval(() => {
  if (timeLeft.value > 0) timeLeft.value--
}, 1000)

// When reaches 0, still send to server
// Server validates actual time taken
```

**Answer submission:**
```typescript
const submitAnswer = async (answerId: string) => {
  const timeTaken = 15 - timeLeft.value

  const { data } = await api.post(`/daily/${gameId}/answer`, {
    questionId: currentQuestion.id,
    answerId,
    timeTaken  // Server validates this
  })

  // Backend tells us what to do next
  if (data.isGameCompleted) {
    router.push(`/daily/results/${gameId}`)
  } else {
    // Move to next question
    currentQuestionIndex.value++
  }
}
```

---

### ResultsScreen.vue

**Only does:**
- Fetch results from completion response
- Display exactly what backend sent:
  - `finalScore` → "920 очков"
  - `chestType` + `chestIcon` → "🏆 Золотой сундук"
  - `rankLabel` → "#847 из 12,847"
- Button to open chest

**Does NOT:**
- Calculate score (uses `finalScore` from API)
- Determine chest type (uses `chestType`)
- Calculate rank (uses `rank` from API)

**Data flow:**
```typescript
// Response from last answer submission
const results = {
  finalScore: 920,
  chestType: "golden",
  chestIcon: "🏆",
  chestLabel: "Золотой сундук",
  rankLabel: "#847 из 12,847"
}

// Just render it
<div>{{ results.chestIcon }} {{ results.chestLabel }}</div>
<div>Твоя позиция: {{ results.rankLabel }}</div>
```

---

### ChestOpening.vue

**Only does:**
- Call `POST /api/v1/daily/:gameId/chest/open`
- Play animation
- Display rewards from response

**Rewards response:**
```json
{
  "rewards": {
    "coins": 420,
    "coinsLabel": "+420 💰",
    "pvpTickets": 5,
    "ticketsLabel": "+5 🎟️ PvP билетов",
    "bonuses": [
      {
        "type": "shield",
        "icon": "🛡️",
        "label": "Щит",
        "description": "Одна бесплатная ошибка в Марафоне"
      }
    ]
  },
  "streakBonus": "+25% от серии 7 дней"
}
```

Frontend just maps over `bonuses` array and displays.

---

## State Management

### Use Vue Query (TanStack Query)

```typescript
// ✅ Good: Server as single source of truth
const { data: status } = useQuery({
  queryKey: ['dailyStatus', playerId],
  queryFn: () => api.get('/daily/status', { params: { playerId } })
})

// Status tells us everything:
// - hasPlayed: true/false
// - canRetry: true/false
// - results: { score, chest, rank }
```

### NO Pinia/Vuex for game state

```typescript
// ❌ BAD: Local game state
const gameStore = {
  currentScore: 0,
  answeredQuestions: [],
  calculateScore() { /* logic */ }
}

// ✅ GOOD: Just track UI state
const uiState = {
  isChestAnimationPlaying: false,
  selectedAnswerId: null,
  showConfetti: false
}
```

---

## API Response Structure

### Principle: Backend returns render-ready data

**Bad (requires frontend logic):**
```json
{
  "correctAnswers": 8,
  "totalQuestions": 10
}
// Frontend calculates: "8/10" or "80%"
```

**Good (ready to display):**
```json
{
  "correctAnswers": 8,
  "totalQuestions": 10,
  "scoreLabel": "8/10 ✓",
  "percentageLabel": "80%",
  "performanceLabel": "Отличный результат!"
}
```

### Localization

Backend includes localized strings:
```json
{
  "chestLabel": "Золотой сундук",
  "retryLabel": "Попробовать ещё раз",
  "shareText": "Я занял #847 место!"
}
```

Frontend just displays. No i18n logic needed for game data.

---

## Error Handling

### Server errors are actionable

```json
{
  "error": {
    "code": "ALREADY_PLAYED_TODAY",
    "message": "Вы уже играли сегодня",
    "action": {
      "type": "navigate",
      "route": "/daily/results",
      "params": { "gameId": "dg_abc123" }
    }
  }
}
```

Frontend:
```typescript
if (error.action?.type === 'navigate') {
  router.push({
    path: error.action.route,
    params: error.action.params
  })
}
```

---

## Validation

### Client-side validation: ONLY for UX

```typescript
// ✅ Show instant feedback (UX)
if (!answerId) {
  toast.error('Выберите ответ')
  return
}

// Still send to server (server validates again)
await api.post('/daily/:gameId/answer', { answerId })
```

**Server ALWAYS validates:**
- Time taken valid (0-15s)
- Question not already answered
- Game still active
- Answer belongs to question

---

## Real-time Updates

### Leaderboard

**Option 1: Polling**
```typescript
useQuery({
  queryKey: ['leaderboard', date],
  queryFn: () => api.get('/daily/leaderboard', { params: { date } }),
  refetchInterval: 10000  // 10s
})
```

**Option 2: WebSocket (future)**
```typescript
// Server pushes rank updates
ws.on('rank_updated', (data) => {
  queryClient.setQueryData(['dailyStatus'], (old) => ({
    ...old,
    rank: data.newRank
  }))
})
```

---

## Testing Frontend

### Mock API responses

```typescript
// tests/daily-challenge.spec.ts
const mockStatus = {
  hasPlayed: false,
  currentStreak: 5,
  canPlayNow: true
}

vi.spyOn(api, 'get').mockResolvedValue({ data: mockStatus })

// Test only rendering logic
expect(screen.getByText('🔥 5 дней подряд')).toBeInTheDocument()
```

### NO business logic tests

```typescript
// ❌ BAD: Testing business logic on frontend
test('calculates golden chest for 8+ correct', () => {
  expect(getChestType(8)).toBe('golden')
})

// ✅ GOOD: Testing rendering
test('displays golden chest when API returns it', () => {
  render(<Results chestType="golden" chestIcon="🏆" />)
  expect(screen.getByText('🏆')).toBeInTheDocument()
})
```

---

## Performance Optimization

### Caching strategy

```typescript
// Cache status for 30s (avoid repeated calls)
useQuery({
  queryKey: ['dailyStatus'],
  queryFn: fetchStatus,
  staleTime: 30_000,
  cacheTime: 5 * 60_000
})

// Invalidate after game completion
await submitAnswer(...)
queryClient.invalidateQueries(['dailyStatus'])
```

### Prefetch next question

```typescript
// While player reads question 3, prefetch question 4
const { prefetchQuery } = useQueryClient()

watchEffect(() => {
  if (currentIndex < 9) {
    prefetchQuery(['question', currentIndex + 1])
  }
})
```

---

## Anti-patterns

### ❌ Don't duplicate backend logic

```typescript
// ❌ BAD
function calculateStreak(lastDate, currentDate) {
  const diff = daysBetween(lastDate, currentDate)
  return diff === 1 ? streak + 1 : 0
}

// ✅ GOOD
const streak = response.data.currentStreak
```

### ❌ Don't store game state locally

```typescript
// ❌ BAD
localStorage.setItem('dailyScore', score)
localStorage.setItem('currentQuestion', index)

// ✅ GOOD
// Backend stores everything
// Frontend refetches on mount
```

### ❌ Don't trust client time

```typescript
// ❌ BAD
const timeTaken = Date.now() - questionStartTime
// User can manipulate this

// ✅ GOOD
// Send client time, but server validates
// Server uses server-side timestamps as source of truth
```

---

## Checklist

Frontend implementation complete when:

- [ ] NO score calculations in components
- [ ] NO game state in Pinia/Vuex/localStorage
- [ ] All data from API responses
- [ ] Timer is visual only (server enforces)
- [ ] Errors handled by server responses
- [ ] Tests only cover rendering, not logic
- [ ] Vue Query for server state management
- [ ] UI state separate from game state
