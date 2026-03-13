# Global Challenge Notifications — Design Spec

**Date:** 2026-03-13
**Status:** Approved

## Problem

When a player sends a direct challenge from the Rivals list, the opponent only receives notification if they are currently on the `duel-lobby` screen. If the opponent is on the home screen or playing another game mode, no WebSocket connection exists and the challenge goes unnoticed until they manually navigate to the lobby.

## Goal

Show a real-time toast notification (Accept / Decline buttons inline) to any logged-in player who receives a direct challenge, regardless of which screen they are on.

---

## Architecture

### Current State

```
App.vue
└── DuelLobbyView → usePvPDuel → useLobbyWebSocket(playerId)
                                  ↑ WS only exists on duel-lobby route
```

### Target State

```
App.vue ──→ useLobbyWebSocket          (global singleton, provide)
        └─→ useGlobalDuelNotifications (challenge_received + game_ready → toast / navigate)

DuelLobbyView → usePvPDuel → inject(LOBBY_WS_KEY)   ← same instance, no duplicate WS
```

The WebSocket lifecycle moves from `DuelLobbyView` to `App.vue`. A single connection covers the entire app session.

---

## Changes

### 1. `useLobbyWebSocket` — refactor

**Change A — `playerId` moves from constructor to `connect(playerId)`:**

In `App.vue`, `playerId` is unavailable at composable init time (auth is async). Deferring to `connect()` allows `App.vue` to call it once auth resolves.

**Change B — `playerId` stored as closure variable for reconnect:**

The internal `ws.onclose` reconnect handler calls `connect()` recursively. After the signature change, the closure must capture `playerId` in a local variable so the recursive call does not require re-passing it.

```ts
// Inside useLobbyWebSocket
let currentPlayerId = ''

const connect = (playerId?: string) => {
  if (playerId) currentPlayerId = playerId
  if (!currentPlayerId) return
  // ... ws setup uses currentPlayerId
  ws.onclose = () => {
    // ...
    reconnectTimeout = setTimeout(() => connect(), delay)  // no arg needed — uses closure
  }
}
```

**Change C — `setupVisibilityHandler` supports multiple callbacks:**

Currently `setupVisibilityHandler` stores a single handler reference, so the second call overwrites the first. App.vue registers a reconnect callback; DuelLobbyView needs a separate `refetchStatus` callback. Fix: accept an array of callbacks internally, or change `setupVisibilityHandler` to `addVisibilityHandler(cb)` that pushes to a set.

```ts
// New API
const addVisibilityHandler = (cb: () => void): (() => void) => {
  visibilityHandlers.add(cb)
  return () => visibilityHandlers.delete(cb)  // returns unsubscribe
}
```

**Change D — `on()` unsubscribe is returned (already done) — callers must use it:**

No change to `useLobbyWebSocket` itself, but callers (see §4) must store and call the returned unsubscribe functions.

**Before / After API:**
```ts
// Before
const lobbyWs = useLobbyWebSocket(playerId)
lobbyWs.connect()
lobbyWs.setupVisibilityHandler(onVisible)

// After
const lobbyWs = useLobbyWebSocket()
lobbyWs.connect(playerId)                        // playerId stored in closure
const off = lobbyWs.addVisibilityHandler(cb)     // returns unsubscribe
```

---

### 2. `App.vue` — global WS + notifications

After `currentUser` resolves:
1. `provide(LOBBY_WS_KEY, lobbyWs)` — typed injection key, available to all descendants
2. Watch `currentUser` → call `lobbyWs.connect(user.id)` once id is known
3. Call `useGlobalDuelNotifications(lobbyWs)` — registers handlers for `challenge_received` and `game_ready`
4. Register reconnect via `lobbyWs.addVisibilityHandler(() => { if (!lobbyWs.isConnected.value) lobbyWs.connect() })`

```ts
// App.vue
import { provide, watch } from 'vue'
import { LOBBY_WS_KEY } from '@/composables/useLobbyWebSocket'

const lobbyWs = useLobbyWebSocket()
provide(LOBBY_WS_KEY, lobbyWs)
useGlobalDuelNotifications(lobbyWs)

lobbyWs.addVisibilityHandler(() => {
  if (!lobbyWs.isConnected.value) lobbyWs.connect()
})

watch(currentUser, (user) => {
  if (user?.id) lobbyWs.connect(user.id)
}, { immediate: true })
```

---

### 3. `useGlobalDuelNotifications` — new composable

