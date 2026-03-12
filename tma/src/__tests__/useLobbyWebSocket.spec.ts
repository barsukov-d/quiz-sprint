import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock WebSocket
class MockWebSocket {
  static OPEN = 1
  static CLOSED = 3
  readyState = MockWebSocket.OPEN
  onopen: (() => void) | null = null
  onmessage: ((e: MessageEvent) => void) | null = null
  onclose: ((e: CloseEvent) => void) | null = null
  onerror: ((e: Event) => void) | null = null
  sentMessages: string[] = []

  send(data: string) {
    this.sentMessages.push(data)
  }
  close() {
    this.readyState = MockWebSocket.CLOSED
  }
  triggerOpen() {
    this.onopen?.()
  }
  triggerMessage(data: object) {
    this.onmessage?.({ data: JSON.stringify(data) } as MessageEvent)
  }
  triggerClose() {
    this.onclose?.({ code: 1006 } as CloseEvent)
  }
}

vi.stubGlobal('WebSocket', MockWebSocket)

beforeEach(() => {
  vi.resetModules()
})

describe('useLobbyWebSocket', () => {
  it('exports connect, disconnect, on, isConnected', async () => {
    const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
    const result = useLobbyWebSocket('player1')

    expect(typeof result.connect).toBe('function')
    expect(typeof result.disconnect).toBe('function')
    expect(typeof result.on).toBe('function')
    expect(typeof result.isConnected).toBe('object') // ref
    expect(typeof result.isPollingFallback).toBe('object') // ref
    expect(typeof result.setupVisibilityHandler).toBe('function')
  })

  it('on() returns unsubscribe function', async () => {
    const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
    const { on } = useLobbyWebSocket('player1')

    const unsub = on('challenge_accepted', () => {})
    expect(typeof unsub).toBe('function')
  })

  it('isConnected is false initially', async () => {
    const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
    const { isConnected } = useLobbyWebSocket('player1')
    expect(isConnected.value).toBe(false)
  })
})
