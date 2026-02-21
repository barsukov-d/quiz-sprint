# Marathon Endless Mode — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove the time-gate between marathon runs and add instant restart so players can play as long as they want, renaming "lives" to "energy" throughout the UI.

**Architecture:** Minimal backend change (set `TimeToNextLife: 0` in mapper — the time-gate was never actually enforced in application code). All meaningful changes are on the frontend: new session composable, "Новый забег" button on game-over screen, cross-run streak, and ⚡ emoji replacing ❤️.

**Tech Stack:** Go (backend mapper), Vue 3 + TypeScript (frontend composables and views), Vitest (unit tests)

---

## Context You Must Know

- `RegenerateLives()` exists in domain but is **never called** anywhere in application/infrastructure — the 4-hour time-gate was never enforced in code. We only need to set `TimeToNextLife: 0` to stop the frontend timer from showing.
- Current "Play Again" in `MarathonGameOverView.vue` calls `reset()` + navigates to `home`, NOT to `marathon-category`. We're adding a "Новый забег" button that goes to `marathon-category` (starts a new run immediately).
- Session state is client-only (no backend changes needed for session tracking).
- Cross-run streak is a pure UI feature — it carries the last `streakCount` value from the composable across runs within a session.
- Run `go test ./...` from `backend/` to verify backend tests.
- Run `pnpm test:unit` from `tma/` to verify frontend tests.
- The test file for frontend lives at `tma/src/__tests__/App.spec.ts` — existing tests must stay green.

---

## Task 1: Backend — Remove Time-Gate from Lives DTO

**Files:**
- Modify: `backend/internal/application/marathon/mapper.go:102-109`

### Step 1: Write a test asserting TimeToNextLife is always 0

Add to `backend/internal/application/marathon/use_cases_test.go` (or create a new file `backend/internal/application/marathon/mapper_test.go`):

```go
package marathon_test

import (
	"testing"
	"time"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

func TestToLivesDTO_TimeToNextLifeIsAlwaysZero(t *testing.T) {
	now := time.Now().Unix()

	// Player lost 2 lives — normally would show a timer
	lives := solo_marathon.ReconstructLivesSystem(3, now-7200)

	dto := ToLivesDTO(lives, now)

	if dto.TimeToNextLife != 0 {
		t.Errorf("Expected TimeToNextLife=0 (no time-gate), got %d", dto.TimeToNextLife)
	}
}
```

### Step 2: Run test — verify it fails

```bash
cd backend && go test ./internal/application/marathon/... -run TestToLivesDTO_TimeToNextLifeIsAlwaysZero -v
```

Expected: `FAIL` — TimeToNextLife is currently non-zero for lost lives.

### Step 3: Fix the mapper

In `backend/internal/application/marathon/mapper.go`, change `ToLivesDTO`:

```go
// ToLivesDTO converts LivesSystem to DTO
// TimeToNextLife is always 0 — no time-gate between runs (instant restart available)
func ToLivesDTO(lives solo_marathon.LivesSystem, now int64) LivesDTO {
	return LivesDTO{
		CurrentLives:   lives.CurrentLives(),
		MaxLives:       lives.MaxLives(),
		TimeToNextLife: 0,
		Label:          lives.Label(),
	}
}
```

Remove the `now int64` parameter usage (keep the parameter signature to avoid breaking callers).

### Step 4: Run test — verify it passes

```bash
cd backend && go test ./internal/application/marathon/... -v
```

Expected: `PASS`

### Step 5: Run full backend test suite

```bash
cd backend && go test ./...
```

Expected: all tests pass.

### Step 6: Commit

```bash
git add backend/internal/application/marathon/mapper.go backend/internal/application/marathon/mapper_test.go
git commit -m "fix(marathon): set TimeToNextLife=0, remove time-gate between runs"
```

---

## Task 2: Frontend — Create Session Composable

**Files:**
- Create: `tma/src/composables/useMarathonSession.ts`

Session tracks: current run number, best score this session, cross-run streak (carries across runs until a wrong answer breaks it).

### Step 1: Write tests for the session composable

Create `tma/src/__tests__/useMarathonSession.spec.ts`:

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { useMarathonSession } from '@/composables/useMarathonSession'

