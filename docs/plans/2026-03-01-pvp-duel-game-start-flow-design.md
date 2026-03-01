# Design: PvP Duel — Game Start Flow & Invite Link

**Date:** 2026-03-01
**Status:** Approved
**Scope:** Документация флоу начала игры, поиска соперника, invite-ссылки

---

## Контекст

Существующая документация (`02_gameplay.md`, `05_api.md`) описывает UI-вайрфреймы и API-контракты,
но не описывает:
- Полный пошаговый UX нового пользователя по invite-ссылке
- Технический механизм передачи `startParam` через Telegram SDK
- Статус inviter-а в лобби пока друг не принял вызов
- Edge cases приёма ссылки

---

## Решения

| Вопрос | Решение |
|--------|---------|
| Приоритетный сценарий | Новый пользователь (первое открытие по invite-ссылке) |
| Жизнь ссылки пока новый пользователь регистрируется | 24ч, auto-accept после регистрации |
| Что видит inviter пока ждёт | Карточка в лобби с обратным отсчётом + WebSocket auto-start |
| Билет нового пользователя | Welcome bonus 3 🎟️ при регистрации, 1 тратится на дуэль |

---

## Технический механизм

### Как Telegram передаёт данные в TMA

```
1. POST /api/v1/duel/challenge/link
   → { challengeLink: "https://t.me/quiz_sprint_dev_bot?startapp=duel_abc123" }

2. Telegram share sheet открывается с этой ссылкой

3. Получатель кликает → Telegram открывает TMA
   SDK (@telegram-apps/sdk): launchParams.startParam = "duel_abc123"

4. App.vue / router (при инициализации):
   if (startParam?.startsWith('duel_')) {
     sessionStorage.setItem('pendingDuelCode', startParam)
     router.push('/duel?challenge=' + startParam)
   }
   // sessionStorage нужен на случай перезагрузки TMA до завершения авторизации

5. DuelLobbyView читает route.query.challenge
   → запускает handleAcceptByLinkCode()
```

**Важно:** `startParam` сохраняется в `sessionStorage` до успешного `accept-by-code`,
чтобы не потерять его если TMA перезагрузится во время регистрации.

---

## UX Flow: Новый пользователь по invite-ссылке

```
Шаг 1: Игрок A делится ссылкой
  ├── нажимает "Поделиться" в лобби
  ├── POST /api/v1/duel/challenge/link → ссылка
  └── открывается Telegram share sheet

Шаг 2: Новый пользователь кликает ссылку
  ├── Telegram открывает TMA
  ├── Splash screen / загрузка SDK
  ├── startParam = "duel_abc123" → сохраняется в sessionStorage
  └── Начинается Telegram Auth

Шаг 3: Регистрация (автоматически через Telegram Auth)
  ├── POST /api/v1/user/register
  └── Welcome bonus: +3 🎟️ (баланс: 3)

Шаг 4: Роутер видит pendingDuelCode в sessionStorage
  └── redirect → /duel?challenge=duel_abc123

Шаг 5: DuelLobbyView
  ├── Баннер: "Принимаем вызов от [имя inviter]..."
  ├── POST /api/v1/duel/challenge/accept-by-code
  │     { playerId, linkCode: "duel_abc123" }
  └── Билет списывается (баланс: 2 🎟️)

Шаг 6: Ответ сервера
  ├── { gameId: "g_xyz", startsIn: 3 }
  └── sessionStorage.removeItem('pendingDuelCode')

Шаг 7: router.push('/duel/g_xyz')
  └── DuelPlayView: "Соперник найден" → 3...2...1 → игра
```

---

## UX Flow: Статус inviter в лобби

После создания ссылки inviter видит в лобби новый блок
(данные из `outgoingChallenges` в `GET /api/v1/duel/status`):

```
┌────────────────────────────────────┐
│  📤 Ожидание ответа на вызов       │
│  Ссылка активна ещё: 23ч 45мин     │
│                      [ Отменить ]  │
└────────────────────────────────────┘
```

**Автостарт:** когда друг принял → WebSocket event `challenge_accepted`
→ DuelLobbyView автоматически переходит в `DuelPlayView`.

### Новое поле в GET /api/v1/duel/status

```json
"outgoingChallenges": [
  {
    "challengeId": "ch_abc",
    "type": "link",
    "expiresAt": 1706515200,
    "expiresInSeconds": 82800,
    "status": "pending"
  }
]
```

### Новый WebSocket event (server → inviter)

```json
{
  "type": "challenge_accepted",
  "data": {
    "challengeId": "ch_abc",
    "gameId": "g_xyz789",
    "acceptedBy": {
      "id": "user_new",
      "username": "NewFriend"
    },
    "startsIn": 3
  }
}
```

---

## Edge Cases

| Ситуация | HTTP | Что показать пользователю |
|----------|------|--------------------------|
| Ссылка истекла (>24ч) | 409 `CHALLENGE_EXPIRED` | "Ссылка устарела. Попроси друга прислать новую" |
| Ссылка уже использована | 409 `CHALLENGE_ACCEPTED` | "Вызов уже принят другим игроком" |
| Inviter уже в игре | 409 `ALREADY_IN_GAME` | "Твой друг сейчас в игре. Дождись окончания" |
| Inviter отменил пока шла регистрация | 404/409 | "Вызов отменён. Хочешь бросить вызов первым?" → кнопка "Создать ссылку" |
| startParam потерян (TMA reload) | — | sessionStorage восстанавливает код |
| Нет билетов при accept | 400 `INSUFFICIENT_TICKETS` | Невозможно: welcome bonus 3 🎟️ выдаётся до accept |
| Сам себе отправил ссылку | 400 `SELF_CHALLENGE` | "Нельзя вызвать самого себя" |
| Принимающий уже в очереди | 409 `ALREADY_IN_QUEUE` | "Ты уже ищешь соперника. Отмени поиск и попробуй снова" |

---

## Что нужно добавить в документацию

### В `02_gameplay.md`

Добавить новый раздел **"0. Invite Link Flow"** перед разделом "1. Pre-Game Screen":

- Полный пошаговый флоу нового пользователя (как выше)
- Вайрфрейм баннера "Принимаем вызов..."
- Уточнить "Share link behavior" — добавить про sessionStorage и startParam

### В `05_api.md`

- `GET /api/v1/duel/status` → добавить поле `outgoingChallenges`
- WebSocket → добавить event `challenge_accepted` (server → inviter)
- Error codes → добавить `SELF_CHALLENGE`, уточнить `CHALLENGE_ACCEPTED`

### В `01_concept.md`

Уточнить раздел "Friend Challenge (Direct Invite)":
- Билет нового пользователя берётся из welcome bonus
- Ссылка: 24ч, one-time use, but waits through registration

---

## Что НЕ входит в этот дизайн

- Matchmaking (случайный соперник) — уже описан в `02_gameplay.md`, без изменений
- Rematch flow — уже описан
- Push-уведомления через Telegram Bot API — отдельная фича (Phase 2, не реализована)
- Referral vs Invite разграничение — отдельная тема
