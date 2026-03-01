# PvP Duel Game Start Flow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Добавить `outgoingChallenges` в статус дуэли + UI-карточку ожидания в лобби + обновить документацию invite-флоу

**Architecture:**
- Backend: `GetDuelStatusUseCase` вызывает `FindPendingByChallenger` (репо уже реализован), добавляет поле в DTO
- Frontend: `usePvPDuel` читает новое поле + polling каждые 5с пока есть ожидающие вызовы, `DuelLobbyView` показывает карточку
- Docs: `02_gameplay.md` и `05_api.md` обновляются новыми разделами

**Tech Stack:** Go/Fiber backend, Vue 3 + TypeScript frontend, существующий Postgres ChallengeRepository

---

## Task 1: Backend — добавить `outgoingChallenges` в DTO

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go:145-154`
- Modify: `backend/internal/application/quick_duel/use_cases.go:83-93`
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go:1301-1312`

### Step 1: Написать failing тест

Добавить в `backend/internal/application/quick_duel/use_cases_test.go` (после строки с `TestGetDuelStatus`):

```go
func TestGetDuelStatus_IncludesOutgoingChallenges(t *testing.T) {
    f := newDuelFixture(t)
    challengerID := f.user1.id

    // Create a link challenge sent by user1
    challenge, err := quick_duel.NewDuelChallenge(
        challengerID,
        nil,
        "link",
        "duel_testcode123",
        time.Now().UTC().Unix()+86400,
    )
    require.NoError(t, err)
    f.challengeRepo.challenges[challenge.ID()] = challenge

    uc := f.newGetDuelStatusUC()
    output, err := uc.Execute(GetDuelStatusInput{PlayerID: challengerID.String()})
    require.NoError(t, err)

    assert.Len(t, output.OutgoingChallenges, 1)
    assert.Equal(t, "link", output.OutgoingChallenges[0].Type)
    assert.Equal(t, "pending", output.OutgoingChallenges[0].Status)
}
```

### Step 2: Запустить тест — убедиться что он падает

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetDuelStatus_IncludesOutgoingChallenges -v
```

Expected: FAIL — `output.OutgoingChallenges` поле не существует

### Step 3: Добавить поле в DTO

В `backend/internal/application/quick_duel/dto.go`, структура `GetDuelStatusOutput`:

```go
// Было:
type GetDuelStatusOutput struct {
    HasActiveDuel     bool              `json:"hasActiveDuel"`
    ActiveGameID      *string           `json:"activeGameId,omitempty"`
    Player            PlayerRatingDTO   `json:"player"`
    Tickets           int               `json:"tickets"`
    FriendsOnline     []FriendDTO       `json:"friendsOnline"`
    PendingChallenges []ChallengeDTO    `json:"pendingChallenges"`
    SeasonID          string            `json:"seasonId"`
    SeasonEndsAt      int64             `json:"seasonEndsAt"`
}

// Стало (добавить одну строку):
type GetDuelStatusOutput struct {
    HasActiveDuel      bool           `json:"hasActiveDuel"`
    ActiveGameID       *string        `json:"activeGameId,omitempty"`
    Player             PlayerRatingDTO `json:"player"`
    Tickets            int            `json:"tickets"`
    FriendsOnline      []FriendDTO    `json:"friendsOnline"`
    PendingChallenges  []ChallengeDTO `json:"pendingChallenges"`
    OutgoingChallenges []ChallengeDTO `json:"outgoingChallenges"`
    SeasonID           string         `json:"seasonId"`
    SeasonEndsAt       int64          `json:"seasonEndsAt"`
}
```

### Step 4: Заполнить поле в use case

В `backend/internal/application/quick_duel/use_cases.go`, метод `GetDuelStatusUseCase.Execute`, после блока `pendingChallenges` (~строка 83):

```go
// Было:
    return GetDuelStatusOutput{
        HasActiveDuel:     activeGameID != nil,
        ActiveGameID:      activeGameID,
        Player:            ToPlayerRatingDTO(rating),
        Tickets:           10, // TODO: get from user wallet
        FriendsOnline:     []FriendDTO{},
        PendingChallenges: challengeDTOs,
        SeasonID:          seasonID,
        SeasonEndsAt:      seasonEndsAt,
    }, nil

