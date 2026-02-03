import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
	usePostMarathonStart,
	usePostMarathonGameidAnswer,
	usePostMarathonGameidBonus,
	usePostMarathonGameidContinue,
	useDeleteMarathonGameid,
	useGetMarathonStatus,
	useGetMarathonPersonalBests,
} from '@/api/generated'
import type {
	InternalInfrastructureHttpHandlersMarathonGameDTO,
	InternalInfrastructureHttpHandlersQuestionDTO,
	InternalInfrastructureHttpHandlersMarathonPersonalBestDTO,
	InternalInfrastructureHttpHandlersMarathonBonusInventoryDTO,
	InternalInfrastructureHttpHandlersMarathonGameOverResultDTO,
	InternalInfrastructureHttpHandlersMarathonMilestoneDTO,
	InternalInfrastructureHttpHandlersSubmitMarathonAnswerData,
} from '@/api/generated'

export type MarathonStatus = 'idle' | 'loading' | 'playing' | 'game-over' | 'error'
export type BonusType = 'shield' | 'fifty_fifty' | 'skip' | 'freeze'

interface MarathonState {
	status: MarathonStatus
	game: InternalInfrastructureHttpHandlersMarathonGameDTO | null
	currentQuestion: InternalInfrastructureHttpHandlersQuestionDTO | null
	bonusInventory: InternalInfrastructureHttpHandlersMarathonBonusInventoryDTO
	score: number
	totalQuestions: number
	timeLimit: number
	shieldActive: boolean
	personalBest: number | null
	categoryId: string | null
	lastAnswerResult: InternalInfrastructureHttpHandlersSubmitMarathonAnswerData | null
	gameOverResult: InternalInfrastructureHttpHandlersMarathonGameOverResultDTO | null
	milestone: InternalInfrastructureHttpHandlersMarathonMilestoneDTO | null
	hiddenAnswerIds: string[]
}

/**
 * Composable for Marathon game management (thin client)
 *
 * All game state comes from server. Frontend only renders.
 */