describe('useMarathonSession', () => {
  it('starts with zero runs and zero best', () => {
    const session = useMarathonSession()
    expect(session.runCount.value).toBe(0)
    expect(session.sessionBest.value).toBe(0)
  })

  it('recordRunResult increments runCount', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    expect(session.runCount.value).toBe(1)
  })

  it('recordRunResult tracks session best', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    session.recordRunResult(23, 5)
    expect(session.sessionBest.value).toBe(47)
  })

  it('motivational prompt mentions deficit to record when close', () => {
    const session = useMarathonSession()
    session.recordRunResult(75, 10)
    // Second run: score 63, personal best 87
    const prompt = session.getMotivationalPrompt(63, 87)
    expect(prompt).toContain('24')  // 87 - 63 = 24
  })

  it('resetSession zeroes all state', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    session.resetSession()
    expect(session.runCount.value).toBe(0)
    expect(session.sessionBest.value).toBe(0)
  })
})
```

### Step 2: Run tests — verify they fail

```bash
cd tma && pnpm test:unit -- useMarathonSession
```

Expected: `FAIL` — module not found.

### Step 3: Implement the composable

Create `tma/src/composables/useMarathonSession.ts`:

```typescript
import { ref, computed } from 'vue'

const runCount = ref(0)
const sessionBest = ref(0)

// Module-level singletons so session persists across component remounts within same app session
export function useMarathonSession() {
  const recordRunResult = (score: number, _streak: number) => {
    runCount.value++
    if (score > sessionBest.value) {
      sessionBest.value = score
    }
  }

  const getMotivationalPrompt = (currentScore: number, personalBest: number | null): string => {
    if (personalBest && personalBest > currentScore) {
      const deficit = personalBest - currentScore
      return `До рекорда ${deficit} ответов. Ещё один забег?`
    }
    if (currentScore > 0 && (!personalBest || currentScore >= personalBest)) {
      return 'Новый рекорд! Сможешь побить его снова?'
    }
    if (runCount.value >= 2) {
      return `Забег #${runCount.value} — лучший в сессии: ${sessionBest.value}`
    }
    return 'Ещё один забег?'
  }

  const resetSession = () => {
    runCount.value = 0
    sessionBest.value = 0
  }

  const sessionLabel = computed(() =>
    runCount.value > 0 ? `Забег #${runCount.value} | Лучший: ${sessionBest.value}` : null
  )

  return {
    runCount,
    sessionBest,
    sessionLabel,
    recordRunResult,
    getMotivationalPrompt,
    resetSession,
  }
}
```

### Step 4: Run tests — verify they pass

```bash
cd tma && pnpm test:unit -- useMarathonSession
```

Expected: `PASS`

### Step 5: Commit

```bash
git add tma/src/composables/useMarathonSession.ts tma/src/__tests__/useMarathonSession.spec.ts
git commit -m "feat(marathon): add useMarathonSession composable for multi-run session tracking"
```

---

## Task 3: Frontend — Add "Новый забег" Button + Session Stats to Game Over Screen

**Files:**
- Modify: `tma/src/views/Marathon/MarathonGameOverView.vue`

### Step 1: Read the current file

Open `tma/src/views/Marathon/MarathonGameOverView.vue`. Understand that `handlePlayAgain()` currently calls `reset()` then `router.push({ name: 'home' })`.

### Step 2: Update the script section

Replace the `<script setup>` block. Key changes:
1. Import `useMarathonSession`
2. Add `handleStartNewRun()` that records result then navigates to `marathon-category`
3. Call `recordRunResult()` once on mount (guard with flag to avoid double-call)
4. Compute `motivationalPrompt`

```typescript
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useMarathonSession } from '@/composables/useMarathonSession'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

const {
  state,
  isLoading,
  canContinue,
  continueOffer,
  continueGame,
  reset,
  initialize,
} = useMarathon(playerId)

const session = useMarathonSession()
const resultRecorded = ref(false)

const gameOverResult = computed(() => state.value.gameOverResult)

const motivationalPrompt = computed(() =>
  session.getMotivationalPrompt(
    gameOverResult.value?.finalScore ?? state.value.score,
    state.value.personalBest,
  )
)

const handleContinueWithCoins = async () => {
  try {
    await continueGame('coins')
  } catch (error) {
    console.error('Failed to continue with coins:', error)
  }
}

const handleContinueWithAd = async () => {
  try {
    await continueGame('ad')
  } catch (error) {
    console.error('Failed to continue with ad:', error)
  }
}

const handleStartNewRun = () => {
  reset()
  router.push({ name: 'marathon-category' })
}

const handleBackToHome = () => {
  reset()
  session.resetSession()
  router.push({ name: 'home' })
}

