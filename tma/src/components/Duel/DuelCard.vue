<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePvPDuel } from '@/composables/usePvPDuel'

interface Props {
  playerId: string
}

const props = defineProps<Props>()
const router = useRouter()

// ===========================
// Composables
// ===========================

const {
  tickets,
  pendingChallenges,
  hasActiveDuel,
  mmr,
  leagueLabel,
  leagueIcon,
  seasonWins,
  seasonLosses,
  winRate,
  isLoading,
  canPlay,
  goToActiveDuel,
  initialize,
} = usePvPDuel(props.playerId)

// ===========================
// Computed
// ===========================

const buttonText = computed(() => {
  if (hasActiveDuel.value) return 'Continue Duel'
  if (pendingChallenges.value.length > 0) return `Challenges (${pendingChallenges.value.length})`
  return 'Find Opponent'
})

const buttonIcon = computed(() => {
  if (hasActiveDuel.value) return 'i-heroicons-play'
  if (pendingChallenges.value.length > 0) return 'i-heroicons-bell-alert'
  return 'i-heroicons-magnifying-glass'
})

const buttonColor = computed(() => {
  if (pendingChallenges.value.length > 0) return 'orange'
  return 'primary'
})

const totalGames = computed(() => seasonWins.value + seasonLosses.value)

const winRateFormatted = computed(() => {
  if (totalGames.value === 0) return '0%'
  return `${Math.round(winRate.value)}%`
})

// ===========================
// Actions
// ===========================

const handleClick = () => {
  if (hasActiveDuel.value) {
    goToActiveDuel()
  } else {
    router.push({ name: 'duel-lobby' })
  }
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  try {
    await initialize()
  } catch (error) {
    console.error('Failed to initialize PvP Duel:', error)
  }
})
</script>

<template>
  <UCard>
    <!-- Header -->
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <UIcon name="i-heroicons-bolt" class="size-5 text-orange-500" />
          <h3 class="text-base font-semibold">PvP Duel</h3>
        </div>
        <UBadge v-if="pendingChallenges.length > 0" color="orange" variant="soft" size="sm">
          {{ pendingChallenges.length }} challenges
        </UBadge>
      </div>
    </template>

    <!-- Body -->
    <div class="space-y-4">
      <!-- Player Rating -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-2xl">{{ leagueIcon }}</span>
          <div>
            <p class="font-semibold">{{ leagueLabel }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ mmr }} MMR</p>
          </div>
        </div>
        <div class="text-right">
          <p class="text-sm font-medium">
            <span class="text-green-600 dark:text-green-400">{{ seasonWins }}W</span>
            <span class="text-gray-400 mx-1">/</span>
            <span class="text-red-600 dark:text-red-400">{{ seasonLosses }}L</span>
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ winRateFormatted }} win rate</p>
        </div>
      </div>

      <!-- Meta Info -->
      <div class="grid grid-cols-2 gap-4 pt-3 border-t border-gray-200 dark:border-gray-700">
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Tickets</p>
          <p class="text-sm font-semibold">
            <UIcon name="i-heroicons-ticket" class="inline size-3.5 text-primary" />
            {{ tickets }}
          </p>
        </div>

        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">This Season</p>
          <p class="text-sm font-semibold">
            <UIcon name="i-heroicons-trophy" class="inline size-3.5 text-yellow-500" />
            {{ totalGames }} games
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
        :disabled="!canPlay && !hasActiveDuel && pendingChallenges.length === 0"
        block
        size="lg"
        @click="handleClick"
      >
        {{ buttonText }}
      </UButton>
    </template>
  </UCard>
</template>
