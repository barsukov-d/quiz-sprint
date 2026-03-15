# Solo Marathon - Frontend Integration

> **Статус реализации (аудит 2026-03-15, обновлено 2026-03-15)**
> ✅ Реализовано: 8 | ⚠️ Расходится: 7 | ❌ Не реализовано: 3
>
> - ✅ Frontend has ZERO game logic
> - ✅ Lives from API
> - ✅ Bonuses trigger API calls
> - ✅ Vue Query for server state
> - ✅ Tests only cover rendering
> - ✅ QuestionScreen — реализован как `MarathonPlayView.vue`
> - ✅ GameOverScreen — реализован как `MarathonGameOverView.vue`
> - ✅ DifficultyTransition toast — реализован; backend возвращает `difficultyChanged`/`difficultyMessage` в question DTO
> - ✅ CorrectAnswerText — доступен в answer response (`correctAnswerText` field)
> - ✅ CanStart — доступен в GET /marathon/status response
> - ⚠️ Timer visual only, server validates — таймер визуальный, но сервер НЕ валидирует `timeTaken` против `timeLimit`
> - ⚠️ MarathonView.vue (pre-start) — существует только `MarathonCategoryView.vue`, нет отдельного pre-start экрана
> - ⚠️ BonusControls — встроен inline в `MarathonPlayView.vue`, не отдельный компонент
> - ⚠️ AnswerFeedback — inline alerts, не отдельный компонент
> - ⚠️ ResultsScreen — не отдельный; GameOver screen совмещает обе функции
> - ⚠️ MilestoneProgress — inline, не отдельный компонент
> - ❌ PersonalBestProgress — не реализован
> - ❌ OnboardingOverlay — не реализован
> - ❌ NetworkOverlay — не реализован
> - ❌ ShareCard — не реализован

## Thin Client Architecture

**Critical:** Frontend has ZERO game logic. Backend owns everything.

---

## Component Responsibilities

### MarathonView.vue

> ⚠️ Существует только `MarathonCategoryView.vue`. Отдельного pre-start экрана с personal best, бонусами и weekly rank нет.

**Only does:**
- Fetch status: `GET /api/v1/marathon/status`
- Show "Start" or "Resume" based on `hasActiveGame`
- Display personal best, bonuses, weekly rank

**Does NOT:**
- Calculate if player can start
- Track lives locally
- Determine if game is over

---

### QuestionScreen.vue

> ✅ Реализован как `MarathonPlayView.vue`. ⚠️ Сервер не валидирует `timeTaken` против `timeLimit` на вопрос.

**Only does:**
- Display question from API
- Show lives counter (from API: `livesLabel`)
- Show score (from API: `scoreLabel`)
- Render bonus buttons (enabled/disabled by API)
- Track timer visually
- Submit answer

**Does NOT:**
- Calculate lives remaining
- Validate answer correctness
- Determine game over state
- Calculate difficulty level

**Timer:**
```typescript
// ✅ Visual countdown only
const timeLimit = ref(currentQuestion.timeLimit)  // From API
const timeLeft = ref(timeLimit.value)

setInterval(() => {
  if (timeLeft.value > 0) timeLeft.value--
  if (timeLeft.value === 0) submitAnswer()  // Auto-submit
}, 1000)
```

**Answer submission:**
```typescript
const submitAnswer = async (answerId: string) => {
  const timeTaken = timeLimit.value - timeLeft.value

  const { data } = await api.post(`/marathon/${gameId}/answer`, {
    questionId: currentQuestion.id,
    answerId,
    timeTaken,
    shieldActive: isShieldActive.value  // Frontend tracks UI state only
  })

  // Backend tells us what happened
  if (data.isGameOver) {
    showGameOverScreen(data.gameOverData)
  } else {
    showFeedback(data)
    setTimeout(() => {
      currentQuestion.value = data.nextQuestion
    }, 2000)
  }
}
```

---

### BonusControls.vue

> ⚠️ Встроен inline в `MarathonPlayView.vue`, не отдельный компонент.

**Only does:**
- Display bonus buttons with quantities (from API)
- Send bonus usage to backend
- Update UI based on backend response