onMounted(async () => {
  await initialize()

  if (!state.value.gameOverResult && state.value.status !== 'game-over') {
    router.push({ name: 'home' })
    return
  }

  // Record this run result in session (once only)
  if (!resultRecorded.value) {
    session.recordRunResult(
      gameOverResult.value?.finalScore ?? state.value.score,
      state.value.streakCount,
    )
    resultRecorded.value = true
  }
})
</script>
```

### Step 3: Update the template

Replace the `<!-- Actions -->` section and add session stats + prompt:

```html
<!-- Session Stats (if 2+ runs) -->
<div v-if="session.runCount.value >= 2" class="w-full text-center text-sm text-gray-500 dark:text-gray-400">
  {{ session.sessionLabel.value }}
</div>

<!-- Motivational Prompt -->
<div class="w-full text-center text-sm font-medium text-primary">
  {{ motivationalPrompt }}
</div>

<!-- Actions -->
<div class="w-full flex flex-col gap-2">
  <UButton
    color="primary"
    block
    size="lg"
    icon="i-heroicons-bolt"
    @click="handleStartNewRun"
  >
    ▶ Новый забег
  </UButton>

  <UButton
    color="gray"
    variant="ghost"
    block
    size="lg"
    icon="i-heroicons-home"
    @click="handleBackToHome"
  >
    На главную
  </UButton>
</div>
```

Remove the old `handlePlayAgain` button.

### Step 4: Verify UI manually

Start dev server: `cd tma && pnpm dev`

Play a marathon run until game over. Verify:
- "Новый забег" button appears and navigates to category selection
- Session stats appear after second run ("Забег #2 | Лучший: 47")
- "На главную" resets session and goes home

### Step 5: Run tests

```bash
cd tma && pnpm test:unit
```

Expected: all existing tests pass.

### Step 6: Commit

```bash
git add tma/src/views/Marathon/MarathonGameOverView.vue
git commit -m "feat(marathon): add instant new run button and session stats to game-over screen"
```

---

## Task 4: Frontend — Replace ❤️ with ⚡ and Remove Life Timer

**Files:**
- Modify: `tma/src/components/Marathon/MarathonCard.vue`
- Modify: `tma/src/views/Marathon/MarathonPlayView.vue`

### Step 1: Update MarathonCard.vue

**Change 1 — Remove timer logic** (lines ~38–76). Delete:
- `timerInterval` ref
- `timeToLifeRestore` ref
- `timeToLifeRestoreFormatted` computed
- `startTimer()` function
- `stopTimer()` function
- `showLifeTimer` computed
- `onBeforeUnmount(() => stopTimer())` hook
- `startTimer()` call in `onMounted`

**Change 2 — Update `livesLabel` computed** — replace ❤️ with ⚡:

```typescript
const livesLabel = computed(() => {
  if (isPlaying.value || isGameOver.value) {
    return lives.value.label
      .replace(/❤️/g, '⚡')
      .replace(/🖤/g, '○')
  }
  return '⚡'.repeat(lives.value.maxLives)
})
```

**Change 3 — Remove life timer display** from template. Delete this block:
```html
<!-- Life restore timer -->
<div v-if="showLifeTimer" ...>
  ...
</div>
```

**Change 4 — Update rules hint text**:

```html
<!-- Rules hint (only when no personal best = likely new player) -->
<div v-if="state.personalBest === null || state.personalBest === 0" ...>
  <p>5 энергии, ошибка = −1 ⚡</p>
  <p>5 правильных подряд = +1 ⚡</p>
  <p>Сложность растёт со временем</p>
</div>
```

**Change 5 — Update header title attribute**:

```html
<span class="text-lg" :title="`${lives.currentLives}/${lives.maxLives} энергии`">
```

### Step 2: Update MarathonPlayView.vue — livesDisplay emoji

Find the template where `livesDisplay` is rendered (hearts loop). Change ❤️ → ⚡ and 🖤 → ○:

Search for the template section that renders `livesDisplay` array. It likely renders something like:
```html
<span v-for="(alive, i) in livesDisplay" :key="i">
  {{ alive ? '❤️' : '🖤' }}
</span>
```

Change to:
```html
<span v-for="(alive, i) in livesDisplay" :key="i">
  {{ alive ? '⚡' : '○' }}
