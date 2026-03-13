# Global Challenge Notifications Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show a real-time toast with inline Accept/Decline to any authenticated player who receives a direct challenge, regardless of which screen they are on.

**Architecture:** Backend enriches `ChallengeCreatedEvent` with `challengerUsername` at the application layer. Frontend lifts the WebSocket singleton from `DuelLobbyView` to `App.vue` via `provide/inject`. A new `useGlobalDuelNotifications` composable subscribes to `challenge_received` and shows a Nuxt UI toast; `usePvPDuel` injects the shared WS instance instead of creating its own.

**Tech Stack:** Vue 3.5, TypeScript, Nuxt UI (`useToast` from `@nuxt/ui`), Vitest, Go 1.25, Fiber

**Spec:** `docs/superpowers/specs/2026-03-13-global-challenge-notifications-design.md`

---

## Chunk 1: Backend — ChallengeCreatedEvent enrichment

### Task 1: Add `challengerUsername` to `ChallengeCreatedEvent`

**Files:**
- Modify: `backend/internal/domain/quick_duel/events.go`

- [ ] **Step 1: Write failing Go test**

Check if `events_test.go` exists in the domain package:
```bash
ls backend/internal/domain/quick_duel/events_test.go 2>/dev/null || echo "not found"
```

If not found, create `backend/internal/domain/quick_duel/events_test.go`:

```go
package quick_duel_test

import (
	"testing"

	quick_duel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

func TestChallengeCreatedEvent_WithChallengerUsername(t *testing.T) {
	challengerID, _ := quick_duel.NewUserID("111")
	friendID, _ := quick_duel.NewUserID("222")
	challengeID := quick_duel.NewChallengeID()

	evt := quick_duel.NewChallengeCreatedEvent(
		challengeID,
		challengerID,
		&friendID,
		quick_duel.DirectChallenge,
		0,
		0,
	)

	enriched := evt.WithChallengerUsername("Pavel")

	if enriched.ChallengerUsername() != "Pavel" {
		t.Errorf("expected 'Pavel', got %q", enriched.ChallengerUsername())
	}
	// Original must not be modified (value receiver copy semantics)
	if evt.ChallengerUsername() != "" {
		t.Errorf("original event should not be modified, got %q", evt.ChallengerUsername())
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd backend && go test ./internal/domain/quick_duel/... -run TestChallengeCreatedEvent_WithChallengerUsername -v
```

Expected: `FAIL` — `evt.WithChallengerUsername undefined`

- [ ] **Step 3: Add field + methods to `events.go`**

In `backend/internal/domain/quick_duel/events.go`, find the `ChallengeCreatedEvent` struct (line ~392). Add the `challengerUsername` field and two methods:

```go
// ChallengeCreatedEvent fired when a challenge is created
type ChallengeCreatedEvent struct {
	challengeID        ChallengeID
	challengerID       UserID
	challengedID       *UserID
	challengeType      ChallengeType
	expiresAt          int64
	occurredAt         int64
	challengerUsername string // notification hint; set by application layer
}

// WithChallengerUsername returns a copy with username set. Called by application layer.
func (e ChallengeCreatedEvent) WithChallengerUsername(name string) ChallengeCreatedEvent {
	e.challengerUsername = name
	return e
}

func (e ChallengeCreatedEvent) ChallengerUsername() string { return e.challengerUsername }
```

