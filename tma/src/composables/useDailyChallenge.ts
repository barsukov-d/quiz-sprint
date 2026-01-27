import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  usePostDailyChallengeStart,
  usePostDailyChallengeGameidAnswer,
  useGetDailyChallengeStatus,
  useGetDailyChallengeLeaderboard,
  useGetDailyChallengeStreak
} from '@/api/generated'
import type {
  InternalInfrastructureHttpHandlersDailyGameDTO,
  InternalInfrastructureHttpHandlersQuestionDTO,
  InternalInfrastructureHttpHandlersGameResultsDTO,
  InternalInfrastructureHttpHandlersStreakDTO
} from '@/api/generated'

export type DailyChallengeStatus = 'idle' | 'loading' | 'playing' | 'completed' | 'error'

interface DailyChallengeState {
  status: DailyChallengeStatus
  game: InternalInfrastructureHttpHandlersDailyGameDTO | null
  currentQuestion: InternalInfrastructureHttpHandlersQuestionDTO | null
  questionIndex: number
  totalQuestions: number
  timeLimit: number
  results: InternalInfrastructureHttpHandlersGameResultsDTO | null
  streak: InternalInfrastructureHttpHandlersStreakDTO | null
  timeToExpire: number
  totalPlayers: number
}

const STORAGE_KEY = 'daily-challenge-state'

// ===========================
// Shared State (Singleton)
// ===========================
// This state is shared across all components using useDailyChallenge
// to prevent data loss when navigating between pages

const sharedState = ref<DailyChallengeState>({
  status: 'idle',
  game: null,
  currentQuestion: null,
  questionIndex: 0,
  totalQuestions: 10,
  timeLimit: 15,
  results: null,
  streak: null,
  timeToExpire: 0,
  totalPlayers: 0
})

/**
 * Composable для управления Daily Challenge игрой
 *
 * Возможности:
 * - Старт игры
 * - Отправка ответов
 * - Получение статуса
 * - Локальное сохранение прогресса
 * - Управление таймером до сброса
 *
 * NOTE: Uses shared singleton state to persist data across page navigation
 */