</span>
```

Also update the label attribute if present: `title="lives"` → `title="энергия"`.

### Step 3: Run tests

```bash
cd tma && pnpm test:unit
```

Expected: all tests pass.

### Step 4: Visual check

Navigate to marathon screen in dev server. Verify:
- No life timer countdown visible anywhere
- Lives display shows ⚡ instead of ❤️
- Rules hint shows "5 энергии" text

### Step 5: Commit

```bash
git add tma/src/components/Marathon/MarathonCard.vue tma/src/views/Marathon/MarathonPlayView.vue
git commit -m "feat(marathon): replace lives/hearts with energy/lightning emoji, remove time-gate timer"
```

---

## Task 5: Documentation Update

**Files:**
- Modify: `docs/game_modes/solo_marathon/01_concept.md`
- Modify: `docs/game_modes/solo_marathon/02_gameplay.md`

### Step 1: Update 01_concept.md

**Key Mechanics table** — change lives references:

```markdown
| Parameter | Value |
|-----------|-------|
| Questions | Endless (until energy runs out) |
| Starting energy | 5 ⚡ |
| Time per question | 15s → 8s (adaptive) |
| Wrong answer penalty | −1 ⚡ |
| Energy regen | +1 ⚡ every 5 correct in a row (Marathon Momentum) |
| Run over | 0 ⚡ → Continue (coins/ad) OR instant new run |
| Score | Correct answers count (best run per week) |
```

**Lives System section** — rename to "Energy System":

```markdown
### 1. Energy System
- Start with 5 ⚡⚡⚡⚡⚡
- Wrong answer = −1 ⚡
- 5 correct in a row = +1 ⚡ (Marathon Momentum)
- 0 ⚡ = run over → instant free restart OR pay to continue
- **NO waiting** between runs
```

**Remove** the line: `- **NO life regeneration** (except continue)`

**Continue section** — update:
```markdown
### 4. Continue Mechanic (optional monetization)
At 0 energy:
- **Continue:** 200 coins OR Rewarded Ad → energy resets to 1 ⚡ (resume same run)
- **New run:** Free → 5 ⚡ fresh start (best score from either run counts)
```

### Step 2: Update 02_gameplay.md

**Pre-Start Screen wireframe** — change ❤️ to ⚡, update rules text:

```
│  Правила:                           │
│  • 5 ⚡ энергии, ошибка = −1 ⚡      │
│  • 5 правильных подряд = +1 ⚡       │
│  • Сложность растёт со временем     │
```

**Add between-run screen** after section 5 (Game Over Screen):

```markdown
### 5b. Between-Run Screen (after declining Continue)
┌─────────────────────────────────────┐
│  🏁 Забег завершён                  │
│                                     │
│  ✅ 47 правильных                   │
│  🔥 Лучшая серия: 12                │
│                                     │
│  Эта сессия:                        │
│  Забег #2 | Лучший: 47              │
│                                     │
│  До рекорда 40 ответов. Ещё один?  │
│                                     │
│  [ ⚡ Новый забег    ]              │
│  [ 📊 Лидерборд     ]              │
│  [ 🚪 На главную    ]              │
└─────────────────────────────────────┘
```

**State Management section** — update lives to energy:

```
Lives remaining (0-5)  →  Energy remaining (0-5)
```

### Step 3: Commit

```bash
git add docs/game_modes/solo_marathon/01_concept.md docs/game_modes/solo_marathon/02_gameplay.md
git commit -m "docs(marathon): update concept and gameplay for energy system + instant restart"
```

---

## Task 6: Final Verification

### Step 1: Run all backend tests

```bash
cd backend && go test ./...
```

Expected: all tests pass.

### Step 2: Run all frontend tests

```bash
cd tma && pnpm test:unit
```

Expected: all tests pass.

### Step 3: Full flow test (manual)

1. Start backend: `cd backend && docker compose -f docker-compose.dev.yml up`
2. Start frontend: `cd tma && pnpm dev`
3. Navigate to Marathon
4. Verify ⚡ icons (not ❤️) in MarathonCard
5. Verify no "Next life in X:XX" timer
6. Play until game over (0 energy)
7. Verify "Новый забег" button appears alongside Continue offer
8. Click "Новый забег" → verify navigates to category, starts fresh with 5 ⚡
9. Play second run → verify "Забег #2 | Лучший: 47" appears on game-over screen
10. Click "На главную" → verify session resets

### Step 4: Commit if any cleanup needed, then push

```bash
git push origin marathon-update
```

---

## What Was Intentionally Left Out (Phase 2)

- Daily missions (separate feature, own task list)
- Best streak leaderboard (requires new backend endpoint)
- Cross-run streak persisted to backend (currently frontend-only)
- Motivational push notifications