**Does NOT:**
- Track bonus inventory locally
- Validate if bonus available
- Implement bonus logic (50/50 answer removal, etc.)

**Using Shield:**
```typescript
const useShield = async () => {
  const { data } = await api.post(`/marathon/${gameId}/bonus`, {
    bonusType: 'shield',
    questionId: currentQuestion.id
  })

  // Backend handles inventory deduction
  bonusInventory.value = data.bonusInventory
  isShieldActive.value = data.bonusActive
  showToast(data.statusMessage)  // "🛡️ Щит активирован"
}
```

**Using 50/50:**
```typescript
const useFiftyFifty = async () => {
  const { data } = await api.post(`/marathon/${gameId}/bonus`, {
    bonusType: 'fifty_fifty',
    questionId: currentQuestion.id
  })

  // Backend tells us which answers to keep
  visibleAnswers.value = data.remainingAnswers
  bonusInventory.value = data.bonusInventory
}
```

**Using Skip:**
```typescript
const useSkip = async () => {
  const { data } = await api.post(`/marathon/${gameId}/bonus`, {
    bonusType: 'skip',
    questionId: currentQuestion.id
  })

  // Backend sends next question directly
  currentQuestion.value = data.nextQuestion
  bonusInventory.value = data.bonusInventory
}
```

---

### AnswerFeedback.vue

> ⚠️ Не отдельный компонент. Inline alerts в `MarathonPlayView.vue`.

**Only does:**
- Display feedback from API response
- Show correct answer if wrong
- Animate lives lost (from API: `livesLabel`)

**Receives from backend:**
```json
{
  "isCorrect": false,
  "correctAnswerText": "1147 год",
  "feedbackMessage": "❌ Неправильно",
  "explanation": "Москва основана в 1147 году.",
  "lives": 2,
  "livesLabel": "❤️❤️🖤",
  "livesLost": 1
}
```

**Renders:**
```vue
<div class="feedback">
  <div class="message">{{ data.feedbackMessage }}</div>
  <div class="correct-answer">Правильный ответ: {{ data.correctAnswerText }}</div>
  <div class="explanation">{{ data.explanation }}</div>
  <div class="lives">{{ data.livesLabel }}</div>
</div>
```

---

### GameOverScreen.vue

> ✅ Реализован как `MarathonGameOverView.vue`.

**Only does:**
- Display final score from API
- Show continue offer (if available)
- Handle continue payment

**Receives from backend:**
```json
{
  "gameOverData": {
    "finalScore": 47,
    "personalBest": 87,
    "isNewRecord": false,
    "weeklyRank": 342,
    "continueOffer": {
      "available": true,
      "costCoins": 200,
      "hasAd": true,
      "message": "Хочешь продолжить?"
    }
  }
}
```

**Continue flow:**
```typescript
const continueGame = async (method: 'coins' | 'ad') => {
  if (method === 'ad') {
    await showRewardedAd()
  }

  const { data } = await api.post(`/marathon/${gameId}/continue`, {
    paymentMethod: method
  })

  // Backend handles:
  // - Coin deduction
  // - Lives reset
  // - Next continue cost increment

  if (data.success) {
    router.push(`/marathon/play/${gameId}`)  // Resume
  }
}
```

---

### ResultsScreen.vue

> ⚠️ Не отдельный компонент. `MarathonGameOverView.vue` совмещает GameOver и Results.

**Only does:**
- Display final stats from API
- Show personal best comparison
- Show weekly rank

**Receives:**
```json
{
  "finalScore": 47,
  "totalQuestions": 50,
  "personalBest": 87,
  "isNewRecord": false,
  "newRecordBonus": 0,
  "weeklyRank": 342,
  "weeklyRankLabel": "#342 из 5,847",
  "bonusesUsed": {
    "shield": 2,
    "fiftyFifty": 0,
    "skip": 1,
    "freeze": 3
  },
  "continueCount": 1
}
```

**Just renders** (no calculations).

---

## State Management

### Use Vue Query

```typescript
// ✅ Server as single source of truth
const { data: status } = useQuery({
  queryKey: ['marathonStatus'],
  queryFn: () => api.get('/marathon/status')
})

// Status contains everything:
// - hasActiveGame
// - personalBest
// - weeklyRank
// - bonusInventory
```

