# Marathon Application Layer

Application layer для Solo Marathon игрового режима.

## 📁 Структура

```
marathon/
├── dto.go                       # Input/Output DTOs для всех use cases
├── mapper.go                    # Domain → DTO маппинг
├── event_bus.go                 # EventBus interface (реализация в infrastructure)
├── start_marathon.go            # StartMarathon use case
├── submit_marathon_answer.go   # SubmitMarathonAnswer use case
├── use_marathon_hint.go        # UseMarathonHint use case
├── abandon_marathon.go          # AbandonMarathon use case
├── get_marathon_status.go      # GetMarathonStatus use case
├── get_personal_bests.go       # GetPersonalBests use case
└── get_marathon_leaderboard.go # GetMarathonLeaderboard use case
```

## 🎯 Use Cases

### 1. StartMarathon
**Назначение:** Начать новую игру в Marathon режиме

**Input:**
- `PlayerID` (string) - ID игрока
- `CategoryID` (string, optional) - ID категории или "all" для всех категорий

**Output:**
- `Game` - информация о игре (MarathonGameDTO)
- `FirstQuestion` - первый вопрос
- `TimeLimit` - лимит времени на первый вопрос
- `HasPersonalBest` - есть ли предыдущий рекорд

**Бизнес-логика:**
1. Проверяет, нет ли активной игры у игрока
2. Загружает PersonalBest для категории (если есть)
3. Создает Quiz для Marathon (TODO: domain service для адаптивной сложности)
4. Создает MarathonGame aggregate
5. Публикует `MarathonGameStartedEvent`

---

### 2. SubmitMarathonAnswer
**Назначение:** Отправить ответ на вопрос в Marathon игре

**Input:**
- `GameID` (string) - ID игры
- `QuestionID` (string) - ID вопроса
- `AnswerID` (string) - ID выбранного ответа
- `PlayerID` (string) - ID игрока (для авторизации)
- `TimeTaken` (int64) - время ответа в миллисекундах

**Output:**
- `IsCorrect` (bool) - правильный ли ответ
- `CorrectAnswerID` (string) - ID правильного ответа
- `BasePoints` (int) - базовые очки
- `CurrentStreak` (int) - текущая серия
- `MaxStreak` (int) - максимальная серия в этой игре
- `LifeLost` (bool) - потеряна ли жизнь
- `RemainingLives` (int) - оставшиеся жизни
- `IsGameOver` (bool) - закончилась ли игра
- `NextQuestion` (QuestionDTO, optional) - следующий вопрос
- `GameOverResult` (GameOverResultDTO, optional) - результат игры

**Бизнес-логика:**
1. Проверяет ownership игры
2. Вызывает `game.AnswerQuestion()` (domain logic)
3. Если game over:
   - Обновляет PersonalBest (если новый рекорд)
   - Возвращает финальную статистику
4. Если игра продолжается:
   - Возвращает следующий вопрос
   - Вычисляет адаптивный time limit
5. Публикует события: `MarathonQuestionAnsweredEvent`, `LifeLostEvent`, `MarathonGameOverEvent`

---

### 3. UseMarathonHint
**Назначение:** Использовать подсказку

**Input:**
- `GameID` (string) - ID игры
- `QuestionID` (string) - ID текущего вопроса
- `HintType` (string) - тип подсказки: "fifty_fifty", "extra_time", "skip"
- `PlayerID` (string) - ID игрока (для авторизации)

**Output:**
- `HintType` (string) - тип использованной подсказки
- `RemainingHints` (int) - оставшиеся подсказки этого типа
- `HintResult` (HintResultDTO) - результат применения подсказки:
  - For `fifty_fifty`: `HiddenAnswerIDs` (массив из 2 ID неправильных ответов)
  - For `extra_time`: `NewTimeLimit` (новый лимит времени +10 сек)
  - For `skip`: `NextQuestion` + `NextTimeLimit`

**Бизнес-логика:**
1. Проверяет ownership игры
2. Вызывает `game.UseHint()` (domain logic)
3. В зависимости от типа подсказки:
   - `fifty_fifty`: возвращает 2 ID неправильных ответов для скрытия
   - `extra_time`: возвращает увеличенный time limit
   - `skip`: пропускает вопрос и возвращает следующий (TODO: domain logic)
4. Публикует `HintUsedEvent`

---

### 4. AbandonMarathon
**Назначение:** Завершить игру досрочно (игрок сдался)

**Input:**
- `GameID` (string) - ID игры
- `PlayerID` (string) - ID игрока (для авторизации)

**Output:**
- `GameOverResult` (GameOverResultDTO) - финальная статистика

**Бизнес-логика:**
1. Проверяет ownership игры
2. Вызывает `game.Abandon()` (domain logic)
3. Обновляет PersonalBest (если новый рекорд)
4. Публикует `MarathonGameOverEvent`

---

### 5. GetMarathonStatus
**Назначение:** Получить статус активной Marathon игры игрока

