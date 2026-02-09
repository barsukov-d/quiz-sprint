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

export interface DuelMatch {
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

export interface MatchResult {
  winnerId?: string
  player1Score: number
  player2Score: number
  player1MmrChange: number
  player2MmrChange: number
  player1NewMmr: number
  player2NewMmr: number
}

type WebSocketMessage =
  | { type: 'match_info'; match: DuelMatch }
  | { type: 'countdown'; seconds: number }
  | { type: 'round_start'; question: DuelQuestion; round: number }
  | { type: 'opponent_answered' }
  | { type: 'round_result'; result: RoundResult; scores: { player1: number; player2: number } }
  | { type: 'match_finished'; result: MatchResult }
  | { type: 'opponent_disconnected' }
  | { type: 'opponent_reconnected' }
  | { type: 'error'; message: string }

/**
 * WebSocket composable for real-time duel gameplay
 */
export function useDuelWebSocket(matchId: string, playerId: string) {
  // ===========================
  // State
  // ===========================

  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const isReconnecting = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5

  // Match state
  const match = ref<DuelMatch | null>(null)
  const currentQuestion = ref<DuelQuestion | null>(null)
  const currentRound = ref(0)
  const countdownSeconds = ref(0)
  const opponentAnswered = ref(false)
  const lastRoundResult = ref<RoundResult | null>(null)
  const matchResult = ref<MatchResult | null>(null)
  const error = ref<string | null>(null)

  // ===========================
  // Computed
  // ===========================

  const isPlayer1 = computed(() => match.value?.player1.id === playerId)

  const myPlayer = computed(() => {
    if (!match.value) return null
    return isPlayer1.value ? match.value.player1 : match.value.player2
  })

  const opponent = computed(() => {
    if (!match.value) return null
    return isPlayer1.value ? match.value.player2 : match.value.player1
  })

  const myScore = computed(() => myPlayer.value?.score ?? 0)
  const opponentScore = computed(() => opponent.value?.score ?? 0)

  const isWaiting = computed(() => match.value?.status === 'waiting')
  const isCountdown = computed(() => match.value?.status === 'countdown')
  const isPlaying = computed(() => match.value?.status === 'in_progress')
  const isFinished = computed(() => match.value?.status === 'finished')

  const didWin = computed(() => {
    if (!matchResult.value || !matchResult.value.winnerId) return null
    return matchResult.value.winnerId === playerId
  })

  const myMmrChange = computed(() => {
    if (!matchResult.value) return 0
    return isPlayer1.value
      ? matchResult.value.player1MmrChange
      : matchResult.value.player2MmrChange
  })

  // ===========================
  // WebSocket Connection
  // ===========================

  const connect = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/ws/duel/${matchId}?playerId=${playerId}`

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
      case 'match_info':
        match.value = message.match
        break

      case 'countdown':
        countdownSeconds.value = message.seconds
        if (match.value) {
          match.value.status = 'countdown'
        }
        break

      case 'round_start':
        currentQuestion.value = message.question
        currentRound.value = message.round
        opponentAnswered.value = false
        lastRoundResult.value = null
        if (match.value) {
          match.value.status = 'in_progress'
          match.value.currentRound = message.round
        }
        break

      case 'opponent_answered':
        opponentAnswered.value = true
        break

      case 'round_result':
        lastRoundResult.value = message.result
        if (match.value) {
          match.value.player1.score = message.scores.player1
          match.value.player2.score = message.scores.player2
        }
        break

      case 'match_finished':
        matchResult.value = message.result
        if (match.value) {
          match.value.status = 'finished'
          match.value.winnerId = message.result.winnerId
        }
        break

      case 'opponent_disconnected':
        if (opponent.value) {
          opponent.value.connected = false
        }
        break

      case 'opponent_reconnected':
        if (opponent.value) {
          opponent.value.connected = true
        }
        break

      case 'error':
        error.value = message.message
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

    // Match state
    match,
    currentQuestion,
    currentRound,
    countdownSeconds,
    opponentAnswered,
    lastRoundResult,
    matchResult,

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