### Local State (UI only)

```typescript
// ✅ ONLY UI state
const uiState = reactive({
  isShieldButtonGlowing: false,
  showBonusTooltip: false,
  confettiActive: false,
  selectedAnswerId: null  // Before submit
})

// ❌ NO game state
// Don't store: lives, score, bonuses locally
```

---

## API Response Structure

### Render-Ready Data

Backend includes UI labels:

**Bad:**
```json
{
  "lives": 2
}
// Frontend calculates: "❤️❤️🖤"
```

**Good:**
```json
{
  "lives": 2,
  "livesLabel": "❤️❤️🖤"
}
```

**Bad:**
```json
{
  "score": 47,
  "totalQuestions": 50
}
// Frontend calculates: "✅ 47"
```

**Good:**
```json
{
  "score": 47,
  "totalQuestions": 50,
  "scoreLabel": "✅ 47",
  "questionNumber": 50
}
```

---

## Error Handling

### Actionable Errors

```json
{
  "error": {
    "code": "INSUFFICIENT_BONUSES",
    "message": "У вас нет щитов",
    "action": {
      "type": "show_offer",
      "offerId": "emergency_bonus_pack"
    }
  }
}
```

**Frontend:**
```typescript
if (error.action?.type === 'show_offer') {
  showOfferModal(error.action.offerId)
}
```

### Insufficient Coins

```json
{
  "error": {
    "code": "INSUFFICIENT_COINS",
    "required": 200,
    "current": 50,
    "action": {
      "type": "navigate",
      "route": "/shop"
    }
  }
}
```

**Frontend:**
```typescript
toast.error(`Недостаточно монет! Нужно: ${error.required}💰`)
router.push(error.action.route)
```

---

## Real-Time Updates

### Leaderboard Polling

```typescript
useQuery({
  queryKey: ['marathonLeaderboard', 'weekly'],
  queryFn: () => api.get('/marathon/leaderboard', {
    params: { type: 'weekly' }
  }),
  refetchInterval: 30000  // 30s
})
```

### Personal Rank Updates

After game completion:
```typescript
await submitAnswer(lastAnswer)

// Invalidate queries
queryClient.invalidateQueries(['marathonStatus'])
queryClient.invalidateQueries(['marathonLeaderboard'])
```

---

## Performance Optimization

### Prefetch Next Question

```typescript
// While player reads feedback (2s), prefetch next question
watchEffect(() => {
  if (showingFeedback.value && nextQuestion.value) {
    // Next question already in response, no prefetch needed
    // Just prepare images if any
    preloadImages(nextQuestion.value.imageUrl)
  }
})
```

### Cache Inventory

```typescript
// Cache bonus inventory for 10s
useQuery({
  queryKey: ['bonusInventory'],
  queryFn: fetchInventory,
  staleTime: 10_000,
  cacheTime: 60_000
})

// Invalidate after bonus usage
await useBonus(...)
queryClient.invalidateQueries(['bonusInventory'])
```

---

## Testing Frontend

### Mock API Responses

```typescript
// tests/marathon.spec.ts
const mockGameOver = {
  isGameOver: true,
  gameOverData: {
    finalScore: 47,
    continueOffer: {
      available: true,
      costCoins: 200
    }
  }
}

vi.spyOn(api, 'post').mockResolvedValue({ data: mockGameOver })

// Test rendering only
await submitAnswer('a_001')
expect(screen.getByText('Финальный счёт: 47')).toBeInTheDocument()
```

### NO Logic Tests

```typescript
// ❌ BAD: Testing game logic
test('loses life on wrong answer', () => {
  expect(calculateLives(3, false)).toBe(2)
})

// ✅ GOOD: Testing rendering
test('displays lives from API', () => {
  render(<LivesDisplay livesLabel="❤️❤️🖤" />)
  expect(screen.getByText('❤️❤️🖤')).toBeInTheDocument()
})
```

---

## Anti-Patterns

### ❌ Don't duplicate backend logic

```typescript
// ❌ BAD
function shouldShowContinue(lives: number, gameStatus: string) {
  return lives === 0 && gameStatus === 'game_over'
}

// ✅ GOOD
const shouldShowContinue = response.gameOverData?.continueOffer.available
```

