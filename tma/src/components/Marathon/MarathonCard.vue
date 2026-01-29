<script setup lang="ts">
import { computed, onMounted, ref, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'

interface Props {
  playerId: string
}

const props = defineProps<Props>()
const router = useRouter()

// ===========================
// Composables
// ===========================

const {
  state,
  isPlaying,
  isLoading,
  hasLives,
  canPlay,
  progressToRecord,
  livesPercent,
  timeToLifeRestoreFormatted,
  checkStatus,
  initialize
} = useMarathon(props.playerId)

// ===========================
// Life Restore Timer
// ===========================

const timerInterval = ref<number | null>(null)

const startTimer = () => {
  if (timerInterval.value) return

  timerInterval.value = window.setInterval(() => {
    if (state.value.timeToLifeRestore > 0) {
      state.value.timeToLifeRestore--
    } else {
      // Жизнь восстановилась - обновляем статус
      checkStatus()
      if (timerInterval.value) {
        clearInterval(timerInterval.value)
        timerInterval.value = null
      }
    }
  }, 1000)
}

const stopTimer = () => {
  if (timerInterval.value) {
    clearInterval(timerInterval.value)
    timerInterval.value = null
  }
}

// ===========================
// Computed
// ===========================

const cardTitle = computed(() => {
  if (isPlaying.value) return 'Solo Marathon - In Progress'
  return 'Solo Marathon'
})

const cardDescription = computed(() => {
  if (isPlaying.value) {
    return `Score: ${state.value.score} | Streak: ${state.value.currentStreak}`
  }
  return 'Answer until first mistake'
})

const buttonText = computed(() => {
  if (isPlaying.value) return 'Continue Marathon'
  if (!hasLives.value) return 'No Lives'
  return 'Start Marathon'
})

const buttonIcon = computed(() => {
  if (isPlaying.value) return 'i-heroicons-play'
  return 'i-heroicons-bolt'
})

const statusBadge = computed(() => {
  if (isPlaying.value) {
    return {
      label: 'In Progress',
      color: 'blue' as const,
      icon: 'i-heroicons-bolt'
    }
  }
  if (!hasLives.value) {
    return {
      label: 'No Lives',
      color: 'red' as const,
      icon: 'i-heroicons-heart'
    }
  }
  return {
    label: 'Ready',
    color: 'green' as const,
    icon: 'i-heroicons-check-circle'
  }
})

// Визуализация жизней
const livesDisplay = computed(() => {
  const hearts = []
  for (let i = 0; i < state.value.maxLives; i++) {
    hearts.push(i < state.value.lives)
  }
  return hearts
})

// Цвет прогресс-бара жизней
const livesColor = computed(() => {
  if (livesPercent.value <= 33) return 'red'
  if (livesPercent.value <= 66) return 'orange'
  return 'green'
})

// ===========================
// Actions
// ===========================

const handleClick = () => {
  if (isLoading.value) return

  if (isPlaying.value) {
    // Продолжить игру
    router.push({ name: 'marathon-play' })
  } else if (canPlay.value) {
    // Выбор категории
    router.push({ name: 'marathon-category' })
  }
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  await initialize()
  startTimer()
})

onBeforeUnmount(() => {
  stopTimer()
})
</script>

<template>
  <UCard>
    <!-- Header -->
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <UIcon name="i-heroicons-bolt" class="size-6 text-yellow-500" />
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
      <!-- Lives Display -->
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Lives</span>
          <div class="flex gap-1">
            <UIcon
              v-for="(filled, index) in livesDisplay"
              :key="index"
              :name="filled ? 'i-heroicons-heart-solid' : 'i-heroicons-heart'"
              :class="filled ? 'text-red-500' : 'text-gray-300 dark:text-gray-600'"
              class="size-5"
            />
          </div>
        </div>
        <UProgress :value="livesPercent" :color="livesColor" />
      </div>

      <!-- Life Restore Timer (если не все жизни) -->
      <div v-if="state.lives < state.maxLives && state.timeToLifeRestore > 0" class="text-center py-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
        <p class="text-xs text-gray-600 dark:text-gray-300 mb-1">Next life in</p>
        <p class="text-sm font-mono font-semibold text-blue-600 dark:text-blue-400">
          <UIcon name="i-heroicons-clock" class="inline size-4" />
          {{ timeToLifeRestoreFormatted }}
        </p>
      </div>

      <!-- Personal Best -->
      <div v-if="state.personalBest !== null" class="space-y-2">
        <div class="flex items-center justify-between text-sm">
          <span class="font-medium">Personal Best</span>
          <UBadge color="yellow" variant="subtle">
            <UIcon name="i-heroicons-trophy" class="size-3" />
            {{ state.personalBest }}
          </UBadge>
        </div>

        <!-- Progress to Record (если играем) -->
        <div v-if="isPlaying && state.score > 0">
          <div class="flex justify-between text-xs mb-1">
            <span class="text-gray-500">Current: {{ state.score }}</span>
            <span class="text-gray-500">{{ progressToRecord }}%</span>
          </div>
          <UProgress
            :value="progressToRecord"
            :color="progressToRecord >= 100 ? 'green' : 'blue'"
            size="xs"
          />
          <p v-if="progressToRecord >= 100" class="text-xs text-green-600 dark:text-green-400 mt-1 text-center">
            🎉 New Record!
          </p>
        </div>
      </div>

      <!-- Current Game Stats (если играем) -->
      <div v-if="isPlaying" class="grid grid-cols-2 gap-4 pt-2 border-t border-gray-200 dark:border-gray-700">
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Score</p>
          <p class="text-lg font-bold text-primary">
            {{ state.score }}
          </p>
        </div>
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Streak</p>
          <p class="text-lg font-bold text-yellow-500">
            🎯 {{ state.currentStreak }}
          </p>
        </div>
      </div>

      <!-- Hints Available (если играем) -->
      <div v-if="isPlaying" class="flex gap-2 justify-center pt-2 border-t border-gray-200 dark:border-gray-700">
        <UChip :text="String(state.hints.fiftyFifty)" size="md">
          <UIcon name="i-heroicons-scissors" class="size-4" />
        </UChip>
        <UChip :text="String(state.hints.extraTime)" size="md">
          <UIcon name="i-heroicons-clock" class="size-4" />
        </UChip>
        <UChip :text="String(state.hints.skip)" size="md">
          <UIcon name="i-heroicons-forward" class="size-4" />
        </UChip>
        <UChip :text="String(state.hints.hint)" size="md">
          <UIcon name="i-heroicons-light-bulb" class="size-4" />
        </UChip>
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <UButton
        :icon="buttonIcon"
        :color="hasLives ? 'primary' : 'gray'"
        :loading="isLoading"
        :disabled="!canPlay && !isPlaying"
        block
        size="lg"
        @click="handleClick"
      >
        {{ buttonText }}
      </UButton>
    </template>
  </UCard>
</template>
