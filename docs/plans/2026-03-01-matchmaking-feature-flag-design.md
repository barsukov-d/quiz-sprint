# Matchmaking Feature Flag

**Date:** 2026-03-01
**Branch:** pvp-duel
**Status:** Approved

## Problem

На начальном этапе поиск случайных соперников (matchmaking queue) не актуален.
Приоритет — приглашение друзей через Telegram-ссылку.

## Solution

Feature flag `FEATURES.matchmaking` в `tma/src/features.ts`.
При `false` — карточка "Найти" скрыта через `v-if`, логика в composable не удаляется.

## Changes

| File | Change |
|------|--------|
| `tma/src/features.ts` | Новый файл с флагом `matchmaking: false` |
| `tma/src/views/Duel/DuelLobbyView.vue` | `v-if="FEATURES.matchmaking"` на карточку Find Match |

## Not Changed

- Backend API (queue endpoints остаются)
- `usePvPDuel` composable (joinQueue, leaveQueue, isSearching остаются)
- Leaderboard, History, Pending Challenges — без изменений

## How to Enable

В `tma/src/features.ts`:
```ts
matchmaking: true,
```