**Input:**
- `PlayerID` (string) - ID игрока

**Output:**
- `HasActiveGame` (bool) - есть ли активная игра
- `Game` (MarathonGameDTO, optional) - информация о игре
- `TimeLimit` (int, optional) - лимит времени на текущий вопрос

**Бизнес-логика:**
1. Ищет активную игру у игрока
2. Если найдена - возвращает полную информацию
3. Если нет - возвращает `HasActiveGame: false`

---

### 6. GetPersonalBests
**Назначение:** Получить все личные рекорды игрока

**Input:**
- `PlayerID` (string) - ID игрока

**Output:**
- `PersonalBests` ([]PersonalBestDTO) - список рекордов по категориям
- `OverallBest` (PersonalBestDTO, optional) - лучший рекорд среди всех категорий

**Бизнес-логика:**
1. Загружает все PersonalBest записи игрока (по всем категориям)
2. Находит лучший рекорд (по streak)
3. Возвращает список и overall best

---

### 7. GetMarathonLeaderboard
**Назначение:** Получить таблицу лидеров для категории

**Input:**
- `CategoryID` (string, optional) - ID категории или "all"
- `TimeFrame` (string, optional) - "all_time", "weekly", "daily" (пока только all_time)
- `Limit` (int) - количество записей (макс 100)

**Output:**
- `Category` (CategoryDTO) - категория
- `TimeFrame` (string) - временной период
- `Entries` ([]LeaderboardEntryDTO) - записи лидерборда
- `PlayerRank` (int, optional) - ранг игрока (TODO)

**Бизнес-логика:**
1. Загружает топ PersonalBest для категории
2. Для каждого record:
   - Загружает username из user repository
   - Создает LeaderboardEntryDTO с рангом
3. TODO: Фильтрация по timeFrame (weekly/daily)
4. TODO: Поиск ранга игрока

---

## 🔧 Зависимости

Каждый use case требует следующие репозитории:

- `marathonRepo` - `solo_marathon.Repository` (всегда)
- `personalBestRepo` - `solo_marathon.PersonalBestRepository` (для рекордов)
- `quizRepo` - `quiz.QuizRepository` (для StartMarathon)
- `categoryRepo` - `quiz.CategoryRepository` (для StartMarathon, GetMarathonLeaderboard)
- `userRepo` - `user.Repository` (для GetMarathonLeaderboard - usernames)
- `eventBus` - `EventBus` (для публикации domain events)

---

## ✅ V2 Updates (2026-01-26)

### Completed
1. **✅ QuestionSelector Domain Service**
   - Location: `backend/internal/domain/solo_marathon/question_selector.go`
   - Weighted random selection based on difficulty distribution
   - Excludes recently shown questions (sliding window of 20)
   - Tests: `question_selector_test.go`

2. **✅ MarathonGameV2 Aggregate**
   - Removed dependency on `kernel.QuizGameplaySession`
   - Uses dynamic question loading via `currentQuestion *quiz.Question`
   - Stores `recentQuestionIDs` for exclusion logic
   - Stores `baseScore` directly (no session)

3. **✅ QuestionRepository Interface**
   - Location: `backend/internal/domain/quiz/question_repository.go`
   - Single source of questions for all game modes
   - Supports filtering by category, difficulty, exclusion

4. **✅ Updated All Use Cases**
   - StartMarathon: Loads first question via QuestionSelector
   - SubmitMarathonAnswer: Loads next question after correct answer
   - AbandonMarathon: Uses V2 baseScore
   - GetMarathonStatus: Uses V2 mapper

See **[ARCHITECTURE.md](./ARCHITECTURE.md)** for detailed V2 architecture.

---

## ⚠️ TODOs

### Высокий приоритет
1. **Skip hint domain logic**
   - Сейчас: возвращает текущий вопрос (не пропускает)
   - Нужно: реализовать skip в `MarathonGameV2.UseHint()`

2. **TimeFrame фильтрация в GetMarathonLeaderboard**
   - Сейчас: только "all_time"
   - Нужно: фильтрация по weekly/daily

### Средний приоритет
4. **Global rank для игрока**
   - Добавить в GameOverResult
   - Требует: leaderboard query после завершения игры

5. **PlayerRank в GetMarathonLeaderboard**
   - Найти ранг текущего игрока в лидерборде

6. **Logging**
   - Добавить structured logging для ошибок
   - Особенно: PersonalBest update failures

---

## 🧪 Тестирование

Для тестирования use cases потребуются моки:
- `marathonRepo` mock
- `personalBestRepo` mock
- `quizRepo` mock
- `categoryRepo` mock
- `userRepo` mock
- `eventBus` mock

См. примеры в `backend/internal/application/quiz/*_test.go`

---

## 📚 Связанные документы

- **Domain model**: `backend/internal/domain/solo_marathon/`
- **Specification**: `docs/game_modes/solo_marathon/`
- **Glossary**: `docs/GLOSSARY.md`
- **Architecture**: `CLAUDE.md`
