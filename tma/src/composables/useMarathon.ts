import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  usePostMarathonStart,
  usePostMarathonGameidAnswer,
  usePostMarathonGameidHint,
  useDeleteMarathonGameid,
  useGetMarathonStatus,
  useGetMarathonPersonalBests
} from '@/api/generated'
import type {
  InternalInfrastructureHttpHandlersMarathonGameDTO,
  InternalInfrastructureHttpHandlersQuestionDTO,
  InternalInfrastructureHttpHandlersMarathonPersonalBestDTO
} from '@/api/generated'

export type MarathonStatus = 'idle' | 'loading' | 'playing' | 'game-over' | 'error'
export type HintType = 'fifty_fifty' | 'extra_time' | 'skip' | 'hint'

interface MarathonState {
  status: MarathonStatus
  game: InternalInfrastructureHttpHandlersMarathonGameDTO | null
  currentQuestion: InternalInfrastructureHttpHandlersQuestionDTO | null
  lives: number
  maxLives: number
  hints: {
    fiftyFifty: number
    extraTime: number
    skip: number
    hint: number
  }
  currentStreak: number
  score: number
  personalBest: number | null
  categoryId: string | null
  timeToLifeRestore: number // секунд до восстановления жизни
  lastAnswerCorrect: boolean | null
}

const STORAGE_KEY = 'marathon-state'

/**
 * Composable для управления Marathon игрой
 *
 * Возможности:
 * - Старт игры с выбором категории
 * - Отправка ответов с немедленным feedback
 * - Система жизней с восстановлением
 * - Подсказки (50/50, +10сек, Skip, Hint)
 * - Личные рекорды по категориям
 * - Адаптивная сложность
 * - Локальное сохранение незавершённой игры
 */
