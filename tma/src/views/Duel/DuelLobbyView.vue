<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { usePvPDuel } from '@/composables/usePvPDuel'

const router = useRouter()
const { currentUser } = useAuth()

const playerId = computed(() => currentUser.value?.id ?? '')

const {
  tickets,
  friendsOnline,
  pendingChallenges,
  hasActiveDuel,
  activeMatchId,
  mmr,
  leagueLabel,
  leagueIcon,
  seasonWins,
  seasonLosses,
  winRate,
  leaderboard,
  playerRank,
  matchHistory,
  isSearching,
  searchTime,
  isLoading,
  canPlay,
  joinQueue,
  leaveQueue,
  sendChallenge,
  respondChallenge,
  createChallengeLink,
  goToActiveDuel,
  refetchStatus,
  refetchLeaderboard,
  refetchHistory,
} = usePvPDuel(playerId.value)

// ===========================
// UI State
// ===========================

const activeTab = ref('play')
const showChallengeLink = ref(false)
const challengeLink = ref('')

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

const handleChallengeFriend = async (friendId: string) => {
  await sendChallenge(friendId)
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  await refetchStatus()
  await refetchLeaderboard()
  await refetchHistory()

  // If has active duel, redirect
  if (hasActiveDuel.value && activeMatchId.value) {
    goToActiveDuel()
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

      <!-- Challenge Link -->
      <UCard>
        <h3 class="font-semibold mb-3">Challenge a Friend</h3>
        <UButton
          icon="i-heroicons-link"
          color="gray"
          variant="soft"
          block
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
              <UBadge v-if="friend.inMatch" size="xs" color="orange">In Match</UBadge>
            </div>
            <UButton
              v-if="!friend.inMatch"
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
      <div v-if="matchHistory.length === 0" class="text-center py-8 text-gray-500">
        No matches yet
      </div>
      <div v-else class="space-y-2">
        <UCard
          v-for="match in matchHistory"
          :key="match.matchId"
          class="!p-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <UIcon
                :name="match.result === 'win' ? 'i-heroicons-trophy' : 'i-heroicons-x-circle'"
                :class="match.result === 'win' ? 'text-green-500' : 'text-red-500'"
                class="size-6"
              />
              <div>
                <p class="font-medium">vs {{ match.opponent }}</p>
                <p class="text-xs text-gray-500">
                  {{ match.playerScore }} - {{ match.opponentScore }}
                </p>
              </div>
            </div>
            <div class="text-right">
              <p
                :class="match.mmrChange! >= 0 ? 'text-green-600' : 'text-red-600'"
                class="font-semibold"
              >
                {{ match.mmrChange! >= 0 ? '+' : '' }}{{ match.mmrChange }} MMR
              </p>
              <UBadge v-if="match.isFriendMatch" size="xs" color="blue" variant="soft">
                Friend
              </UBadge>
            </div>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