### ❌ Don't track lives locally

```typescript
// ❌ BAD
const lives = ref(3)
function loseLife() {
  lives.value--
}

// ✅ GOOD
const lives = computed(() => gameState.value?.lives)
```

### ❌ Don't implement bonus logic

```typescript
// ❌ BAD
function use5050(answers: Answer[]) {
  const wrongAnswers = answers.filter(a => !a.isCorrect)
  return wrongAnswers.slice(0, 2)  // Remove 2 wrong
}

// ✅ GOOD
const visibleAnswers = response.data.remainingAnswers  // Backend decided
```

---

---

## Additional UI Components

### MilestoneProgress.vue

> ⚠️ Inline в `MarathonPlayView.vue`, не отдельный компонент.

Displays progress toward next milestone (25, 50, 100, 200, 500).

**Receives from backend:**
```json
{
  "milestone": {
    "next": 50,
    "current": 47,
    "remaining": 3,
    "label": "Следующая цель: 50 ✅ (ещё 3)"
  }
}
```

**Renders:** Inline progress text below question area. Non-intrusive.

---

### PersonalBestProgress.vue

> ❌ Не реализован.

Visual progress bar comparing current score to personal best.

**Receives from backend:**
```json
{
  "personalBest": 87,
  "currentScore": 47,
  "progressPercent": 54,
  "progressLabel": "47/87 рекорда"
}
```

**Used in:** Game Over screen, Results screen.

---

### DifficultyTransition.vue

> ✅ Реализован. Backend возвращает `difficultyChanged`/`difficultyMessage` в question DTO; frontend toast component показывает уведомление.

Brief toast notification when timer limit changes.

**Receives from backend (in question response):**
```json
{
  "difficultyChanged": true,
  "difficultyMessage": "⚡ Сложность растёт! Время: 12 сек"
}
```

**Behavior:** Show for 2 seconds, auto-dismiss, non-blocking.

---

### OnboardingOverlay.vue

> ❌ Не реализован.

Highlights bonus buttons for first-time players.

**Receives from backend:**
```json
{
  "isOnboarding": true,
  "onboardingStep": 2,
  "onboardingHint": "Нажми 🛡️ чтобы защититься от ошибки",
  "highlightBonus": "shield"
}
```

**Behavior:** Animated pulse on target button + tooltip text. Player can dismiss or ignore.

---

### NetworkOverlay.vue

> ❌ Не реализован.

Shown on connection loss during game.

**Triggered by:** Frontend detecting network failure (no backend involvement).

```vue
<div class="network-overlay" v-if="isDisconnected">
  <div class="spinner" />
  <p>🔄 Переподключение...</p>
  <p>Твой прогресс сохранён. Таймер на паузе.</p>
</div>
```

**Behavior:** Blocks all interaction. Auto-dismisses on reconnect.

---

### ShareCard

> ❌ Не реализован.

Format for "Поделиться" button on Results screen.

**Generated by backend:**
```json
{
  "shareText": "🏃 Мой марафон в Quiz Sprint!\n✅ 47 правильных ответов\n🏆 #127 на этой неделе\nПопробуй побить мой рекорд!",
  "shareUrl": "https://quiz-sprint-tma.online/marathon"
}
```

**Frontend:** Uses Telegram `shareUrl` API or clipboard fallback.

---

## Checklist

Frontend complete when:

- [ ] NO lives calculation in components
- [ ] NO bonus logic (50/50 removal, shield effect)
- [ ] NO score tracking locally
- [ ] All game state from API responses
- [ ] Bonuses trigger API calls, not local changes
- [ ] Timer is visual only (server validates)
- [ ] Continue flow via API
- [ ] Tests only cover rendering
- [ ] Vue Query for server state
- [ ] Error handling uses API action hints
- [ ] Milestone progress displayed from API
- [ ] Personal best progress bar on game over / results
- [ ] Difficulty transition toasts
- [ ] Bonus tooltips on long-press
- [ ] First-time onboarding overlay
- [ ] Network disconnect overlay
- [ ] Share card functionality
