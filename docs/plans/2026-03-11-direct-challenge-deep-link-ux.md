# Direct Challenge Deep Link UX — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** When Player B taps "⚔️ Принять вызов" in a Telegram notification, the app navigates to the duel lobby and shows a prominent hero banner with the challenger's name and Accept/Decline buttons.

**Architecture:** Handle `challenge_<uuid>` prefix in `App.vue` → navigate to `duel-lobby?directChallenge=<uuid>`. `DuelLobbyView` reads the query param, finds the challenge in `pendingChallenges` after status loads, and renders a hero banner above all other content.

**Tech Stack:** Vue 3, TypeScript, Vue Router, vue-i18n, Nuxt UI (UCard, UButton, UIcon, UBadge)

---

### Task F1: Handle `challenge_` deep link in App.vue

**Files:**
- Modify: `tma/src/App.vue`

**Step 1: Add `challenge_` case to `handleDeepLink`**

In `App.vue`, find `handleDeepLink` (currently handles `duel_` and `ref_`).
Add a new branch **before** the `duel_` check:

```typescript
// Direct challenge notification: challenge_<uuid>
if (startParam.startsWith('challenge_')) {
  const challengeId = startParam.replace('challenge_', '')
  console.log('⚔️ Direct challenge deep link, navigating to lobby')
  router.push({
    name: 'duel-lobby',
    query: { directChallenge: challengeId },
  })
  return
}
```

**Step 2: Build to verify no TypeScript errors**
```bash
cd tma && pnpm build 2>&1 | tail -20
```
Expected: no errors

**Step 3: Commit**
```bash
git add tma/src/App.vue
git commit -m "feat(pvp-duel): handle challenge_ deep link prefix in App.vue"
```

---

### Task F2: Add i18n keys

**Files:**
- Modify: `tma/src/i18n/locales/ru.ts`
- Modify: `tma/src/i18n/locales/en.ts`

**Step 1: Add two new keys to `ru.ts`**

In `tma/src/i18n/locales/ru.ts`, inside the `duel: {` block, after `friendReady`:
```typescript
challengerInvites: '{name} бросает тебе вызов',
challengeNotFound: 'Вызов истёк или уже принят',
```

**Step 2: Add same keys to `en.ts`**

Find the `duel: {` block in `tma/src/i18n/locales/en.ts` and add after the equivalent last key:
```typescript
challengerInvites: '{name} challenges you to a duel',
challengeNotFound: 'Challenge expired or already accepted',
```

**Step 3: Commit**
```bash
git add tma/src/i18n/locales/ru.ts tma/src/i18n/locales/en.ts
git commit -m "feat(pvp-duel): add i18n keys for direct challenge hero banner"
```

---

### Task F3: Hero banner logic in DuelLobbyView

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Add `directChallengeId` ref and `directChallenge` computed**

In the `<script setup>` section, after the existing `deepLinkChallenge` and `deepLinkError` refs:

```typescript
// Direct challenge deep link state
const directChallengeId = ref<string | null>(null)

const directChallenge = computed(() => {
  if (!directChallengeId.value) return null
  return pendingChallenges.value.find(c => c.id === directChallengeId.value) ?? null
})

const directChallengeNotFound = computed(() =>
  directChallengeId.value !== null && !isLoading.value && directChallenge.value === null
)
```

**Step 2: Add dismiss and action handlers**

After `dismissDeepLinkError`:
```typescript
const dismissDirectChallenge = () => {
  directChallengeId.value = null
  router.replace({ name: 'duel-lobby' })
}

const handleDirectAccept = async () => {
  if (!directChallenge.value?.id) return
  await respondChallenge(directChallenge.value.id, 'accept')
  dismissDirectChallenge()
  await refetchStatus()
  if (hasActiveDuel.value && activeGameId.value) {
    goToActiveDuel()
  }
}

const handleDirectDecline = async () => {
  if (!directChallenge.value?.id) return
  await respondChallenge(directChallenge.value.id, 'decline')
  dismissDirectChallenge()
}
```

**Step 3: Read query param in `onMounted`**

In `onMounted`, after `await refetchStatus()` and before the existing active duel check:

```typescript
// Check for direct challenge deep link (from Telegram notification)
const directChallengeParam = route.query.directChallenge as string | undefined
if (directChallengeParam) {
  directChallengeId.value = directChallengeParam
  window.scrollTo(0, 0)
}
```

**Step 4: Build to verify**
```bash
cd tma && pnpm build 2>&1 | tail -20
```
Expected: no errors

**Step 5: Commit**
```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): add direct challenge hero banner logic"
```

---

### Task F4: Hero banner template

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

**Step 1: Add hero banner to template**

In the `<template>`, after the `<!-- Header -->` block and **before** `<!-- Deep Link Challenge Loading -->`, add:

```html
<!-- Direct Challenge Hero Banner -->
<template v-if="directChallengeId">
  <!-- Loading state -->
  <div v-if="isLoading" class="mb-4 rounded-2xl bg-primary-50 dark:bg-primary-900/30 border border-primary-200 dark:border-primary-700 p-5">
    <div class="animate-pulse space-y-3">
      <div class="h-5 bg-primary-200 dark:bg-primary-700 rounded w-2/3" />
      <div class="h-4 bg-primary-100 dark:bg-primary-800 rounded w-1/2" />
      <div class="grid grid-cols-2 gap-3 mt-4">
        <div class="h-10 bg-primary-200 dark:bg-primary-700 rounded-lg" />
        <div class="h-10 bg-primary-100 dark:bg-primary-800 rounded-lg" />
      </div>
    </div>
  </div>

  <!-- Found state -->
  <div
    v-else-if="directChallenge"
    class="mb-4 rounded-2xl bg-orange-50 dark:bg-orange-900/20 border-2 border-orange-300 dark:border-orange-600 p-5"
  >
    <div class="flex items-center gap-3 mb-1">
      <span class="text-3xl">⚔️</span>
      <div>
        <h2 class="text-lg font-bold text-orange-700 dark:text-orange-300">
          {{ t('duel.incomingChallenge') }}
        </h2>
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('duel.challengerInvites', { name: directChallenge.challengerUsername || t('duel.friend') }) }}
        </p>
      </div>
    </div>
    <div class="grid grid-cols-2 gap-3 mt-4">
      <UButton
        color="primary"
        block
        size="lg"
        @click="handleDirectAccept"
      >
        {{ t('duel.acceptChallenge') }}
      </UButton>
      <UButton
        color="gray"
        variant="soft"
        block
        size="lg"
        @click="handleDirectDecline"
      >
        {{ t('duel.decline') }}
      </UButton>
    </div>
  </div>

  <!-- Not found / expired state -->
  <div
    v-else-if="directChallengeNotFound"
    class="mb-4 rounded-2xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-5"
  >
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-2xl">⏰</span>
        <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
          {{ t('duel.challengeNotFound') }}
        </p>
      </div>
      <UButton size="xs" color="gray" variant="ghost" icon="i-heroicons-x-mark" @click="dismissDirectChallenge" />
    </div>
  </div>
</template>
```

**Step 2: Build to verify**
```bash
cd tma && pnpm build 2>&1 | tail -20
```
Expected: no errors

**Step 3: Verify locally**

Start dev server and open: `http://localhost:5173/duel?directChallenge=<any-uuid>`

Expected:
- If status loads and no matching challenge: expired state (grey banner)
- If matching `pendingChallenge` exists: orange banner with inviter name + buttons

**Step 4: Commit**
```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): hero banner template for direct challenge deep link"
```
