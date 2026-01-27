import { ref, computed, onBeforeUnmount } from 'vue'

export interface GameTimerOptions {
  /**
   * Начальное время в секундах
   */
  initialTime: number

  /**
   * Callback при окончании времени
   */
  onTimeout?: () => void

  /**
   * Callback каждую секунду
   */
  onTick?: (remainingTime: number) => void

  /**
   * Автоматический старт таймера
   */
  autoStart?: boolean

  /**
   * Включить звуковое предупреждение на последних секундах
   */
  soundWarning?: boolean

  /**
   * За сколько секунд включить предупреждение
   */
  warningThreshold?: number
}

/**
 * Composable для управления игровым таймером
 *
 * Возможности:
 * - Обратный отсчёт времени
 * - Пауза/возобновление
 * - Добавление времени (для подсказки +10сек)
 * - Форматирование времени
 * - Автоматическое срабатывание callback при окончании
 * - Визуальные/звуковые предупреждения
 */
export function useGameTimer(options: GameTimerOptions) {
  const {
    initialTime,
    onTimeout,
    onTick,
    autoStart = false,
    soundWarning = false,
    warningThreshold = 5
  } = options

  // ===========================
  // State
  // ===========================

  const remainingTime = ref(initialTime)
  const isRunning = ref(false)
  const isPaused = ref(false)
  const startTime = ref<number | null>(null)
  const pausedAt = ref<number | null>(null)
  const intervalId = ref<number | null>(null)

  // ===========================
  // Computed Properties
  // ===========================

  // Форматированное время (MM:SS)
  const formattedTime = computed(() => {
    const minutes = Math.floor(remainingTime.value / 60)
    const seconds = remainingTime.value % 60
    return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
  })

  // Прогресс (0-100%)
  const progress = computed(() => {
    if (initialTime === 0) return 0
    return Math.round((remainingTime.value / initialTime) * 100)
  })

  // Осталось процентов времени
  const percentRemaining = computed(() => progress.value)

  // Критическое ли время (последние секунды)
  const isWarning = computed(() => remainingTime.value <= warningThreshold && remainingTime.value > 0)

  // Закончилось ли время
  const isExpired = computed(() => remainingTime.value <= 0)

  // Общее затраченное время (в миллисекундах)
  const elapsedTime = computed(() => {
    if (!startTime.value) return 0
    const now = pausedAt.value || Date.now()
    return now - startTime.value
  })

  // ===========================
  // Timer Logic
  // ===========================

  const tick = () => {
    if (!isRunning.value || isPaused.value) return

    remainingTime.value--

    // Callback каждую секунду
    if (onTick) {
      onTick(remainingTime.value)
    }

    // Звуковое предупреждение
    if (soundWarning && isWarning.value && remainingTime.value > 0) {
      // TODO: Добавить звук
      // playWarningSound()
    }

    // Проверка окончания времени
    if (remainingTime.value <= 0) {
      stop()

      if (onTimeout) {
        onTimeout()
      }
    }
  }

  /**
   * Запустить таймер
   */
  const start = () => {
    if (isRunning.value) return

    isRunning.value = true
    isPaused.value = false

    if (!startTime.value) {
      startTime.value = Date.now()
    } else if (pausedAt.value) {
      // Возобновление после паузы - корректируем startTime
      const pauseDuration = Date.now() - pausedAt.value
      startTime.value += pauseDuration
      pausedAt.value = null
    }

    intervalId.value = window.setInterval(tick, 1000)
  }

  /**
   * Остановить таймер
   */
  const stop = () => {
    if (!isRunning.value) return

    isRunning.value = false
    isPaused.value = false

    if (intervalId.value !== null) {
      clearInterval(intervalId.value)
      intervalId.value = null
    }
  }

  /**
   * Пауза таймера
   */
  const pause = () => {
    if (!isRunning.value || isPaused.value) return

    isPaused.value = true
    pausedAt.value = Date.now()

    if (intervalId.value !== null) {
      clearInterval(intervalId.value)
      intervalId.value = null
    }
  }

  /**
   * Возобновить таймер после паузы
   */
  const resume = () => {
    if (!isPaused.value) return

    isPaused.value = false

    if (pausedAt.value) {
      const pauseDuration = Date.now() - pausedAt.value
      if (startTime.value) {
        startTime.value += pauseDuration
      }
      pausedAt.value = null
    }

    intervalId.value = window.setInterval(tick, 1000)
  }

  /**
   * Перезапустить таймер
   */
  const reset = (newTime?: number) => {
    stop()

    remainingTime.value = newTime ?? initialTime
    startTime.value = null
    pausedAt.value = null

    if (autoStart) {
      start()
    }
  }

  /**
   * Добавить время (для подсказки +10сек)
   */
  const addTime = (seconds: number) => {
    remainingTime.value += seconds

    // Не превышаем начальное время
    if (remainingTime.value > initialTime * 2) {
      remainingTime.value = initialTime * 2
    }
  }

  /**
   * Установить новое время
   */
  const setTime = (seconds: number) => {
    remainingTime.value = seconds
  }

  // ===========================
  // Lifecycle
  // ===========================

  // Автоматический старт
  if (autoStart) {
    start()
  }

  // Очистка при размонтировании компонента
  onBeforeUnmount(() => {
    stop()
  })

  // ===========================
  // Return
  // ===========================

  return {
    // State
    remainingTime,
    isRunning,
    isPaused,
    elapsedTime,

    // Computed
    formattedTime,
    progress,
    percentRemaining,
    isWarning,
    isExpired,

    // Actions
    start,
    stop,
    pause,
    resume,
    reset,
    addTime,
    setTime
  }
}
