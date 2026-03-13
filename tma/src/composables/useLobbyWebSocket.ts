import { ref, onUnmounted } from 'vue'

export type LobbyEventType =
	| 'connected'
	| 'challenge_received'
	| 'challenge_accepted'
	| 'challenge_declined'
	| 'challenge_expired'
	| 'game_ready'
	| 'queue_matched'

export interface LobbyEvent {
	type: LobbyEventType
	data?: Record<string, unknown>
}

type EventHandler = (data: Record<string, unknown> | undefined) => void

const getWsBase = (): string => {
	const h = window.location.hostname
	if (h === 'dev.quiz-sprint-tma.online') return 'wss://api-dev.quiz-sprint-tma.online'
	if (h === 'staging.quiz-sprint-tma.online') return 'wss://api-staging.quiz-sprint-tma.online'
	if (h === 'quiz-sprint-tma.online') return 'wss://api.quiz-sprint-tma.online'
	const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
	return `${proto}//${window.location.host}`
}

export function useLobbyWebSocket(playerId: string) {
	const isConnected = ref(false)
	const isPollingFallback = ref(false)

	let ws: WebSocket | null = null
	let reconnectAttempts = 0
	const maxReconnectAttempts = 5
	const handlers = new Map<string, Set<EventHandler>>()
	let visibilityHandler: (() => void) | null = null
	let reconnectTimeout: ReturnType<typeof setTimeout> | null = null
	let pingInterval: ReturnType<typeof setInterval> | null = null

	const emit = (type: string, data: Record<string, unknown> | undefined) => {
		handlers.get(type)?.forEach((h) => h(data))
	}

	const on = (type: LobbyEventType, handler: EventHandler) => {
		if (!handlers.has(type)) handlers.set(type, new Set())
		handlers.get(type)!.add(handler)
		return () => handlers.get(type)?.delete(handler)
	}

	const connect = () => {
		if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING))
			return
		if (!playerId) return

		const url = `${getWsBase()}/ws/duel/lobby?playerId=${playerId}`
		ws = new WebSocket(url)

		ws.onopen = () => {
			isConnected.value = true
			isPollingFallback.value = false
			reconnectAttempts = 0
			pingInterval = setInterval(() => {
				if (ws && ws.readyState === WebSocket.OPEN) {
					ws.send(JSON.stringify({ type: 'ping' }))
				}
			}, 30000)
		}

		ws.onmessage = (event) => {
			try {
				const msg: LobbyEvent = JSON.parse(event.data)
				emit(msg.type, msg.data)
			} catch {
				// ignore malformed messages
			}
		}

		ws.onclose = () => {
			isConnected.value = false
			ws = null
			if (pingInterval) {
				clearInterval(pingInterval)
				pingInterval = null
			}
			if (reconnectAttempts < maxReconnectAttempts) {
				const delay = Math.min(500 * Math.pow(2, reconnectAttempts), 8000)
				reconnectAttempts++
				reconnectTimeout = setTimeout(connect, delay)
			} else {
				isPollingFallback.value = true
			}
		}

		ws.onerror = () => {
			// onclose fires after onerror; reconnect logic handled there
		}
	}

	const disconnect = () => {
		if (reconnectTimeout) {
			clearTimeout(reconnectTimeout)
			reconnectTimeout = null
		}
		if (pingInterval) {
			clearInterval(pingInterval)
			pingInterval = null
		}
		if (ws) {
			ws.onclose = null // prevent reconnect loop
			ws.close()
			ws = null
		}
		isConnected.value = false
	}

	// TMA lifecycle: disconnect when app is hidden, reconnect when it returns.
	// Disconnecting on hide triggers SetOffline on the backend immediately,
	// so rivals see the player go offline as soon as they switch away.
	const setupVisibilityHandler = (onVisible: () => void) => {
		visibilityHandler = () => {
			if (document.hidden) {
				disconnect()
			} else {
				reconnectAttempts = 0
				connect()
				onVisible()
			}
		}
		document.addEventListener('visibilitychange', visibilityHandler)
	}

	onUnmounted(() => {
		disconnect()
		if (visibilityHandler) document.removeEventListener('visibilitychange', visibilityHandler)
	})

	return {
		isConnected,
		isPollingFallback,
		connect,
		disconnect,
		on,
		setupVisibilityHandler,
	}
}
