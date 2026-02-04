# Testing Guide

## Окружения

| Env | Base URL | Admin Key |
|-----|----------|-----------|
| Dev (local) | `http://localhost:3000` | `dev-admin-key-2026` |
| Dev (tunnel) | `https://dev.quiz-sprint-tma.online` | `dev-admin-key-2026` |
| Staging | `https://staging.quiz-sprint-tma.online` | GitHub Secret `STAGING_ADMIN_API_KEY` |

## Setup

```bash
# === Dev ===
BASE=http://localhost:3000/api/v1
KEY="X-Admin-Key: dev-admin-key-2026"
PLAYER="1121083057"

# === Staging ===
BASE=https://staging.quiz-sprint-tma.online/api/v1
KEY="X-Admin-Key: staging-admin-key-2026"
PLAYER="1121083057"
```

> **Staging Admin Key:** хранится в GitHub Secrets → `STAGING_ADMIN_API_KEY`.
> Чтобы добавить/изменить: GitHub → Settings → Secrets → Actions → `STAGING_ADMIN_API_KEY`.
> После изменения нужен редеплой staging.

---

## 1. Полный игровой цикл

```bash
# Начать daily challenge
curl -s -X POST "$BASE/daily-challenge/start" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\"}" | jq

# Ответить на вопрос (повторить 10 раз с разными questionId/answerId)
curl -s -X POST "$BASE/daily-challenge/GAME_ID/answer" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"answerId\": \"A_ID\", \"timeTaken\": 5}" | jq

# Открыть сундук
curl -s -X POST "$BASE/daily-challenge/GAME_ID/chest/open" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\"}" | jq

# Повторная попытка (за монеты или рекламу)
curl -s -X POST "$BASE/daily-challenge/GAME_ID/retry" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"paymentMethod\": \"ad\"}" | jq
```

## 2. Проверка статуса

```bash
# Текущий статус игры
curl -s "$BASE/daily-challenge/status?playerId=$PLAYER" | jq

# Streak игрока
curl -s "$BASE/daily-challenge/streak?playerId=$PLAYER" | jq

# Leaderboard за сегодня
curl -s "$BASE/daily-challenge/leaderboard?limit=10" | jq
```

---

## 3. Admin: Управление streak

```bash
# Установить streak = 30 дней (тест бонуса 1.5x)
curl -s -X PATCH "$BASE/admin/daily-challenge/streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"currentStreak\": 30, \"bestStreak\": 30}" | jq

# Сбросить streak в 0 (тест потери streak)
curl -s -X PATCH "$BASE/admin/daily-challenge/streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"currentStreak\": 0}" | jq

# Установить lastPlayedDate на позавчера (тест прерывания streak)
curl -s -X PATCH "$BASE/admin/daily-challenge/streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"lastPlayedDate\": \"2026-01-29\"}" | jq
```

### Пороги streak-бонусов

| Streak | Множитель | Команда |
|--------|-----------|---------|
| 3 дня  | 1.1x      | `"currentStreak": 3` |
| 7 дней | 1.25x     | `"currentStreak": 7` |
| 30 дней| 1.5x      | `"currentStreak": 30` |

## 4. Admin: Симуляция streak

```bash
# Создать 7 completed-игр за последние 7 дней
curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"days\": 7, \"baseScore\": 50}" | jq

# Создать 30-дневный streak с высоким счётом
curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"days\": 30, \"baseScore\": 80}" | jq
```

## 5. Admin: Полный сброс игрока

```bash
# Полный сброс — удаляет ВСЕ данные: daily games, marathon, quiz sessions, stats
# Профиль пользователя сохраняется
curl -s -X DELETE "$BASE/admin/player/reset?playerId=$PLAYER" \
  -H "$KEY" | jq
```

Ответ покажет что именно удалено:
```json
{
  "data": {
    "playerId": "1121083057",
    "totalDeleted": 8,
    "deleted": {
      "quiz_sessions": 2,
      "daily_games": 5,
      "marathon_games": 0,
      "marathon_personal_bests": 0,
      "user_stats": 1
    }
  }
}
```