export function useMarathon(playerId: string) {
	const router = useRouter()

	// ===========================
	// State
	// ===========================

	const state = ref<MarathonState>({
		status: 'idle',
		game: null,
		currentQuestion: null,
		bonusInventory: { shield: 0, fiftyFifty: 0, skip: 0, freeze: 0 },
		score: 0,
		totalQuestions: 0,
		timeLimit: 15,
		shieldActive: false,
		personalBest: null,
		categoryId: null,
		lastAnswerResult: null,
		gameOverResult: null,
		milestone: null,
		hiddenAnswerIds: [],
	})

	// ===========================
	// API Hooks
	// ===========================

	const startMutation = usePostMarathonStart()
	const answerMutation = usePostMarathonGameidAnswer()
	const bonusMutation = usePostMarathonGameidBonus()
	const continueMutation = usePostMarathonGameidContinue()
	const abandonMutation = useDeleteMarathonGameid()

	const {
		data: statusData,
		refetch: refetchStatus,
		isLoading: isLoadingStatus,
	} = useGetMarathonStatus(
		computed(() => ({ playerId })),
		{
			query: {
				enabled: computed(() => !!playerId),
			},
		},
	)

	const { data: personalBestsData, refetch: refetchPersonalBests } = useGetMarathonPersonalBests(
		computed(() => ({ playerId })),
		{
			query: {
				enabled: computed(() => !!playerId),
			},
		},
	)

	// ===========================
	// Computed Properties
	// ===========================

	const isPlaying = computed(() => state.value.status === 'playing')
	const isGameOver = computed(() => state.value.status === 'game-over')
	const isLoading = computed(
		() =>
			state.value.status === 'loading' ||
			startMutation.isPending.value ||
			answerMutation.isPending.value ||
			bonusMutation.isPending.value ||
			continueMutation.isPending.value ||
			abandonMutation.isPending.value ||
			isLoadingStatus.value,
	)

	const lives = computed(() => state.value.game?.lives ?? { currentLives: 0, maxLives: 3, label: '', timeToNextLife: 0 })
	const hasLives = computed(() => lives.value.currentLives > 0)
	const canPlay = computed(() => state.value.status !== 'playing')

	const livesPercent = computed(() =>
		Math.round((lives.value.currentLives / lives.value.maxLives) * 100),
	)

	const progressToRecord = computed(() => {
		if (!state.value.personalBest) return 0
		if (state.value.score >= state.value.personalBest) return 100
		return Math.round((state.value.score / state.value.personalBest) * 100)
	})

	// Bonus availability
	const canUseShield = computed(() => state.value.bonusInventory.shield > 0 && isPlaying.value)
	const canUseFiftyFifty = computed(() => state.value.bonusInventory.fiftyFifty > 0 && isPlaying.value)
	const canUseSkip = computed(() => state.value.bonusInventory.skip > 0 && isPlaying.value)
	const canUseFreeze = computed(() => state.value.bonusInventory.freeze > 0 && isPlaying.value)

	// Continue offer from game over result
	const continueOffer = computed(() => state.value.gameOverResult?.continueOffer)
	const canContinue = computed(() => continueOffer.value?.available === true)

	// ===========================
	// Helpers
	// ===========================

	const syncFromGame = (game: InternalInfrastructureHttpHandlersMarathonGameDTO) => {
		state.value.game = game
		state.value.currentQuestion = game.currentQuestion ?? null
		state.value.bonusInventory = game.bonusInventory
		state.value.score = game.score
		state.value.totalQuestions = game.totalQuestions
		state.value.timeLimit = game.timeLimit
		state.value.shieldActive = game.shieldActive
		state.value.personalBest = game.personalBest ?? null
	}

	// ===========================
	// Game Actions
	// ===========================

	const startGame = async (categoryId: string) => {
		try {
			state.value.status = 'loading'
			state.value.categoryId = categoryId

			const response = await startMutation.mutateAsync({
				data: {
					playerId,
					categoryId,
				},
			})

			const gameData = response.data

			syncFromGame(gameData.game)
			state.value.status = 'playing'
			state.value.lastAnswerResult = null
			state.value.gameOverResult = null
			state.value.milestone = null
			state.value.hiddenAnswerIds = []

			router.push({ name: 'marathon-play' })

			return true
		} catch (error: unknown) {
			state.value.status = 'error'
			console.error('Failed to start Marathon:', error)

			if (error && typeof error === 'object' && 'response' in error) {
				const axiosError = error as { response?: { status?: number } }
				if (axiosError.response?.status === 409) {
					await refetchStatus()
				}
			}

			throw error
		}
	}

	const submitAnswer = async (answerId: string, timeTaken: number) => {
		if (!state.value.game?.id || !state.value.currentQuestion?.id) {
			throw new Error('No active game or question')
		}

		try {
			const response = await answerMutation.mutateAsync({
				gameId: state.value.game.id,
				data: {
					questionId: state.value.currentQuestion.id,
					answerId,
					playerId,
					timeTaken,
				},
			})

			const answerData = response.data
			state.value.lastAnswerResult = answerData

			// Update state from server response
			state.value.score = answerData.score
			state.value.totalQuestions = answerData.totalQuestions
			state.value.bonusInventory = answerData.bonusInventory
			state.value.shieldActive = false // Reset after answer
			state.value.milestone = answerData.milestone ?? null
			state.value.hiddenAnswerIds = []

			if (answerData.isGameOver) {
				state.value.status = 'game-over'
				state.value.currentQuestion = null
				state.value.gameOverResult = answerData.gameOverResult ?? null

				await refetchPersonalBests()

				router.push({ name: 'marathon-gameover' })
			} else if (answerData.nextQuestion) {
				state.value.currentQuestion = answerData.nextQuestion
				state.value.timeLimit = answerData.nextTimeLimit ?? state.value.timeLimit

				// Update lives from server
				if (state.value.game) {
					state.value.game = {
						...state.value.game,
						lives: answerData.lives,
					}
				}
			}

			return answerData
		} catch (error) {
			console.error('Failed to submit answer:', error)
			throw error
		}
	}

	const useBonus = async (bonusType: BonusType) => {
		if (!state.value.game?.id || !state.value.currentQuestion?.id) {
			throw new Error('No active game or question')
		}

		try {
			const response = await bonusMutation.mutateAsync({
				gameId: state.value.game.id,
				data: {
					bonusType,
					playerId,
					questionId: state.value.currentQuestion.id,
				},
			})

			const bonusData = response.data

			// Update bonus inventory from server
			state.value.bonusInventory = bonusData.bonusInventory

			// Handle bonus result
			const result = bonusData.bonusResult
			if (bonusType === 'skip' && result.nextQuestion) {
				state.value.currentQuestion = result.nextQuestion
				state.value.hiddenAnswerIds = []
				if (result.nextTimeLimit) {
					state.value.timeLimit = result.nextTimeLimit
				}
			} else if (bonusType === 'fifty_fifty' && result.hiddenAnswerIds) {
				state.value.hiddenAnswerIds = result.hiddenAnswerIds
			} else if (bonusType === 'shield') {
				state.value.shieldActive = result.shieldActive ?? true
			} else if (bonusType === 'freeze' && result.newTimeLimit) {
				state.value.timeLimit = result.newTimeLimit
			}

			return bonusData
		} catch (error) {
			console.error('Failed to use bonus:', error)
			throw error
		}
	}

	const continueGame = async (paymentMethod: 'coins' | 'ad') => {
		if (!state.value.game?.id) {
			throw new Error('No game to continue')
		}

		try {
			state.value.status = 'loading'

			const response = await continueMutation.mutateAsync({
				gameId: state.value.game.id,
				data: {
					playerId,
					paymentMethod,
				},
			})

			const continueData = response.data

			syncFromGame(continueData.game)
			state.value.status = 'playing'
			state.value.lastAnswerResult = null
			state.value.gameOverResult = null
			state.value.hiddenAnswerIds = []

			router.push({ name: 'marathon-play' })

			return continueData
		} catch (error) {
			state.value.status = 'game-over'
			console.error('Failed to continue Marathon:', error)
			throw error
		}
	}

	const abandonGame = async () => {
		if (!state.value.game?.id) {
			throw new Error('No active game')
		}

		try {
			await abandonMutation.mutateAsync({
				gameId: state.value.game.id,
				data: {
					playerId,
				},
			})

			state.value.status = 'game-over'
			state.value.currentQuestion = null

			await refetchStatus()
			await refetchPersonalBests()

			router.push({ name: 'marathon-gameover' })

			return true
		} catch (error) {
			console.error('Failed to abandon game:', error)
			throw error
		}
	}

	const checkStatus = async () => {
		try {
			await refetchStatus()

			if (statusData.value?.data) {
				const data = statusData.value.data
				if (data.hasActiveGame && data.game) {
					syncFromGame(data.game)

					if (data.game.status === 'game_over') {
						state.value.status = 'game-over'
					} else {
						state.value.status = 'playing'
					}
				}
			}
		} catch (error) {
			console.error('Failed to check status:', error)
		}
	}

	const loadPersonalBests = async () => {
		try {
			await refetchPersonalBests()

			if (personalBestsData.value?.data?.personalBests) {
				const currentCategoryBest = personalBestsData.value.data.personalBests.find(
					(pb: InternalInfrastructureHttpHandlersMarathonPersonalBestDTO) =>
						pb.category.id === state.value.categoryId,
				)

				if (currentCategoryBest) {
					state.value.personalBest = currentCategoryBest.bestScore
				}
			}
		} catch (error) {
			console.error('Failed to load personal bests:', error)
		}
	}

	const reset = () => {
		state.value = {
			status: 'idle',
			game: null,
			currentQuestion: null,
			bonusInventory: { shield: 0, fiftyFifty: 0, skip: 0, freeze: 0 },
			score: 0,
			totalQuestions: 0,
			timeLimit: 15,
			shieldActive: false,
			personalBest: null,
			categoryId: null,
			lastAnswerResult: null,
			gameOverResult: null,
			milestone: null,
			hiddenAnswerIds: [],
		}
	}

	// ===========================
	// Lifecycle
	// ===========================

	const initialized = ref(false)

	const initialize = async () => {
		if (initialized.value) return

		await checkStatus()
		await loadPersonalBests()

		initialized.value = true
	}

	// ===========================
	// Return
	// ===========================

	return {
		// State
		state,

		// Computed
		isPlaying,
		isGameOver,
		isLoading,
		lives,
		hasLives,
		canPlay,
		progressToRecord,
		livesPercent,
		canUseShield,
		canUseFiftyFifty,
		canUseSkip,
		canUseFreeze,
		continueOffer,
		canContinue,

		// Actions
		startGame,
		submitAnswer,
		useBonus,
		continueGame,
		abandonGame,
		checkStatus,
		loadPersonalBests,
		reset,
		initialize,

		// Data from API
		statusData,
		personalBestsData,
	}
}
