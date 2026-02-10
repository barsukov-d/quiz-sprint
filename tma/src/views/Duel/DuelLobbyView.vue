<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { usePvPDuel } from '@/composables/usePvPDuel'
import { usePostDuelChallengeAcceptByCode } from '@/api/generated/hooks/duelController/usePostDuelChallengeAcceptByCode'

const router = useRouter()
const route = useRoute()
const { currentUser } = useAuth()

const playerId = computed(() => currentUser.value?.id ?? '')

const {
  tickets,
  friendsOnline,
  pendingChallenges,
  hasActiveDuel,
  activeGameId,
  mmr,
  leagueLabel,
  leagueIcon,
  seasonWins,
  seasonLosses,
  winRate,
  leaderboard,
  playerRank,
  gameHistory,
  isSearching,
  searchTime,
  isLoading,
  canPlay,
  joinQueue,
  leaveQueue,
  sendChallenge,
  respondChallenge,
  createChallengeLink,
  shareChallengeToTelegram,
  goToActiveDuel,
  refetchStatus,
  refetchLeaderboard,
  refetchHistory,
} = usePvPDuel(playerId.value)

// Accept by link code mutation
const { mutateAsync: acceptByLinkCode, isPending: isAcceptingChallenge } = usePostDuelChallengeAcceptByCode()

// ===========================
// UI State
// ===========================

const activeTab = ref('play')
const showChallengeLink = ref(false)
const challengeLink = ref('')
const deepLinkChallenge = ref<string | null>(null)
const deepLinkError = ref<string | null>(null)

// ===========================
// Computed
// ===========================

const searchTimeFormatted = computed(() => {
  const minutes = Math.floor(searchTime.value / 60)
  const seconds = searchTime.value % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
})

// ===========================
// Actions
// ===========================

const handleFindMatch = async () => {
  if (isSearching.value) {
    await leaveQueue()
  } else {
    await joinQueue()
  }
}

const handleAcceptChallenge = async (challengeId: string) => {
  await respondChallenge(challengeId, 'accept')
}

const handleDeclineChallenge = async (challengeId: string) => {
  await respondChallenge(challengeId, 'decline')
}

const handleCreateLink = async () => {
  const result = await createChallengeLink()
  if (result?.challengeLink) {
    challengeLink.value = result.challengeLink
    showChallengeLink.value = true
  }
}

const handleCopyLink = () => {
  navigator.clipboard.writeText(challengeLink.value)
}

const isSharing = ref(false)

const handleShareToTelegram = async () => {
  try {
    isSharing.value = true
    await shareChallengeToTelegram()
  } catch (error) {
    console.error('Failed to share:', error)
  } finally {
    isSharing.value = false
  }
}

const handleChallengeFriend = async (friendId: string) => {
  await sendChallenge(friendId)
}

// ===========================
// Deep Link Handling
// ===========================

