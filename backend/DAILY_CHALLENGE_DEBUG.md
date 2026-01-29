# Daily Challenge - Debug Guide

## Проблемы и решения

### ✅ Исправленные проблемы UI

1. **Score не отображался**
   - **Было**: `{{ state.game.score }}` (поле не существует)
   - **Стало**: `{{ state.game.finalScore || 0 }}`
   - **Файл**: `tma/src/components/DailyChallenge/DailyChallengeCard.vue:253`

2. **Progress показывал 10% вместо 100%**
   - **Было**: Прогресс вычислялся только для `isPlaying`
   - **Стало**: Показывается 100% для завершённых игр с зелёным цветом
   - **Файл**: `tma/src/components/DailyChallenge/DailyChallengeCard.vue:193-200`

## Скрипт для сброса Daily Challenge

### Использование

```bash
# Сбросить сегодняшний Daily Challenge
cd backend
make reset-daily-challenge

# Сбросить конкретную дату
make reset-daily-challenge DATE=2026-01-27

# Или напрямую через скрипт
./scripts/reset-daily-challenge.sh
./scripts/reset-daily-challenge.sh 2026-01-27
```

### Что делает скрипт

1. Удаляет все игры (`daily_games`) для указанной даты
2. Удаляет daily quiz (`daily_quizzes`) для указанной даты
3. Показывает статистику удалённых записей

### Workflow для отладки

```bash
# 1. Сбросить Daily Challenge
make reset-daily-challenge

# 2. Обновить браузер
# - Очистить localStorage (или сделает автоматически миграция)
# - Обновить страницу (Cmd+R)

# 3. Начать новый Daily Challenge
# - Кликнуть "Start Challenge"
# - Играть и проверять
```

## Архитектура Daily Challenge

### База данных

**Таблица `daily_quizzes`:**
- `id` - UUID daily quiz
- `date` - DATE (уникальный)
- `question_ids` - JSONB массив из 10 question IDs
- `expires_at` - BIGINT timestamp (следующий день 00:00 UTC)
- `created_at` - BIGINT timestamp

**Таблица `daily_games`:**
- `id` - UUID игры
- `player_id` - TEXT user ID
- `daily_quiz_id` - UUID ссылка на daily_quizzes
- `date` - DATE
- `status` - TEXT ('in_progress', 'completed')
- `session_state` - JSONB (сериализованная сессия)
- `current_streak` - INT
- `best_streak` - INT
- `last_played_date` - DATE (nullable)
- `rank` - INT (nullable)

### Важные особенности

1. **Ephemeral Quiz**: Quiz создаётся в памяти, НЕ сохраняется в таблицу `quizzes`
   - Причина: Временный quiz только на день
   - Решение: Custom десериализатор `deserializeDailyChallengeSession()` восстанавливает quiz из question_ids

2. **Streak System**: Автоматически вычисляется при старте игры
   - Загружает предыдущую игру (yesterday)
   - Копирует streak или создаёт новый

3. **Rank Calculation**: Вычисляется после завершения игры
   - Формула: `finalScore = baseScore * streakMultiplier`
   - Сортировка: по `finalScore DESC`, потом по `completed_at ASC`

## Типичные ошибки

### 1. `gameId: undefined`
**Причина**: Старые данные в localStorage без поля `gameId`
**Решение**: Автоматическая миграция в `useDailyChallenge.ts:159-164`

### 2. `sql: converting NULL to string is unsupported`
**Причина**: `last_played_date` может быть NULL
**Решение**: Использовать `sql.NullString` при Scan/Insert

### 3. `failed to load quiz: quiz not found`
**Причина**: Quiz не сохранён в БД (ephemeral)
**Решение**: `deserializeDailyChallengeSession()` восстанавливает quiz из daily_quiz

## Тестирование

### Ручное тестирование

```bash
# 1. Полный цикл
make reset-daily-challenge
# - Обновить app
# - Start Challenge
# - Ответить на все 10 вопросов
# - Проверить Results page
# - Проверить Home card (Completed status)

# 2. Streak тестирование
# День 1
make reset-daily-challenge DATE=2026-01-25
# - Сыграть и завершить
# - Проверить streak = 1

# День 2
make reset-daily-challenge DATE=2026-01-26
# - Сыграть и завершить
# - Проверить streak = 2, bonus = +10%

# День 3
make reset-daily-challenge DATE=2026-01-27
# - Сыграть и завершить
# - Проверить streak = 3, bonus = +10%
```

### Unit тесты

```bash
# Запустить тесты question selector
go test ./internal/domain/solo_marathon -v -run TestQuestionSelector

# Все тесты
go test ./...
```

## API Endpoints

```
POST   /api/v1/daily-challenge/start
POST   /api/v1/daily-challenge/:gameId/answer
GET    /api/v1/daily-challenge/status?playerId=X
GET    /api/v1/daily-challenge/leaderboard?date=YYYY-MM-DD
GET    /api/v1/daily-challenge/streak?playerId=X
```

## Логи для отладки

```bash
# Просмотр логов backend
docker compose -f docker-compose.dev.yml logs -f api | grep -i daily

# Паттерны в логах
📝 [SubmitDailyAnswer] Received request
✅ [StartDailyChallenge] Created daily game
❌ [SubmitDailyAnswer] Failed to find game
```

## Полезные SQL запросы

```sql
-- Показать все daily quizzes
SELECT id, date, created_at FROM daily_quizzes ORDER BY date DESC LIMIT 5;

-- Показать все игры для сегодня
SELECT id, player_id, status, current_streak, rank
FROM daily_games
WHERE date = CURRENT_DATE
ORDER BY rank NULLS LAST;

-- Показать streak игрока
SELECT player_id, date, current_streak, best_streak, last_played_date
FROM daily_games
WHERE player_id = '1121083057'
ORDER BY date DESC
LIMIT 7;

-- Удалить все для отладки
DELETE FROM daily_games WHERE date = '2026-01-27';
DELETE FROM daily_quizzes WHERE date = '2026-01-27';
```
