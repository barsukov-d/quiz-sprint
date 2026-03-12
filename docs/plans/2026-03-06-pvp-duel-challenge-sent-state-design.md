# Design: PvP Duel — запрет дублей и UI состояние "Вызов отправлен"

Date: 2026-03-06

## Problem

1. Игрок может отправить несколько вызовов одному и тому же сопернику — бэкенд не проверяет дубли.
2. Кнопка "Вызов" в списке соперников не отражает состояние: не видно, что вызов уже отправлен.

## Thin Client Principle

Состояние хранится на бэкенде. Фронт только рендерит. Поэтому флаг `hasPendingChallenge` должен приходить в ответе API, а не вычисляться на фронте.

## Solution

### Backend — 1: запрет дублей в `SendChallengeUseCase`

Перед созданием нового challenge проверить `FindPendingByChallenger(challengerID)`.
Если уже есть pending challenge к тому же `friendID` — вернуть `ErrChallengeAlreadySent`.

Handler маппит `ErrChallengeAlreadySent` → HTTP 409.

### Backend — 2: `hasPendingChallenge` в `RivalItemDTO`

`GetRivalsUseCase` получает `challengeRepo` как зависимость.
При построении списка соперников — для каждого rival проверяется наличие pending challenge
от `playerID` к `rival.id`. Результат записывается в `hasPendingChallenge bool`.

Изменения:
- `RivalItemDTO` (swagger_models.go): добавить поле `HasPendingChallenge bool json:"hasPendingChallenge"`
- `RivalDTO` (dto.go): добавить поле `HasPendingChallenge bool`
- `GetRivalsUseCase`: принять `challengeRepo`, заполнять поле для каждого rival
- Handler wire: передать `challengeRepo` в конструктор use case
- Регенерировать Swagger + TypeScript типы

### Frontend

После регенерации типов — обновить кнопку в `DuelLobbyView.vue`:

```html
<UButton
  size="xs"
  :disabled="rival.hasPendingChallenge"
  :color="rival.hasPendingChallenge ? 'gray' : 'primary'"
  @click="() => !rival.hasPendingChallenge && handleChallengeFriend(rival.id!)"
>
  {{ rival.hasPendingChallenge ? t('duel.challengeSent') : t('duel.challenge') }}
</UButton>
```

Добавить i18n ключ `duel.challengeSent` = `"Вызов отправлен"` во все локали.

## Data Flow

```
GET /duel/rivals?playerId=X
  → GetRivalsUseCase
  → для каждого rival: FindPendingByChallenger(X) → фильтр по challengedId
  → RivalItemDTO { ..., hasPendingChallenge: true/false }
  → фронт рендерит кнопку disabled/active
```

```
POST /duel/challenge { playerId, friendId }
  → SendChallengeUseCase
  → FindPendingByChallenger(playerId) → проверить дубль
  → если дубль: ErrChallengeAlreadySent → 409
  → иначе: создать, сохранить → 201
```

## Files to Change

### Backend
- `backend/internal/domain/quick_duel/errors.go` — добавить `ErrChallengeAlreadySent`
- `backend/internal/application/quick_duel/use_cases.go` — дублепроверка в `SendChallengeUseCase`, `challengeRepo` в `GetRivalsUseCase`
- `backend/internal/application/quick_duel/dto.go` — `HasPendingChallenge` в `RivalDTO`
- `backend/internal/infrastructure/http/handlers/swagger_models.go` — `HasPendingChallenge` в `RivalItemDTO`
- `backend/internal/infrastructure/http/handlers/duel_handlers.go` — wire + error mapper

### Frontend
- `tma/` — `pnpm run generate:all`
- `tma/src/views/Duel/DuelLobbyView.vue` — обновить кнопку
- `tma/src/locales/*.json` — добавить `duel.challengeSent`
