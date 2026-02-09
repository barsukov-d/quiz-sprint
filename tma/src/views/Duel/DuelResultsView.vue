<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { usePvPDuel } from '@/composables/usePvPDuel'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()

const duelId = computed(() => route.params.duelId as string)
const playerId = computed(() => currentUser.value?.id ?? '')

const {
  matchHistory,
  requestRematch,
  refetchHistory,
  isLoading,
} = usePvPDuel(playerId.value)

// ===========================
// State
// ===========================

const rematchStatus = ref<'idle' | 'pending' | 'accepted' | 'declined'>('idle')
const rematchError = ref<string | null>(null)

// ===========================
// Computed
// ===========================

// Find this match in history
const matchData = computed(() => {
  return matchHistory.value.find((m) => m.matchId === duelId.value)
})

const didWin = computed(() => matchData.value?.result === 'win')
const isDraw = computed(() => matchData.value?.result === 'draw')

const resultIcon = computed(() => {
  if (isDraw.value) return 'i-heroicons-minus-circle'
  return didWin.value ? 'i-heroicons-trophy' : 'i-heroicons-x-circle'
})

const resultColor = computed(() => {
  if (isDraw.value) return 'text-gray-500'
  return didWin.value ? 'text-yellow-500' : 'text-red-500'
})

const resultText = computed(() => {
  if (isDraw.value) return 'Draw!'
  return didWin.value ? 'Victory!' : 'Defeat'
})

const mmrChange = computed(() => matchData.value?.mmrChange ?? 0)
const mmrChangeColor = computed(() => (mmrChange.value >= 0 ? 'text-green-600' : 'text-red-600'))

// ===========================
// Actions
// ===========================

const handleRematch = async () => {
  try {
    rematchStatus.value = 'pending'
    rematchError.value = null

    const result = await requestRematch(duelId.value)

    if (result?.status === 'accepted') {
      rematchStatus.value = 'accepted'
      // Will be redirected by the composable
    } else {
      rematchStatus.value = 'pending'
    }
  } catch (error) {
    console.error('Rematch failed:', error)
    rematchStatus.value = 'idle'
    rematchError.value = 'Failed to request rematch'
  }
}

const handleBackToLobby = () => {
  router.push({ name: 'duel-lobby' })
}

const handleHome = () => {
  router.push({ name: 'home' })
}

const handleShare = () => {
  // TODO: Share victory card
  const text = didWin.value
    ? `I just won a PvP Duel! ${matchData.value?.playerScore} - ${matchData.value?.opponentScore}`
    : `Just finished a PvP Duel: ${matchData.value?.playerScore} - ${matchData.value?.opponentScore}`

  if (navigator.share) {
    navigator.share({
      title: 'Quiz Sprint Duel',
      text,
    })
  }
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  await refetchHistory()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex flex-col">
    <!-- Header -->
    <div class="px-4 py-3 flex items-center justify-between">
      <button class="p-2 -ml-2" @click="handleHome">
        <UIcon name="i-heroicons-x-mark" class="size-6" />
      </button>
      <h1 class="text-lg font-semibold">Match Result</h1>
      <div class="w-10" />
    </div>

    <!-- Result Hero -->
    <div class="flex-1 flex flex-col items-center justify-center p-6">
      <!-- Result Icon -->
      <div class="mb-6">
        <UIcon :name="resultIcon" :class="resultColor" class="size-24" />
      </div>

      <!-- Result Text -->
      <h2 class="text-4xl font-bold mb-4">{{ resultText }}</h2>

      <!-- Score -->
      <div class="flex items-center gap-6 mb-6">
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">You</p>
          <p class="text-5xl font-bold text-primary">
            {{ matchData?.playerScore ?? 0 }}
          </p>
        </div>
        <span class="text-2xl text-gray-400">-</span>
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ matchData?.opponent ?? 'Opponent' }}
          </p>
          <p class="text-5xl font-bold text-orange-500">
            {{ matchData?.opponentScore ?? 0 }}
          </p>
        </div>
      </div>

      <!-- MMR Change -->
      <div class="mb-8">
        <p :class="mmrChangeColor" class="text-2xl font-bold">
          {{ mmrChange >= 0 ? '+' : '' }}{{ mmrChange }} MMR
        </p>
      </div>

      <!-- Friend Match Badge -->
      <UBadge
        v-if="matchData?.isFriendMatch"
        color="blue"
        variant="soft"
        size="lg"
        class="mb-6"
      >
        Friend Match
      </UBadge>
    </div>

    <!-- Actions -->
    <div class="p-4 space-y-3">
      <!-- Rematch Button -->
      <UButton
        v-if="rematchStatus === 'idle'"
        icon="i-heroicons-arrow-path"
        color="primary"
        size="xl"
        block
        :loading="isLoading"
        @click="handleRematch"
      >
        Request Rematch
      </UButton>

      <UButton
        v-else-if="rematchStatus === 'pending'"
        icon="i-heroicons-clock"
        color="gray"
        variant="soft"
        size="xl"
        block
        disabled
      >
        Waiting for opponent...
      </UButton>

      <UAlert
        v-if="rematchError"
        color="red"
        variant="soft"
        class="mb-2"
      >
        {{ rematchError }}
      </UAlert>

      <!-- Share Button -->
      <UButton
        icon="i-heroicons-share"
        color="gray"
        variant="soft"
        size="lg"
        block
        @click="handleShare"
      >
        Share Result
      </UButton>

      <!-- Back to Lobby -->
      <UButton
        icon="i-heroicons-arrow-left"
        color="gray"
        variant="ghost"
        size="lg"
        block
        @click="handleBackToLobby"
      >
        Back to Lobby
      </UButton>
    </div>
  </div>
</template>
