# Ubiquitous Language - Quiz Sprint TMA

Словарь терминов и определений для команды разработки. Этот язык используется как в коде, так и в общении.

---

## Core Domain (Quiz Taking)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Quiz Session** | Активная попытка пользователя пройти квиз | Одна активная сессия на квиз на пользователя |
| **User Answer** | Ответ пользователя на конкретный вопрос | Нельзя ответить на один вопрос дважды |
| **Score** | Очки, набранные в сессии. Складываются из базовых очков, бонуса за скорость и бонуса за серию. | Неотрицательное число, увеличивается только |
| **Time Bonus** | Бонусные очки, начисляемые за быстрый правильный ответ. | Зависит от времени ответа. |
| **Streak Bonus** | Бонусные очки, начисляемые за серию правильных ответов. | Начисляется при достижении порога серии. |
| **Correct Answer Streak** | Счетчик последовательных правильных ответов в текущей сессии. | Сбрасывается при неверном ответе. |
| **Session Status** | Состояние сессии (Active, Completed, Abandoned) | Можно отвечать только в Active |
| **Current Question** | Индекс текущего вопроса в сессии | 0-indexed, от 0 до количества вопросов |
| **Time Limit** | Ограничение времени на весь квиз (в секундах) | Положительное число, макс 3600 секунд |
| **Passing Score** | Минимальный процент для прохождения | От 0 до 100% |

---

## Supporting Domain (Quiz Catalog)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Quiz** | Набор вопросов с правилами прохождения | Минимум 5 вопросов, максимум 50 |
| **Question** | Вопрос с вариантами ответов | Ровно 1 правильный ответ из 2-4 вариантов |
| **Answer** | Вариант ответа на вопрос | Текст не пустой, макс 200 символов |
| **Base Points** | Базовые баллы за правильный ответ | Неотрицательное число, макс 1000 |
| **Time Limit Per Question** | Ограничение времени на ответ на один вопрос (в секундах) | Положительное число, от 5 до 60 |
| **Max Time Bonus** | Максимальный бонус в очках за быстрый ответ | Неотрицательное число |
| **Streak Threshold** | Количество правильных ответов для получения бонуса за серию | Положительное число, > 1 |
| **Streak Bonus** | Количество бонусных очков за достижение серии | Неотрицательное число |
| **Category** | Тематическая категория квиза (одна на квиз) | Уникальное название, используется для навигации |
| **Tag** | Дополнительная метка квиза (много на квиз) | Формат `{category}:{value}`, используется для фильтрации |
| **Compact Format** | Оптимизированный формат импорта квизов для LLM | Сокращение токенов на 64% по сравнению с verbose |
| **Batch Import** | Пакетный импорт нескольких квизов одновременно | - |

---

## Supporting Domain (Leaderboard)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Leaderboard** | Таблица лидеров для конкретного квиза | Упорядочена по Score (DESC). Время прохождения используется как вторичный критерий при равенстве очков. |
| **Leaderboard Entry** | Запись о прохождении квиза пользователем | Один пользователь = одна лучшая попытка |
| **Rank** | Позиция в таблице лидеров | Положительное число, начинается с 1 |

---

## Supporting Domain (User Stats)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Current Streak** | Текущая серия дней прохождения Daily Quiz подряд | Сбрасывается при пропуске дня |
| **Longest Streak** | Лучшая серия за все время | Обновляется только если Current > Longest |
| **Last Daily Quiz Date** | Дата последнего прохождения Daily Quiz | Используется для расчета streak |
| **Total Quizzes Completed** | Общее количество завершенных квизов | Неотрицательное число, только увеличивается |
| **Daily Quiz** | Квиз дня, одинаковый для всех пользователей | Детерминированный выбор по дате |

---

## Generic Domain (Identity)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **User** | Пользователь системы | Уникальный UserID |
| **Telegram User** | Пользователь из Telegram | Уникальный TelegramID |
| **Username** | Имя пользователя для отображения | Строка, может быть пустой (anonymous) |

---

## Value Objects (General)

| Термин | Тип | Описание |
|--------|-----|----------|
| **QuizID** | UUID | Уникальный идентификатор квиза |
| **SessionID** | UUID | Уникальный идентификатор сессии |
| **QuestionID** | UUID | Уникальный идентификатор вопроса |
| **AnswerID** | UUID | Уникальный идентификатор варианта ответа |
| **UserID** | UUID | Уникальный идентификатор пользователя |
| **CategoryID** | UUID | Уникальный идентификатор категории |
| **Points** | int | Очки (неотрицательное число) |
| **Timestamp** | int64 | Unix timestamp (секунды с 1970-01-01) |

