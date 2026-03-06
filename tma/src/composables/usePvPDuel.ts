import { computed, ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { shareURL } from '@tma.js/sdk'
import {
	useGetDuelStatus,
	useGetDuelLeaderboard,
	useGetDuelHistory,
	useGetDuelRivals,
	usePostDuelQueueJoin,
	useDeleteDuelQueueLeave,
	usePostDuelChallenge,
	usePostDuelChallengeChallengeidRespond,
	usePostDuelChallengeLink,
	usePostDuelGameGameidRematch,
	usePostDuelChallengeChallengeidStart,
} from '@/api/generated'

/**
 * Composable for PvP Duel game mode
 *
 * SERVER-SIDE STATE ARCHITECTURE:
 * - Backend is the single source of truth
 * - Frontend fetches fresh data from API
 * - Real-time updates via WebSocket (when in game)
 */
export function usePvPDuel(playerId: string) {
	const router = useRouter()

	// ===========================
	// Local UI State
	// ===========================
	const isSearching = ref(false)
	const searchTime = ref(0)
	const searchInterval = ref<ReturnType<typeof setInterval> | null>(null)

	// Polling interval when waiting for a friend to accept a challenge link
	let pollInterval: ReturnType<typeof setInterval> | null = null

	const startOutgoingPoll = () => {
		if (pollInterval) return
		pollInterval = setInterval(async () => {
			if (outgoingChallenges.value.length > 0) {
				await refetchStatus()
				if (hasActiveDuel.value && activeGameId.value) {
					stopOutgoingPoll()
					goToActiveDuel()
				}
			} else {
				stopOutgoingPoll()
			}
		}, 5000)
	}

	const stopOutgoingPoll = () => {
		if (pollInterval) {
			clearInterval(pollInterval)
			pollInterval = null
		}
	}

	// ===========================
	// API Hooks (Server State)
	// ===========================

	const joinQueueMutation = usePostDuelQueueJoin()
	const leaveQueueMutation = useDeleteDuelQueueLeave()
	const sendChallengeMutation = usePostDuelChallenge()
	const respondChallengeMutation = usePostDuelChallengeChallengeidRespond()
	const createLinkMutation = usePostDuelChallengeLink()
	const rematchMutation = usePostDuelGameGameidRematch()
	const startChallengeMutation = usePostDuelChallengeChallengeidStart()

	// Main status endpoint
	const {
		data: statusData,
		refetch: refetchStatus,
		isLoading: isLoadingStatus,
	} = useGetDuelStatus(
		computed(() => ({ playerId })),
		{
			query: {
				enabled: computed(() => !!playerId),
				refetchOnWindowFocus: true,
				staleTime: 0,
				retry: false,
				refetchOnMount: true,
			},
		},
	)

	// Leaderboard
	const { data: leaderboardData, refetch: refetchLeaderboard } = useGetDuelLeaderboard(
		computed(() => ({ playerId, type: 'seasonal', limit: 20 })),
		{
			query: {
				enabled: computed(() => !!playerId),
				staleTime: 60000, // 1 minute cache
			},
		},
	)

	// Match history
	const { data: historyData, refetch: refetchHistory } = useGetDuelHistory(
		computed(() => ({ playerId, limit: 20, offset: 0, filter: 'all' })),
		{
			query: {
				enabled: computed(() => !!playerId),
				staleTime: 30000, // 30 seconds cache
			},
		},
	)

	// Rivals (recent opponents)
	const { data: rivalsData, refetch: refetchRivals } = useGetDuelRivals(
		computed(() => ({ playerId })),
		{
			query: {
				enabled: computed(() => !!playerId),
				staleTime: 30000,
			},
		},
	)

	// ===========================
	// Computed Properties
	// ===========================

	// Status data
	const hasActiveDuel = computed(() => statusData.value?.data?.hasActiveDuel ?? false)
	const activeGameId = computed(() => statusData.value?.data?.activeGameId ?? null)
	const player = computed(() => statusData.value?.data?.player ?? null)
	const tickets = computed(() => statusData.value?.data?.tickets ?? 0)
	const friendsOnline = computed(() => statusData.value?.data?.friendsOnline ?? [])
	const pendingChallenges = computed(() => statusData.value?.data?.pendingChallenges ?? [])
	const outgoingChallenges = computed(() => statusData.value?.data?.outgoingChallenges ?? [])

	const outgoingReadyChallenges = computed(() =>
		outgoingChallenges.value.filter((c) => c.status === 'accepted_waiting_inviter'),
	)

	const outgoingPendingChallenges = computed(() =>
		outgoingChallenges.value.filter((c) => c.status === 'pending'),
	)

	const seasonId = computed(() => statusData.value?.data?.seasonId ?? '')
	const seasonEndsAt = computed(() => statusData.value?.data?.seasonEndsAt ?? 0)

	// Player rating info
	const mmr = computed(() => player.value?.mmr ?? 1000)
	const league = computed(() => player.value?.league ?? 'bronze')
	const leagueLabel = computed(() => player.value?.leagueLabel ?? 'Bronze IV')
	const leagueIcon = computed(() => player.value?.leagueIcon ?? '')
	const division = computed(() => player.value?.division ?? 4)
	const seasonWins = computed(() => player.value?.seasonWins ?? 0)
	const seasonLosses = computed(() => player.value?.seasonLosses ?? 0)
	const winRate = computed(() => player.value?.winRate ?? 0)

	// Leaderboard - handle both array and object response formats
	const leaderboard = computed(() => {
		const data = leaderboardData.value?.data as unknown
		if (!data) return []
		// If data has 'entries' property, use it; otherwise assume data is the array
		if (typeof data === 'object' && data !== null && 'entries' in data) {
			return (data as { entries: unknown[] }).entries ?? []
		}
		if (Array.isArray(data)) return data
		return []
	})
	const playerRank = computed(() => {
		const data = leaderboardData.value?.data as unknown
		if (!data) return 0
		if (typeof data === 'object' && data !== null && 'playerRank' in data) {
			return (data as { playerRank: number }).playerRank ?? 0
		}
		return 0
	})

	// History
	const gameHistory = computed(() => historyData.value?.data?.games ?? [])

	// Rivals
	const rivals = computed(() => rivalsData.value?.data?.rivals ?? [])

	// Loading states
	const isLoading = computed(
		() =>
			isLoadingStatus.value ||
			joinQueueMutation.isPending.value ||
			leaveQueueMutation.isPending.value ||
			sendChallengeMutation.isPending.value ||
			respondChallengeMutation.isPending.value,
	)

	// Can play (has tickets and not in game)
	const canPlay = computed(() => tickets.value > 0 && !hasActiveDuel.value && !isSearching.value)

	// Start/stop polling when outgoing challenges change
	watch(
		outgoingChallenges,
		(challenges) => {
			if (challenges.length > 0) {
				startOutgoingPoll()
			} else {
				stopOutgoingPoll()
			}
		},
		{ immediate: true },
	)

	onUnmounted(() => {
		stopOutgoingPoll()
	})

	// ===========================
	// Actions
	// ===========================

	/**
	 * Join matchmaking queue
	 */
	const joinQueue = async () => {
		try {
			console.log('[usePvPDuel] Joining queue...')

			isSearching.value = true
			searchTime.value = 0

			// Start search timer
			searchInterval.value = setInterval(() => {
				searchTime.value++
			}, 1000)

			const response = await joinQueueMutation.mutateAsync({
				data: { playerId },
			})

			console.log('[usePvPDuel] Joined queue:', response.data)

			// Poll for match found
			pollForMatch()

			return response.data
		} catch (error) {
			console.error('[usePvPDuel] Failed to join queue:', error)
			stopSearching()
			throw error
		}
	}

	/**
	 * Leave matchmaking queue
	 */
	const leaveQueue = async () => {
		try {
			console.log('[usePvPDuel] Leaving queue...')

			await leaveQueueMutation.mutateAsync({
				params: { playerId },
			})

			stopSearching()
			await refetchStatus()
		} catch (error) {
			console.error('[usePvPDuel] Failed to leave queue:', error)
			stopSearching()
			throw error
		}
	}

	/**
	 * Stop searching state
	 */
	const stopSearching = () => {
		isSearching.value = false
		searchTime.value = 0
		if (searchInterval.value) {
			clearInterval(searchInterval.value)
			searchInterval.value = null
		}
	}

	/**
	 * Poll for match found
	 */
	const pollForMatch = async () => {
		const maxAttempts = 60 // 60 seconds max
		let attempts = 0

		const poll = async () => {
			if (!isSearching.value || attempts >= maxAttempts) {
				stopSearching()
				return
			}

			attempts++
			await refetchStatus()

			if (hasActiveDuel.value && activeGameId.value) {
				stopSearching()
				router.push({ name: 'duel-play', params: { duelId: activeGameId.value } })
				return
			}

			setTimeout(poll, 1000)
		}

		poll()
	}

	/**
	 * Send challenge to friend
	 */
	const sendChallenge = async (friendId: string) => {
		try {
			console.log('[usePvPDuel] Sending challenge to:', friendId)

			const response = await sendChallengeMutation.mutateAsync({
				data: { playerId, friendId },
			})

			console.log('[usePvPDuel] Challenge sent:', response.data)
			await refetchStatus()
			await refetchRivals()

			return response.data
		} catch (error) {
			console.error('[usePvPDuel] Failed to send challenge:', error)
			throw error
		}
	}

	/**
	 * Respond to challenge (accept/decline)
	 */
	const respondChallenge = async (challengeId: string, action: 'accept' | 'decline') => {
		try {
			console.log('[usePvPDuel] Responding to challenge:', { challengeId, action })

			const response = await respondChallengeMutation.mutateAsync({
				challengeId,
				data: { playerId, action },
			})

			console.log('[usePvPDuel] Challenge response:', response.data)
			await refetchStatus()

			// If accepted and game started, navigate to play
			if (action === 'accept' && response.data?.gameId) {
				router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
			}

			return response.data
		} catch (error) {
			console.error('[usePvPDuel] Failed to respond to challenge:', error)
			throw error
		}
	}

	/**
	 * Create shareable challenge link
	 */
	const createChallengeLink = async () => {
		try {
			console.log('[usePvPDuel] Creating challenge link...')

			const response = await createLinkMutation.mutateAsync({
				data: { playerId },
			})

			console.log('[usePvPDuel] Challenge link created:', response.data)

			return response.data
		} catch (error) {
			console.error('[usePvPDuel] Failed to create challenge link:', error)
			throw error
		}
	}

	/**
	 * Share challenge link via Telegram
	 * Creates a new challenge link and opens Telegram share dialog
	 */
	const shareChallengeToTelegram = async (message?: string) => {
		try {
			console.log('[usePvPDuel] Sharing challenge to Telegram...')

			// Create a new challenge link
			const result = await createChallengeLink()

			if (!result?.challengeLink) {
				throw new Error('Failed to create challenge link')
			}

			// Share via Telegram using TMA SDK
			const shareMessage = message ?? '⚔️ Вызываю тебя на дуэль в Quiz Sprint!'
			shareURL(result.challengeLink, shareMessage)

			console.log('[usePvPDuel] Challenge shared:', result.challengeLink)

			return result
		} catch (error) {
			console.error('[usePvPDuel] Failed to share challenge:', error)
			throw error
		}
	}

	/**
	 * Request rematch after a game
	 */
	const requestRematch = async (gameId: string) => {
		try {
			console.log('[usePvPDuel] Requesting rematch for:', gameId)

			const response = await rematchMutation.mutateAsync({
				gameId,
				data: { playerId },
			})

			console.log('[usePvPDuel] Rematch response:', response.data)

			// If rematch auto-accepted (opponent already requested)
			if (response.data?.status === 'accepted' && response.data?.gameId) {
				router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
			}

			return response.data
		} catch (error) {
			console.error('[usePvPDuel] Failed to request rematch:', error)
			throw error
		}
	}

	/**
	 * Start challenge game (inviter confirms after invitee accepted via link)
	 */
	const startChallenge = async (challengeId: string) => {
		try {
			console.log('[usePvPDuel] Starting challenge:', challengeId)

			const response = await startChallengeMutation.mutateAsync({
				challengeId,
				data: { playerId },
			})

			console.log('[usePvPDuel] Challenge started:', response)

			if (response?.data?.gameId) {
				router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
			}

			return response
		} catch (error) {
			console.error('[usePvPDuel] Failed to start challenge:', error)
			throw error
		}
	}

	/**
	 * Navigate to active duel
	 */
	const goToActiveDuel = () => {
		if (activeGameId.value) {
			router.push({ name: 'duel-play', params: { duelId: activeGameId.value } })
		}
	}

	/**
	 * Initialize (load data)
	 */
	const initialize = async () => {
		console.log('[usePvPDuel] Initializing...')
		await refetchStatus()
	}

	// ===========================
	// Return
	// ===========================

	return {
		// Status
		hasActiveDuel,
		activeGameId,
		player,
		tickets,
		friendsOnline,
		pendingChallenges,
		outgoingChallenges,
		outgoingReadyChallenges,
		outgoingPendingChallenges,
		seasonId,
		seasonEndsAt,

		// Rating
		mmr,
		league,
		leagueLabel,
		leagueIcon,
		division,
		seasonWins,
		seasonLosses,
		winRate,

		// Leaderboard
		leaderboard,
		playerRank,

		// History
		gameHistory,

		// Rivals
		rivals,

		// UI State
		isSearching,
		searchTime,
		isLoading,
		canPlay,

		// Actions
		joinQueue,
		leaveQueue,
		sendChallenge,
		respondChallenge,
		createChallengeLink,
		shareChallengeToTelegram,
		requestRematch,
		startChallenge,
		goToActiveDuel,
		initialize,
		startOutgoingPoll,
		stopOutgoingPoll,

		// Refetch
		refetchStatus,
		refetchLeaderboard,
		refetchHistory,
		refetchRivals,
	}
}
