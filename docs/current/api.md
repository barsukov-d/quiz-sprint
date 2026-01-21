# API Reference - Current Implementation

> **Автогенерированная документация:** `http://localhost:3000/swagger/index.html`
> **Для доменной модели см.:** [`domain.md`](./domain.md)
> **Для user flows см.:** [`user-flows.md`](./user-flows.md)

---

## 📋 Содержание

1. [Base URL](#base-url)
2. [Authentication](#authentication)
3. [Quiz Endpoints](#quiz-endpoints)
4. [Session Endpoints](#session-endpoints)
5. [Leaderboard Endpoints](#leaderboard-endpoints)
6. [User Endpoints](#user-endpoints)
7. [Category Endpoints](#category-endpoints)
8. [WebSocket](#websocket)
9. [Error Responses](#error-responses)

---

## Base URL

| Environment | URL |
|-------------|-----|
| Development | `https://dev.quiz-sprint-tma.online/api/v1` |
| Staging | `https://staging.quiz-sprint-tma.online/api/v1` |
| Production | `https://quiz-sprint-tma.online/api/v1` |

---

## Authentication

**Method:** Telegram Mini App authentication

**Header:**
```
Authorization: tma <base64-encoded-init-data>
```

**Backend Validation:**
- Криптографическая проверка подписи
- Проверка expiration (1 час)
- Невозможно подделать данные пользователя

**Примечание:** Некоторые endpoints требуют аутентификации (отмечены 🔒)

---

## Quiz Endpoints

### GET /quiz
**Описание:** Получить список всех квизов

**Требует аутентификации:** ❌

**Query Parameters:**
```
?categoryId={uuid}  // Опционально: фильтр по категории
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "General Knowledge",
      "description": "Test your general knowledge",
      "categoryId": "uuid",
      "categoryName": "General",
      "questionCount": 10,
      "estimatedTime": 5,
      "totalPoints": 1000,
      "passingScore": 60
    }
  ]
}
```

---

### GET /quiz/:id
**Описание:** Получить детали квиза

**Требует аутентификации:** ❌

**Path Parameters:**
- `id` - UUID квиза

**Response 200:**
```json
{
  "data": {
    "id": "uuid",
    "title": "General Knowledge",
    "description": "Test your knowledge across various topics",
    "categoryId": "uuid",
    "questionCount": 10,
    "estimatedTime": 5,
    "totalPoints": 1000,
    "passingScore": 60,
    "basePoints": 100,
    "timeLimitPerQuestion": 20,
    "maxTimeBonus": 50,
    "streakThreshold": 3,
    "streakBonus": 100
  }
}
```

**Response 404:**
```json
{
  "error": "Quiz not found"
}
```

---

### GET /quiz/daily
**Описание:** Получить квиз дня (Daily Challenge)

**Требует аутентификации:** 🔒 Да

**Query Parameters:**
```
?userId={uuid}  // UUID пользователя (из Telegram auth)
```

**Response 200 (не пройден сегодня):**
```json
{
  "data": {
    "quiz": {
      "id": "uuid",
      "title": "Столицы мира",
      "description": "...",
      "questionCount": 10,
      "estimatedTime": 5
    },
    "completionStatus": "not_attempted",
    "userResult": null
  }
}
```

**Response 200 (уже пройден сегодня):**
```json
{
  "data": {
    "quiz": { ... },
    "completionStatus": "completed",
    "userResult": {
      "score": 8530,
      "rank": 12,
      "completedAt": "2026-01-21T10:30:00Z"
    }
  }
}
```

---

### GET /quiz/random
**Описание:** Получить случайный квиз

**Требует аутентификации:** ❌

**Query Parameters:**
```
?categoryId={uuid}        // Опционально: фильтр по категории
&excludeCompleted=true    // Опционально: исключить пройденные (требует userId)
&userId={uuid}            // Требуется если excludeCompleted=true
```

**Response 200:**
```json
{
  "data": {
    "id": "uuid",
    "title": "Random Quiz Title",
    "description": "...",
    "questionCount": 15,
    "estimatedTime": 7
  }
}
```

---

## Session Endpoints

### POST /quiz/:id/start
**Описание:** Начать прохождение квиза

**Требует аутентификации:** 🔒 Да

**Path Parameters:**
- `id` - UUID квиза

**Request Body:**
```json
{
  "userId": "uuid"
}
```

**Response 200:**
```json
{
  "data": {
    "session": {
      "sessionId": "uuid",
      "quizId": "uuid",
      "userId": "uuid",
      "status": "active",
      "score": 0,
      "currentQuestionIndex": 0,
      "streak": 0,
      "startedAt": "2026-01-21T10:00:00Z"
    },
    "firstQuestion": {
      "questionId": "uuid",
      "text": "What is the capital of France?",
      "answers": [
        { "id": "uuid-1", "text": "London" },
        { "id": "uuid-2", "text": "Paris" },
        { "id": "uuid-3", "text": "Berlin" },
        { "id": "uuid-4", "text": "Madrid" }
      ],
      "timeLimit": 20
    }
  }
}
```

**Response 409 (уже есть активная сессия):**
```json
{
  "error": "Active session already exists for this quiz",
  "activeSession": {
    "sessionId": "uuid",
    "currentQuestionIndex": 3,
    "score": 245
  }
}
```

---

### GET /quiz/:id/active-session
**Описание:** Получить активную сессию (для восстановления)

**Требует аутентификации:** 🔒 Да

**Path Parameters:**
- `id` - UUID квиза

**Query Parameters:**
```
?userId={uuid}
```

**Response 200:**
```json
{
  "data": {
    "session": {
      "sessionId": "uuid",
      "quizId": "uuid",
      "score": 245,
      "currentQuestionIndex": 3,
      "streak": 2,
      "startedAt": "2026-01-21T09:00:00Z"
    },
    "currentQuestion": {
      "questionId": "uuid",
      "text": "Which planet is known as the Red Planet?",
      "answers": [ ... ],
      "timeLimit": 20
    },
    "totalQuestions": 10
  }
}
```

**Response 404:**
```json
{
  "error": "No active session found"
}
```

---

### POST /quiz/session/:sessionId/answer
**Описание:** Отправить ответ на вопрос

**Требует аутентификации:** 🔒 Да

**Path Parameters:**
- `sessionId` - UUID сессии

**Request Body:**
```json
{
  "questionId": "uuid",
  "answerId": "uuid",
  "userId": "uuid",
  "timeTaken": 8
}
```

**Response 200:**
```json
{
  "data": {
    "isCorrect": true,
    "pointsEarned": {
      "base": 100,
      "timeBonus": 30,
      "streakBonus": 0,
      "total": 130
    },
    "streakInfo": {
      "current": 2,
      "threshold": 3,
      "bonusEarned": false
    },
    "nextQuestion": {
      "questionId": "uuid",
      "text": "Next question text...",
      "answers": [ ... ],
      "timeLimit": 20
    }
  }
}
```

**Response 200 (последний вопрос):**
```json
{
  "data": {
    "isCorrect": true,
    "pointsEarned": { ... },
    "streakInfo": { ... },
    "nextQuestion": null,
    "finalResults": {
      "sessionId": "uuid",
      "totalScore": 8530,
      "passed": true
    }
  }
}
```

**Response 400:**
```json
{
  "error": "Question already answered"
}
```

---

### DELETE /quiz/session/:sessionId
**Описание:** Удалить активную сессию (Abandon)

**Требует аутентификации:** 🔒 Да

**Path Parameters:**
- `sessionId` - UUID сессии

**Query Parameters:**
```
?userId={uuid}
```

**Response 200:**
```json
{
  "message": "Session deleted successfully"
}
```

**Response 403:**
```json
{
  "error": "Unauthorized: not your session"
}
```

---

### GET /quiz/session/:sessionId
**Описание:** Получить результаты сессии (для Results экрана)

**Требует аутентификации:** ❌

**Path Parameters:**
- `sessionId` - UUID сессии

**Response 200:**
```json
{
  "data": {
    "session": {
      "sessionId": "uuid",
      "quizId": "uuid",
      "userId": "uuid",
      "score": 8530,
      "status": "completed",
      "startedAt": "2026-01-21T10:00:00Z",
      "completedAt": "2026-01-21T10:04:23Z"
    },
    "quiz": {
      "title": "General Knowledge",
      "description": "...",
      "totalPoints": 10000,
      "passingScore": 60
    },
    "statistics": {
      "totalQuestions": 10,
      "correctAnswers": 17,
      "timeSpent": 263,
      "passed": true,
      "scorePercentage": 85.3,
      "longestStreak": 8,
      "avgAnswerTime": 3.2
    }
  }
}
```

---

## Leaderboard Endpoints

### GET /quiz/:id/leaderboard
**Описание:** Получить таблицу лидеров квиза

**Требует аутентификации:** ❌

**Path Parameters:**
- `id` - UUID квиза

**Query Parameters:**
```
?limit=50  // default: 50, max: 100
```

**Response 200:**
```json
{
  "data": {
    "entries": [
      {
        "userId": "uuid",
        "username": "@username1",
        "score": 10000,
        "rank": 1,
        "completedAt": "2026-01-20T15:30:00Z"
      },
      {
        "userId": "uuid",
        "username": "@username2",
        "score": 9850,
        "rank": 2,
        "completedAt": "2026-01-21T09:15:00Z"
      }
    ],
    "totalPlayers": 1234
  }
}
```

---

## User Endpoints

### POST /user/register
**Описание:** Регистрация/авторизация пользователя через Telegram

**Требует аутентификации:** 🔒 Да

**Request Body:**
```json
{
  "telegramId": 123456789,
  "username": "john_doe",
  "firstName": "John",
  "lastName": "Doe"
}
```

**Response 200:**
```json
{
  "data": {
    "userId": "uuid",
    "telegramId": 123456789,
    "username": "john_doe",
    "createdAt": "2026-01-21T10:00:00Z"
  }
}
```

---

### GET /user/:id
**Описание:** Получить информацию о пользователе

**Требует аутентификации:** ❌

**Path Parameters:**
- `id` - UUID пользователя

**Response 200:**
```json
{
  "data": {
    "userId": "uuid",
    "username": "john_doe",
    "firstName": "John",
    "createdAt": "2026-01-15T00:00:00Z"
  }
}
```

---

### GET /user/:id/stats
**Описание:** Получить статистику пользователя

**Требует аутентификации:** ❌

**Path Parameters:**
- `id` - UUID пользователя

**Response 200:**
```json
{
  "data": {
    "currentStreak": 5,
    "longestStreak": 12,
    "lastDailyQuizDate": "2026-01-21",
    "totalQuizzesCompleted": 47
  }
}
```

---

### GET /user/:id/sessions/active
**Описание:** Получить все активные сессии пользователя

**Требует аутентификации:** 🔒 Да

**Path Parameters:**
- `id` - UUID пользователя

**Response 200:**
```json
{
  "data": {
    "sessions": [
      {
        "sessionId": "uuid",
        "quizId": "uuid",
        "quizTitle": "General Knowledge",
        "currentQuestion": 3,
        "totalQuestions": 10,
        "score": 245,
        "startedAt": "2026-01-21T09:00:00Z"
      },
      {
        "sessionId": "uuid",
        "quizId": "uuid",
        "quizTitle": "Geography Quiz",
        "currentQuestion": 7,
        "totalQuestions": 15,
        "score": 680,
        "startedAt": "2026-01-20T18:00:00Z"
      }
    ]
  }
}
```

---

## Category Endpoints

### GET /categories
**Описание:** Получить список всех категорий

**Требует аутентификации:** ❌

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "General Knowledge",
      "slug": "general-knowledge",
      "description": "Test your general knowledge",
      "icon": "🧠",
      "quizCount": 12
    },
    {
      "id": "uuid",
      "name": "Geography",
      "slug": "geography",
      "description": "Explore the world",
      "icon": "🌍",
      "quizCount": 8
    }
  ]
}
```

---

### POST /categories
**Описание:** Создать новую категорию (Admin only)

**Требует аутентификации:** 🔒 Да (Admin)

**Request Body:**
```json
{
  "name": "Science",
  "description": "Scientific knowledge",
  "icon": "🧬"
}
```

**Response 201:**
```json
{
  "data": {
    "id": "uuid",
    "name": "Science",
    "slug": "science",
    "description": "Scientific knowledge",
    "icon": "🧬"
  }
}
```

---

## WebSocket

### ws://*/ws/leaderboard/:quizId
**Описание:** Real-time обновления таблицы лидеров

**Connection:**
```javascript
const ws = new WebSocket('wss://quiz-sprint-tma.online/ws/leaderboard/{quizId}');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Leaderboard updated:', data);
};
```

**Server Message (LeaderboardUpdatedEvent):**
```json
{
  "type": "leaderboard_updated",
  "quizId": "uuid",
  "entries": [
    {
      "userId": "uuid",
      "username": "@username1",
      "score": 10000,
      "rank": 1
    }
  ],
  "timestamp": "2026-01-21T10:30:00Z"
}
```

---

## Error Responses

### Standard Error Format
```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {}  // Опциональные детали
}
```

### HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Успешный запрос |
| 201 | Created | Ресурс создан |
| 400 | Bad Request | Невалидные данные |
| 401 | Unauthorized | Нет или невалидная аутентификация |
| 403 | Forbidden | Нет прав доступа |
| 404 | Not Found | Ресурс не найден |
| 409 | Conflict | Конфликт состояния (например, активная сессия уже существует) |
| 500 | Internal Server Error | Ошибка сервера |

---

### Common Errors

**Quiz Not Found (404):**
```json
{
  "error": "Quiz not found",
  "code": "QUIZ_NOT_FOUND"
}
```

**Active Session Exists (409):**
```json
{
  "error": "Active session already exists for this quiz",
  "code": "ACTIVE_SESSION_EXISTS",
  "details": {
    "sessionId": "uuid",
    "currentQuestionIndex": 3
  }
}
```

**Unauthorized (401):**
```json
{
  "error": "Invalid or expired Telegram authentication",
  "code": "UNAUTHORIZED"
}
```

**Question Already Answered (400):**
```json
{
  "error": "Question already answered",
  "code": "QUESTION_ALREADY_ANSWERED"
}
```

---

## Rate Limiting

**Limits:**
- 100 requests / минута per IP
- 1000 requests / час per user

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1642838400  // Unix timestamp
```

**Response 429 (Too Many Requests):**
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retryAfter": 60
}
```

---

## Pagination

Для endpoints возвращающих списки (например, `/quiz`):

**Query Parameters:**
```
?limit=20      // default: 20, max: 100
&offset=0      // default: 0
```

**Response Headers:**
```
X-Total-Count: 156
X-Limit: 20
X-Offset: 0
```

---

## CORS

**Allowed Origins:**
- `https://dev.quiz-sprint-tma.online`
- `https://staging.quiz-sprint-tma.online`
- `https://quiz-sprint-tma.online`

**Allowed Methods:** GET, POST, DELETE, OPTIONS

**Allowed Headers:** Content-Type, Authorization

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 1.0
**Swagger UI:** http://localhost:3000/swagger/index.html
