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
  timeToExpire,
  questionIndex,
  totalQuestions,
  isPlaying,
  isCompleted,
  isLoading,
  hasPlayed,
  canPlay,
  progress,
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

  // Refresh status periodically to keep time accurate
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

const cardTitle = computed(() => {
  if (hasPlayed.value) return 'Daily Challenge - Completed'
  if (isPlaying.value) return 'Daily Challenge - In Progress'
  return 'Daily Challenge'
})

const cardDescription = computed(() => {
  if (hasPlayed.value) {
    return 'Come back tomorrow for a new challenge!'
  }
  if (isPlaying.value) {
    return `Question ${questionIndex.value + 1} of ${totalQuestions.value}`
  }
  return '10 questions, one chance per day'
})

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

const statusBadge = computed(() => {
  if (hasPlayed.value) {
    return {
      label: 'Completed',
      color: 'green' as const,
      icon: 'i-heroicons-check-circle'
    }
  }
  if (isPlaying.value) {
    return {
      label: 'In Progress',
      color: 'blue' as const,
      icon: 'i-heroicons-clock'
    }
  }
  return {
    label: 'Available',
    color: 'yellow' as const,
    icon: 'i-heroicons-sparkles'
  }
})

// ===========================
// Actions
// ===========================

const handleClick = async () => {
  if (isLoading.value) return

  if (hasPlayed.value) {
    // Показать результаты
    router.push({ name: 'daily-challenge-results' })
  } else if (isPlaying.value) {
    // Продолжить игру
    router.push({ name: 'daily-challenge-play' })
  } else if (canPlay.value) {
    // Начать игру
    try {
      await startGame()
      // После успешного старта, перейти на экран игры
      router.push({ name: 'daily-challenge-play' })
    } catch (error) {
      console.error('Failed to start game:', error)
    }
  }
}

const progressDaily = computed(
  () =>{ return hasPlayed.value ? 100 : progress.value}
)

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  try {
    await initialize()
    startCountdown()
  } catch (error) {
    console.error('Failed to initialize Daily Challenge:', error)
    // Continue anyway - the UI will show appropriate states
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
        <div class="flex items-center gap-3">
          <UIcon name="i-heroicons-calendar-days" class="size-6 text-primary" />
          <div>
            <h3 class="text-lg font-semibold">{{ cardTitle }}</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ cardDescription }}
            </p>
          </div>
        </div>
        <UBadge
          :color="statusBadge.color"
          :icon="statusBadge.icon"
          size="lg"
        >
          {{ statusBadge.label }}
        </UBadge>
      </div>
    </template>

    <!-- Body -->
    <div class="space-y-4">
      <!-- Progress Bar (если игра в процессе или завершена) -->
      <div v-if="isPlaying || hasPlayed" class="space-y-2">
        <div class="flex justify-between text-sm">
          <span class="font-medium">Progress</span>
          <span class="text-gray-500">{{ hasPlayed ? 100 : progress }}%</span>
        </div>
        <UProgress v-model="progressDaily"  />
      </div>

      <!-- Streak Info -->
      <div v-if="streak" class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <span class="text-2xl">{{ streaks.getStreakEmoji.value }}</span>
          <div>
            <p class="text-sm font-medium">{{ streaks.formattedStreak.value }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Current Streak
            </p>
          </div>
        </div>

        <!-- Next Milestone -->
        <div v-if="streaks.nextMilestone.value" class="flex-1">
          <div class="flex justify-between text-xs mb-1">
            <span class="text-gray-500">Next: {{ streaks.nextMilestoneInfo.value?.label }}</span>
            <span class="text-gray-500">{{ streaks.daysToNextMilestone.value }} days</span>
          </div>
          <UProgress
            :value="streaks.progressToNextMilestone.value"
            :color="streaks.getProgressColor.value.replace('bg-', '')"
            size="xs"
          />
        </div>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-2 gap-4 pt-2 border-t border-gray-200 dark:border-gray-700">
        <!-- Countdown to Reset -->
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Resets in</p>
          <p class="text-sm font-mono font-semibold">
            <UIcon name="i-heroicons-clock" class="inline size-4" />
            {{ timeToExpireFormatted }}
          </p>
        </div>

        <!-- Players Today -->
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Players Today</p>
          <p class="text-sm font-semibold">
            <UIcon name="i-heroicons-user-group" class="inline size-4" />
            {{ totalPlayers }}
          </p>
        </div>
      </div>

      <!-- Score (если завершено) -->
      <div v-if="hasPlayed && game" class="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
        <p class="text-sm text-gray-600 dark:text-gray-300 mb-1">Your Score</p>
        <p class="text-2xl font-bold text-green-600 dark:text-green-400">
          {{ game?.finalScore || 0 }} points
        </p>
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