---

## Domain Events

| Event | Контекст | Payload | Subscribers |
|-------|---------|---------|-------------|
| **QuizStartedEvent** | Quiz Taking | quizID, sessionID, userID, timestamp | Analytics |
| **AnswerSubmittedEvent** | Quiz Taking | sessionID, questionID, answerID, isCorrect, points | Analytics |
| **QuizCompletedEvent** | Quiz Taking | quizID, sessionID, userID, finalScore, timestamp | Leaderboard, User Stats, Analytics |
| **LeaderboardUpdatedEvent** | Leaderboard | quizID, topEntries[] | WebSocket (real-time) |
| **QuizImportedEvent** | Quiz Catalog | quizID, categoryID, tags[], importBatchID, generatedAt, source | Analytics, Search Index |

---

## Session States

```
┌──────────┐
│  Active  │  ← Можно отвечать на вопросы
└─────┬────┘
      │
      ├───────► Completed  (все вопросы отвечены)
      │
      └───────► Abandoned  (пользователь удалил сессию)
```

---

## Scoring Formula

```
Total Score = Σ(Question Score)

Question Score (if correct) = Base Points + Time Bonus + Streak Bonus (если достигнут порог)

Time Bonus = Max Time Bonus × (Time Remaining / Time Limit Per Question)
  - Linear decay: чем быстрее ответ, тем больше бонус
  - Если Time Remaining ≤ 0, бонус = 0

Streak Bonus:
  - Начисляется при достижении Streak Threshold (например, 3 правильных подряд)
  - Одноразовый бонус при достижении порога
  - Streak сбрасывается при неверном ответе

Question Score (if incorrect) = 0
  - Streak сбрасывается в 0
```

**Пример:**
```
Quiz Settings:
- Base Points: 100
- Max Time Bonus: 50
- Time Limit Per Question: 20s
- Streak Threshold: 3
- Streak Bonus: 100

Question 1 (correct, answered in 5s):
  Base: 100
  Time Bonus: 50 × (15/20) = 37.5 ≈ 38
  Streak: 0 (порог не достигнут)
  Total: 138

Question 2 (correct, answered in 8s):
  Base: 100
  Time Bonus: 50 × (12/20) = 30
  Streak: 0 (порог не достигнут)
  Total: 130

Question 3 (correct, answered in 10s):
  Base: 100
  Time Bonus: 50 × (10/20) = 25
  Streak: 100 (порог достигнут! 3 правильных подряд)
  Total: 225

Question 4 (incorrect):
  Total: 0
  Streak сброшен в 0
```

---

## Category vs Tag

**Category** (Одна на квиз):
- Основная классификация
- Используется для навигации в UI
- Примеры: `Programming`, `Geography`, `History`, `Movies`
- Обязательное поле

**Tags** (Много на квиз):
- Дополнительные метки для фильтрации
- Формат: `{category}:{value}`
- Примеры:
  - `language:go`, `language:python`, `language:javascript`
  - `difficulty:easy`, `difficulty:medium`, `difficulty:hard`
  - `topic:variables`, `topic:functions`, `topic:concurrency`
  - `domain:web-development`, `domain:data-structures`
- Опциональное поле (0-10 тегов)

**Примеры использования:**

Quiz: "Go Basics"
- Category: `Programming` (для навигации)
- Tags: `["language:go", "difficulty:easy", "topic:syntax"]` (для фильтрации)

Quiz: "World Capitals"
- Category: `Geography`
- Tags: `["difficulty:medium", "topic:capitals", "domain:world-geography"]`

---

## Telegram Integration

**Telegram Auth:**
- Клиент отправляет Base64-encoded init data в header `Authorization: tma <base64>`
- Backend валидирует криптографическую подпись
- Проверка expiration (1 час)
- Невозможно подделать данные пользователя

**Telegram Notifications:**
- Используются для уведомлений о событиях
- Примеры: Daily Quiz available, Leaderboard position changed

---

## Common Abbreviations

| Abbreviation | Full Term |
|--------------|-----------|
| **TMA** | Telegram Mini App |
| **DDD** | Domain-Driven Design |
| **CQRS** | Command Query Responsibility Segregation |
| **ACL** | Anti-Corruption Layer |
| **DTO** | Data Transfer Object |
| **UUID** | Universally Unique Identifier |

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 1.0
**Проект:** Quiz Sprint TMA
