# PvP Challenge Flow — Redesign

**Date:** 2026-03-01
**Approach:** B — Full challenge lifecycle
**Scope:** Backend + Frontend

---

## Проблемы текущей реализации

| # | Проблема |
|---|----------|
| 1 | Принятие вызова по ссылке — без подтверждения (авто-принятие) |
| 2 | Исходящие вызовы не отображаются нигде |
| 3 | Нет уведомлений: ни когда вызывают, ни когда друг готов |
| 4 | Нет асинхронной логики: друг принял вызов, но инвайтер офлайн |

---

## Правила билетов

| Действие | Билет |
|----------|-------|
| Создать вызов (инвайтер) | Проверяем ≥1 🎟️ при создании. Нет — блокируем |
| Принять вызов (инвайти) | Не нужен |
| Завершение дуэли | -1 🎟️ у инвайтера |

---

## Challenge Lifecycle (новый)

```
LINK_CREATED → ACCEPTED_WAITING_INVITER → BOTH_READY → IN_GAME
     ↓                    ↓
  EXPIRED              DECLINED
```

| Статус | Описание |
|--------|----------|
| `LINK_CREATED` | Инвайтер создал ссылку, ждём инвайти |
| `ACCEPTED_WAITING_INVITER` | Инвайти принял, инвайтер должен подтвердить старт |
| `BOTH_READY` | Оба подтвердили — создаётся `DuelGame` |
| `EXPIRED` | Ссылка истекла (24ч) |
| `DECLINED` | Инвайти отклонил |

---

## Уведомления через Telegram Bot

### Когда бот пишет

| Событие | Кому | Сообщение |
|---------|------|-----------|
| Создан **прямой** вызов (друг известен) | Инвайти | "@inviter вызывает тебя на дуэль! [Принять →]" |
| Нет реакции 30 мин (прямой вызов) | Инвайти | Напоминание |
| Инвайти принял вызов | Инвайтер | "@friend готов к дуэли! [Зайти в лобби →]" |
| Инвайтер нажал "Начать", инвайти офлайн | Инвайти | "@inviter ждёт тебя! [Зайти →]" |

> **Link-флоу:** бот не знает Telegram ID инвайти заранее (ссылка анонимная).
> Уведомление инвайти возможно только если он уже зарегистрирован в системе.
> Для нового пользователя — только ручной шаринг ссылки инвайтером.

### TelegramNotifier интерфейс (бэкенд)

```go
type TelegramNotifier interface {
    NotifyChallengeCreated(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
    NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
    NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
}
```

Реализуется через Telegram Bot API `sendMessage`. Вызывается из use case'ов, не из хендлеров.

---

## Backend изменения

### 1. Новые статусы в `DuelChallenge`

```go
const (
    ChallengeStatusLinkCreated          = "LINK_CREATED"
    ChallengeStatusAcceptedWaitingInviter = "ACCEPTED_WAITING_INVITER"
    ChallengeStatusBothReady            = "BOTH_READY"
    ChallengeStatusExpired              = "EXPIRED"
    ChallengeStatusDeclined             = "DECLINED"
)
```

### 2. `GET /duel/status` — добавить `outgoingChallenges`

```json
{
  "incomingChallenges": [...],
  "outgoingChallenges": [
    {
      "id": "ch_abc",
      "status": "ACCEPTED_WAITING_INVITER",
      "inviteeName": "@friend",
      "inviteeTelegramID": 123456,
      "acceptedAt": 1706429000,
      "expiresAt": 1706515400
    },
    {
      "id": "ch_xyz",
      "status": "LINK_CREATED",
      "inviteeName": null,
      "acceptedAt": null,
      "expiresAt": 1706515400
    }
  ]
}
```

### 3. Новый эндпоинт: `POST /duel/challenge/{challengeId}/start`

Инвайтер подтверждает старт когда инвайти готов.

```
POST /api/v1/duel/challenge/{challengeId}/start
Body: { "playerId": "..." }
→ 200 { "gameId": "g_xyz" }
→ 404 Challenge not found
→ 409 Challenge not in ACCEPTED_WAITING_INVITER state
```

### 4. `POST /duel/challenge/accept-by-code` — изменение поведения

**Было:** авто-принятие → сразу создаёт игру и возвращает `gameId`.

**Стало:**
- Устанавливает статус `ACCEPTED_WAITING_INVITER`
- Вызывает `TelegramNotifier.NotifyChallengeAccepted` для инвайтера
- Возвращает `{ "status": "ACCEPTED_WAITING_INVITER", "challengeId": "..." }` (без `gameId`)
- Игра создаётся только после `/start`

### 5. Блокировка создания вызова без билета

В `CreateChallengeLinkUseCase` и `SendChallengeUseCase`:
```go
if player.Tickets < 1 {
    return ErrInsufficientTickets
}
```

Билет **не резервируется** при создании — только проверяется наличие. Списание — после завершения дуэли.

---

## Frontend изменения

### 1. Главный экран — бейдж на PvP карте

Бейдж = `incomingChallenges.length + outgoingChallenges.filter(ACCEPTED_WAITING_INVITER).length`

