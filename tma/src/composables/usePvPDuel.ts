import { type MaybeRef, computed, ref, toValue, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useLobbyWebSocket } from '@/composables/useLobbyWebSocket'
import { shareURL, shareMessage } from '@tma.js/sdk'
import { apiClient } from '@/api/client'
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

const DEBUG = import.meta.env.DEV
const log = (...args: unknown[]) => {
	if (DEBUG) console.log('[usePvPDuel]', ...args)
}
const logError = (...args: unknown[]) => {
	if (DEBUG) console.error('[usePvPDuel]', ...args)
}

let pollTimeout: ReturnType<typeof setTimeout> | null = null

export function usePvPDuel(playerIdRef: MaybeRef<string>) {
	const router = useRouter()

	// ===========================
	// Lobby WebSocket
	// ===========================
	const lobbyWs = useLobbyWebSocket(toValue(playerIdRef))

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
		computed(() => ({ playerId: toValue(playerIdRef) })),
		{
			query: {
				enabled: computed(() => !!toValue(playerIdRef)),
				refetchOnWindowFocus: true,
				staleTime: 0,
				retry: false,
				refetchOnMount: true,
			},
		},
	)

	// Leaderboard
	const { data: leaderboardData, refetch: refetchLeaderboard } = useGetDuelLeaderboard(
		computed(() => ({ playerId: toValue(playerIdRef), type: 'seasonal', limit: 20 })),
		{
			query: {
				enabled: computed(() => !!toValue(playerIdRef)),
				staleTime: 60000, // 1 minute cache
			},
		},
	)

	// Match history
	const { data: historyData, refetch: refetchHistory } = useGetDuelHistory(
		computed(() => ({ playerId: toValue(playerIdRef), limit: 20, offset: 0, filter: 'all' })),
		{
			query: {
				enabled: computed(() => !!toValue(playerIdRef)),
				staleTime: 30000, // 30 seconds cache
			},
		},
	)

	// Rivals (recent opponents)
	const { data: rivalsData, refetch: refetchRivals } = useGetDuelRivals(
		computed(() => ({ playerId: toValue(playerIdRef) })),
		{
			query: {
				enabled: computed(() => !!toValue(playerIdRef)),
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
	const acceptedChallenges = computed(() => statusData.value?.data?.acceptedChallenges ?? [])

	const outgoingReadyChallenges = computed(() =>
		outgoingChallenges.value.filter((c) => c.status === 'accepted_waiting_inviter'),
	)

	const outgoingPendingChallenges = computed(() =>
		outgoingChallenges.value.filter((c) => c.status === 'pending' && c.type === 'direct'),
	)

	const outgoingLinkChallenges = computed(() =>
		outgoingChallenges.value.filter((c) => c.status === 'pending' && c.type === 'link'),
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
	// @ts-expect-error seasonDraws will be added to DTO when backend migration 021 is deployed
	const seasonDraws = computed(() => player.value?.seasonDraws ?? 0)
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
	// Also invalidate rivals cache when challenge state changes (decline/expire/accept)
	watch(
		outgoingChallenges,
		(challenges, prevChallenges) => {
			if (challenges.length > 0) {
				startOutgoingPoll()
			} else {
				stopOutgoingPoll()
			}
			if (prevChallenges !== undefined && challenges.length < prevChallenges.length) {
				refetchRivals()
			}
		},
		{ immediate: true },
	)

	// WS event handlers: refetch state and navigate on game_ready
	lobbyWs.on('challenge_received', async () => {
		await refetchStatus()
	})

	lobbyWs.on('challenge_accepted', async () => {
		await refetchStatus()
		if (hasActiveDuel.value && activeGameId.value) {
			router.push({ name: 'duel-play', params: { duelId: activeGameId.value } })
		}
	})

	lobbyWs.on('challenge_declined', async () => {
		await refetchStatus()
		await refetchRivals()
	})

	lobbyWs.on('challenge_expired', async () => {
		await refetchStatus()
	})

	lobbyWs.on('game_ready', async (data) => {
		const gameId = data?.gameId as string | undefined
		// Always verify with fresh status before navigating — stale WS events
		// can arrive after a game is already finished and would redirect the user
		// back into a dead game session.
		await refetchStatus()
		if (hasActiveDuel.value && activeGameId.value) {
			if (!gameId || activeGameId.value === gameId) {
				router.push({ name: 'duel-play', params: { duelId: activeGameId.value } })
			}
		}
	})

	lobbyWs.on('queue_matched', async () => {
		await refetchStatus()
		if (hasActiveDuel.value && activeGameId.value) {
			router.push({ name: 'duel-play', params: { duelId: activeGameId.value } })
		}
	})

	onUnmounted(() => {
		stopOutgoingPoll()
		if (pollTimeout) {
			clearTimeout(pollTimeout)
			pollTimeout = null
		}
	})

	// ===========================
	// Actions
	// ===========================

	/**
	 * Join matchmaking queue
	 */
	const joinQueue = async () => {
		try {
			log('Joining queue...')

			isSearching.value = true
			searchTime.value = 0

			// Start search timer
			searchInterval.value = setInterval(() => {
				searchTime.value++
			}, 1000)

			const response = await joinQueueMutation.mutateAsync({
				data: { playerId: toValue(playerIdRef) },
			})

			log('Joined queue:', response.data)

			// Poll for match found
			pollForMatch()

			return response.data
		} catch (error) {
			logError('Failed to join queue:', error)
			stopSearching()
			throw error
		}
	}

	/**
	 * Leave matchmaking queue
	 */
	const leaveQueue = async () => {
		try {
			log('Leaving queue...')

			await leaveQueueMutation.mutateAsync({
				params: { playerId: toValue(playerIdRef) },
			})

			stopSearching()
			await refetchStatus()
		} catch (error) {
			logError('Failed to leave queue:', error)
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
		if (pollTimeout) {
			clearTimeout(pollTimeout)
			pollTimeout = null
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

			pollTimeout = setTimeout(poll, lobbyWs.isPollingFallback.value ? 3000 : 1000)
		}

		poll()
	}

	/**
	 * Send challenge to friend
	 */
	const sendChallenge = async (friendId: string) => {
		try {
			log('Sending challenge to:', friendId)

			const response = await sendChallengeMutation.mutateAsync({
				data: { playerId: toValue(playerIdRef), friendId },
			})

			log('Challenge sent:', response.data)
			await refetchStatus()
			await refetchRivals()

			return response.data
		} catch (error) {
			logError('Failed to send challenge:', error)
			throw error
		}
	}

	/**
	 * Respond to challenge (accept/decline)
	 */
	const respondChallenge = async (challengeId: string, action: 'accept' | 'decline') => {
		try {
			log('Responding to challenge:', { challengeId, action })

			const response = await respondChallengeMutation.mutateAsync({
				challengeId,
				data: { playerId: toValue(playerIdRef), action },
			})

			log('Challenge response:', response.data)
			await refetchStatus()

			// If accepted and game started, navigate to play
			if (action === 'accept' && response.data?.gameId) {
				router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
			}

			return response.data
		} catch (error) {
			logError('Failed to respond to challenge:', error)
			throw error
		}
	}

	/**
	 * Create shareable challenge link
	 */
	const createChallengeLink = async () => {
		try {
			log('Creating challenge link...')

			const response = await createLinkMutation.mutateAsync({
				data: { playerId: toValue(playerIdRef) },
			})

			log('Challenge link created:', response.data)

			return response.data
		} catch (error) {
			logError('Failed to create challenge link:', error)
			throw error
		}
	}

	/**
	 * Share challenge link via Telegram
	 * Creates a new challenge link and opens Telegram share dialog.
	 * Uses shareMessage (TMA v8.0+, sends message with inline button) when available,
	 * falls back to shareURL for older clients.
	 */
	const shareChallengeToTelegram = async (message?: string) => {
		try {
			log('Sharing challenge to Telegram...')

			// Create a new challenge link
			const result = await createChallengeLink()

			if (!result?.challengeLink) {
				throw new Error('Failed to create challenge link')
			}

			const shareText = message ?? '⚔️ Вызываю тебя на дуэль в Quiz Sprint!'

			// Try modern approach: shareMessage with inline keyboard (TMA v8.0+)
			if (shareMessage.isSupported()) {
				try {
					const { data } = await apiClient.post<{
						data: { preparedMessageId: string; expiresAt: number }
					}>('/duel/challenge/prepare-share', {
						playerId: toValue(playerIdRef),
						challengeLink: result.challengeLink,
					})
					shareMessage(data.data.preparedMessageId)
					log(
						'[usePvPDuel] Challenge shared via shareMessage:',
						data.data.preparedMessageId,
					)
					return result
				} catch (prepareErr) {
					log(
						'[usePvPDuel] shareMessage prepare failed, falling back to shareURL:',
						prepareErr,
					)
				}
			}

			// Fallback for older clients
			shareURL(result.challengeLink, shareText)
			log('Challenge shared via shareURL:', result.challengeLink)

			return result
		} catch (error) {
			logError('Failed to share challenge:', error)
			throw error
		}
	}

	/**
	 * Request rematch after a game
	 */
	const requestRematch = async (gameId: string) => {
		try {
			log('Requesting rematch for:', gameId)

			const response = await rematchMutation.mutateAsync({
				gameId,
				data: { playerId: toValue(playerIdRef) },
			})

			log('Rematch response:', response.data)

			// If rematch auto-accepted (opponent already requested)
			if (response.data?.status === 'accepted' && response.data?.gameId) {
				router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
			}

			return response.data
		} catch (error) {
			logError('Failed to request rematch:', error)
			throw error
		}
	}

	/**
	 * Start challenge game (inviter confirms after invitee accepted via link)
	 */
	const startChallenge = async (challengeId: string) => {
		try {
			log('Starting challenge:', challengeId)

			const response = await startChallengeMutation.mutateAsync({
				challengeId,
				data: { playerId: toValue(playerIdRef) },
			})

			log('Challenge started:', response)

			if (response?.data?.gameId) {
				router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
			}

			return response
		} catch (error) {
			logError('Failed to start challenge:', error)
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
		log('Initializing...')
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
		acceptedChallenges,
		outgoingReadyChallenges,
		outgoingPendingChallenges,
		outgoingLinkChallenges,
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
		seasonDraws,
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
		isLoadingStatus,
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

		// Lobby WebSocket
		lobbyWs,
	}
}