**File:** `tma/src/composables/useGlobalDuelNotifications.ts`

**Responsibilities:**
- Subscribe to `challenge_received` → show toast with Accept / Decline
- Subscribe to `game_ready` → navigate to `duel-play` (handles the case where user accepted from toast but is not on `duel-lobby`)
- Suppress `challenge_received` toast when current route is `duel-lobby` (native UI handles it there)
- Store all unsubscribe functions and call them in `onUnmounted`

**Toast structure:**
```
┌─────────────────────────────────────┐
│ ⚔️  Вызов на дуэль                  │
│  <challengerUsername> вызывает вас  │
│  [Принять]  [Отклонить]             │
└─────────────────────────────────────┘
```

**Accept flow:**
1. Disable buttons (in-flight state)
2. `POST /duel/challenge/{challengeId}/respond { action: "accept" }`
3. Dismiss toast
4. `router.push({ name: 'duel-lobby' })`
5. When game starts, `game_ready` WS event fires → `useGlobalDuelNotifications` global handler navigates to `duel-play`

**Decline flow:**
1. Disable buttons (in-flight state)
2. `POST /duel/challenge/{challengeId}/respond { action: "decline" }`
3. Dismiss toast

**Error handling:** On API error — toast switches to `color: 'error'`, error message shown, buttons re-enable.

**Timeout:** 60 seconds. The challenge itself is valid for much longer (`DirectChallengeExpirySeconds = 3600`), but 60s is a reasonable notification window. A dismissed toast does not invalidate the challenge — it remains visible in the lobby's Pending Challenges list.

**Handler cleanup:**
```ts
export function useGlobalDuelNotifications(lobbyWs: LobbyWsInstance) {
  const offChallengeReceived = lobbyWs.on('challenge_received', handler)
  const offGameReady = lobbyWs.on('game_ready', gameReadyHandler)

  onUnmounted(() => {
    offChallengeReceived()
    offGameReady()
  })
}
```

---

### 4. `usePvPDuel` — inject instead of create

**Change:** Replace `useLobbyWebSocket(playerId)` with `inject(LOBBY_WS_KEY)`.

The injected instance is already connected — no `connect()` call needed in `usePvPDuel`.

DuelLobbyView's `onMounted` currently calls `lobbyWs.connect()` and `lobbyWs.setupVisibilityHandler(refetchStatus)`. After the migration:
- `lobbyWs.connect()` call is removed (App.vue already connected)
- `lobbyWs.setupVisibilityHandler` becomes `lobbyWs.addVisibilityHandler(refetchStatus)` (additive, not overwriting)

**Handler cleanup — critical:** `usePvPDuel` registers multiple `lobbyWs.on(...)` handlers. The returned unsubscribe functions must be stored and called in `onUnmounted` to prevent duplicate handler accumulation when the user navigates away from and back to `duel-lobby`:

```ts
const unsubs: Array<() => void> = []
unsubs.push(lobbyWs.on('challenge_received', async () => { await refetchStatus() }))
unsubs.push(lobbyWs.on('challenge_accepted', async () => { await refetchStatus() }))
// ... etc

onUnmounted(() => {
  unsubs.forEach(off => off())
  stopOutgoingPoll()
})
```

**Dual handler on `challenge_received` when on `duel-lobby`:**
Both `useGlobalDuelNotifications` (challenge_received → suppressed, route === duel-lobby) and `usePvPDuel` (challenge_received → refetchStatus) subscribe when the user is on `duel-lobby`. This is correct behavior: global handler skips toast, local handler refreshes state. No conflict.

---

### 5. Typed injection key

Use a typed `InjectionKey` to avoid `inject()` returning `T | undefined` without assertions:

```ts
// useLobbyWebSocket.ts
import type { InjectionKey } from 'vue'
export type LobbyWsInstance = ReturnType<typeof useLobbyWebSocket>
export const LOBBY_WS_KEY: InjectionKey<LobbyWsInstance> = Symbol('lobbyWs')
```

---

## Backend Change

### Concrete mechanism for `challengerUsername`

`ChallengeCreatedEvent` is constructed inside the domain aggregate (`duel_challenge.go`) via `NewChallengeCreatedEvent` — which only knows about IDs. The domain factory must not take a username.

**Solution: post-creation enrichment via `WithChallengerUsername`.**