const handleAcceptByLinkCode = async (linkCode: string) => {
  if (!playerId.value) {
    deepLinkError.value = 'Пожалуйста, авторизуйтесь'
    return
  }

  try {
    console.log('[DuelLobby] Accepting challenge by link code:', linkCode)
    const response = await acceptByLinkCode({
      data: {
        playerId: playerId.value,
        linkCode,
      },
    })

    if (response.data?.success) {
      console.log('[DuelLobby] Challenge accepted, game starting...')
      deepLinkChallenge.value = null
      // Clear the query param
      router.replace({ name: 'duel-lobby' })
      // Refresh status to get the new game
      await refetchStatus()
      if (hasActiveDuel.value && activeGameId.value) {
        goToActiveDuel()
      }
    }
  } catch (error: unknown) {
    console.error('[DuelLobby] Failed to accept challenge:', error)
    deepLinkError.value = 'Не удалось принять вызов. Возможно, ссылка устарела.'
  }
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  await refetchStatus()
  await refetchLeaderboard()
  await refetchHistory()

  // If has active duel, redirect
  if (hasActiveDuel.value && activeGameId.value) {
    goToActiveDuel()
    return
  }

  // Check for deep link challenge
  const challengeCode = route.query.challenge as string
  if (challengeCode) {
    console.log('[DuelLobby] Deep link challenge detected:', challengeCode)
    deepLinkChallenge.value = challengeCode
    // Auto-accept the challenge
    await handleAcceptByLinkCode(challengeCode)
  }
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <button class="p-2 -ml-2" @click="router.push({ name: 'home' })">
        <UIcon name="i-heroicons-arrow-left" class="size-6" />
      </button>
      <h1 class="text-xl font-bold">PvP Duel</h1>
      <div class="w-10" />
    </div>

    <!-- Deep Link Challenge Loading -->
    <UCard v-if="isAcceptingChallenge" class="mb-4">
      <div class="flex items-center justify-center gap-3 py-4">
        <div class="animate-spin">
          <UIcon name="i-heroicons-arrow-path" class="size-6 text-primary" />
        </div>
        <p class="font-medium">Принимаем вызов...</p>
      </div>
    </UCard>

    <!-- Deep Link Error -->
    <UCard v-if="deepLinkError" class="mb-4 border-red-200 dark:border-red-800">
      <div class="flex items-center gap-3">
        <UIcon name="i-heroicons-exclamation-circle" class="size-6 text-red-500" />
        <div>
          <p class="font-medium text-red-600 dark:text-red-400">{{ deepLinkError }}</p>
          <UButton
            size="xs"
            color="gray"
            variant="link"
            class="mt-1"
            @click="deepLinkError = null; router.replace({ name: 'duel-lobby' })"
          >
            Закрыть
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Player Rating Card -->
    <UCard class="mb-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span class="text-4xl">{{ leagueIcon }}</span>
          <div>
            <p class="text-lg font-bold">{{ leagueLabel }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ mmr }} MMR</p>
          </div>
        </div>
        <div class="text-right">
          <div class="flex items-center gap-1 text-sm">
            <UIcon name="i-heroicons-ticket" class="size-4 text-primary" />
            <span class="font-semibold">{{ tickets }}</span>
          </div>
          <p class="text-xs text-gray-500 dark:text-gray-400">tickets</p>
        </div>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
        <div class="text-center">
          <p class="text-lg font-bold text-green-600 dark:text-green-400">{{ seasonWins }}</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">Wins</p>
        </div>
        <div class="text-center">
          <p class="text-lg font-bold text-red-600 dark:text-red-400">{{ seasonLosses }}</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">Losses</p>
        </div>
        <div class="text-center">
          <p class="text-lg font-bold">{{ Math.round(winRate) }}%</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">Win Rate</p>
        </div>
      </div>
    </UCard>

    <!-- Pending Challenges -->
    <div v-if="pendingChallenges.length > 0" class="mb-4">
      <h2 class="text-sm font-semibold text-gray-600 dark:text-gray-400 mb-2">
        Pending Challenges
      </h2>
      <div class="space-y-2">
        <UCard v-for="challenge in pendingChallenges" :key="challenge.id" class="!p-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-bolt" class="size-5 text-orange-500" />
              <span class="font-medium">Challenge</span>
            </div>
            <div class="flex gap-2">
              <UButton
                size="xs"
                color="green"
                @click="handleAcceptChallenge(challenge.id!)"
              >
                Accept
              </UButton>
              <UButton
                size="xs"
                color="red"
                variant="soft"
                @click="handleDeclineChallenge(challenge.id!)"
              >
                Decline
              </UButton>
            </div>
          </div>
        </UCard>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 mb-4">
      <UButton
        :color="activeTab === 'play' ? 'primary' : 'gray'"
        :variant="activeTab === 'play' ? 'solid' : 'ghost'"
        size="sm"
        @click="activeTab = 'play'"
      >
        Play
      </UButton>
      <UButton
        :color="activeTab === 'leaderboard' ? 'primary' : 'gray'"
        :variant="activeTab === 'leaderboard' ? 'solid' : 'ghost'"
        size="sm"
        @click="activeTab = 'leaderboard'"
      >
        Leaderboard
      </UButton>
      <UButton
        :color="activeTab === 'history' ? 'primary' : 'gray'"
        :variant="activeTab === 'history' ? 'solid' : 'ghost'"
        size="sm"
        @click="activeTab = 'history'"
      >
        History
      </UButton>
    </div>

    <!-- Play Tab -->
    <div v-if="activeTab === 'play'" class="space-y-4">
      <!-- Find Match Button -->
      <UCard class="text-center">
        <div v-if="isSearching" class="py-4">
          <div class="animate-pulse mb-4">
            <UIcon name="i-heroicons-magnifying-glass" class="size-12 text-primary" />
          </div>
          <p class="text-lg font-semibold mb-1">Searching for opponent...</p>
          <p class="text-2xl font-mono font-bold text-primary">{{ searchTimeFormatted }}</p>
          <UButton
            color="gray"
            variant="soft"
            class="mt-4"
            @click="handleFindMatch"
          >
            Cancel
          </UButton>
        </div>

        <div v-else class="py-2">
          <UButton
            icon="i-heroicons-magnifying-glass"
            size="xl"
            :disabled="!canPlay"
            :loading="isLoading"
            block
            @click="handleFindMatch"
          >
            Find Random Opponent
          </UButton>
          <p v-if="tickets === 0" class="text-sm text-red-500 mt-2">
            No tickets available
          </p>
        </div>
      </UCard>

      <!-- Invite Friend -->
      <UCard>
        <h3 class="font-semibold mb-3">Challenge a Friend</h3>

        <!-- Primary: Share via Telegram -->
        <UButton
          icon="i-heroicons-paper-airplane"
          color="primary"
          block
          :loading="isSharing"
          @click="handleShareToTelegram"
        >
          Invite Friend via Telegram
        </UButton>

        <!-- Secondary: Create link manually -->
        <UButton
          icon="i-heroicons-link"
          color="gray"
          variant="soft"
          block
          class="mt-2"
          @click="handleCreateLink"
        >
          Create Challenge Link
        </UButton>

        <div v-if="showChallengeLink" class="mt-3 p-3 bg-gray-100 dark:bg-gray-800 rounded-lg">
          <p class="text-xs text-gray-500 mb-2">Share this link:</p>
          <div class="flex gap-2">
            <input
              :value="challengeLink"
              readonly
              class="flex-1 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded px-2 py-1"
            />
            <UButton size="xs" @click="handleCopyLink">Copy</UButton>
          </div>
        </div>
      </UCard>

      <!-- Friends Online -->
      <UCard v-if="friendsOnline.length > 0">
        <h3 class="font-semibold mb-3">Friends Online</h3>
        <div class="space-y-2">
          <div
            v-for="friend in friendsOnline"
            :key="friend.id"
            class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 bg-green-500 rounded-full" />
              <span>{{ friend.username }}</span>
              <UBadge v-if="friend.inGame" size="xs" color="orange">In Game</UBadge>
            </div>
            <UButton
              v-if="!friend.inGame"
              size="xs"
              @click="handleChallengeFriend(friend.id!)"
            >
              Challenge
            </UButton>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Leaderboard Tab -->
    <div v-else-if="activeTab === 'leaderboard'">
      <UCard>
        <div v-if="playerRank > 0" class="mb-4 p-3 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
          <p class="text-sm text-gray-600 dark:text-gray-400">Your Rank</p>
          <p class="text-2xl font-bold text-primary">#{{ playerRank }}</p>
        </div>

        <div class="space-y-2">
          <div
            v-for="(entry, index) in (leaderboard as any[])"
            :key="entry.playerId ?? index"
            class="flex items-center justify-between p-2 rounded"
            :class="entry.playerId === playerId ? 'bg-primary-50 dark:bg-primary-900/20' : ''"
          >
            <div class="flex items-center gap-3">
              <span class="w-6 text-center font-bold" :class="index < 3 ? 'text-yellow-500' : ''">
                {{ index + 1 }}
              </span>
              <span class="text-lg">{{ entry.leagueIcon ?? '' }}</span>
              <div>
                <p class="font-medium">{{ entry.username ?? 'Player' }}</p>
                <p class="text-xs text-gray-500">{{ entry.mmr ?? 0 }} MMR</p>
              </div>
            </div>
            <div class="text-right text-sm">
              <span class="text-green-600">{{ entry.wins ?? 0 }}W</span>
              <span class="text-gray-400 mx-1">/</span>
              <span class="text-red-600">{{ entry.losses ?? 0 }}L</span>
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <!-- History Tab -->
    <div v-else-if="activeTab === 'history'">
      <div v-if="gameHistory.length === 0" class="text-center py-8 text-gray-500">
        No games yet
      </div>
      <div v-else class="space-y-2">
        <UCard
          v-for="game in gameHistory"
          :key="game.gameId"
          class="!p-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <UIcon
                :name="game.result === 'win' ? 'i-heroicons-trophy' : 'i-heroicons-x-circle'"
                :class="game.result === 'win' ? 'text-green-500' : 'text-red-500'"
                class="size-6"
              />
              <div>
                <p class="font-medium">vs {{ game.opponent }}</p>
                <p class="text-xs text-gray-500">
                  {{ game.playerScore }} - {{ game.opponentScore }}
                </p>
              </div>
            </div>
            <div class="text-right">
              <p
                :class="game.mmrChange! >= 0 ? 'text-green-600' : 'text-red-600'"
                class="font-semibold"
              >
                {{ game.mmrChange! >= 0 ? '+' : '' }}{{ game.mmrChange }} MMR
              </p>
              <UBadge v-if="game.isFriendGame" size="xs" color="blue" variant="soft">
                Friend
              </UBadge>
            </div>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
