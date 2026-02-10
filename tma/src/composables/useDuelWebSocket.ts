import { ref, onUnmounted, computed } from 'vue'

export interface DuelQuestion {
  id: string
  questionNumber: number
  text: string
  answers: { id: string; text: string }[]
  timeLimit: number
  serverTime: number
}

export interface DuelPlayer {
  id: string
  username: string
  avatar?: string
  mmr: number
  league: string
  leagueIcon: string
  score: number
  connected: boolean
}

export interface DuelGame {
  id: string
  status: 'waiting' | 'countdown' | 'in_progress' | 'finished'
  player1: DuelPlayer
  player2: DuelPlayer
  currentRound: number
  totalRounds: number
  startedAt: number
  finishedAt?: number
  winnerId?: string
}

export interface RoundResult {
  questionNumber: number
  player1Correct: boolean
  player1Time: number
  player2Correct: boolean
  player2Time: number
  correctAnswerId: string
}

export interface GameResult {
  winnerId?: string
  player1Score: number
  player2Score: number
  player1MmrChange: number
  player2MmrChange: number
  player1NewMmr: number
  player2NewMmr: number
}

// WebSocket message types matching backend implementation
type WebSocketMessage =
  | { type: 'connected'; data: { gameId: string; playerId: string } }
  | {
      type: 'game_ready'
      data: {
        gameId: string
        player1Id: string
        player2Id: string
        startsIn: number
        totalRounds: number
      }
    }
  | {
      type: 'new_question'
      data: {
        roundNum: number
        totalRounds: number
        question: { id: string; text: string; answers: { id: string; text: string }[]; timeLimit: number }
        serverTime: number
      }
    }
  | {
      type: 'answer_result'
      data: {
        playerId: string
        questionId: string
        isCorrect: boolean
        correctAnswer: string
        pointsEarned: number
        timeTaken: number
        player1Score: number
        player2Score: number
      }
    }
  | {
      type: 'round_complete'
      data: { roundNum: number; player1Score: number; player2Score: number; nextRoundIn: number }
    }
  | { type: 'round_timeout'; data: { roundNum: number } }
  | {
      type: 'game_complete'
      data: {
        winnerId?: string
        player1Score: number
        player2Score: number
        player1MMRChange: number
        player2MMRChange: number
        player1NewMMR: number
        player2NewMMR: number
      }
    }
  | { type: 'opponent_disconnected'; data: { playerId: string; reconnectIn: number } }
  | { type: 'error'; error: string }
  | { type: 'pong' }

/**
 * WebSocket composable for real-time duel gameplay
 */