## 6. Admin: Сброс daily challenge

```bash
# Удалить сегодняшнюю игру (чтобы переиграть)
curl -s -X DELETE "$BASE/admin/daily-challenge/games?playerId=$PLAYER&date=$(date +%Y-%m-%d)" \
  -H "$KEY" | jq

# Удалить ВСЕ daily games (но не marathon/quiz sessions)
curl -s -X DELETE "$BASE/admin/daily-challenge/games?playerId=$PLAYER" \
  -H "$KEY" | jq
```

## 7. Admin: Дебаг

```bash
# Посмотреть все игры игрока (последние 20)
curl -s "$BASE/admin/daily-challenge/games?playerId=$PLAYER" \
  -H "$KEY" | jq

# Посмотреть больше записей
curl -s "$BASE/admin/daily-challenge/games?playerId=$PLAYER&limit=50" \
  -H "$KEY" | jq
```

---

## Типичные тест-сценарии

### A. Тест streak-бонусов
```bash
# 1. Полный сброс
curl -s -X DELETE "$BASE/admin/player/reset?playerId=$PLAYER" -H "$KEY"

# 2. Симулировать 6 дней
curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"days\": 6}"

# 3. Сыграть сегодня — streak станет 7, бонус 1.25x
curl -s -X POST "$BASE/daily-challenge/start" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\"}"
```

### B. Тест потери streak
```bash
# 1. Установить streak = 15, но lastPlayedDate = позавчера
curl -s -X PATCH "$BASE/admin/daily-challenge/streak" \
  -H "$KEY" -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"currentStreak\": 15, \"lastPlayedDate\": \"$(date -v-2d +%Y-%m-%d)\"}"

# 2. Начать игру — streak должен сброситься
curl -s -X POST "$BASE/daily-challenge/start" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\"}"
```

### C. Тест retry-механики
```bash
# 1. Удалить сегодняшние игры
curl -s -X DELETE "$BASE/admin/daily-challenge/games?playerId=$PLAYER&date=$(date +%Y-%m-%d)" -H "$KEY"

# 2. Сыграть и завершить первую попытку
# 3. Вызвать retry — проверить что attempt_number = 2
# 4. Попробовать третий retry — должен получить ошибку
```

### D. Тест сундуков
```bash
# 1. Удалить сегодняшнюю игру
curl -s -X DELETE "$BASE/admin/daily-challenge/games?playerId=$PLAYER&date=$(date +%Y-%m-%d)" -H "$KEY"

# 2. Сыграть, завершить, открыть сундук
# 3. Проверить тип сундука в debug view:
curl -s "$BASE/admin/daily-challenge/games?playerId=$PLAYER&limit=1" -H "$KEY" | jq '.data.games[0] | {chestType, chestCoins, baseScore}'
```

---
---

# Marathon (Solo)

## 1. Полный игровой цикл

```bash
# Начать марафон (categoryId = null → все категории)
curl -s -X POST "$BASE/marathon/start" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"categoryId\": null}" | jq

# Сохранить GAME_ID из ответа
GAME_ID="id-from-response"

# Ответить на вопрос (повторять с разными questionId/answerId)
curl -s -X POST "$BASE/marathon/$GAME_ID/answer" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"answerId\": \"A_ID\", \"timeTaken\": 5}" | jq

# Использовать бонус (shield / fifty_fifty / skip / freeze)
curl -s -X POST "$BASE/marathon/$GAME_ID/bonus" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"bonusType\": \"shield\"}" | jq

# Продолжить после game over (за монеты или рекламу)
curl -s -X POST "$BASE/marathon/$GAME_ID/continue" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"paymentMethod\": \"ad\"}" | jq

# Сдаться (abandon)
curl -s -X DELETE "$BASE/marathon/$GAME_ID" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\"}" | jq
```

## 2. Проверка статуса

