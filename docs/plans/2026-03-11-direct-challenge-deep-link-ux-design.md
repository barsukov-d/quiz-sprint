# Direct Challenge Deep Link UX — Design

**Problem:** When Player B taps "⚔️ Принять вызов" in the Telegram notification, the app opens but shows nothing actionable. The `challenge_<id>` deep link prefix is not handled — App.vue only handles `duel_` (link challenges). Player B sees the home screen with no guidance.

**Goal:** Navigate directly to the duel lobby with a prominent hero banner at the top showing the challenger's name and Accept/Decline buttons.

---

## Approved Design

### Variant A: Hero banner in DuelLobbyView

**Routing change (App.vue)**

Add handler for `challenge_` prefix alongside the existing `duel_` handler:

```
challenge_<uuid>  →  router.push({ name: 'duel-lobby', query: { directChallenge: <uuid> } })
duel_<code>       →  router.push({ name: 'duel-lobby', query: { challenge: <code> } })  // unchanged
```

**Hero banner (DuelLobbyView)**

- Rendered **above** the player rating card, only when `?directChallenge=<id>` is present
- Data source: `pendingChallenges.find(c => c.id === directChallengeId)` — no new API call
- `window.scrollTo(0, 0)` on mount when banner is active

**Banner states:**

| State | Content |
|-------|---------|
| Loading (`isLoading`) | Skeleton + "Загружаем вызов..." |
| Found | Inviter name + "⚔️ Вызов на дуэль!" + Принять / Отказать buttons |
| Not found / expired | "⏰ Вызов истёк или уже принят" + close button |

**Visual structure (found state):**
```
┌─────────────────────────────────────┐
│  ⚔️  Вызов на дуэль!               │  ← accent bg (primary/orange tones)
│  <InviterName> бросает тебе вызов   │
│                                     │
│  [ ✅ Принять ]  [ ❌ Отказать ]    │
└─────────────────────────────────────┘
```

**After action:**
- Remove `directChallenge` from query (`router.replace({ name: 'duel-lobby' })`)
- Accept: `respondChallenge(id, 'accept')` → `refetchStatus()` → if active game → `goToActiveDuel()`
- Decline: `respondChallenge(id, 'decline')` → banner disappears, user stays in lobby

**Edge cases:**

| Situation | Behavior |
|-----------|----------|
| Already in game | Existing `onMounted` redirect fires before banner is shown |
| Challenge not in `pendingChallenges` | Show "expired" banner state |
| Status still loading | Show skeleton in banner |
| User declines | Banner removed, lobby shown normally |
| User accepts, game created | `goToActiveDuel()` |

---

## Scope

**Frontend only** — 2 files:
- `tma/src/App.vue` — add `challenge_` case to `handleDeepLink`
- `tma/src/views/Duel/DuelLobbyView.vue` — add hero banner logic + template