```
┌──────────────────────────────────┐
│  ⚔️  Дуэль              🔴 3     │
│  🥇 Gold III · 1650 MMR          │
│  [ Играть ]                      │
└──────────────────────────────────┘
```

### 2. Лобби — карточки вызовов

#### Входящие вызовы
```
── ВХОДЯЩИЕ ВЫЗОВЫ ──────────────────
┌─────────────────────────────────┐
│  ⚔️  @ProGamer вызывает тебя     │
│  🥇 Gold II · осталось 23ч      │
│  [ ПРИНЯТЬ ]  [ ОТКЛОНИТЬ ]     │
└─────────────────────────────────┘
```

#### Исходящие вызовы
```
── ИСХОДЯЩИЕ ВЫЗОВЫ ─────────────────

// LINK_CREATED — ждём инвайти
┌─────────────────────────────────┐
│  ✈  Ожидаем ответа...           │
│  Ссылка активна ещё 22ч 10мин   │
│  [ ОТПРАВИТЬ СНОВА ]  [ ✕ ]     │
└─────────────────────────────────┘

// ACCEPTED_WAITING_INVITER — друг готов, ждём нас
┌─────────────────────────────────┐
│  ✅  @Kolya готов к дуэли!       │
│  [ НАЧАТЬ ДУЭЛЬ → ]             │
└─────────────────────────────────┘
```

### 3. Модальное окно подтверждения

**Триггер:** пользователь переходит по ссылке (`route.query.challenge` существует).

**Было:** авто-принятие без подтверждения.

**Стало:** показываем модал перед `accept-by-code`.

```
┌─────────────────────────────────┐
│  ⚔️  Тебя вызывают на дуэль!    │
│                                 │
│  @ProGamer (🥇 Gold II)          │
│  хочет сразиться с тобой        │
│                                 │
│  [ ПРИНЯТЬ ВЫЗОВ ]              │
│  [ Отклонить ]                  │
└─────────────────────────────────┘
```

После "ПРИНЯТЬ":
- Вызов `POST /duel/challenge/accept-by-code`
- Ответ `ACCEPTED_WAITING_INVITER` → переходим в лобби с карточкой "Ждём @ProGamer..."
- Если инвайтер онлайн и уже нажал "Начать" → сразу `gameId` → переходим в игру

### 4. Флоу после принятия (инвайти)

```
Принял вызов
      ↓
Polling GET /duel/status каждые 5с
      ↓
outgoingChallenges[n].status === "BOTH_READY"?
      ↓ да
router.push('/duel/{gameId}')
```

---

## Полные UX-флоу

### Новый пользователь по ссылке

| Шаг | Экран | Действие |
|-----|-------|---------|
| 1 | Telegram чат | Кликает invite-ссылку |
| 2 | TMA splash | Загрузка + Telegram Auth |
| 3 | (фон) | `POST /user/register` |
| 4 | Модал подтверждения | "Принять вызов?" |
| 5 | Жмёт "Принять" | `POST /duel/challenge/accept-by-code` |
| 6 | Лобби | Карточка "Ждём @ProGamer..." + polling |
| 7 | Инвайтер жмёт "Начать" | Polling ловит `BOTH_READY` → игра |

### Инвайтер (асинхронный сценарий)

| Шаг | Действие |
|-----|---------|
| 1 | Создаёт ссылку → делится в Telegram вручную |
| 2 | Уходит из приложения |
| 3 | Друг принимает вызов |
| 4 | Бот пишет инвайтеру: "@friend готов к дуэли! [Зайти →]" |
| 5 | Инвайтер заходит → видит карточку "✅ @friend готов!" |
| 6 | Жмёт "Начать дуэль" → `POST /duel/challenge/{id}/start` |
| 7 | Бот пишет инвайти (если офлайн): "@inviter ждёт тебя!" |
| 8 | Оба в игре |

---

## Крайние случаи

| Ситуация | Поведение |
|----------|-----------|
| Инвайти принял, инвайтер офлайн | Статус `ACCEPTED_WAITING_INVITER`. Бот→инвайтер. Лобби показывает карточку "✅ @friend готов!" при следующем входе |
| Инвайтер нажал "Начать", инвайти офлайн | Бот→инвайти. Лобби инвайтера показывает "Ожидаем @friend..." |
| Несколько входящих вызовов | Список в лобби, принимаем по одному |
| Вызов истёк пока играли | При возврате карточка: "Истёк" + кнопка удалить |
| У инвайтера нет билета | Блокируем создание ссылки: "Нет билетов" |
| Инвайти отклонил | Карточка инвайтера: "❌ @friend отклонил". Бот не пишет |
| Link-флоу, инвайти новый пользователь | Бот не может уведомить заранее (нет Telegram ID). Только ручной шаринг |
| Ссылка открыта в браузере | Редирект на `t.me/quiz_sprint_bot?start=duel_abc123` |
| Инвайти пытается принять дважды | 409 — вызов уже принят |
| Инвайтер отменяет ссылку | `DELETE /duel/challenge/{id}` → 404 для инвайти |

---

## Что НЕ меняется

- WebSocket для игрового процесса (только для лобби добавляем polling)
- Matchmaking (за feature flag, не затрагиваем)
- Механика самой дуэли (вопросы, таймер, счёт)
- MMR система