export function useMarathon(playerId: string) {
  const router = useRouter()

  // ===========================
  // State Management
  // ===========================

  const state = ref<MarathonState>({
    status: 'idle',
    game: null,
    currentQuestion: null,
    lives: 3,
    maxLives: 3,
    hints: {
      fiftyFifty: 1,
      extraTime: 1,
      skip: 1,
      hint: 1
    },
    currentStreak: 0,
    score: 0,
    personalBest: null,
    categoryId: null,
    timeToLifeRestore: 0,
    lastAnswerCorrect: null
  })

  // ===========================
  // API Hooks
  // ===========================

  const startMutation = usePostMarathonStart()
  const answerMutation = usePostMarathonGameidAnswer()
  const hintMutation = usePostMarathonGameidHint()
  const abandonMutation = useDeleteMarathonGameid()

  // Pass playerId as params (first argument), then options (second argument)
  const { data: statusData, refetch: refetchStatus, isLoading: isLoadingStatus } = useGetMarathonStatus(
    computed(() => ({ playerId })),
    {
      query: {
        enabled: computed(() => !!playerId)
      }
    }
  )

  const { data: personalBestsData, refetch: refetchPersonalBests } = useGetMarathonPersonalBests(
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
  const isGameOver = computed(() => state.value.status === 'game-over')
  const isLoading = computed(() =>
    state.value.status === 'loading' ||
    startMutation.isPending.value ||
    answerMutation.isPending.value ||
    hintMutation.isPending.value ||
    abandonMutation.isPending.value ||
    isLoadingStatus.value
  )

  const hasLives = computed(() => state.value.lives > 0)
  const canPlay = computed(() => hasLives.value && state.value.status !== 'playing')

  // Прогресс до личного рекорда (0-100%)
  const progressToRecord = computed(() => {
    if (!state.value.personalBest) return 0
    if (state.value.score >= state.value.personalBest) return 100
    return Math.round((state.value.score / state.value.personalBest) * 100)
  })

  // Процент оставшихся жизней
  const livesPercent = computed(() =>
    Math.round((state.value.lives / state.value.maxLives) * 100)
  )

  // Доступность подсказок
  const canUseFiftyFifty = computed(() => state.value.hints.fiftyFifty > 0 && isPlaying.value)
  const canUseExtraTime = computed(() => state.value.hints.extraTime > 0 && isPlaying.value)
  const canUseSkip = computed(() => state.value.hints.skip > 0 && isPlaying.value)
  const canUseHint = computed(() => state.value.hints.hint > 0 && isPlaying.value)

  // Форматированное время до восстановления жизни
  const timeToLifeRestoreFormatted = computed(() => {
    const seconds = state.value.timeToLifeRestore
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${minutes}:${secs.toString().padStart(2, '0')}`
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
        lives: state.value.lives,
        hints: state.value.hints,
        currentStreak: state.value.currentStreak,
        score: state.value.score,
        categoryId: state.value.categoryId,
        timestamp: Date.now()
      }))
    } catch (error) {
      console.error('Failed to save Marathon state:', error)
    }
  }

  const loadFromLocalStorage = () => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (!saved) return false

      const data = JSON.parse(saved)

      // Проверяем, что данные для текущего игрока
      if (data.playerId !== playerId) return false

      // Проверяем, что данные не старше 7 дней
      const age = Date.now() - data.timestamp
      if (age > 7 * 24 * 60 * 60 * 1000) {
        clearLocalStorage()
        return false
      }

      // Восстанавливаем состояние
      state.value.game = data.game
      state.value.currentQuestion = data.currentQuestion
      state.value.lives = data.lives
      state.value.hints = data.hints
      state.value.currentStreak = data.currentStreak
      state.value.score = data.score
      state.value.categoryId = data.categoryId
      state.value.status = 'playing'

      return true
    } catch (error) {
      console.error('Failed to load Marathon state:', error)
      return false
    }
  }

  const clearLocalStorage = () => {
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch (error) {
      console.error('Failed to clear Marathon state:', error)
    }
  }

  // ===========================
  // Game Actions
  // ===========================

  /**
   * Начать игру
   */
  const startGame = async (categoryId: string) => {
    try {
      state.value.status = 'loading'
      state.value.categoryId = categoryId

      const response = await startMutation.mutateAsync({
        data: {
          playerId,
          categoryId
        }
      })

      const gameData = response.data

      state.value.game = gameData.game
      state.value.currentQuestion = gameData.firstQuestion
      state.value.lives = gameData.game.lives.currentLives
      state.value.maxLives = gameData.game.lives.maxLives
      state.value.hints = {
        fiftyFifty: gameData.game.hints.fiftyFifty,
        extraTime: gameData.game.hints.extraTime,
        skip: gameData.game.hints.skip,
        hint: 0 // hint not in DTO
      }
      state.value.currentStreak = gameData.game.currentStreak
      state.value.score = gameData.game.baseScore
      state.value.personalBest = gameData.hasPersonalBest ? 0 : null
      state.value.status = 'playing'
      state.value.lastAnswerCorrect = null

      saveToLocalStorage()

      // Переходим на экран игры
      router.push({ name: 'marathon-play' })

      return true
    } catch (error: unknown) {
      state.value.status = 'error'
      console.error('Failed to start Marathon:', error)

      // Обработка ошибок
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { status?: number } }
        if (axiosError.response?.status === 409) {
          // Уже есть активная игра
          await refetchStatus()
        }
      }

      throw error
    }
  }

  /**
   * Отправить ответ на текущий вопрос
   */
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
          timeTaken
        }
      })

      const answerData = response.data

      // Сохраняем правильность ответа для feedback UI
      state.value.lastAnswerCorrect = answerData.isCorrect

      if (answerData.isGameOver) {
        // Игра окончена
        state.value.status = 'game-over'
        state.value.currentQuestion = null
        clearLocalStorage()

        // Обновляем рекорды
        await refetchPersonalBests()

        // Переходим на экран Game Over
        router.push({ name: 'marathon-gameover' })
      } else if (answerData.nextQuestion) {
        // Следующий вопрос
        state.value.currentQuestion = answerData.nextQuestion
        state.value.currentStreak = answerData.currentStreak
        state.value.score = state.value.score + answerData.basePoints
        state.value.lives = answerData.remainingLives

        saveToLocalStorage()
      }

      return answerData
    } catch (error) {
      console.error('Failed to submit answer:', error)
      throw error
    }
  }

  /**
   * Использовать подсказку
   */
  const useHint = async (hintType: HintType) => {
    if (!state.value.game?.id) {
      throw new Error('No active game')
    }

    // Проверяем доступность подсказки
    if (state.value.hints.fiftyFifty <= 0) {
      throw new Error(`No ${hintType} hints available`)
    }

    try {
      const response = await hintMutation.mutateAsync({
        gameId: state.value.game.id,
        data: {
          hintType,
          playerId,
          questionId: state.value.currentQuestion?.id || ''
        }
      })

      const hintData = response.data

      // Обновляем оставшиеся подсказки (remainingHints is now a number, not an object)
      // We'll decrement the specific hint type
      const hintKeyMap: Record<HintType, keyof typeof state.value.hints> = {
        'fifty_fifty': 'fiftyFifty',
        'extra_time': 'extraTime',
        'skip': 'skip',
        'hint': 'hint'
      }
      const hintKey = hintKeyMap[hintType]
      state.value.hints = {
        ...state.value.hints,
        [hintKey]: state.value.hints[hintKey] - 1
      }

      // Обработка результата подсказки
      if (hintType === 'skip' && hintData.hintResult.nextQuestion) {
        // Skip - показываем следующий вопрос
        state.value.currentQuestion = hintData.hintResult.nextQuestion
      } else if (hintType === 'fifty_fifty' && hintData.hintResult.hiddenAnswerIds) {
        // 50/50 - обновляем текущий вопрос с убранными вариантами
        // (UI должен отфильтровать hiddenAnswerIds)
      } else if (hintType === 'extra_time') {
        // +10 сек - таймер обновится автоматически
      }

      saveToLocalStorage()

      return hintData
    } catch (error) {
      console.error('Failed to use hint:', error)
      throw error
    }
  }

  /**
   * Завершить игру досрочно
   */
  const abandonGame = async () => {
    if (!state.value.game?.id) {
      throw new Error('No active game')
    }

    try {
      await abandonMutation.mutateAsync({
        gameId: state.value.game.id,
        data: {
          playerId
        }
      })

      state.value.status = 'game-over'
      state.value.currentQuestion = null
      clearLocalStorage()

      // Обновляем статус
      await refetchStatus()
      await refetchPersonalBests()

      // Переходим на главную или экран game over
      router.push({ name: 'marathon-gameover' })

      return true
    } catch (error) {
      console.error('Failed to abandon game:', error)
      throw error
    }
  }

  /**
   * Получить статус (жизни, таймер восстановления)
   */
  const checkStatus = async () => {
    try {
      await refetchStatus()

      if (statusData.value?.data) {
        const data = statusData.value.data
        if (data.game?.lives) {
          state.value.lives = data.game.lives.currentLives
          state.value.maxLives = data.game.lives.maxLives
          state.value.timeToLifeRestore = data.game.lives.timeToNextLife ?? 0
        }

        // Если есть активная игра - восстанавливаем
        if (data.hasActiveGame) {
          loadFromLocalStorage()
        }
      }
    } catch (error) {
      console.error('Failed to check status:', error)
    }
  }

  /**
   * Загрузить личные рекорды
   */
  const loadPersonalBests = async () => {
    try {
      await refetchPersonalBests()

      if (personalBestsData.value?.data?.personalBests) {
        // Находим рекорд для текущей категории
        const currentCategoryBest = personalBestsData.value.data.personalBests.find(
          (pb: InternalInfrastructureHttpHandlersMarathonPersonalBestDTO) =>
            pb.category.id === state.value.categoryId
        )

        if (currentCategoryBest) {
          state.value.personalBest = currentCategoryBest.bestScore
        }
      }
    } catch (error) {
      console.error('Failed to load personal bests:', error)
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
      lives: 3,
      maxLives: 3,
      hints: {
        fiftyFifty: 1,
        extraTime: 1,
        skip: 1,
        hint: 1
      },
      currentStreak: 0,
      score: 0,
      personalBest: null,
      categoryId: null,
      timeToLifeRestore: 0,
      lastAnswerCorrect: null
    }
    clearLocalStorage()
  }

  // ===========================
  // Lifecycle
  // ===========================

  const initialized = ref(false)

  const initialize = async () => {
    if (initialized.value) return

    // Пытаемся восстановить из localStorage
    const restored = loadFromLocalStorage()

    // Загружаем статус с сервера
    await checkStatus()
    await loadPersonalBests()

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
    isGameOver,
    isLoading,
    hasLives,
    canPlay,
    progressToRecord,
    livesPercent,
    canUseFiftyFifty,
    canUseExtraTime,
    canUseSkip,
    canUseHint,
    timeToLifeRestoreFormatted,

    // Actions
    startGame,
    submitAnswer,
    useHint,
    abandonGame,
    checkStatus,
    loadPersonalBests,
    reset,
    initialize,

    // Data from API
    statusData,
    personalBestsData
  }
}
