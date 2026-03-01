# Design: PvP Duel — "Кто зовёт — тот платит"

**Date:** 2026-03-01
**Status:** Approved

---

## Правило

> **Challenger (тот, кто создаёт вызов) платит 1 🎟️. Challenged (тот, кого зовут) играет бесплатно.**

Применяется ко **всем** входящим вызовам:
- Прямой вызов онлайн-другу (`POST /duel/challenge`)
- Invite-ссылка (`POST /duel/challenge/link` → `accept-by-code`)

Случайный поиск (queue) не меняется — каждый платит сам за себя.

---

## Момент списания билета у challenger

**При старте игры** (когда challenged принял), не при создании вызова.

Причина: если ссылка истекла или друг отклонил — билет не тратится.

---

## Изменения по слоям

### Backend

| Use Case | Было | Стало |
|----------|------|-------|
| `RespondChallengeUseCase` (action=accept) | consumeTicket(challengedID) | удалить эту строку |
| `AcceptByLinkCodeUseCase` | consumeTicket(accepterID) | удалить; consumeTicket(challengerID) |
| `SendChallengeUseCase` | нет списания | нет изменений (пока не реализовано) |

**`AcceptByLinkCodeOutput`** — добавить поле:
```go
type AcceptByLinkCodeOutput struct {
    Success          bool   `json:"success"`
    GameID           string `json:"gameId"`
    StartsIn         int    `json:"startsIn"`
    ChallengerID     string `json:"challengerId"`
    FreeForAccepter  bool   `json:"freeForAccepter"` // всегда true
    TicketConsumed   bool   `json:"ticketConsumed"`   // чей билет списан
    TicketOwnerID    string `json:"ticketOwnerId"`    // challengerID
}
```

**Error при нехватке билета у challenger:**
```
400 INSUFFICIENT_TICKETS → challenged видит "Друг не смог начать игру"
Challenge.Status = "failed"
```

### Frontend — UI

**При принятии вызова (любой тип):**
Показать баннер на экране "Соперник найден" перед 3-2-1:

```
┌────────────────────────────────────────────┐
│  🎟️ Бесплатная игра!                       │
│  @PlayerName угощает тебя дуэлью           │
│  Этот матч тебе ничего не стоит            │
└────────────────────────────────────────────┘
```

Длительность: 2-3 секунды, затем обратный отсчёт.

**Изменить label кнопки "Вызвать друга" в лобби:**
```
До:  Стоимость: 1 🎟️ (обоим)
После: Стоимость: 1 🎟️ (с тебя)
```

**Изменить label в карточке pending challenge для challenged:**
```
До:  Стоимость: 1 🎟️
После: Бесплатно для тебя 🎁
```

---

## Таблица: кто платит

| Сценарий | Challenger | Challenged |
|----------|-----------|-----------|
| Случайный поиск | 1 🎟️ | 1 🎟️ |
| Прямой вызов другу | 1 🎟️ при принятии | **0 🎟️** |
| Invite-ссылка | 1 🎟️ при принятии | **0 🎟️** |
| Реванш | 1 🎟️ каждый | 1 🎟️ каждый |

> Реванш — симметричная инициатива, оба соглашаются играть снова.

---

## Edge Cases

| Ситуация | Поведение |
|----------|-----------|
| У challenger нет билета в момент принятия | 400 `INSUFFICIENT_TICKETS`; challenged видит "Друг не смог начать (нет билетов)" |
| Challenged принял, но игра уже есть у challenger | 409 `ALREADY_IN_GAME` |
| Challenger отменил ссылку до принятия | 404/409; ни у кого билет не списывается |
| Новый пользователь принимает ссылку | Welcome bonus → его собственные билеты не тратятся на эту игру; welcome bonus сохраняется целым |

---

## Что обновить в существующем плане реализации

Файл: `docs/plans/2026-03-01-pvp-duel-game-start-flow.md`

**Task 1** (backend outgoingChallenges) — без изменений.

**Добавить Task 2b: Изменить логику списания билетов:**
- `RespondChallengeUseCase`: убрать `consumeTicket(challengedID)`
- `AcceptByLinkCodeUseCase`: убрать `consumeTicket(accepterID)`, добавить `consumeTicket(challengerID)`
- Добавить поле `freeForAccepter: true` в оба output DTO
- Добавить обработку `INSUFFICIENT_TICKETS` на стороне challenger

**Task 4** (DuelLobbyView): добавить баннер + изменить label кнопок.