// Стало (добавить блок outgoing + поле):
    outgoingChallenges, err := uc.challengeRepo.FindPendingByChallenger(playerID)
    if err != nil {
        outgoingChallenges = []*quick_duel.DuelChallenge{}
    }

    outgoingDTOs := make([]ChallengeDTO, 0, len(outgoingChallenges))
    for _, c := range outgoingChallenges {
        outgoingDTOs = append(outgoingDTOs, ToChallengeDTO(c, now))
    }

    return GetDuelStatusOutput{
        HasActiveDuel:      activeGameID != nil,
        ActiveGameID:       activeGameID,
        Player:             ToPlayerRatingDTO(rating),
        Tickets:            10, // TODO: get from user wallet
        FriendsOnline:      []FriendDTO{},
        PendingChallenges:  challengeDTOs,
        OutgoingChallenges: outgoingDTOs,
        SeasonID:           seasonID,
        SeasonEndsAt:       seasonEndsAt,
    }, nil
```

### Step 5: Обновить Swagger-модель

В `backend/internal/infrastructure/http/handlers/swagger_models.go`, структура `GetDuelStatusResponse`:

```go
// Добавить строку OutgoingChallenges:
type GetDuelStatusResponse struct {
    Data struct {
        HasActiveDuel      bool                 `json:"hasActiveDuel"`
        ActiveGameID       *string              `json:"activeGameId,omitempty"`
        Player             DuelPlayerRatingDTO  `json:"player"`
        Tickets            int                  `json:"tickets"`
        FriendsOnline      []DuelFriendDTO      `json:"friendsOnline"`
        PendingChallenges  []DuelChallengeDTO   `json:"pendingChallenges"`
        OutgoingChallenges []DuelChallengeDTO   `json:"outgoingChallenges"`
        SeasonID           string               `json:"seasonId"`
        SeasonEndsAt       int64                `json:"seasonEndsAt"`
    } `json:"data"`
}
```

### Step 6: Запустить тест — убедиться что проходит

```bash
cd backend && go test ./internal/application/quick_duel/... -run TestGetDuelStatus_IncludesOutgoingChallenges -v
```

Expected: PASS

### Step 7: Запустить все тесты

```bash
cd backend && go test ./...
```

Expected: все PASS, никаких новых ошибок

### Step 8: Коммит

```bash
git add backend/internal/application/quick_duel/dto.go \
        backend/internal/application/quick_duel/use_cases.go \
        backend/internal/application/quick_duel/use_cases_test.go \
        backend/internal/infrastructure/http/handlers/swagger_models.go
git commit -m "feat(pvp-duel): add outgoingChallenges to GET /duel/status response"
```

---

## Task 2: Регенерировать TypeScript-типы

После изменения бэкенда нужно обновить сгенерированные типы, чтобы фронт видел `outgoingChallenges`.

**Files:**
- Generate: `tma/src/api/generated/` (авто)

### Step 1: Поднять бэкенд и сгенерировать Swagger + TypeScript

```bash
# Терминал 1 — бэкенд
cd backend && docker compose -f docker-compose.dev.yml up

