# Rivals Cache Invalidation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Сбрасывать rivals-кэш при изменении состояния challenge, чтобы кнопка "Вызов отправлен" корректно переходила обратно в "Вызов" после decline/expire/accept.

**Architecture:** Только фронтенд. Два изменения в `usePvPDuel.ts`: вызов `refetchRivals()` в `sendChallenge()` (мгновенное обновление) и в `watch(outgoingChallenges)` (автоматическая инвалидация при любом изменении challenge).

**Tech Stack:** Vue 3, Vue Query (TanStack Query), TypeScript

---

### Task 1: Добавить `refetchRivals()` в `sendChallenge`

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts:313-329`

**Step 1: Убедиться что тест на type-check чистый до изменений**

```bash
cd /Users/barsukov/projects/quiz-sprint/tma && pnpm run type-check
```
Expected: no errors

**Step 2: Добавить `await refetchRivals()` после `await refetchStatus()`**

Текущий код `sendChallenge` (строки 313-329):
```typescript
const sendChallenge = async (friendId: string) => {
    try {
        console.log('[usePvPDuel] Sending challenge to:', friendId)

        const response = await sendChallengeMutation.mutateAsync({
            data: { playerId, friendId },
        })

        console.log('[usePvPDuel] Challenge sent:', response.data)
        await refetchStatus()

        return response.data
    } catch (error) {
        console.error('[usePvPDuel] Failed to send challenge:', error)
        throw error
    }
}
```

Заменить на:
```typescript
const sendChallenge = async (friendId: string) => {
    try {
        console.log('[usePvPDuel] Sending challenge to:', friendId)

        const response = await sendChallengeMutation.mutateAsync({
            data: { playerId, friendId },
        })

        console.log('[usePvPDuel] Challenge sent:', response.data)
        await refetchStatus()
        await refetchRivals()

        return response.data
    } catch (error) {
        console.error('[usePvPDuel] Failed to send challenge:', error)
        throw error
    }
}
```

**Step 3: Type-check**

```bash
cd /Users/barsukov/projects/quiz-sprint/tma && pnpm run type-check
```
Expected: no errors

**Step 4: Commit**

```bash
git add tma/src/composables/usePvPDuel.ts
git commit -m "fix(pvp-duel): refetch rivals immediately after sending challenge"
```

---

### Task 2: Инвалидировать rivals при изменении `outgoingChallenges`

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts:197-208`

**Step 1: Обновить `watch(outgoingChallenges)`**

Текущий код (строки 197-208):
```typescript
// Start/stop polling when outgoing challenges change
watch(
    outgoingChallenges,
    (challenges) => {
        if (challenges.length > 0) {
            startOutgoingPoll()
        } else {
            stopOutgoingPoll()
        }
    },
    { immediate: true },
)
```

Заменить на:
```typescript
// Start/stop polling when outgoing challenges change
// Also invalidate rivals cache when challenge state changes (decline/expire/accept)
watch(
    outgoingChallenges,
    (challenges, prevChallenges) => {
        if (challenges.length > 0) {
            startOutgoingPoll()
        } else {
            stopOutgoingPoll()
        }
        if (prevChallenges !== undefined) {
            refetchRivals()
        }
    },
    { immediate: true },
)
```

Ключевое: `prevChallenges !== undefined` пропускает initial trigger (immediate: true), вызывая `refetchRivals()` только при реальных изменениях.

**Step 2: Type-check и lint**

```bash
cd /Users/barsukov/projects/quiz-sprint/tma && pnpm run type-check && pnpm lint
```
Expected: no errors

**Step 3: Commit**

```bash
git add tma/src/composables/usePvPDuel.ts
git commit -m "fix(pvp-duel): invalidate rivals cache when outgoing challenge state changes"
```

---

### Task 3: Финальная проверка

**Step 1: Запустить unit-тесты**

```bash
cd /Users/barsukov/projects/quiz-sprint/tma && pnpm test:unit
```
Expected: все тесты кроме pre-existing App.spec.ts PASS

**Step 2: Сборка**

```bash
cd /Users/barsukov/projects/quiz-sprint/tma && pnpm build
```
Expected: build successful

**Step 3: Push**

```bash
git push
```

---

### Ручное тестирование (после push)

1. Открыть TMA с аккаунта A — в списке соперников нажать "Вызов" для аккаунта B
2. Убедиться что кнопка сразу переходит в "Вызов отправлен" (Task 1)
3. С аккаунта B отклонить вызов
4. С аккаунта A подождать ≤5 секунд — кнопка должна вернуться к "Вызов" (Task 2)
5. Повторить с истекающим вызовом (подождать 60 секунд) — кнопка возвращается к "Вызов"