The `NewChallengeCreatedEvent` factory signature is **unchanged** — domain aggregate stays ID-only.

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/domain/quick_duel/... -v
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/quick_duel/events.go backend/internal/domain/quick_duel/events_test.go
git commit -m "feat(duel): add ChallengerUsername notification hint to ChallengeCreatedEvent"
```

---

### Task 2: Enrich event in `SendChallengeUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go` (lines 383–403)
- Modify: `backend/internal/application/quick_duel/use_cases_test.go`

The `Execute` method currently publishes events at line 384 (before user fetch), then fetches the challenger user at line 391 only inside a Telegram notification guard. We need the username for WS enrichment regardless, so hoist the lookup before the event loop — but **keep the Telegram notification block guarded by `inviteeTgID > 0`**.

- [ ] **Step 1: Write failing test in `use_cases_test.go`**

The fixture pre-populates `testPlayer1ID = "player111"` with username `"Player1"`. Add this test after the existing `TestSendChallenge_Success` block:

```go
func TestSendChallenge_PublishesEventWithChallengerUsername(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSendChallengeUC()

	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, evt := range f.eventBus.events {
		if e, ok := evt.(quick_duel.ChallengeCreatedEvent); ok {
			found = true
			if e.ChallengerUsername() == "" {
				t.Error("ChallengerUsername should not be empty after enrichment")
			}
			if e.ChallengerUsername() != "Player1" {
				t.Errorf("ChallengerUsername = %q, want %q", e.ChallengerUsername(), "Player1")
			}
		}
	}
	if !found {
		t.Fatal("no ChallengeCreatedEvent found in published events")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestSendChallenge_PublishesEventWithChallengerUsername -v
```

Expected: FAIL — `ChallengerUsername = "", want "Player1"`

- [ ] **Step 3: Replace lines 383–403**

The existing block (lines 383–403) is:
```go
// Publish events
for _, event := range challenge.Events() {
    uc.eventBus.Publish(event)
}

// Send Telegram notification to invitee (best-effort)
if inviteeTgID, err := strconv.ParseInt(friendID.String(), 10, 64); err == nil && inviteeTgID > 0 {
    inviterName := challengerID.String()
    if u, err := uc.userRepo.FindByID(challengerID); err == nil && u != nil {
        if n := u.TelegramUsername().String(); n != "" {
            inviterName = n
        } else if n := u.Username().String(); n != "" {
            inviterName = n
        }
    }
    deepLink := ...
    ...
}
```

Replace with:
```go
// Fetch challenger display name — used for WS enrichment and Telegram notification
inviterName := challengerID.String() // ID as fallback
if u, err := uc.userRepo.FindByID(challengerID); err == nil && u != nil {
	if n := u.TelegramUsername().String(); n != "" {
		inviterName = n
	} else if n := u.Username().String(); n != "" {
		inviterName = n
	}
}

// Publish events — enrich ChallengeCreatedEvent with challenger username
for _, event := range challenge.Events() {
	if e, ok := event.(quick_duel.ChallengeCreatedEvent); ok {
		event = e.WithChallengerUsername(inviterName)
	}
	uc.eventBus.Publish(event)
}

// Send Telegram notification to invitee (best-effort)
if inviteeTgID, err := strconv.ParseInt(friendID.String(), 10, 64); err == nil && inviteeTgID > 0 {
	deepLink := "https://t.me/" + uc.botUsername + "?startapp=challenge_" + challenge.ID().String()
	if msgID, err := uc.notifier.NotifyChallengeReceived(context.Background(), inviteeTgID, inviterName, deepLink); err == nil && msgID > 0 {
		challenge.SetTelegramMessageID(msgID)
		_ = uc.challengeRepo.Save(challenge)
	}
}
```

Key points:
- `inviterName` defaults to `challengerID.String()` (not empty string) — same fallback as the original
- Telegram notification block stays inside `inviteeTgID > 0` guard — preserved
- `context.Background()` in `NotifyChallengeReceived` — preserved

- [ ] **Step 4: Build to verify no compile errors**

```bash
cd backend && go build ./...
```

- [ ] **Step 5: Run the new test**

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestSendChallenge_PublishesEventWithChallengerUsername -v
```

Expected: PASS

- [ ] **Step 6: Run all backend tests**

```bash
cd backend && go test ./...
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go backend/internal/application/quick_duel/use_cases_test.go
git commit -m "feat(duel): enrich ChallengeCreatedEvent with challenger username before WS publish"
```

---

### Task 3: Add `challengerUsername` to `challenge_received` WS payload

**Files:**
- Modify: `backend/internal/infrastructure/messaging/lobby_event_bus.go` (lines 23–33)

- [ ] **Step 1: Update the `ChallengeCreatedEvent` case**

In `lobby_event_bus.go`, find the `ChallengeCreatedEvent` case and update the Data map:

```go
case domainDuel.ChallengeCreatedEvent:
	// Direct challenge: notify invitee if connected
	if e.ChallengedID() != nil {
		b.hub.Notify(e.ChallengedID().String(), appDuel.LobbyEvent{
			Type: "challenge_received",
			Data: map[string]interface{}{
				"challengeId":        e.ChallengeID().String(),
				"expiresIn":          domainDuel.DirectChallengeExpirySeconds,
				"challengerUsername": e.ChallengerUsername(),
			},
		})
	}
```

- [ ] **Step 2: Build + run all backend tests**

```bash
cd backend && go build ./... && go test ./...
```

Expected: `ok` on all packages

- [ ] **Step 3: Commit**

```bash
git add backend/internal/infrastructure/messaging/lobby_event_bus.go
git commit -m "feat(duel): include challengerUsername in challenge_received WS payload"
```

---

## Chunk 2: Frontend — `useLobbyWebSocket` refactor

### Task 4: Update tests to the new API (write failing tests first)

**Files:**
- Modify: `tma/src/__tests__/useLobbyWebSocket.spec.ts`

The current test file (3 tests) uses the old API:
- `useLobbyWebSocket('player1')` — positional arg
- checks `setupVisibilityHandler` exists

Replace with the new API tests:

- [ ] **Step 1: Rewrite `useLobbyWebSocket.spec.ts`**

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock WebSocket
class MockWebSocket {
	static OPEN = 1
	static CLOSED = 3
	readyState = MockWebSocket.OPEN
	url = ''
	onopen: (() => void) | null = null
	onmessage: ((e: MessageEvent) => void) | null = null
	onclose: ((e: CloseEvent) => void) | null = null
	onerror: ((e: Event) => void) | null = null

	constructor(url: string) {
		this.url = url
	}
	close() {
		this.readyState = MockWebSocket.CLOSED
	}
	triggerOpen() { this.onopen?.() }
	triggerMessage(data: object) {
		this.onmessage?.({ data: JSON.stringify(data) } as MessageEvent)
	}
	triggerClose() { this.onclose?.({ code: 1006 } as CloseEvent) }
}

vi.stubGlobal('WebSocket', MockWebSocket)

let lastWs: MockWebSocket | null = null
const OriginalMockWS = MockWebSocket
vi.stubGlobal('WebSocket', class extends OriginalMockWS {
	constructor(url: string) {
		super(url)
		lastWs = this
	}
})

beforeEach(() => {
	vi.resetModules()
	lastWs = null
})

describe('useLobbyWebSocket', () => {
	it('exports connect, disconnect, on, addVisibilityHandler, isConnected', async () => {
		const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
		const result = useLobbyWebSocket()

		expect(typeof result.connect).toBe('function')
		expect(typeof result.disconnect).toBe('function')
		expect(typeof result.on).toBe('function')
		expect(typeof result.addVisibilityHandler).toBe('function')
		expect(typeof result.isConnected).toBe('object') // ref
		expect(typeof result.isPollingFallback).toBe('object') // ref
	})

	it('connect(playerId) opens WebSocket to correct URL', async () => {
		vi.stubGlobal('location', { hostname: 'localhost', protocol: 'http:', host: 'localhost:5173' })
		const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
		const { connect } = useLobbyWebSocket()

		connect('player42')

		expect(lastWs?.url).toContain('playerId=player42')
	})

	it('isConnected is false initially, true after onopen', async () => {
		const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
		const { connect, isConnected } = useLobbyWebSocket()

		expect(isConnected.value).toBe(false)
		connect('p1')
		lastWs?.triggerOpen()
		expect(isConnected.value).toBe(true)
	})

	it('on() returns unsubscribe function that removes handler', async () => {
		const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
		const { connect, on } = useLobbyWebSocket()
		connect('p1')
		lastWs?.triggerOpen()

		const handler = vi.fn()
		const unsub = on('challenge_accepted', handler)

		lastWs?.triggerMessage({ type: 'challenge_accepted' })
		expect(handler).toHaveBeenCalledOnce()

		unsub()
		lastWs?.triggerMessage({ type: 'challenge_accepted' })
		expect(handler).toHaveBeenCalledOnce() // still once — handler removed
	})

	it('addVisibilityHandler supports multiple callbacks and returns individual unsubscribe', async () => {
		const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
		const { addVisibilityHandler } = useLobbyWebSocket()

		const cb1 = vi.fn()
		const cb2 = vi.fn()
		const off1 = addVisibilityHandler(cb1)
		addVisibilityHandler(cb2)

		// Simulate visibilitychange with document.hidden = false
		Object.defineProperty(document, 'hidden', { value: false, configurable: true })
		document.dispatchEvent(new Event('visibilitychange'))

		expect(cb1).toHaveBeenCalledOnce()
		expect(cb2).toHaveBeenCalledOnce()

		// Unsubscribe cb1 only
		off1()
		document.dispatchEvent(new Event('visibilitychange'))

		expect(cb1).toHaveBeenCalledOnce() // still once
		expect(cb2).toHaveBeenCalledTimes(2)
	})

	it('reconnect after close uses stored playerId without re-passing', async () => {
		vi.useFakeTimers()
		const { useLobbyWebSocket } = await import('@/composables/useLobbyWebSocket')
		const { connect } = useLobbyWebSocket()

		connect('player99')
		const firstWs = lastWs

		firstWs?.triggerClose()
		vi.advanceTimersByTime(600) // past first reconnect delay (500ms)

		expect(lastWs).not.toBe(firstWs) // new WS created
		expect(lastWs?.url).toContain('playerId=player99')

		vi.useRealTimers()
	})
})
```

- [ ] **Step 2: Run to verify tests fail**

```bash
cd tma && pnpm test:unit --reporter=verbose src/__tests__/useLobbyWebSocket.spec.ts
```

Expected: FAIL — `addVisibilityHandler is not a function` and constructor arg errors

---

### Task 5: Refactor `useLobbyWebSocket`

**Files:**
- Modify: `tma/src/composables/useLobbyWebSocket.ts`

- [ ] **Step 1: Replace the composable**

```typescript
import { ref, onUnmounted } from 'vue'
import type { InjectionKey } from 'vue'

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

export function useLobbyWebSocket() {
	const isConnected = ref(false)
	const isPollingFallback = ref(false)

	let ws: WebSocket | null = null
	let currentPlayerId = ''
	let reconnectAttempts = 0
	const maxReconnectAttempts = 5
	const handlers = new Map<string, Set<EventHandler>>()
	const visibilityHandlers = new Set<() => void>()
	let globalVisibilityListener: (() => void) | null = null
	let reconnectTimeout: ReturnType<typeof setTimeout> | null = null

	const emit = (type: string, data: Record<string, unknown> | undefined) => {
		handlers.get(type)?.forEach((h) => h(data))
	}

	const on = (type: LobbyEventType, handler: EventHandler): (() => void) => {
		if (!handlers.has(type)) handlers.set(type, new Set())
		handlers.get(type)!.add(handler)
		return () => handlers.get(type)?.delete(handler)
	}

	const connect = (playerId?: string) => {
		if (playerId) currentPlayerId = playerId
		if (!currentPlayerId) return
		if (ws && ws.readyState === WebSocket.OPEN) return

		const url = `${getWsBase()}/ws/duel/lobby?playerId=${currentPlayerId}`
		ws = new WebSocket(url)

		ws.onopen = () => {
			isConnected.value = true
			isPollingFallback.value = false
			reconnectAttempts = 0
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
			if (reconnectAttempts < maxReconnectAttempts) {
				const delay = Math.min(500 * Math.pow(2, reconnectAttempts), 8000)
				reconnectAttempts++
				reconnectTimeout = setTimeout(() => connect(), delay)
			} else {
				isPollingFallback.value = true
			}
		}

		ws.onerror = () => {
			// onclose fires after onerror; reconnect handled there
		}
	}

	const disconnect = () => {
		if (reconnectTimeout) {
			clearTimeout(reconnectTimeout)
			reconnectTimeout = null
		}
		if (ws) {
			ws.onclose = null // prevent reconnect loop
			ws.close()
			ws = null
		}
		isConnected.value = false
	}

	// Multi-callback visibility handler — safe to call from multiple places.
	// Does NOT call connect() itself — each caller is responsible for reconnect logic.
	const addVisibilityHandler = (cb: () => void): (() => void) => {
		visibilityHandlers.add(cb)
		if (!globalVisibilityListener) {
			globalVisibilityListener = () => {
				if (!document.hidden) {
					visibilityHandlers.forEach((handler) => handler())
				}
			}
			document.addEventListener('visibilitychange', globalVisibilityListener)
		}
		return () => visibilityHandlers.delete(cb)
	}

	onUnmounted(() => {
		disconnect()
		if (globalVisibilityListener) {
			document.removeEventListener('visibilitychange', globalVisibilityListener)
			globalVisibilityListener = null
		}
	})

	return {
		isConnected,
		isPollingFallback,
		connect,
		disconnect,
		on,
		addVisibilityHandler,
	}
}

export type LobbyWsInstance = ReturnType<typeof useLobbyWebSocket>
export const LOBBY_WS_KEY: InjectionKey<LobbyWsInstance> = Symbol('lobbyWs')
```

- [ ] **Step 2: Run the tests**

```bash
cd tma && pnpm test:unit --reporter=verbose src/__tests__/useLobbyWebSocket.spec.ts
```

Expected: all 5 tests PASS

- [ ] **Step 3: Commit**

```bash
git add tma/src/composables/useLobbyWebSocket.ts tma/src/__tests__/useLobbyWebSocket.spec.ts
git commit -m "refactor(duel): useLobbyWebSocket — connect(playerId?), addVisibilityHandler, LOBBY_WS_KEY"
```

---

## Chunk 3: Frontend integration

### Task 6: Create `useGlobalDuelNotifications`

**Files:**
- Create: `tma/src/composables/useGlobalDuelNotifications.ts`
- Create: `tma/src/__tests__/useGlobalDuelNotifications.spec.ts`

- [ ] **Step 1: Write the test**

```typescript
// tma/src/__tests__/useGlobalDuelNotifications.spec.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { defineComponent, provide } from 'vue'
import { mount } from '@vue/test-utils'

// ── Mocks ──────────────────────────────────────────────────────────────────

const mockAdd = vi.fn()
const mockRemove = vi.fn()
vi.mock('@nuxt/ui', () => ({
	useToast: () => ({ add: mockAdd, remove: mockRemove }),
}))

const mockPush = vi.fn()
let mockRouteName = 'home'
vi.mock('vue-router', () => ({
	useRouter: () => ({
		push: mockPush,
		currentRoute: { value: { name: mockRouteName } },
	}),
}))

const mockMutateAsync = vi.fn().mockResolvedValue({})
vi.mock('@/api/generated', () => ({
	usePostDuelChallengeChallengeidRespond: () => ({
		mutateAsync: mockMutateAsync,
	}),
}))

vi.mock('@/composables/useAuth', () => ({
	useAuth: () => ({ currentUser: { value: { id: 'user-me' } } }),
}))

vi.mock('vue-i18n', () => ({
	useI18n: () => ({ t: (key: string, args?: Record<string, unknown>) => {
		if (args?.name) return `${args.name} challenges you`
		return key
	}},
}))

// ── Helpers ────────────────────────────────────────────────────────────────

function buildMockWs() {
	const handlerMap = new Map<string, Set<(data: unknown) => void>>()
	return {
		on: vi.fn((type: string, cb: (data: unknown) => void) => {
			if (!handlerMap.has(type)) handlerMap.set(type, new Set())
			handlerMap.get(type)!.add(cb)
			return () => handlerMap.get(type)?.delete(cb)
		}),
		trigger: (type: string, data: unknown) => {
			handlerMap.get(type)?.forEach((cb) => cb(data))
		},
	}
}

function mountWithComposable(composableFn: () => void) {
	const Wrapper = defineComponent({
		setup() { composableFn() },
		template: '<div/>',
	})
	return mount(Wrapper)
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe('useGlobalDuelNotifications', () => {
	beforeEach(() => {
		mockAdd.mockClear()
		mockRemove.mockClear()
		mockPush.mockClear()
		mockMutateAsync.mockClear()
		mockRouteName = 'home'
	})

	it('shows toast when challenge_received fires on non-lobby route', async () => {
		const { useGlobalDuelNotifications } = await import('@/composables/useGlobalDuelNotifications')
		const mockWs = buildMockWs()

		mountWithComposable(() => useGlobalDuelNotifications(mockWs as any))

		mockWs.trigger('challenge_received', {
			challengeId: 'cid-1',
			challengerUsername: 'Pavel',
			expiresIn: 3600,
		})

		expect(mockAdd).toHaveBeenCalledOnce()
		const call = mockAdd.mock.calls[0][0]
		expect(call.id).toBe('cid-1')
		expect(call.description).toContain('Pavel')
		expect(call.actions).toHaveLength(2)
		expect(call.duration).toBe(60000)
	})

	it('suppresses toast when route is duel-lobby', async () => {
		mockRouteName = 'duel-lobby'
		const { useGlobalDuelNotifications } = await import('@/composables/useGlobalDuelNotifications')
		const mockWs = buildMockWs()

		mountWithComposable(() => useGlobalDuelNotifications(mockWs as any))

		mockWs.trigger('challenge_received', { challengeId: 'cid-2', challengerUsername: 'Ana' })

		expect(mockAdd).not.toHaveBeenCalled()
	})

	it('navigates to duel-play on game_ready when not on duel-lobby', async () => {
		const { useGlobalDuelNotifications } = await import('@/composables/useGlobalDuelNotifications')
		const mockWs = buildMockWs()

		mountWithComposable(() => useGlobalDuelNotifications(mockWs as any))

		mockWs.trigger('game_ready', { gameId: 'game-xyz' })

		expect(mockPush).toHaveBeenCalledWith({ name: 'duel-play', params: { duelId: 'game-xyz' } })
	})

	it('skips game_ready navigation when on duel-lobby', async () => {
		mockRouteName = 'duel-lobby'
		const { useGlobalDuelNotifications } = await import('@/composables/useGlobalDuelNotifications')
		const mockWs = buildMockWs()

		mountWithComposable(() => useGlobalDuelNotifications(mockWs as any))

		mockWs.trigger('game_ready', { gameId: 'game-xyz' })

		expect(mockPush).not.toHaveBeenCalled()
	})
})
```

- [ ] **Step 2: Run to verify failure**

```bash
cd tma && pnpm test:unit --reporter=verbose src/__tests__/useGlobalDuelNotifications.spec.ts
```

Expected: FAIL — module not found

- [ ] **Step 3: Implement `useGlobalDuelNotifications.ts`**

```typescript
// tma/src/composables/useGlobalDuelNotifications.ts
import { onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from '@nuxt/ui'
import { useI18n } from 'vue-i18n'
import { useAuth } from '@/composables/useAuth'
import { usePostDuelChallengeChallengeidRespond } from '@/api/generated'
import type { LobbyWsInstance } from '@/composables/useLobbyWebSocket'

export function useGlobalDuelNotifications(lobbyWs: LobbyWsInstance) {
	const router = useRouter()
	const toast = useToast()
	const { t } = useI18n()
	const { currentUser } = useAuth()
	const respondMutation = usePostDuelChallengeChallengeidRespond()

	const offChallengeReceived = lobbyWs.on('challenge_received', (data) => {
		// Handled natively by DuelLobbyView when on that screen
		if (router.currentRoute.value.name === 'duel-lobby') return

		const challengeId = data?.challengeId as string
		const challengerUsername = (data?.challengerUsername as string) || t('duel.friend')
		let isPending = false

		const respond = async (action: 'accept' | 'decline') => {
			if (isPending) return
			isPending = true
			const playerId = currentUser.value?.id ?? ''
			try {
				await respondMutation.mutateAsync({ challengeId, data: { playerId, action } })
				toast.remove(challengeId)
				if (action === 'accept') {
					router.push({ name: 'duel-lobby' })
				}
			} catch {
				toast.remove(challengeId)
				toast.add({
					id: challengeId + '-err',
					title: t('duel.acceptFailed'),
					color: 'error',
					duration: 5000,
				})
				isPending = false
			}
		}

		toast.add({
			id: challengeId,
			title: t('duel.incomingChallenge'),
			description: t('duel.challengerInvites', { name: challengerUsername }),
			icon: 'i-heroicons-bolt',
			color: 'warning',
			duration: 60000,
			actions: [
				{
					label: t('duel.accept'),
					color: 'primary' as const,
					onClick: () => respond('accept'),
				},
				{
					label: t('duel.decline'),
					color: 'neutral' as const,
					variant: 'ghost' as const,
					onClick: () => respond('decline'),
				},
			],
		})
	})

	const offGameReady = lobbyWs.on('game_ready', (data) => {
		// usePvPDuel handles this when DuelLobbyView is mounted
		if (router.currentRoute.value.name === 'duel-lobby') return

		const gameId = data?.gameId as string | undefined
		if (gameId) router.push({ name: 'duel-play', params: { duelId: gameId } })
	})

	onUnmounted(() => {
		offChallengeReceived()
		offGameReady()
	})
}
```

- [ ] **Step 4: Run tests**

```bash
cd tma && pnpm test:unit --reporter=verbose src/__tests__/useGlobalDuelNotifications.spec.ts
```

Expected: 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add tma/src/composables/useGlobalDuelNotifications.ts tma/src/__tests__/useGlobalDuelNotifications.spec.ts
git commit -m "feat(duel): useGlobalDuelNotifications — toast with Accept/Decline for incoming challenges"
```

---

### Task 7: Update `usePvPDuel` — inject WS and cleanup subscriptions

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts`

- [ ] **Step 1: Replace `useLobbyWebSocket` with `inject`**

At the top of `usePvPDuel.ts`, update the import and the WS initialization:

```typescript
// Remove this import:
import { useLobbyWebSocket } from '@/composables/useLobbyWebSocket'

// Add these imports:
import { inject } from 'vue'
import { LOBBY_WS_KEY, type LobbyWsInstance } from '@/composables/useLobbyWebSocket'
```

Inside `usePvPDuel(playerId)`, replace:
```typescript
// Remove:
const lobbyWs = useLobbyWebSocket(playerId)

// Add:
const lobbyWs = inject(LOBBY_WS_KEY) as LobbyWsInstance
```

- [ ] **Step 2: Collect all WS handler unsubscribes**

Find the "WS event handlers" comment block (~line 221) and replace all bare `lobbyWs.on(...)` calls:

```typescript
// WS event handlers: refetch state and navigate on game_ready
const wsUnsubs: Array<() => void> = []

wsUnsubs.push(
	lobbyWs.on('challenge_received', async () => {
		await refetchStatus()
	}),
)

wsUnsubs.push(
	lobbyWs.on('challenge_accepted', async () => {
		await refetchStatus()
	}),
)

wsUnsubs.push(
	lobbyWs.on('challenge_declined', async () => {
		await refetchStatus()
		await refetchRivals()
	}),
)

wsUnsubs.push(
	lobbyWs.on('challenge_expired', async () => {
		await refetchStatus()
	}),
)

wsUnsubs.push(
	lobbyWs.on('game_ready', (data) => {
		const gameId = data?.gameId as string | undefined
		if (gameId) {
			router.push({ name: 'duel-play', params: { duelId: gameId } })
		} else {
			refetchStatus().then(() => {
				if (hasActiveDuel.value && activeGameId.value) goToActiveDuel()
			})
		}
	}),
)

wsUnsubs.push(
	lobbyWs.on('queue_matched', (data) => {
		const gameId = data?.gameId as string | undefined
		if (gameId) router.push({ name: 'duel-play', params: { duelId: gameId } })
	}),
)
```

- [ ] **Step 3: Update `onUnmounted` to clean up WS subscriptions**

Find the existing `onUnmounted` (~line 251):
```typescript
onUnmounted(() => {
	stopOutgoingPoll()
})
```

Replace with:
```typescript
onUnmounted(() => {
	wsUnsubs.forEach((off) => off())
	stopOutgoingPoll()
})
```

- [ ] **Step 4: Type-check**

```bash
cd tma && pnpm run type-check
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add tma/src/composables/usePvPDuel.ts
git commit -m "refactor(duel): usePvPDuel inject lobbyWs via LOBBY_WS_KEY, clean up on() subscriptions in onUnmounted"
```

---

### Task 8: Update `DuelLobbyView`

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

- [ ] **Step 1: Remove `lobbyWs.connect()`, switch to `addVisibilityHandler`**

In `onMounted` (lines 241–246), change:

```typescript
// Remove:
lobbyWs.connect()
lobbyWs.setupVisibilityHandler(async () => {
	await refetchStatus()
})

// Replace with:
const offVisibility = lobbyWs.addVisibilityHandler(async () => {
	await refetchStatus()
})
```

- [ ] **Step 2: Add cleanup for visibility handler**

The view already has `onUnmounted` (line 220) for `nowInterval`. Add the visibility unsubscribe:

```typescript
onUnmounted(() => {
	if (nowInterval) clearInterval(nowInterval)
	offVisibility()
})
```

Note: `offVisibility` must be declared at the top of `<script setup>` scope (outside `onMounted`), initialized with a no-op then set in `onMounted`:

```typescript
let offVisibility: () => void = () => {}

// inside onMounted:
offVisibility = lobbyWs.addVisibilityHandler(async () => {
	await refetchStatus()
})
```

- [ ] **Step 3: Type-check + lint**

```bash
cd tma && pnpm run type-check && pnpm lint
```

- [ ] **Step 4: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "refactor(duel): DuelLobbyView — remove connect(), use addVisibilityHandler with cleanup"
```

---

### Task 9: Update `App.vue` — provide WS singleton and init notifications

**Files:**
- Modify: `tma/src/App.vue`

- [ ] **Step 1: Add imports**

In `<script setup>`, after existing imports, add:

```typescript
import { provide, watch } from 'vue'
import { useLobbyWebSocket, LOBBY_WS_KEY } from '@/composables/useLobbyWebSocket'
import { useGlobalDuelNotifications } from '@/composables/useGlobalDuelNotifications'
```

Also add `currentUser` to the `useAuth` destructure (line 28):

```typescript
const { isInitialized, getRawInitData, setCurrentUser, consumeStartParam, currentUser } = useAuth()
```

- [ ] **Step 2: Initialize WS and notifications**

After the `useAuth` line, add:

```typescript
// Global lobby WebSocket — single connection for entire app session
const lobbyWs = useLobbyWebSocket()
provide(LOBBY_WS_KEY, lobbyWs)
useGlobalDuelNotifications(lobbyWs)

// Reconnect when TMA returns from background
lobbyWs.addVisibilityHandler(() => {
	if (!lobbyWs.isConnected.value) lobbyWs.connect()
})

// Connect once user is authenticated
watch(
	currentUser,
	(user) => {
		if (user?.id) lobbyWs.connect(user.id)
	},
	{ immediate: true },
)
```

- [ ] **Step 3: Type-check**

```bash
cd tma && pnpm run type-check
```

- [ ] **Step 4: Run all frontend tests**

```bash
cd tma && pnpm test:unit
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
git add tma/src/App.vue
git commit -m "feat(duel): lift lobby WS to App.vue, provide singleton, init global challenge notifications"
```

---

### Task 10: Smoke test

- [ ] Start the dev environment:

```bash
# Terminal 1
cd backend && docker compose -f docker-compose.dev.yml up

# Terminal 2
cloudflared tunnel run quiz-sprint-dev

# Terminal 3
cd tma && pnpm dev
```

Access at `https://dev.quiz-sprint-tma.online`

- [ ] Open app as **User A**, navigate to Rivals list
- [ ] Open app as **User B** on the **home screen** (not duel-lobby)
- [ ] User A sends a direct challenge to User B
- [ ] **Verify:** toast appears on User B's screen with ⚔️, challenger name, Accept + Decline buttons
- [ ] **Verify:** Accept button navigates to duel-lobby and shows "Ожидаем инвайтера" banner
- [ ] **Verify:** if User B is on duel-lobby when challenge arrives, no toast appears (native pending challenges UI shows instead)
- [ ] **Verify:** WS indicator in lobby header is green (connected)

- [ ] Push to remote

```bash
git push
```
