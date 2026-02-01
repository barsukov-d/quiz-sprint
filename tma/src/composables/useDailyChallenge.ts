import { computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  usePostDailyChallengeStart,
  usePostDailyChallengeGameidAnswer,
  useGetDailyChallengeStatus,
  useGetDailyChallengeStreak,
  usePostDailyChallengeGameidRetry,
  usePostDailyChallengeGameidChestOpen
} from '@/api/generated'

/**
 * Composable для управления Daily Challenge игрой
 *
 * ✅ SERVER-SIDE STATE ARCHITECTURE:
 * - Backend DB is the single source of truth
 * - Frontend always fetches fresh data from API (/status)
 * - Results are included in /status response when game is completed
 * - No local state management (stateless)
 *
 * Возможности:
 * - Старт игры
 * - Отправка ответов
 * - Получение статуса с сервера (всегда актуальный)
 */
export function useDailyChallenge(playerId: string) {
  const router = useRouter()

  // ===========================
  // API Hooks (Server State)
  // ===========================

  const startMutation = usePostDailyChallengeStart()
  const answerMutation = usePostDailyChallengeGameidAnswer()
  const retryMutation = usePostDailyChallengeGameidRetry()
  const openChestMutation = usePostDailyChallengeGameidChestOpen()

  // Main API endpoint - single source of truth
  const { data: statusData, refetch: refetchStatus, isLoading: isLoadingStatus } = useGetDailyChallengeStatus(
    computed(() => ({ playerId })),
    {
      query: {
        enabled: computed(() => !!playerId),
        refetchOnWindowFocus: true, // Always fetch fresh data
        staleTime: 0 // No caching, always fresh
      }
    }
  )

  const { data: streakData, refetch: refetchStreak } = useGetDailyChallengeStreak(
    computed(() => ({ playerId })),
    {
      query: {
        enabled: computed(() => !!playerId)
      }
    }
  )

  // ===========================
  // Computed Properties (from API)
  // ===========================

  // Game state from API
  const game = computed(() => statusData.value?.data?.game ?? null)
  const results = computed(() => statusData.value?.data?.results ?? null)
  const hasPlayed = computed(() => statusData.value?.data?.hasPlayed ?? false)
  const timeToExpire = computed(() => statusData.value?.data?.timeToExpire ?? 0)
  const totalPlayers = computed(() => statusData.value?.data?.totalPlayers ?? 0)
  const timeLimit = computed(() => statusData.value?.data?.timeLimit ?? 15)

  // Game details
  const currentQuestion = computed(() => game.value?.currentQuestion ?? null)
  const questionIndex = computed(() => game.value?.questionIndex ?? 0)
  const totalQuestions = computed(() => game.value?.totalQuestions ?? 10)
  const streak = computed(() => streakData.value?.data?.streak ?? game.value?.streak ?? null)

  // Status flags
  const isPlaying = computed(() => game.value?.status === 'in_progress')
  const isCompleted = computed(() => game.value?.status === 'completed')
  const canPlay = computed(() => !hasPlayed.value && !isPlaying.value)

  const isLoading = computed(() =>
    startMutation.isPending.value ||
    answerMutation.isPending.value ||
    retryMutation.isPending.value ||
    openChestMutation.isPending.value ||
    isLoadingStatus.value
  )

  // Progress (0-100%)
  const progress = computed(() => {
    if (totalQuestions.value === 0) return 0
    return Math.round((questionIndex.value / totalQuestions.value) * 100)
  })

  // Remaining questions
  const remainingQuestions = computed(() =>
    totalQuestions.value - questionIndex.value
  )

  // Formatted time to expire
  const timeToExpireFormatted = computed(() => {
    const seconds = timeToExpire.value
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  })

  // ===========================
  // Game Actions
  // ===========================

  /**
   * Начать игру
   */
  const startGame = async (date?: string) => {
    try {
      console.log('[useDailyChallenge] Starting game...')

      const response = await startMutation.mutateAsync({
        data: {
          playerId,
          date
        }
      })

      console.log('[useDailyChallenge] Game started:', response.data)

      // Refresh status to get updated game state
      await refetchStatus()

      // Navigate to play page
      router.push({ name: 'daily-challenge-play' })

      return true
    } catch (error: unknown) {
      console.error('[useDailyChallenge] Failed to start game:', error)

      // Refresh status on error
      await refetchStatus()

      throw error
    }
  }

  /**
   * Отправить ответ на текущий вопрос
   */
  const submitAnswer = async (answerId: string, timeTaken: number) => {
    console.log('[useDailyChallenge] submitAnswer called', {
      gameId: game.value?.gameId,
      questionId: currentQuestion.value?.id,
      answerId,
      timeTaken
    })

    if (!game.value?.gameId || !currentQuestion.value?.id) {
      console.error('[useDailyChallenge] No active game or question!', {
        hasGame: !!game.value,
        gameId: game.value?.gameId,
        hasQuestion: !!currentQuestion.value
      })
      throw new Error('No active game or question')
    }

    try {
      console.log('[useDailyChallenge] Sending answer mutation...')
      const response = await answerMutation.mutateAsync({
        gameId: game.value.gameId,
        data: {
          questionId: currentQuestion.value.id,
          answerId,
          playerId,
          timeTaken
        }
      })

      console.log('[useDailyChallenge] Answer mutation successful:', response.data)

      const answerData = response.data

      console.log('[useDailyChallenge] Answer result:', {
        isCorrect: answerData.isCorrect,
        correctAnswerId: answerData.correctAnswerId,
        isGameCompleted: answerData.isGameCompleted
      })

      if (answerData.isGameCompleted) {
        // Game completed - refresh status to get results
        console.log('[useDailyChallenge] Game completed! Fetching results from server...')
        await refetchStatus()
        await refetchStreak()
      } else {
        // Game continues - refresh status to get next question
        console.log('[useDailyChallenge] Game continues, fetching next question...')
        await refetchStatus()
      }

      // Return answer data with feedback (isCorrect, correctAnswerId)
      // View handles navigation after showing feedback
      return answerData
    } catch (error) {
      console.error('[useDailyChallenge] Failed to submit answer:', error)
      // Refresh status on error to sync with server
      await refetchStatus()
      throw error
    }
  }

  /**
   * Обновить статус с сервера
   */
  const checkStatus = async () => {
    try {
      await refetchStatus()
      await refetchStreak()
    } catch (error) {
      console.error('[useDailyChallenge] Failed to check status:', error)
    }
  }

  /**
   * Инициализация (загрузка данных с сервера)
   */
  const initialize = async () => {
    console.log('[useDailyChallenge] Initializing...')
    await checkStatus()
  }

  /**
   * Открыть сундук (получить награды)
   */
  const openChest = async () => {
    if (!game.value?.gameId) {
      throw new Error('No game to open chest for')
    }

    try {
      console.log('[useDailyChallenge] Opening chest...')
      const response = await openChestMutation.mutateAsync({
        gameId: game.value.gameId,
        data: {
          playerId
        }
      })

      console.log('[useDailyChallenge] Chest opened:', response.data)
      return response.data
    } catch (error) {
      console.error('[useDailyChallenge] Failed to open chest:', error)
      throw error
    }
  }

  /**
   * Повторить челлендж (создать вторую попытку)
   */
  const retryChallenge = async (paymentMethod: 'coins' | 'ad') => {
    if (!game.value?.gameId) {
      throw new Error('No game to retry')
    }

    try {
      console.log('[useDailyChallenge] Retrying challenge...', { paymentMethod })
      const response = await retryMutation.mutateAsync({
        gameId: game.value.gameId,
        data: {
          playerId,
          paymentMethod
        }
      })

      console.log('[useDailyChallenge] Retry successful:', response.data)

      // Refresh status to get new game
      await refetchStatus()

      // Navigate to play page
      router.push({ name: 'daily-challenge-play' })

      return response.data
    } catch (error) {
      console.error('[useDailyChallenge] Failed to retry challenge:', error)
      throw error
    }
  }

  // ===========================
  // Return
  // ===========================

  return {
    // Computed (from API)
    game,
    results,
    hasPlayed,
    timeToExpire,
    totalPlayers,
    timeLimit,
    currentQuestion,
    questionIndex,
    totalQuestions,
    streak,
    isPlaying,
    isCompleted,
    canPlay,
    isLoading,
    progress,
    remainingQuestions,
    timeToExpireFormatted,

    // Actions
    startGame,
    submitAnswer,
    checkStatus,
    initialize,
    openChest,
    retryChallenge,

    // Raw API data (for advanced usage)
    statusData,
    streakData,
    refetchStatus,
    refetchStreak
  }
}
