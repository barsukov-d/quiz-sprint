<script setup lang="ts">
import { computed, onMounted, ref, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useStreaks } from '@/composables/useStreaks'

interface Props {
  playerId: string
}

const props = defineProps<Props>()
const router = useRouter()

// ===========================
// Composables
// ===========================

const {
  game,
  streak,
  totalPlayers,
  questionIndex,
  totalQuestions,
  isPlaying,
  isLoading,
  hasPlayed,
  canPlay,
  timeToExpireFormatted,
  startGame,
  checkStatus,
  initialize
} = useDailyChallenge(props.playerId)

const streaks = useStreaks(streak)

// ===========================
// Countdown Timer
// ===========================

const countdownInterval = ref<number | null>(null)

const startCountdown = () => {
  if (countdownInterval.value) return

  countdownInterval.value = window.setInterval(() => {
    checkStatus()
  }, 60000) // Refresh every minute
}

const stopCountdown = () => {
  if (countdownInterval.value) {
    clearInterval(countdownInterval.value)
    countdownInterval.value = null
  }
}

// ===========================
// Computed
// ===========================

const buttonText = computed(() => {
  if (isPlaying.value) return 'Continue'
  if (hasPlayed.value) return 'View Results'
  return 'Start Challenge'
})

const buttonIcon = computed(() => {
  if (isPlaying.value) return 'i-heroicons-play'
  if (hasPlayed.value) return 'i-heroicons-chart-bar'
  return 'i-heroicons-play'
})

const buttonColor = computed(() => {
  if (hasPlayed.value) return 'gray'
  return 'primary'
})

const statusIcon = computed(() => {
  if (hasPlayed.value) return 'i-heroicons-check-circle'
  if (isPlaying.value) return 'i-heroicons-clock'
  return null
})

const questionProgress = computed(() =>
  Math.round((questionIndex.value / totalQuestions.value) * 100)
)

// ===========================
// Actions
// ===========================

const handleClick = async () => {
  if (isLoading.value) return

  if (hasPlayed.value) {
    router.push({ name: 'daily-challenge-results' })
  } else if (isPlaying.value) {
    router.push({ name: 'daily-challenge-play' })
  } else if (canPlay.value) {
    try {
      await startGame()
      router.push({ name: 'daily-challenge-play' })
    } catch (error) {
      console.error('Failed to start game:', error)
    }
  }
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  try {
    await initialize()
    startCountdown()
  } catch (error) {
    console.error('Failed to initialize Daily Challenge:', error)
  }
})

onBeforeUnmount(() => {
  stopCountdown()
})
</script>

<template>
  <UCard>
    <!-- Header -->
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <UIcon name="i-heroicons-calendar-days" class="size-5 text-primary" />
          <h3 class="text-base font-semibold">Today's Challenge</h3>
        </div>
        <UIcon
          v-if="statusIcon"
          :name="statusIcon"
          :class="[
            'size-5',
            hasPlayed ? 'text-green-500' : 'text-blue-500'
          ]"
        />
      </div>
    </template>

    <!-- Body -->
    <div class="space-y-4">
      <!-- Completed State: Score + Streak -->
      <div v-if="hasPlayed && game" class="text-center space-y-2 py-2">
        <p class="text-3xl font-bold text-green-600 dark:text-green-400">
          {{ game.finalScore || 0 }} points
        </p>
        <p v-if="streak" class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ streaks.formattedStreak.value }}
        </p>
      </div>

      <!-- In Progress State: Question Progress -->
      <div v-else-if="isPlaying" class="space-y-2">
        <div class="flex justify-between text-sm">
          <span class="font-medium">Question {{ questionIndex + 1 }}/{{ totalQuestions }}</span>
          <span class="text-gray-500 dark:text-gray-400">{{ questionProgress }}%</span>
        </div>
        <UProgress v-model="questionProgress" color="primary" size="sm" />
      </div>

      <!-- Not Played State: Info + Streak -->
      <div v-else class="space-y-2 py-1">
        <p class="text-sm text-gray-700 dark:text-gray-300">
          10 questions • 15s each
        </p>
        <p v-if="streak" class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ streaks.formattedStreak.value }}
        </p>
      </div>

      <!-- Meta Info (always shown) -->
      <div class="grid grid-cols-2 gap-4 pt-3 border-t border-gray-200 dark:border-gray-700">
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Resets in</p>
          <p class="text-sm font-mono font-semibold tabular-nums">
            <UIcon name="i-heroicons-clock" class="inline size-3.5" />
            {{ timeToExpireFormatted }}
          </p>
        </div>

        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Players Today</p>
          <p class="text-sm font-semibold">
            <UIcon name="i-heroicons-user-group" class="inline size-3.5" />
            {{ totalPlayers }}
          </p>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <UButton
        :icon="buttonIcon"
        :color="buttonColor"
        :loading="isLoading"
        :disabled="!canPlay && !hasPlayed && !isPlaying"
        block
        size="lg"
        @click="handleClick"
      >
        {{ buttonText }}
      </UButton>
    </template>
  </UCard>
</template>