export function useDailyChallenge(playerId: string) {
  const router = useRouter()

  // Use shared state instead of local state
  const state = sharedState

  // ===========================
  // API Hooks
  // ===========================

  const startMutation = usePostDailyChallengeStart()
  const answerMutation = usePostDailyChallengeGameidAnswer()

  // Pass playerId as params (first argument), then options (second argument)
  const { data: statusData, refetch: refetchStatus, isLoading: isLoadingStatus } = useGetDailyChallengeStatus(
    computed(() => ({ playerId })),
    {
      query: {
        enabled: computed(() => !!playerId)
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
  // Computed Properties
  // ===========================

  const isPlaying = computed(() => state.value.status === 'playing')
  const isCompleted = computed(() => state.value.status === 'completed')
  const isLoading = computed(() =>
    state.value.status === 'loading' ||
    startMutation.isPending.value ||
    answerMutation.isPending.value ||
    isLoadingStatus.value
  )
  const hasPlayed = computed(() => statusData.value?.data?.hasPlayed ?? false)
  const canPlay = computed(() => !hasPlayed.value && state.value.status !== 'playing')

  // Прогресс (0-100%)
  const progress = computed(() => {
    if (state.value.totalQuestions === 0) return 0
    return Math.round((state.value.questionIndex / state.value.totalQuestions) * 100)
  })

  // Осталось вопросов
  const remainingQuestions = computed(() =>
    state.value.totalQuestions - state.value.questionIndex
  )

  // Форматированное время до сброса
  const timeToExpireFormatted = computed(() => {
    const seconds = state.value.timeToExpire
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  })

  // ===========================
  // Local Storage
  // ===========================

  const saveToLocalStorage = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({
        playerId,
        game: state.value.game,
        currentQuestion: state.value.currentQuestion,
        questionIndex: state.value.questionIndex,
        timestamp: Date.now()
      }))
    } catch (error) {
      console.error('Failed to save Daily Challenge state:', error)
    }
  }

  const loadFromLocalStorage = () => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (!saved) return false

      const data = JSON.parse(saved)

      // Проверяем, что данные для текущего игрока
      if (data.playerId !== playerId) return false

      // Проверяем, что данные не старше 24 часов
      const age = Date.now() - data.timestamp
      if (age > 24 * 60 * 60 * 1000) {
        clearLocalStorage()
        return false
      }

      // MIGRATION: Проверяем, что игра имеет gameId (добавлен в новой версии)
      if (data.game && !data.game.gameId) {
        console.warn('[useDailyChallenge] Old localStorage data detected (missing gameId), clearing...')
        clearLocalStorage()
        return false
      }

      // Восстанавливаем состояние
      state.value.game = data.game
      state.value.currentQuestion = data.currentQuestion
      state.value.questionIndex = data.questionIndex
      state.value.status = 'playing'

      return true
    } catch (error) {
      console.error('Failed to load Daily Challenge state:', error)
      return false
    }
  }

  const clearLocalStorage = () => {
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch (error) {
      console.error('Failed to clear Daily Challenge state:', error)
    }
  }

  // ===========================
  // Game Actions
  // ===========================

  /**
   * Начать игру
   */
  const startGame = async (date?: string) => {
    try {
      state.value.status = 'loading'

      const response = await startMutation.mutateAsync({
        data: {
          playerId,
          date
        }
      })

      const gameData = response.data

      state.value.game = gameData.game
      state.value.currentQuestion = gameData.firstQuestion
      state.value.questionIndex = 1
      state.value.totalQuestions = 10
      state.value.timeLimit = gameData.timeLimit
      state.value.timeToExpire = gameData.timeToExpire
      state.value.totalPlayers = gameData.totalPlayers
      state.value.status = 'playing'

      saveToLocalStorage()

      // Переходим на экран игры
      router.push({ name: 'daily-challenge-play' })

      return true
    } catch (error: any) {
      state.value.status = 'error'
      console.error('Failed to start Daily Challenge:', error)

      // Обработка ошибок
      if (error.response?.status === 409) {
        // Уже играл сегодня
        await refetchStatus()
      }

      throw error
    }
  }

  /**
   * Отправить ответ на текущий вопрос
   */
  const submitAnswer = async (answerId: string, timeTaken: number) => {
    console.log('[useDailyChallenge] submitAnswer called', {
      gameId: state.value.game?.gameId,
      questionId: state.value.currentQuestion?.id,
      answerId,
      timeTaken
    })

    if (!state.value.game?.gameId || !state.value.currentQuestion?.id) {
      console.error('[useDailyChallenge] No active game or question!', {
        hasGame: !!state.value.game,
        gameId: state.value.game?.gameId,
        hasQuestion: !!state.value.currentQuestion
      })

      // If gameId is missing, clear localStorage and reload
      if (state.value.game && !state.value.game.gameId) {
        console.error('[useDailyChallenge] Game data is corrupted (missing gameId). Clearing localStorage...')
        clearLocalStorage()
        state.value.status = 'idle'
        throw new Error('Game data corrupted. Please start a new game.')
      }

      throw new Error('No active game or question')
    }

    try {
      console.log('[useDailyChallenge] Sending answer mutation...')
      const response = await answerMutation.mutateAsync({
        gameId: state.value.game.gameId,
        data: {
          questionId: state.value.currentQuestion.id,
          answerId,
          playerId,
          timeTaken
        }
      })

      console.log('[useDailyChallenge] Answer mutation successful:', response.data)

      const answerData = response.data

      // Обновляем индекс вопроса
      state.value.questionIndex = answerData.questionIndex + 1

      console.log('[useDailyChallenge] Checking game completion:', {
        isGameCompleted: answerData.isGameCompleted,
        hasGameResults: !!answerData.gameResults,
        gameResults: answerData.gameResults
      })

      if (answerData.isGameCompleted && answerData.gameResults) {
        // Игра завершена
        console.log('[useDailyChallenge] Game completed! Navigating to results...')
        state.value.results = answerData.gameResults
        state.value.currentQuestion = null
        state.value.status = 'completed'
        clearLocalStorage()

        // Переходим на экран результатов
        router.push({ name: 'daily-challenge-results' })
      } else if (answerData.nextQuestion) {
        // Следующий вопрос
        state.value.currentQuestion = answerData.nextQuestion
        state.value.timeLimit = answerData.nextTimeLimit ?? 15
        saveToLocalStorage()
      }

      return answerData
    } catch (error) {
      console.error('Failed to submit answer:', error)
      throw error
    }
  }

  /**
   * Получить статус игры
   */
  const checkStatus = async () => {
    try {
      await refetchStatus()

      if (statusData.value?.data) {
        const data = statusData.value.data
        state.value.timeToExpire = data.timeToExpire
        state.value.totalPlayers = data.totalPlayers

        if (data.hasPlayed && data.game?.status === 'completed') {
          state.value.status = 'completed'
          state.value.game = data.game
        } else if (data.game?.status === 'in_progress') {
          // Есть незавершённая игра - пытаемся восстановить
          loadFromLocalStorage()
        }
      }
    } catch (error) {
      console.error('Failed to check status:', error)
    }
  }

  /**
   * Получить серию игрока
   */
  const loadStreak = async () => {
    try {
      await refetchStreak()
      if (streakData.value?.data?.streak) {
        state.value.streak = streakData.value.data.streak
      }
    } catch (error) {
      console.error('Failed to load streak:', error)
    }
  }

  /**
   * Сбросить состояние
   */
  const reset = () => {
    state.value = {
      status: 'idle',
      game: null,
      currentQuestion: null,
      questionIndex: 0,
      totalQuestions: 10,
      timeLimit: 15,
      results: null,
      streak: null,
      timeToExpire: 0,
      totalPlayers: 0
    }
    clearLocalStorage()
  }

  // ===========================
  // Lifecycle
  // ===========================

  // При создании composable пытаемся восстановить состояние
  const initialized = ref(false)

  const initialize = async () => {
    if (initialized.value) return

    // Пытаемся восстановить из localStorage
    const restored = loadFromLocalStorage()

    // Загружаем статус с сервера
    await checkStatus()
    await loadStreak()

    initialized.value = true

    return restored
  }

  // ===========================
  // Return
  // ===========================

  return {
    // State
    state,

    // Computed
    isPlaying,
    isCompleted,
    isLoading,
    hasPlayed,
    canPlay,
    progress,
    remainingQuestions,
    timeToExpireFormatted,

    // Actions
    startGame,
    submitAnswer,
    checkStatus,
    loadStreak,
    reset,
    initialize,

    // Data from API
    statusData,
    streakData
  }
}