# Терминал 2 — генерация
cd tma && pnpm run generate:all
```

Expected: обновлённые файлы в `tma/src/api/generated/types/`, включая `outgoingChallenges` в типе ответа `/duel/status`

### Step 2: Проверить что тип появился

```bash
grep -r "outgoingChallenges" tma/src/api/generated/
```

Expected: найдено в одном или более файлах типов

### Step 3: Коммит

```bash
git add tma/src/api/generated/
git commit -m "chore: regenerate API types with outgoingChallenges field"
```

---

## Task 3: Frontend — `usePvPDuel` — добавить `outgoingChallenges` + polling

**Files:**
- Modify: `tma/src/composables/usePvPDuel.ts:94-96` (computed) и конец файла (return)

### Step 1: Добавить computed для `outgoingChallenges`

В `tma/src/composables/usePvPDuel.ts`, после строки `const pendingChallenges = ...` (≈строка 95):

```typescript
// Было:
const pendingChallenges = computed(() => statusData.value?.data?.pendingChallenges ?? [])

// Стало (добавить следующую строку):
const pendingChallenges = computed(() => statusData.value?.data?.pendingChallenges ?? [])
const outgoingChallenges = computed(() => statusData.value?.data?.outgoingChallenges ?? [])
```

### Step 2: Добавить polling при наличии ожидающих вызовов

В `tma/src/composables/usePvPDuel.ts`, найти блок `// Local UI State` и добавить ниже:

```typescript
// Polling interval when waiting for a friend to accept a challenge link
let pollInterval: ReturnType<typeof setInterval> | null = null

const startOutgoingPoll = () => {
    if (pollInterval) return
    pollInterval = setInterval(async () => {
        if (outgoingChallenges.value.length > 0) {
            await refetchStatus()
            // If an outgoing challenge was accepted, hasActiveDuel will be true
            if (hasActiveDuel.value && activeGameId.value) {
                stopOutgoingPoll()
                goToActiveDuel()
            }
        } else {
            stopOutgoingPoll()
        }
    }, 5000)
}

const stopOutgoingPoll = () => {
    if (pollInterval) {
        clearInterval(pollInterval)
        pollInterval = null
    }
}
```

### Step 3: Запускать polling когда есть ожидающие вызовы

В том же файле, добавить `watch` после объявления `outgoingChallenges`:

```typescript
import { computed, ref, watch } from 'vue'

// ... в теле функции usePvPDuel, после outgoingChallenges:
watch(outgoingChallenges, (challenges) => {
    if (challenges.length > 0) {
        startOutgoingPoll()
    } else {
        stopOutgoingPoll()
    }
}, { immediate: true })
```

### Step 4: Добавить `stopOutgoingPoll` в cleanup и экспортировать `outgoingChallenges`

В `tma/src/composables/usePvPDuel.ts`, найти раздел `return {`:

```typescript
// Добавить в return объект:
return {
    // ... существующие поля ...
    outgoingChallenges,
    startOutgoingPoll,
    stopOutgoingPoll,
    // ... остальные поля ...
}
```

Также убедиться что `stopOutgoingPoll()` вызывается при размонтировании:

```typescript
import { computed, ref, watch, onUnmounted } from 'vue'

// В теле usePvPDuel:
onUnmounted(() => {
    stopOutgoingPoll()
})
```

### Step 5: Проверить типы

```bash
cd tma && pnpm run type-check
```

Expected: no errors

### Step 6: Коммит

```bash
git add tma/src/composables/usePvPDuel.ts
git commit -m "feat(pvp-duel): expose outgoingChallenges with 5s polling in usePvPDuel"
```

---

## Task 4: Frontend — DuelLobbyView — карточка "Ожидание ответа"

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue`

### Step 1: Добавить `outgoingChallenges` в деструктуризацию composable

В `tma/src/views/Duel/DuelLobbyView.vue` (≈строка 16), в деструктуризации `usePvPDuel`:

```typescript
const {
    // ... существующие поля ...
    outgoingChallenges,  // ← добавить
    // ...
} = usePvPDuel(playerId.value)
```

### Step 2: Добавить computed для форматирования времени истечения

В секции `// Computed` DuelLobbyView:

```typescript
const outgoingChallengeExpiry = computed(() => {
    const challenge = outgoingChallenges.value[0]
    if (!challenge) return ''
    const secondsLeft = (challenge.expiresAt ?? 0) - Math.floor(Date.now() / 1000)
    if (secondsLeft <= 0) return 'Истекла'
    const hours = Math.floor(secondsLeft / 3600)
    const minutes = Math.floor((secondsLeft % 3600) / 60)
    if (hours > 0) return `${hours}ч ${minutes}мин`
    return `${minutes} мин`
})
```

### Step 3: Добавить UI-карточку в template

В `<template>` файла `DuelLobbyView.vue`, найти блок `<!-- Pending Challenges -->` и добавить **перед ним** блок для исходящих вызовов:

```html
<!-- Outgoing Challenge (waiting for friend to accept link) -->
<div v-if="outgoingChallenges.length > 0" class="mb-4">
    <UCard class="border-primary-200 dark:border-primary-800">
        <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-paper-airplane" class="size-5 text-primary" />
                <div>
                    <p class="font-medium text-sm">{{ t('duel.waitingForFriend') }}</p>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                        {{ t('duel.linkExpiresIn', { time: outgoingChallengeExpiry }) }}
                    </p>
                </div>
            </div>
            <div class="flex items-center gap-2">
                <div class="w-2 h-2 bg-primary rounded-full animate-pulse" />
                <span class="text-xs text-gray-500">{{ t('duel.waiting') }}</span>
            </div>
        </div>
    </UCard>
</div>
```

### Step 4: Добавить i18n-ключи

Найти файл локализации (скорее всего `tma/src/i18n/` или определён inline). Добавить новые ключи:

```
duel.waitingForFriend = "Ожидание ответа на вызов"
duel.linkExpiresIn = "Ссылка активна ещё: {time}"
duel.waiting = "ожидание..."
```

Если локализация inline в компоненте через `useI18n` → добавить ключи в соответствующий объект messages.

Если через внешний файл locale → найти `tma/src/locales/*.json` или `*.ts` и добавить там.

### Step 5: Проверить что нет TypeScript-ошибок

```bash
cd tma && pnpm run type-check
```

Expected: no errors

### Step 6: Запустить линтер

```bash
cd tma && pnpm lint
```

Expected: no errors

### Step 7: Коммит

```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): add outgoing challenge status card to duel lobby"
```

---

## Task 5: Docs — обновить `02_gameplay.md`

**Files:**
- Modify: `docs/game_modes/pvp_duel/02_gameplay.md`

### Step 1: Добавить раздел "0. Invite Link Flow" в начало файла

Вставить после `## Entry Point` и перед `## Game Flow`:

```markdown
## Invite Link Flow (Friend Challenge по ссылке)

### Технический механизм

```
POST /api/v1/duel/challenge/link → { challengeLink: "t.me/bot?startapp=duel_abc123" }
Telegram share sheet → получатель кликает

TMA открывается: SDK launchParams.startParam = "duel_abc123"
useAuth.ts → сохраняет startParam в памяти

App.vue (onMounted):
  registerUser() → Welcome bonus +3 🎟️ (новый пользователь)
  consumeStartParam() → "duel_abc123"
  handleDeepLink() → router.push('/duel?challenge=duel_abc123')

DuelLobbyView (onMounted):
  route.query.challenge = "duel_abc123"
  POST /duel/challenge/accept-by-code { linkCode: "duel_abc123" }
  → { gameId: "g_xyz", startsIn: 3 }
  router.push('/duel/g_xyz')
```

### UX Flow — новый пользователь

| Шаг | Экран | Действие |
|-----|-------|---------|
| 1 | Telegram чат | Кликает invite-ссылку |
| 2 | TMA splash | Загрузка + Telegram Auth |
| 3 | (фон) | `POST /user/register` → +3 🎟️ welcome bonus |
| 4 | DuelLobby | Баннер "Принимаем вызов..." |
| 5 | (фон) | `POST /duel/challenge/accept-by-code` → -1 🎟️ |
| 6 | DuelPlay | "Соперник найден" → 3...2...1 → игра |

### Состояния баннера "Принимаем вызов"

```
┌───────────────────────────────────┐
│  ⟳  Принимаем вызов от друга...   │  ← isPending
└───────────────────────────────────┘

