# Testing Guide: Daily Challenge

> Admin API key для dev: `dev-admin-key-2026`
> Base URL: `http://localhost:3000` (Docker) или `https://dev.quiz-sprint-tma.online`

## Setup

```bash
# Переменные для удобства
BASE=http://localhost:3000/api/v1
KEY="X-Admin-Key: dev-admin-key-2026"
PLAYER="your-telegram-id"
```

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