```bash
# Статус активной игры (есть ли незавершённая)
curl -s "$BASE/marathon/status?playerId=$PLAYER" | jq

# Персональные рекорды по категориям
curl -s "$BASE/marathon/personal-bests?playerId=$PLAYER" | jq

# Leaderboard (all_time | weekly | daily)
curl -s "$BASE/marathon/leaderboard?categoryId=all&timeFrame=all_time&limit=10" | jq
```

## 3. Сброс данных марафона

```bash
# Полный сброс игрока (удаляет ВСЁ: daily, marathon, quiz sessions, stats)
curl -s -X DELETE "$BASE/admin/player/reset?playerId=$PLAYER" \
  -H "$KEY" | jq
```

> Admin-эндпоинтов специфичных для марафона пока нет. Используй полный сброс.

---

## Типичные тест-сценарии (Marathon)

### A. Старт → ответы → game over
```bash
# 1. Сброс
curl -s -X DELETE "$BASE/admin/player/reset?playerId=$PLAYER" -H "$KEY"

# 2. Старт
curl -s -X POST "$BASE/marathon/start" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"categoryId\": null}" | jq
# → Сохранить game.id, game.currentQuestion.id, game.currentQuestion.answers[0].id

# 3. Отвечать неправильно 3 раза подряд (потеря всех жизней)
# → isGameOver: true, gameOverResult с continueOffer
```

### B. Тест бонусов
```bash
# 1. Начать игру
# 2. Ответить правильно на 5+ вопросов (набрать бонусы на milestones)
# 3. Проверить bonusInventory в ответе
# 4. Использовать fifty_fifty:
curl -s -X POST "$BASE/marathon/$GAME_ID/bonus" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"bonusType\": \"fifty_fifty\"}" | jq
# → bonusResult.hiddenAnswerIds — 2 неправильных ответа скрыты

# 5. Использовать skip:
curl -s -X POST "$BASE/marathon/$GAME_ID/bonus" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"bonusType\": \"skip\"}" | jq
# → bonusResult.nextQuestion — новый вопрос

# 6. Использовать freeze:
curl -s -X POST "$BASE/marathon/$GAME_ID/bonus" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"bonusType\": \"freeze\"}" | jq
# → bonusResult.newTimeLimit — увеличенный таймер
```

### C. Тест continue после game over
```bash
# 1. Довести игру до game over (3 неправильных ответа)
# 2. Проверить gameOverResult.continueOffer:
#    - available: true
#    - costCoins, hasAd, continueCount
# 3. Продолжить:
curl -s -X POST "$BASE/marathon/$GAME_ID/continue" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"paymentMethod\": \"ad\"}" | jq
# → game с восстановленной жизнью, continueCount + 1

# 4. Довести до game over снова → continue повторно
# → costCoins увеличивается с каждым continue
```

### D. Тест resume (продолжение активной игры)
```bash
# 1. Начать марафон, ответить на пару вопросов
# 2. Проверить статус:
curl -s "$BASE/marathon/status?playerId=$PLAYER" | jq
# → hasActiveGame: true, game.currentQuestion не null

# 3. Нажать "Continue Marathon" в UI — должен загрузить текущий вопрос
# → Если бесконечная загрузка: проверить что currentQuestion есть в ответе status
```

### E. Тест shield
```bash
# 1. Набрать shield бонус (milestone)
# 2. Активировать shield:
curl -s -X POST "$BASE/marathon/$GAME_ID/bonus" \
  -H "Content-Type: application/json" \
  -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"Q_ID\", \"bonusType\": \"shield\"}" | jq
# → shieldActive: true

# 3. Ответить неправильно:
# → shieldConsumed: true, lifeLost: false (щит поглотил удар)
```

### F. Тест personal best
```bash
# 1. Сброс
curl -s -X DELETE "$BASE/admin/player/reset?playerId=$PLAYER" -H "$KEY"

# 2. Сыграть первую игру → завершить → проверить personal-bests
curl -s "$BASE/marathon/personal-bests?playerId=$PLAYER" | jq

# 3. Сыграть вторую игру с лучшим результатом
# → isNewPersonalBest: true в gameOverResult
```
