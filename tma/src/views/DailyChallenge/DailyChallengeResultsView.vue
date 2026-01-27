<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useAuth } from '@/composables/useAuth'
import { useStreaks } from '@/composables/useStreaks'
import DailyChallengeLeaderboard from '@/components/DailyChallenge/DailyChallengeLeaderboard.vue'

// ===========================
// Auth & Router
// ===========================

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

// ===========================
// Daily Challenge Composable
// ===========================

const {
  results,
  game,
  streak,
  isCompleted,
  timeToExpireFormatted,
  initialize
} = useDailyChallenge(playerId)

const streaks = useStreaks(streak)

const scorePercentage = computed(() => {
  if (!results.value) return 0
  return Math.round((results.value.correctAnswers / results.value.totalQuestions) * 100)
})

const performanceLevel = computed(() => {
  const pct = scorePercentage.value
  if (pct >= 90) return { label: 'Excellent!', color: 'green', emoji: '🌟' }
  if (pct >= 70) return { label: 'Great!', color: 'blue', emoji: '👏' }
  if (pct >= 50) return { label: 'Good!', color: 'yellow', emoji: '👍' }
  return { label: 'Keep trying!', color: 'gray', emoji: '💪' }
})

const hasNewStreakRecord = computed(() => {
  if (!streak.value) return false
  return streak.value.currentStreak > streak.value.bestStreak
})

// ===========================
// Methods
// ===========================

const handleReviewAnswers = () => {
  router.push({ name: 'daily-challenge-review' })
}

const handleGoHome = () => {
  router.push({ name: 'home' })
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  console.log('[DailyChallengeResults] onMounted', {
    isCompleted: isCompleted.value,
    hasResults: !!results.value
  })

  // Only initialize if we don't have results yet
  // (e.g., user refreshed the page)
  if (!results.value) {
    console.log('[DailyChallengeResults] No results in state, calling initialize...')
    await initialize()
  }

  // Redirect if game is not completed
  if (!isCompleted.value || !results.value) {
    console.log('[DailyChallengeResults] Redirecting to home - missing results')
    router.push({ name: 'home' })
  }
})
</script>

<template>
  <div class="results-container">
    <!-- Loading State -->
    <div v-if="!results" class="loading-container">
      <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
      <p class="text-gray-500 dark:text-gray-400 mt-4">Loading results...</p>
    </div>

    <!-- Results View -->
    <div v-else class="results-content">
      <!-- Header: Score Card -->
      <UCard class="score-card">
        <div class="score-content">
          <!-- Performance Level -->
          <div class="performance-badge">
            <span class="performance-emoji">{{ performanceLevel.emoji }}</span>
            <h2 class="performance-label">{{ performanceLevel.label }}</h2>
          </div>

          <!-- Score -->
          <div class="score-display">
            <div class="score-value">{{ game?.finalScore || 0 }}</div>
            <div class="score-label">points</div>
          </div>

          <!-- Accuracy -->
          <div class="accuracy-display">
            <UProgress :value="scorePercentage" :color="performanceLevel.color" size="lg" />
            <p class="accuracy-text">
              {{ results.correctAnswers }} / {{ results.totalQuestions }} correct
              <span class="accuracy-percentage">({{ scorePercentage }}%)</span>
            </p>
          </div>
        </div>
      </UCard>

      <!-- Streak Info (if new record) -->
      <UAlert
        v-if="hasNewStreakRecord"
        color="yellow"
        variant="soft"
        title="New Streak Record!"
        icon="i-heroicons-fire"
      >
        <template #description>
          <p>
            You've reached a {{ streak!.currentStreak }} day streak!
            {{ streaks.getStreakEmoji.value }}
          </p>
        </template>
      </UAlert>

      <!-- Rank Card -->
      <UCard>
        <div class="rank-content">
          <div class="rank-header">
            <UIcon name="i-heroicons-chart-bar" class="size-6 text-primary" />
            <h3 class="text-lg font-semibold">Your Ranking</h3>
          </div>
          <div class="rank-stats">
            <div class="rank-stat">
              <div class="stat-value">
                <UBadge color="primary" size="xl" variant="soft">
                  #{{ results.rank }}
                </UBadge>
              </div>
              <div class="stat-label">Your Rank</div>
            </div>
            <div class="rank-stat">
              <div class="stat-value text-2xl font-bold text-gray-700 dark:text-gray-300">
                {{ results.totalPlayers }}
              </div>
              <div class="stat-label">Total Players</div>
            </div>
          </div>
        </div>
      </UCard>

      <!-- Leaderboard -->
      <UCard>
        <DailyChallengeLeaderboard
          :leaderboard="results.leaderboard"
          :current-player-id="playerId"
          :max-entries="10"
        />
      </UCard>

      <!-- Action Buttons -->
      <div class="actions">
        <UButton
          color="primary"
          size="xl"
          icon="i-heroicons-document-text"
          block
          @click="handleReviewAnswers"
        >
          Review Answers
        </UButton>

        <UButton
          color="gray"
          size="xl"
          icon="i-heroicons-home"
          variant="outline"
          block
          @click="handleGoHome"
        >
          Back to Home
        </UButton>
      </div>

      <!-- Next Challenge Info -->
      <div class="next-challenge-info">
        <UIcon name="i-heroicons-calendar-days" class="size-5 text-gray-400" />
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Next challenge available in <span class="font-semibold">{{ timeToExpireFormatted }}</span>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.results-container {
  min-height: 100vh;
  padding: 1rem;
  padding-top: 6rem;
  padding-bottom: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
}