**Step 1 — Add optional field + enrichment method to domain event** (`events.go`):
```go
type ChallengeCreatedEvent struct {
    challengeID        ChallengeID
    challengerID       UserID
    challengedID       *UserID
    challengeType      ChallengeType
    expiresAt          int64
    occurredAt         int64
    challengerUsername string  // notification hint; empty = unknown
}

// WithChallengerUsername returns a copy with username set (called by app layer)
func (e ChallengeCreatedEvent) WithChallengerUsername(name string) ChallengeCreatedEvent {
    e.challengerUsername = name
    return e
}

func (e ChallengeCreatedEvent) ChallengerUsername() string { return e.challengerUsername }
```

The domain aggregate creates the event without username (unchanged). The field is purely a notification hint — no domain logic depends on it.

**Step 2 — Application layer enriches before publishing** (`use_cases.go`):

`SendChallengeUseCase.Execute` already has the challenger's `User` object. After saving the challenge, it iterates `challenge.Events()` and enriches `ChallengeCreatedEvent` before calling `eventBus.Publish`:

```go
// SendChallengeUseCase.Execute (after challenge.Save)
challengerUser, _ := uc.userRepo.FindByID(ctx, challengerID)
for _, evt := range challenge.Events() {
    if e, ok := evt.(quick_duel.ChallengeCreatedEvent); ok {
        evt = e.WithChallengerUsername(challengerUser.Username())
    }
    uc.eventBus.Publish(evt)
}
```

**Step 3 — EventBus reads the field** (`lobby_event_bus.go`):

```go
case domainDuel.ChallengeCreatedEvent:
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

No domain factory changes. Domain aggregate stays ID-only. Application layer is the enrichment point.

---

## WS Event Format

```json
{
  "type": "challenge_received",
  "data": {
    "challengeId": "uuid",
    "expiresIn": 3600,
    "challengerUsername": "Pavel"
  }
}
```

---

## Edge Cases

| Scenario | Behavior |
|---|---|
| Invitee on `duel-lobby` | Global handler suppresses toast (route check); `usePvPDuel` handler calls `refetchStatus()` |
| Invitee accepts from toast, not on `duel-lobby` | Navigates to `duel-lobby`; `game_ready` WS event handled by global `useGlobalDuelNotifications` handler → navigates to `duel-play` |
| WS disconnected (offline) | Toast does not appear; pending challenge visible on next `refetchStatus()` when lobby opens |
| Accept API returns error | Toast color → error, message shown, buttons re-enable |
| Two simultaneous challenges | Two independent toasts, each with own in-flight state |
| Challenge expires while toast open | Toast auto-dismisses after 60s; challenge remains in server state until server expires it |
| User navigates away/back to `duel-lobby` | `usePvPDuel.onUnmounted` unsubscribes handlers; remount re-subscribes once — no duplicates |
| User not yet authenticated | `lobbyWs.connect()` not called until `currentUser` resolves |
| `addVisibilityHandler` called multiple times | Each call pushes to a set; all callbacks fire on visibility change; each returns its own unsubscribe |

---

## Files Changed

| File | Change |
|---|---|
| `tma/src/composables/useLobbyWebSocket.ts` | `playerId` → `connect(playerId?)` closure; `setupVisibilityHandler` → `addVisibilityHandler` (multi-callback) |
| `tma/src/composables/usePvPDuel.ts` | `inject(LOBBY_WS_KEY)`; store + call `on()` unsubscribes in `onUnmounted` |
| `tma/src/composables/useGlobalDuelNotifications.ts` | **New** — toast + game_ready handler |
| `tma/src/App.vue` | Init WS, provide, call `useGlobalDuelNotifications`, watch user |
| `tma/src/views/Duel/DuelLobbyView.vue` | Remove `lobbyWs.connect()`; change `setupVisibilityHandler` → `addVisibilityHandler` |
| `tma/src/__tests__/useLobbyWebSocket.spec.ts` | Update: no-arg constructor; `setupVisibilityHandler` → `addVisibilityHandler` |
| `backend/internal/domain/quick_duel/events.go` | Add `challengerUsername` field, `WithChallengerUsername()`, `ChallengerUsername()` to `ChallengeCreatedEvent` |
| `backend/internal/infrastructure/messaging/lobby_event_bus.go` | Read `e.ChallengerUsername()` in `challenge_received` payload |
| `backend/internal/application/quick_duel/use_cases.go` | Enrich `ChallengeCreatedEvent` with username before `eventBus.Publish` |

---

## Out of Scope

- Push notifications when app is fully closed (requires Telegram Bot API)
- Toast for `queue_matched` (matchmaking)
- Notification history / inbox
