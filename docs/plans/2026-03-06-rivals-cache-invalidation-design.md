# Design: Rivals cache invalidation on challenge state change

Date: 2026-03-06

## Problem

`hasPendingChallenge` остаётся `true` после decline/expire/accept challenge, потому что rivals-кэш (staleTime: 30с) не инвалидируется при изменении состояния challenge.

## Edge Cases

| Сценарий | DB статус | hasPendingChallenge | До фикса | После фикса |
|----------|-----------|---------------------|----------|-------------|
| Вызов отправлен | `pending` | `true` | ✅ | ✅ |
| Противник отклонил | `declined` | `false` | ❌ кэш 30с | ✅ ≤5с |
| Вызов истёк | `expired` | `false` | ❌ кэш 30с | ✅ ≤5с |
| Игра началась | `accepted` | `false` | ❌ кэш 30с | ✅ ≤5с |
| Ждёт старта (link) | `accepted_waiting_inviter` | `true` | ✅ | ✅ |

## Solution

**Только frontend — `tma/src/composables/usePvPDuel.ts`**

StaleTime остаётся 30с (оптимизация). Инвалидация через два механизма:

### 1. Вызов `refetchRivals()` в `sendChallenge()`

После `refetchStatus()` добавить `await refetchRivals()` — мгновенная реакция при отправке.

### 2. `watch(outgoingChallenges)` → `refetchRivals()`

Существующий watch дополнить вызовом `refetchRivals()` при изменении (не на initial trigger):

```typescript
watch(
  outgoingChallenges,
  (challenges, prevChallenges) => {
    if (challenges.length > 0) startOutgoingPoll()
    else stopOutgoingPoll()
    // Инвалидировать rivals при любом изменении outgoing challenges
    if (prevChallenges !== undefined) refetchRivals()
  },
  { immediate: true },
)
```

## Data Flow

```
Challenge declined by opponent
  → outgoingPoll (5s interval) → refetchStatus()
  → outgoingChallenges computed updated (challenge gone)
  → watch fires (prevChallenges !== undefined)
  → refetchRivals() called
  → rivals API returns hasPendingChallenge: false
  → button shows "Вызов" ✅

Challenge sent by player
  → sendChallenge()
  → refetchStatus() + refetchRivals() (immediate)
  → rivals API returns hasPendingChallenge: true
  → button shows "Вызов отправлен" ✅ (instantly, not waiting for watch)
```

## Files to Change

- `tma/src/composables/usePvPDuel.ts` — только 2 места
