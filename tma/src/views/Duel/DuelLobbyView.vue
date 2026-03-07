<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { usePvPDuel } from '@/composables/usePvPDuel'
import { usePostDuelChallengeAcceptByCode } from '@/api/generated/hooks/duelController/usePostDuelChallengeAcceptByCode'
import { useI18n } from 'vue-i18n'
import { FEATURES } from '@/features'

const router = useRouter()
const route = useRoute()
const { currentUser } = useAuth()

const playerId = computed(() => currentUser.value?.id ?? '')
const { t } = useI18n()

const {
	tickets,
	pendingChallenges,
	outgoingReadyChallenges,
	outgoingPendingChallenges,
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
	rivals,
	isSearching,
	searchTime,
	isLoading,
	canPlay,
	joinQueue,
	leaveQueue,
	sendChallenge,
	respondChallenge,
	shareChallengeToTelegram,
	startChallenge,
	goToActiveDuel,
	refetchStatus,
	refetchLeaderboard,
	refetchHistory,
	refetchRivals,
} = usePvPDuel(playerId.value)

// Accept by link code mutation
const { mutateAsync: acceptByLinkCode, isPending: isAcceptingChallenge } =
	usePostDuelChallengeAcceptByCode()

// ===========================
// UI State
// ===========================

const activeTab = ref('play')
const showChallengeLink = ref(false)
const challengeLink = ref('')
const deepLinkChallenge = ref<string | null>(null)
const deepLinkError = ref<string | null>(null)
const showConfirmModal = ref(false)
const pendingLinkCode = ref<string | null>(null)

function navigateToHome() {
	router.push({ name: 'home' })
}

function setActiveTab(tab: string) {
	activeTab.value = tab
}

function dismissDeepLinkError() {
	deepLinkError.value = null
	router.replace({ name: 'duel-lobby' })
}

// ===========================
// Computed
// ===========================

const searchTimeFormatted = computed(() => {
	const minutes = Math.floor(searchTime.value / 60)
	const seconds = searchTime.value % 60
	return `${minutes}:${seconds.toString().padStart(2, '0')}`
})

const formatExpiresIn = (seconds: number): string => {
	const hours = Math.floor(seconds / 3600)
	const minutes = Math.floor((seconds % 3600) / 60)
	if (hours > 0) return `${hours}ч ${minutes}мин`
	return `${minutes}мин`
}

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

const handleCopyLink = () => {
	navigator.clipboard.writeText(challengeLink.value)
}

const isSharing = ref(false)
const sendingChallengeId = ref<string | null>(null)

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
	sendingChallengeId.value = friendId
	try {
		await sendChallenge(friendId)
	} finally {
		sendingChallengeId.value = null
	}
}

const handleConfirmChallenge = async () => {
	if (!pendingLinkCode.value) return
	showConfirmModal.value = false
	await handleAcceptByLinkCode(pendingLinkCode.value)
	pendingLinkCode.value = null
}

const handleDeclineChallengeModal = () => {
	showConfirmModal.value = false
	pendingLinkCode.value = null
	router.replace({ name: 'duel-lobby' })
}

// ===========================
// Deep Link Handling
// ===========================

const handleAcceptByLinkCode = async (linkCode: string) => {
	if (!playerId.value) {
		deepLinkError.value = t('duel.pleaseLogin')
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
			console.log('[DuelLobby] Challenge accepted, waiting for inviter to start...')
			deepLinkChallenge.value = null
			router.replace({ name: 'duel-lobby' })
			await refetchStatus()
			if (hasActiveDuel.value && activeGameId.value) {
				goToActiveDuel()
			}
		}
	} catch (error: unknown) {
		console.error('[DuelLobby] Failed to accept challenge:', error)
		deepLinkError.value = t('duel.acceptFailed')
	}
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	await refetchStatus()
	await refetchLeaderboard()
	await refetchHistory()
	await refetchRivals()

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
		// Show confirmation modal instead of auto-accepting
		pendingLinkCode.value = challengeCode
		showConfirmModal.value = true
	}
})
</script>