.results-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Score Card */
.score-card {
  background: linear-gradient(135deg, rgb(var(--color-primary-50)) 0%, rgb(var(--color-primary-100)) 100%);
}

.score-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  padding: 1rem;
}

.performance-badge {
  text-align: center;
}

.performance-emoji {
  font-size: 3rem;
  display: block;
  margin-bottom: 0.5rem;
}

.performance-label {
  font-size: 1.5rem;
  font-weight: 700;
  color: rgb(var(--color-gray-900));
}

.score-display {
  text-align: center;
}

.score-value {
  font-size: 4rem;
  font-weight: 900;
  color: rgb(var(--color-primary-600));
  line-height: 1;
}

.score-label {
  font-size: 0.875rem;
  color: rgb(var(--color-gray-600));
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-top: 0.25rem;
}

.accuracy-display {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.accuracy-text {
  text-align: center;
  font-size: 0.875rem;
  color: rgb(var(--color-gray-700));
  font-weight: 600;
}

.accuracy-percentage {
  color: rgb(var(--color-gray-500));
}

/* Rank Card */
.rank-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.rank-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgb(var(--color-gray-200));
}

.rank-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  padding: 1rem 0;
}

.rank-stat {
  text-align: center;
}

.stat-value {
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.875rem;
  color: rgb(var(--color-gray-500));
}

/* Actions */
.actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1rem;
}

/* Next Challenge Info */
.next-challenge-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  text-align: center;
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .score-card {
    background: linear-gradient(135deg, rgb(var(--color-primary-900) / 0.3) 0%, rgb(var(--color-primary-800) / 0.3) 100%);
  }

  .performance-label {
    color: rgb(var(--color-gray-100));
  }

  .score-value {
    color: rgb(var(--color-primary-400));
  }

  .score-label {
    color: rgb(var(--color-gray-400));
  }

  .accuracy-text {
    color: rgb(var(--color-gray-300));
  }

  .rank-header {
    border-bottom-color: rgb(var(--color-gray-700));
  }
}

/* Mobile optimizations */
@media (max-width: 640px) {
  .results-container {
    padding: 0.75rem;
    padding-top: 5rem;
  }

  .score-value {
    font-size: 3rem;
  }

  .performance-emoji {
    font-size: 2.5rem;
  }

  .rank-stats {
    gap: 1rem;
  }
}
</style>