┌───────────────────────────────────┐
│  ✗  Ссылка устарела               │  ← 409 CHALLENGE_EXPIRED
│     Попроси друга прислать новую  │
│                          [Закрыть]│
└───────────────────────────────────┘
```

### Статус для Inviter (тот, кто создал ссылку)

GET /duel/status возвращает `outgoingChallenges`.
Лобби показывает карточку пока ожидается ответ, polling каждые 5с:

```
┌──────────────────────────────────────┐
│  ✈ Ожидание ответа на вызов          │
│  Ссылка активна ещё: 23ч 45мин       │  ● ожидание...
└──────────────────────────────────────┘
```

Когда друг принял → polling обнаруживает `hasActiveDuel: true` → авто-переход в дуэль.

### Edge Cases

| Ситуация | HTTP | UI |
|----------|------|----|
| Ссылка истекла | 409 | "Ссылка устарела. Попроси новую" |
| Ссылка уже использована | 409 | "Вызов уже принят другим игроком" |
| Inviter уже в игре | 409 | "Твой друг сейчас в игре" |
| Inviter отменил ссылку | 404 | "Вызов отменён. Хочешь создать новый?" |
| Себе отправил ссылку | 400 | "Нельзя вызвать самого себя" |
| Уже в очереди | 409 | "Отмени поиск и попробуй снова" |
```

### Step 2: Сохранить файл и проверить что всё на месте

```bash
grep -n "Invite Link Flow\|UX Flow\|Edge Cases" docs/game_modes/pvp_duel/02_gameplay.md
```

Expected: все три раздела найдены

### Step 3: Коммит

```bash
git add docs/game_modes/pvp_duel/02_gameplay.md
git commit -m "docs(pvp-duel): add invite link flow section to 02_gameplay.md"
```

---

## Task 6: Docs — обновить `05_api.md`

**Files:**
- Modify: `docs/game_modes/pvp_duel/05_api.md`

### Step 1: Обновить GET /duel/status response

В `docs/game_modes/pvp_duel/05_api.md`, найти раздел `### 1. Get Duel Status` и добавить `outgoingChallenges` в Response 200:

```json
// Добавить в блок "data" после "pendingChallenges":
"outgoingChallenges": [
  {
    "id": "ch_xyz",
    "challengerId": "user_123",
    "type": "link",
    "status": "pending",
    "challengeLink": "https://t.me/quiz_sprint_dev_bot?startapp=duel_abc123",
    "expiresAt": 1706515200,
    "expiresIn": 82800,
    "createdAt": 1706428800
  }
]
```

### Step 2: Добавить ошибку SELF_CHALLENGE в Error Codes

В `docs/game_modes/pvp_duel/05_api.md`, найти таблицу `## Error Codes` и добавить строку:

```markdown
| 400 | `SELF_CHALLENGE` | Cannot challenge yourself |
```

### Step 3: Коммит

```bash
git add docs/game_modes/pvp_duel/05_api.md
git commit -m "docs(pvp-duel): add outgoingChallenges to status spec and SELF_CHALLENGE error"
```

---

## Verification

После всех задач — финальная проверка:

```bash
# Backend tests
cd backend && go test ./...

# Frontend checks
cd tma && pnpm run type-check && pnpm lint

# Убедиться что новые поля видны в Swagger
# Открыть: http://localhost:3000/swagger/index.html → GET /duel/status
```

Ожидаемый результат: все тесты проходят, нет TypeScript-ошибок, в Swagger виден `outgoingChallenges`.