<template>
	<div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
		<!-- Header -->
		<div class="flex items-center justify-between mb-6">
			<button class="p-2 -ml-2" @click="navigateToHome">
				<UIcon name="i-heroicons-arrow-left" class="size-6" />
			</button>
			<h1 class="text-xl font-bold">{{ t('duel.title') }}</h1>
			<div class="w-10" />
		</div>

		<!-- Deep Link Challenge Loading -->
		<UCard v-if="isAcceptingChallenge" class="mb-4">
			<div class="flex items-center justify-center gap-3 py-4">
				<div class="animate-spin">
					<UIcon name="i-heroicons-arrow-path" class="size-6 text-primary" />
				</div>
				<p class="font-medium">{{ t('duel.acceptingChallenge') }}</p>
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
						@click="dismissDeepLinkError"
					>
						{{ t('duel.close') }}
					</UButton>
				</div>
			</div>
		</UCard>

		<!-- Confirmation Modal -->
		<UModal v-model:open="showConfirmModal" :dismissible="false">
			<template #content>
				<div class="p-6 text-center">
					<div class="text-4xl mb-4">⚔️</div>
					<h3 class="text-xl font-bold mb-2">{{ t('duel.incomingChallenge') }}</h3>
					<p class="text-gray-600 dark:text-gray-400 mb-6">
						{{ t('duel.wantsToFight') }}
					</p>
					<div class="space-y-3">
						<UButton
							block
							size="lg"
							color="primary"
							:loading="isAcceptingChallenge"
							@click="handleConfirmChallenge"
						>
							{{ t('duel.acceptChallenge') }}
						</UButton>
						<UButton
							block
							size="lg"
							color="gray"
							variant="ghost"
							@click="handleDeclineChallengeModal"
						>
							{{ t('duel.decline') }}
						</UButton>
					</div>
				</div>
			</template>
		</UModal>

		<!-- Player Rating Card -->
		<UCard class="mb-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<span class="text-4xl">{{ leagueIcon }}</span>
					<div>
						<p class="text-lg font-bold">{{ leagueLabel }}</p>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{{ t('duel.mmrValue', { mmr }) }}
						</p>
					</div>
				</div>
				<div class="text-right">
					<div class="flex items-center gap-1 text-sm">
						<UIcon name="i-heroicons-ticket" class="size-4 text-primary" />
						<span class="font-semibold">{{ tickets }}</span>
					</div>
					<p class="text-xs text-gray-500 dark:text-gray-400">{{ t('duel.tickets') }}</p>
				</div>
			</div>

			<!-- Stats -->
			<div
				class="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-gray-200 dark:border-gray-700"
			>
				<div class="text-center">
					<p class="text-lg font-bold text-green-600 dark:text-green-400">
						{{ seasonWins }}
					</p>
					<p class="text-xs text-gray-500 dark:text-gray-400">{{ t('duel.wins') }}</p>
				</div>
				<div class="text-center">
					<p class="text-lg font-bold text-red-600 dark:text-red-400">
						{{ seasonLosses }}
					</p>
					<p class="text-xs text-gray-500 dark:text-gray-400">{{ t('duel.losses') }}</p>
				</div>
				<div class="text-center">
					<p class="text-lg font-bold">{{ Math.round(winRate) }}%</p>
					<p class="text-xs text-gray-500 dark:text-gray-400">{{ t('duel.winRate') }}</p>
				</div>
			</div>
		</UCard>

		<!-- Pending Challenges -->
		<UCard v-if="pendingChallenges.length > 0" class="mb-4">
			<template #header>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2.5">
						<UIcon name="i-heroicons-bolt" class="size-5 text-orange-500" />
						<h3 class="text-base font-semibold">{{ t('duel.pendingChallenges') }}</h3>
					</div>
					<UBadge color="orange" variant="soft" size="sm">
						{{ t('duel.pendingCount', { count: pendingChallenges.length }) }}
					</UBadge>
				</div>
			</template>

			<div class="space-y-4">
				<div v-for="(challenge, index) in pendingChallenges" :key="challenge.id">
					<UDivider v-if="index > 0" class="mb-4" />
					<!-- Challenger identity -->
					<div class="flex items-center gap-4 mb-3">
						<div
							class="flex-shrink-0 w-12 h-12 rounded-lg flex items-center justify-center bg-orange-100 dark:bg-orange-900/30"
						>
							<UIcon name="i-heroicons-bolt" class="size-6 text-orange-500" />
						</div>
						<div>
							<p class="font-semibold">
								{{ challenge.challengerUsername || t('duel.challenge') }}
							</p>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								{{ t('duel.challengesYou') }}
							</p>
						</div>
					</div>
					<!-- Action buttons -->
					<div class="grid grid-cols-2 gap-3">
						<UButton
							color="green"
							block
							@click="() => handleAcceptChallenge(challenge.id!)"
						>
							{{ t('duel.accept') }}
						</UButton>
						<UButton
							color="red"
							variant="soft"
							block
							@click="() => handleDeclineChallenge(challenge.id!)"
						>
							{{ t('duel.decline') }}
						</UButton>
					</div>
				</div>
			</div>
		</UCard>

		<!-- Tabs -->
		<div class="flex gap-2 mb-4">
			<UButton
				:color="activeTab === 'play' ? 'primary' : 'gray'"
				:variant="activeTab === 'play' ? 'solid' : 'ghost'"
				size="sm"
				@click="() => setActiveTab('play')"
			>
				{{ t('duel.tabPlay') }}
			</UButton>
			<UButton
				:color="activeTab === 'leaderboard' ? 'primary' : 'gray'"
				:variant="activeTab === 'leaderboard' ? 'solid' : 'ghost'"
				size="sm"
				@click="() => setActiveTab('leaderboard')"
			>
				{{ t('duel.tabLeaderboard') }}
			</UButton>
			<UButton
				:color="activeTab === 'history' ? 'primary' : 'gray'"
				:variant="activeTab === 'history' ? 'solid' : 'ghost'"
				size="sm"
				@click="() => setActiveTab('history')"
			>
				{{ t('duel.tabHistory') }}
			</UButton>
		</div>

		<!-- Play Tab -->
		<div v-if="activeTab === 'play'" class="space-y-4">
			<!-- Find Match Button -->
			<UCard v-if="FEATURES.matchmaking" class="text-center">
				<div v-if="isSearching" class="py-4">
					<div class="animate-pulse mb-4">
						<UIcon name="i-heroicons-magnifying-glass" class="size-12 text-primary" />
					</div>
					<p class="text-lg font-semibold mb-1">{{ t('duel.searching') }}</p>
					<p class="text-2xl font-mono font-bold text-primary">
						{{ searchTimeFormatted }}
					</p>
					<UButton color="gray" variant="soft" class="mt-4" @click="handleFindMatch">
						{{ t('duel.cancel') }}
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
						{{ t('duel.findOpponent') }}
					</UButton>
					<p v-if="tickets === 0" class="text-sm text-red-500 mt-2">
						{{ t('duel.noTickets') }}
					</p>
				</div>
			</UCard>

			<!-- Outgoing Challenges -->
			<div v-if="outgoingReadyChallenges.length > 0" class="space-y-2">
				<h3 class="text-sm font-semibold text-gray-600 dark:text-gray-400">
					{{ t('duel.outgoingChallenges') }}
				</h3>

				<!-- Ready to start -->
				<UCard
					v-for="challenge in outgoingReadyChallenges"
					:key="challenge.id"
					class="border-green-200 dark:border-green-800"
				>
					<div class="flex items-center gap-2 mb-3">
						<span class="text-green-500">✅</span>
						<p class="font-medium">
							{{ challenge.inviteeName || t('duel.friend') }} {{ t('duel.isReady') }}
						</p>
					</div>
					<UButton
						color="green"
						block
						@click="
							() => {
								startChallenge(challenge.id!)
							}
						"
					>
						{{ t('duel.startDuel') }}
					</UButton>
				</UCard>
			</div>

			<!-- Outgoing pending link challenges (inviter waiting for someone to click) -->
			<div v-if="outgoingPendingChallenges.length > 0" class="space-y-2">
				<UCard
					v-for="challenge in outgoingPendingChallenges"
					:key="challenge.id"
					class="border-blue-200 dark:border-blue-800"
				>
					<div class="flex items-center gap-2 mb-1">
						<span class="text-blue-500">⏳</span>
						<p class="font-medium">{{ t('duel.waitingForResponse') }}</p>
					</div>
					<p class="text-sm text-gray-500 dark:text-gray-400">
						{{ t('duel.linkExpiresIn', { time: formatExpiresIn(challenge.expiresIn ?? 0) }) }}
					</p>
				</UCard>
			</div>

			<!-- Rivals Section -->
			<UCard>
				<h3 class="font-semibold mb-3">{{ t('duel.rivals') }}</h3>

				<!-- Rivals list -->
				<div v-if="rivals.length > 0" class="space-y-3 mb-4">
					<div v-for="rival in rivals" :key="rival.id" class="flex items-center gap-3">
						<!-- Avatar with online indicator -->
						<div class="relative flex-shrink-0">
							<div
								class="w-10 h-10 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center text-lg"
							>
								{{ rival.leagueIcon }}
							</div>
							<span
								class="absolute bottom-0 right-0 w-3 h-3 rounded-full ring-2 ring-white dark:ring-gray-900"
								:class="
									rival.isOnline ? 'bg-green-500' : 'bg-gray-400 dark:bg-gray-600'
								"
							/>
						</div>

						<!-- Info -->
						<div class="flex-1 min-w-0">
							<p class="font-semibold truncate">{{ rival.username }}</p>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								{{ rival.mmr }} MMR
							</p>
						</div>

						<!-- Action button -->
						<UButton
							size="sm"
							:disabled="rival.hasPendingChallenge || sendingChallengeId === rival.id"
							:loading="sendingChallengeId === rival.id"
							:color="rival.hasPendingChallenge ? 'gray' : 'primary'"
							:variant="rival.hasPendingChallenge ? 'soft' : 'solid'"
							@click="
								() => {
									if (!rival.hasPendingChallenge && !sendingChallengeId)
										handleChallengeFriend(rival.id!)
								}
							"
						>
							{{
								rival.hasPendingChallenge
									? t('duel.challengeSent')
									: t('duel.challenge')
							}}
						</UButton>
					</div>
				</div>

				<!-- Empty state -->
				<p v-else class="text-sm text-gray-500 dark:text-gray-400 mb-4">
					{{ t('duel.noRivalsYet') }}
				</p>

				<!-- Divider -->
				<div class="flex items-center gap-3 my-3">
					<div class="flex-1 h-px bg-gray-200 dark:bg-gray-700" />
					<span class="text-xs text-gray-400">{{ t('duel.orInvite') }}</span>
					<div class="flex-1 h-px bg-gray-200 dark:bg-gray-700" />
				</div>

				<!-- Invite via Telegram -->
				<UButton
					icon="i-heroicons-paper-airplane"
					color="primary"
					block
					:loading="isSharing"
					@click="handleShareToTelegram"
				>
					{{ t('duel.inviteFriend') }}
				</UButton>

				<div
					v-if="showChallengeLink"
					class="mt-3 p-3 bg-gray-100 dark:bg-gray-800 rounded-lg"
				>
					<p class="text-xs text-gray-500 mb-2">{{ t('duel.shareLink') }}</p>
					<div class="flex gap-2">
						<input
							:value="challengeLink"
							readonly
							class="flex-1 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded px-2 py-1"
						/>
						<UButton size="xs" @click="handleCopyLink">{{ t('duel.copy') }}</UButton>
					</div>
				</div>
			</UCard>
		</div>

		<!-- Leaderboard Tab -->
		<div v-else-if="activeTab === 'leaderboard'">
			<UCard>
				<div
					v-if="playerRank > 0"
					class="mb-4 p-3 bg-primary-50 dark:bg-primary-900/20 rounded-lg"
				>
					<p class="text-sm text-gray-600 dark:text-gray-400">{{ t('duel.yourRank') }}</p>
					<p class="text-2xl font-bold text-primary">#{{ playerRank }}</p>
				</div>

				<div class="space-y-2">
					<div
						v-for="(entry, index) in leaderboard as any[]"
						:key="entry.playerId ?? index"
						class="flex items-center justify-between p-2 rounded"
						:class="
							entry.playerId === playerId
								? 'bg-primary-50 dark:bg-primary-900/20'
								: ''
						"
					>
						<div class="flex items-center gap-3">
							<span
								class="w-6 text-center font-bold"
								:class="index < 3 ? 'text-yellow-500' : ''"
							>
								{{ index + 1 }}
							</span>
							<span class="text-lg">{{ entry.leagueIcon ?? '' }}</span>
							<div>
								<p class="font-medium">{{ entry.username ?? 'Player' }}</p>
								<p class="text-xs text-gray-500">{{ entry.mmr ?? 0 }} MMR</p>
							</div>
						</div>
						<div class="text-right text-sm">
							<span class="text-green-600"
								>{{ entry.wins ?? 0 }}{{ t('duel.wIndicator') }}</span
							>
							<span class="text-gray-400 mx-1">/</span>
							<span class="text-red-600"
								>{{ entry.losses ?? 0 }}{{ t('duel.lIndicator') }}</span
							>
						</div>
					</div>
				</div>
			</UCard>
		</div>

		<!-- History Tab -->
		<div v-else-if="activeTab === 'history'">
			<div v-if="gameHistory.length === 0" class="text-center py-8 text-gray-500">
				{{ t('duel.noGames') }}
			</div>
			<div v-else class="space-y-2">
				<UCard v-for="game in gameHistory" :key="game.gameId" class="!p-3">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<UIcon
								:name="
									game.result === 'win'
										? 'i-heroicons-trophy'
										: 'i-heroicons-x-circle'
								"
								:class="game.result === 'win' ? 'text-green-500' : 'text-red-500'"
								class="size-6"
							/>
							<div>
								<p class="font-medium">
									{{ t('duel.vsOpponent', { name: game.opponent }) }}
								</p>
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
								{{
									t('duel.mmrChange', {
										sign: game.mmrChange! >= 0 ? '+' : '',
										amount: game.mmrChange,
									})
								}}
							</p>
							<UBadge v-if="game.isFriendGame" size="xs" color="blue" variant="soft">
								{{ t('duel.friendBadge') }}
							</UBadge>
						</div>
					</div>
				</UCard>
			</div>
		</div>
	</div>
</template>
