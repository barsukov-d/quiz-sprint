# Matchmaking Feature Flag Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Скрыть карточку matchmaking ("Найти") в DuelLobbyView через feature flag, сохранив всю логику.

**Architecture:** Новый файл `tma/src/features.ts` с центральным реестром флагов. `DuelLobbyView.vue` импортирует `FEATURES` и скрывает блок через `v-if`.

**Tech Stack:** Vue 3, TypeScript

---

### Task 1: Создать `tma/src/features.ts`

**Files:**
- Create: `tma/src/features.ts`

**Step 1: Создать файл**

```ts
// tma/src/features.ts
export const FEATURES = {
  matchmaking: false, // включить когда будем готовы к поиску соперников
}
```

**Step 2: Commit**

```bash
git add tma/src/features.ts
git commit -m "feat: add features.ts with matchmaking flag"
```

---

### Task 2: Скрыть карточку matchmaking в DuelLobbyView.vue

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue` (строки 1–8, 354–383)

**Step 1: Импортировать FEATURES в script setup**

В блоке `<script setup>` добавить импорт после существующих:
```ts
import { FEATURES } from '@/features'
```

**Step 2: Добавить v-if на карточку Find Match**

Найти строку 354 (`<!-- Find Match Button -->`).
Изменить открывающий тег `<UCard`:

```html
<!-- Find Match Button -->
<UCard v-if="FEATURES.matchmaking" class="text-center">
```

Всё содержимое карточки (строки 355–383) остаётся без изменений.

**Step 3: Проверить в браузере**

Запустить dev-сервер:
```bash
cd tma && pnpm dev
```

Открыть страницу `/duel` — карточка "Найти" должна отсутствовать.
Блок "Вызвать друга" должен быть виден.

**Step 4: Убедиться что включение работает**

В `tma/src/features.ts` временно поменять `matchmaking: true`.
Карточка "Найти" появляется. Вернуть `false`.

**Step 5: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): hide matchmaking via feature flag"
```
