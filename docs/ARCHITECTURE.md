# System Architecture - Quiz Sprint TMA

## 📋 Содержание

1. [Описание системы](#описание-системы)
2. [Bounded Contexts](#bounded-contexts)
3. [Context Map](#context-map)
4. [Tech Stack](#tech-stack)
5. [Dependency Diagram](#dependency-diagram)

---

## Описание системы

**Quiz Sprint TMA** - это Telegram Mini Application для прохождения интерактивных викторин в режиме реального времени.

### Бизнес-цели:
- Предоставить пользователям увлекательный опыт прохождения викторин
- Создать соревновательную среду через таблицу лидеров
- Мотивировать пользователей проходить квизы быстро и точно
- Интеграция с Telegram для легкого доступа и социального взаимодействия

### Ключевые характеристики домена:
- **Ограничение по времени**: Каждый квиз имеет временной лимит
- **Мгновенная обратная связь**: Пользователь сразу узнает правильность ответа
- **Подсчет очков**: Баллы начисляются за правильные ответы с учетом скорости
- **Соревнование**: Результаты сравниваются в реальном времени
- **Неизменяемость**: Ответы нельзя изменить после отправки

---

## Bounded Contexts

### 1. Quiz Taking Context (Core Domain) 🎯

**Ответственность:**
- Процесс прохождения квизов
- Управление игровыми сессиями
- Отслеживание ответов пользователя
- Подсчет очков и времени

**Ubiquitous Language:**
- Quiz Session (Игровая сессия)
- User Answer (Ответ пользователя)
- Score (Очки)
- Time Limit (Временной лимит)

**Почему Core Domain?**
Это сердце бизнес-логики. Именно здесь происходит основное взаимодействие пользователя с системой.

---

### 2. Quiz Catalog Context (Supporting) 📚

**Ответственность:**
- Хранение и управление контентом квизов
- Управление вопросами и ответами
- **Категоризация квизов (одна категория на квиз - для навигации)**
- **Теговая система (множественные теги на квиз - для фильтрации)**
- Публикация квизов
- Импорт квизов из внешних источников (LLM, JSON)

**Ubiquitous Language:**
- Quiz (Квиз)
- Question (Вопрос)
- Answer (Вариант ответа)
- Category (Категория) - **одна на квиз, основная классификация**
- Tag (Тег) - **много на квиз, дополнительные метки**
- Quiz Metadata (Метаданные квиза)
- Compact Format (Компактный формат) - оптимизированный формат для LLM
- Batch Import (Пакетный импорт) - импорт нескольких квизов одновременно

**Почему Supporting?**
Необходим для работы Core Domain, но не является уникальным конкурентным преимуществом.

**Hybrid Approach: Category + Tags:**
- **Category** = главная полка в библиотеке (Programming, History, Movies)
  - Одна категория на квиз
  - Используется для основной навигации в UI (CategoriesView → QuizListView)
  - Обязательное поле
- **Tags** = ярлыки на книге (language:go, difficulty:easy, topic:concurrency)
  - Много тегов на квиз (0-10)
  - Используются для фильтрации и поиска
  - Опциональное поле
  - Формат: `{category}:{value}` (например: `language:go`, `difficulty:medium`)

---

### 3. Leaderboard Context (Supporting) 🏆

**Ответственность:**
- Отображение рейтинга игроков
- Вычисление позиций в таблице
- Real-time обновления результатов
- Хранение исторических данных

**Ubiquitous Language:**
- Leaderboard (Таблица лидеров)
- Rank (Позиция/Ранг)
- Leaderboard Entry (Запись в таблице)

**Особенность:**
Использует CQRS pattern - это Read Model, обновляется через Domain Events.

---

### 4. Identity Context (Generic) 👤

**Ответственность:**
- Управление пользователями
- Авторизация через Telegram
- Профили пользователей

**Ubiquitous Language:**
- User (Пользователь)
- Telegram User (Telegram пользователь)
- User Profile (Профиль пользователя)

**Почему Generic?**
Типовая функциональность, не специфичная для квиз-приложения.

---

### 5. User Stats Context (Supporting) 📊

**Ответственность:**
- Отслеживание прогресса пользователя
- Streak tracking (серии ежедневных активностей)
- Статистика по Daily Quiz
- Мотивация пользователей через достижения

**Ubiquitous Language:**
- Current Streak (Текущая серия)
- Longest Streak (Лучшая серия)
- Last Daily Quiz Date (Дата последнего Daily Quiz)
- Total Quizzes Completed (Всего завершено квизов)

---

## Context Map

```
┌─────────────────────────────────────────────────────────────┐
│                    QUIZ SPRINT SYSTEM                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────┐
│   Identity Context      │
│   (Generic Subdomain)   │
│                         │
│ • User                  │
│ • TelegramAuth          │
│ • UserProfile           │
└────────────┬────────────┘
             │ ACL (Anti-Corruption Layer)
             │ Exposes: UserID, Username
             │
    ┌────────┴──────────────────────────────────┐
    │                                            │
    ▼                                            ▼
┌─────────────────────────┐        ┌─────────────────────────┐
│  Quiz Catalog Context   │        │  Quiz Taking Context    │
│  (Supporting Subdomain) │◄───────│  (Core Domain) 🎯       │
│                         │ Uses   │                         │
│ • Quiz                  │        │ • QuizSession           │
│ • Question              │        │ • UserAnswer            │
│ • Answer                │        │ • SessionProgress       │
│ • Category              │        │                         │
│ • Tag                   │        │                         │
└─────────────────────────┘        └──────────┬──────────────┘
                                              │
                                              │ Domain Events:
                                              │ • QuizStarted
                                              │ • AnswerSubmitted
                                              │ • QuizCompleted
                                              │
                      ┌───────────────────────┴───────────────────┐
                      │                                           │
                      ▼                                           ▼
           ┌─────────────────────────┐              ┌─────────────────────────┐
           │ Leaderboard Context     │              │ User Stats Context      │
           │ (Supporting Subdomain)  │              │ (Supporting Subdomain)  │
           │                         │              │                         │
           │ • LeaderboardEntry      │              │ • UserStats             │
           │ • Ranking               │              │ • StreakTracking        │
           │ • EventHandlers         │              │ • EventHandlers         │
           └─────────────────────────┘              └─────────────────────────┘
```

### Типы взаимодействий:

1. **Shared Kernel**: Quiz Catalog ↔ Quiz Taking
   - Делят QuizID, QuestionID
   - Quiz Taking читает Quiz (read-only)

2. **Published Language**: Identity → All
   - UserID - общий идентификатор
   - Username - для отображения

3. **Event-Driven**: Quiz Taking → Leaderboard, User Stats
   - Асинхронное обновление через Domain Events
   - Eventual consistency

4. **ACL (Anti-Corruption Layer)**: Quiz Taking → Identity
   - Защита от изменений в Identity Context
   - Минимальная зависимость (только UserID)

---

## Tech Stack

### Frontend
- **Framework**: Vue 3.5 (Composition API)
- **Language**: TypeScript 5.9 (strict mode)
- **Build Tool**: Vite 6
- **Router**: Vue Router 4
- **State**: Vue Query (TanStack Query)
- **Telegram**: @telegram-apps/sdk
- **Testing**: Vitest, Playwright
- **Validation**: Zod (from generated schemas)

### Backend
- **Language**: Go 1.25
- **Web Framework**: Fiber v3
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **API Docs**: Swagger (swaggo/swag)
- **Hot Reload**: Air (development)

### Infrastructure
- **Containerization**: Docker, Docker Compose
- **Web Server**: nginx (reverse proxy)
- **SSL**: Let's Encrypt
- **CI/CD**: GitHub Actions
- **VPS**: Ubuntu 22.04 LTS

### Architecture Pattern
- **Backend**: Domain-Driven Design (DDD)
  - Layered Architecture: Domain → Application → Infrastructure
  - CQRS for Leaderboard (Read Model)
  - Event-Driven for cross-context communication
- **Frontend**: Feature-Sliced Design
  - Auto-generated API client from Swagger
  - Runtime hostname detection for multi-environment

---

## Dependency Diagram

```
┌────────────────────────────────────────────────────────┐
│                   Application Layer                     │
│  (Use Cases - orchestration only)                      │
├────────────────────────────────────────────────────────┤
│ • StartQuizUseCase                                     │
│ • SubmitAnswerUseCase                                  │
│ • GetLeaderboardUseCase                                │
│ • GetDailyQuizUseCase                                  │
│ • UpdateUserStatsUseCase                               │
└────────────────┬───────────────────────────────────────┘
                 │ depends on
                 ▼
┌────────────────────────────────────────────────────────┐
│                    Domain Layer                         │
│  (Business logic, rules, invariants)                   │
├────────────────────────────────────────────────────────┤
│ Aggregates:                                            │
│ • Quiz (with Questions, Answers)                       │
│ • QuizSession (with UserAnswers)                       │
│ • Category                                             │
│ • Tag                                                  │
│ • UserStats                                            │
│                                                         │
│ Value Objects:                                         │
│ • IDs, Points, TimeLimit, Streak, etc                  │
│                                                         │
│ Domain Events:                                         │
│ • QuizStarted, AnswerSubmitted, QuizCompleted         │
│                                                         │
│ Interfaces (defined, not implemented):                │
│ • QuizRepository                                       │
│ • SessionRepository                                    │
│ • UserStatsRepository                                  │
│ • EventBus                                             │
└────────────────┬───────────────────────────────────────┘
                 │ implemented by
                 ▼
┌────────────────────────────────────────────────────────┐
│                Infrastructure Layer                     │
│  (HTTP, Database, WebSocket, External services)        │
├────────────────────────────────────────────────────────┤
│ • Fiber HTTP Handlers                                  │
│ • PostgreSQL Repository Implementations                │
│ • Redis Cache                                          │
│ • WebSocket Hub (Leaderboard real-time)                │
│ • In-Memory Event Bus                                  │
│ • Swagger Documentation                                │
└────────────────────────────────────────────────────────┘
```

---

## DDD Layer Responsibilities

### Domain Layer (`internal/domain/`)
**Pure business logic - NO external dependencies:**
- ✅ Use: Value Objects, Factory methods (`NewQuiz`), `ReconstructEntity()` for DB loading
- ❌ NO: `context.Context`, JSON tags, database imports, `time.Time` (use `int64` Unix timestamps)

### Application Layer (`internal/application/`)
**Use Cases:**
- ✅ Use: Input/Output DTOs, `context.Context`, orchestration
- ❌ NO: Business logic (delegate to domain), HTTP concerns

### Infrastructure Layer (`internal/infrastructure/`)
**Technical implementations:**
- ✅ Use: HTTP handlers (thin adapters), Repository implementations, DB/SQL
- ❌ NO: Business logic

**Error Mapping**: Each handler has domain-specific error mapper (e.g., `quiz.ErrQuizNotFound` → HTTP 404)

---

## Database Schema

**Tables** (PostgreSQL):
- `users` - User profiles (Telegram auth)
- `quizzes`, `questions`, `answers` - Quiz data
- `quiz_sessions` - User attempts
- `categories` - Quiz categories
- `quiz_tags` - Many-to-many tags
- `leaderboard_entries` - Read model for rankings
- `user_stats` - Streaks, totals

---

## API Structure

**Code-first approach:**
```
Go Handlers (@annotations) → swag → swagger.json → kubb → TypeScript types + Vue Query hooks
```

**Workflow:**
1. Update Go handler annotations in `backend/internal/infrastructure/http/handlers/`
2. Define DTOs in `swagger_models.go` (use concrete types, never `map[string]interface{}`)
3. Run `pnpm run generate:all` from `tma/` (generates Swagger + TypeScript)
4. Use generated hooks: `import { useGetQuizId } from '@/api/generated/hooks/quizController'`

**Endpoints:**
- **Quiz**: `GET /api/v1/quiz`, `GET /api/v1/quiz/:id`, `POST /api/v1/quiz/:id/start`
- **Session**: `POST /api/v1/quiz/session/:sessionId/answer`, `DELETE /api/v1/quiz/session/:sessionId`
- **User**: `POST /api/v1/user/register`, `GET /api/v1/user/:id`
- **Categories**: `GET /api/v1/categories`, `POST /api/v1/categories`
- **Leaderboard**: `GET /api/v1/quiz/:id/leaderboard`
- **Daily Quiz**: `GET /api/v1/quiz/daily`
- **WebSocket**: `wss://<domain>/ws/leaderboard/:id`

---

## Environments

| Environment | URL | API Port | Database |
|-------------|-----|----------|----------|
| Development | `dev.quiz-sprint-tma.online` | 3000 (local) | PostgreSQL (Docker) |
| Staging | `staging.quiz-sprint-tma.online` | 3001 (Docker) | PostgreSQL (Docker) |
| Production | `quiz-sprint-tma.online` | 3000 (Docker) | PostgreSQL (Docker) |

**API Endpoints**: `https://<domain>/api/v1/*`, WebSocket: `wss://<domain>/ws/leaderboard/:id`

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 1.0
**Проект:** Quiz Sprint TMA