export function useDuelWebSocket(gameId: string, playerId: string) {
  // ===========================
  // State
  // ===========================

  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const isReconnecting = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5

  // Game state
  const game = ref<DuelGame | null>(null)
  const currentQuestion = ref<DuelQuestion | null>(null)
  const currentRound = ref(0)
  const countdownSeconds = ref(0)
  const opponentAnswered = ref(false)
  const lastRoundResult = ref<RoundResult | null>(null)
  const gameResult = ref<GameResult | null>(null)
  const error = ref<string | null>(null)

  // ===========================
  // Computed
  // ===========================

  const isPlayer1 = computed(() => game.value?.player1.id === playerId)

  const myPlayer = computed(() => {
    if (!game.value) return null
    return isPlayer1.value ? game.value.player1 : game.value.player2
  })

  const opponent = computed(() => {
    if (!game.value) return null
    return isPlayer1.value ? game.value.player2 : game.value.player1
  })

  const myScore = computed(() => myPlayer.value?.score ?? 0)
  const opponentScore = computed(() => opponent.value?.score ?? 0)

  const isWaiting = computed(() => game.value?.status === 'waiting')
  const isCountdown = computed(() => game.value?.status === 'countdown')
  const isPlaying = computed(() => game.value?.status === 'in_progress')
  const isFinished = computed(() => game.value?.status === 'finished')

  const didWin = computed(() => {
    if (!gameResult.value || !gameResult.value.winnerId) return null
    return gameResult.value.winnerId === playerId
  })

  const myMmrChange = computed(() => {
    if (!gameResult.value) return 0
    return isPlayer1.value
      ? gameResult.value.player1MmrChange
      : gameResult.value.player2MmrChange
  })

  // ===========================
  // WebSocket Connection
  // ===========================

  const connect = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/ws/duel/${gameId}?playerId=${playerId}`

    console.log('[DuelWS] Connecting to:', wsUrl)

    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('[DuelWS] Connected')
      isConnected.value = true
      isReconnecting.value = false
      reconnectAttempts.value = 0
      error.value = null
    }

    ws.value.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)
        handleMessage(message)
      } catch (e) {
        console.error('[DuelWS] Failed to parse message:', e)
      }
    }

    ws.value.onclose = (event) => {
      console.log('[DuelWS] Disconnected:', event.code, event.reason)
      isConnected.value = false

      // Try to reconnect if not finished
      if (!isFinished.value && reconnectAttempts.value < maxReconnectAttempts) {
        isReconnecting.value = true
        reconnectAttempts.value++
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), 10000)
        console.log(`[DuelWS] Reconnecting in ${delay}ms (attempt ${reconnectAttempts.value})`)
        setTimeout(connect, delay)
      }
    }

    ws.value.onerror = (event) => {
      console.error('[DuelWS] Error:', event)
      error.value = 'Connection error'
    }
  }

  const disconnect = () => {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    isConnected.value = false
  }

  // ===========================
  // Message Handling
  // ===========================

  const handleMessage = (message: WebSocketMessage) => {
    console.log('[DuelWS] Message:', message.type, message)

    switch (message.type) {
      case 'connected':
        console.log('[DuelWS] Connected to game:', message.data.gameId)
        break

      case 'game_ready':
        // Initialize game state when both players are ready
        game.value = {
          id: message.data.gameId,
          status: 'countdown',
          player1: {
            id: message.data.player1Id,
            username: '',
            mmr: 0,
            league: '',
            leagueIcon: '',
            score: 0,
            connected: true,
          },
          player2: {
            id: message.data.player2Id,
            username: '',
            mmr: 0,
            league: '',
            leagueIcon: '',
            score: 0,
            connected: true,
          },
          currentRound: 0,
          totalRounds: message.data.totalRounds,
          startedAt: Date.now(),
        }
        countdownSeconds.value = message.data.startsIn
        // Countdown timer
        const countdownInterval = setInterval(() => {
          countdownSeconds.value--
          if (countdownSeconds.value <= 0) {
            clearInterval(countdownInterval)
          }
        }, 1000)
        break

      case 'new_question':
        currentQuestion.value = {
          id: message.data.question.id,
          questionNumber: message.data.roundNum,
          text: message.data.question.text,
          answers: message.data.question.answers,
          timeLimit: message.data.question.timeLimit,
          serverTime: message.data.serverTime,
        }
        currentRound.value = message.data.roundNum
        opponentAnswered.value = false
        lastRoundResult.value = null
        if (game.value) {
          game.value.status = 'in_progress'
          game.value.currentRound = message.data.roundNum
          game.value.totalRounds = message.data.totalRounds
        }
        break

      case 'answer_result':
        // Check if this is opponent's answer (mark opponent as answered)
        if (message.data.playerId !== playerId) {
          opponentAnswered.value = true
        }
        // Update scores
        if (game.value) {
          game.value.player1.score = message.data.player1Score
          game.value.player2.score = message.data.player2Score
        }
        // Store last result for feedback
        lastRoundResult.value = {
          questionNumber: currentRound.value,
          player1Correct: false, // Will be filled from specific player data
          player1Time: 0,
          player2Correct: false,
          player2Time: 0,
          correctAnswerId: message.data.correctAnswer,
        }
        break

      case 'round_complete':
        // Both players answered, round is complete
        if (game.value) {
          game.value.player1.score = message.data.player1Score
          game.value.player2.score = message.data.player2Score
        }
        break

      case 'round_timeout':
        // Time ran out for the round
        console.log('[DuelWS] Round timeout:', message.data.roundNum)
        break

      case 'game_complete':
        gameResult.value = {
          winnerId: message.data.winnerId,
          player1Score: message.data.player1Score,
          player2Score: message.data.player2Score,
          player1MmrChange: message.data.player1MMRChange,
          player2MmrChange: message.data.player2MMRChange,
          player1NewMmr: message.data.player1NewMMR,
          player2NewMmr: message.data.player2NewMMR,
        }
        if (game.value) {
          game.value.status = 'finished'
          game.value.winnerId = message.data.winnerId
        }
        break

      case 'opponent_disconnected':
        if (opponent.value) {
          opponent.value.connected = false
        }
        break

      case 'error':
        error.value = message.error
        break

      case 'pong':
        // Heartbeat response
        break
    }
  }

  // ===========================
  // Actions
  // ===========================

  const sendAnswer = (answerId: string, timeTaken: number) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      console.error('[DuelWS] Cannot send answer: not connected')
      return
    }

    ws.value.send(
      JSON.stringify({
        type: 'answer',
        questionId: currentQuestion.value?.id,
        answerId,
        timeTaken,
      }),
    )
  }

  const sendReady = () => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return

    ws.value.send(JSON.stringify({ type: 'ready' }))
  }

  // ===========================
  // Lifecycle
  // ===========================

  onUnmounted(() => {
    disconnect()
  })

  // ===========================
  // Return
  // ===========================

  return {
    // Connection
    isConnected,
    isReconnecting,
    error,
    connect,
    disconnect,

    // Game state
    game,
    currentQuestion,
    currentRound,
    countdownSeconds,
    opponentAnswered,
    lastRoundResult,
    gameResult,

    // Computed
    isPlayer1,
    myPlayer,
    opponent,
    myScore,
    opponentScore,
    isWaiting,
    isCountdown,
    isPlaying,
    isFinished,
    didWin,
    myMmrChange,

    // Actions
    sendAnswer,
    sendReady,
  }
}
