# 🏗️ Архитектура Quiz-приложения на Go

## Обзор

Для приложения с real-time режимами (Quick Duel, Party Mode) и асинхронными (Daily Challenge, Solo Marathon) рекомендуется **гибридная архитектура** на основе Clean Architecture с выделением WebSocket-компонента.

---

## 📐 Высокоуровневая схема

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              КЛИЕНТЫ                                         │
│              iOS App    │    Android App    │    Web App                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API GATEWAY / LOAD BALANCER                        │
│                          (Traefik / Nginx / Kong)                           │
│              Rate Limiting • Auth • SSL • Routing • Load Balancing          │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
        ┌───────────────────────┐       ┌───────────────────────┐
        │     REST API          │       │    WebSocket Hub      │
        │   (Stateless)         │       │    (Stateful)         │
        │                       │       │                       │
        │ • Auth/Registration   │       │ • Quick Duel          │
        │ • User Profile        │       │ • Party Mode          │
        │ • Leaderboards        │       │ • Real-time updates   │
        │ • Daily Challenge     │       │ • Matchmaking         │
        │ • Shop/Payments       │       │                       │
        │ • Solo Marathon       │       │                       │
        └───────────────────────┘       └───────────────────────┘
                    │                               │
                    └───────────────┬───────────────┘
                                    ▼
        ┌─────────────────────────────────────────────────────────┐
        │                    CORE SERVICES                         │
        │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
        │  │    Game     │ │  Question   │ │    User     │        │
        │  │   Service   │ │   Service   │ │   Service   │        │
        │  └─────────────┘ └─────────────┘ └─────────────┘        │
        │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
        │  │ Matchmaking │ │ Leaderboard │ │Notification │        │
        │  │   Service   │ │   Service   │ │   Service   │        │
        │  └─────────────┘ └─────────────┘ └─────────────┘        │
        └─────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
        ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
        │  PostgreSQL   │   │     Redis     │   │   NATS/Kafka  │
        │ (основные     │   │ (кеш, сессии, │   │  (очереди,    │
        │  данные)      │   │  лидерборды,  │   │   события)    │
        │               │   │  matchmaking) │   │               │
        └───────────────┘   └───────────────┘   └───────────────┘
```

---

## 🧱 Архитектурный паттерн: Clean Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│    ┌─────────────────────────────────────────────────────┐     │
│    │                    DOMAIN                            │     │
│    │         (Entities: User, Game, Question)            │     │
│    │              Чистая бизнес-логика                   │     │
│    │              Без зависимостей                       │     │
│    └─────────────────────────────────────────────────────┘     │
│                            ▲                                    │
│    ┌─────────────────────────────────────────────────────┐     │
│    │                   USE CASES                          │     │
│    │    (GameService, MatchmakingService, UserService)   │     │
│    │           Бизнес-правила приложения                 │     │
│    │           Оркестрация domain-сущностей              │     │
│    └─────────────────────────────────────────────────────┘     │
│                            ▲                                    │
│    ┌─────────────────────────────────────────────────────┐     │
│    │                   ADAPTERS                           │     │
│    │  HTTP Handlers │ WS Handlers │ Repositories │ gRPC  │     │
│    │        Преобразование данных между слоями           │     │
│    └─────────────────────────────────────────────────────┘     │
│                            ▲                                    │
│    ┌─────────────────────────────────────────────────────┐     │
│    │                 INFRASTRUCTURE                       │     │
│    │    PostgreSQL │ Redis │ NATS │ Firebase │ Stripe    │     │
│    │              Внешние системы и драйверы             │     │
│    └─────────────────────────────────────────────────────┘     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Почему Clean Architecture:**
- Независимость бизнес-логики от фреймворков и БД
- Легко тестировать (моки на уровне интерфейсов)
- Легко менять инфраструктуру (Redis → Memcached, PostgreSQL → MySQL)
- Понятная структура для команды

---

## 📁 Структура проекта

```
quiz-app/
├── cmd/                      # Точки входа
│   ├── api/main.go          # REST API сервер
│   ├── ws/main.go           # WebSocket сервер
│   └── worker/main.go       # Background jobs
│
├── internal/
│   ├── domain/              # Сущности и бизнес-правила
│   │   ├── user.go
│   │   ├── game.go
│   │   ├── question.go
│   │   └── room.go
│   │
│   ├── usecase/             # Бизнес-логика (сервисы)
│   │   ├── game/
│   │   │   ├── quick_duel.go
│   │   │   ├── daily_challenge.go
│   │   │   ├── solo_marathon.go
│   │   │   └── party_mode.go
│   │   ├── matchmaking/
│   │   ├── leaderboard/
│   │   └── user/
│   │
│   ├── adapter/             # Адаптеры
│   │   ├── http/            # REST handlers
│   │   ├── websocket/       # WS hub и handlers
│   │   └── repository/      # Реализации репозиториев
│   │       ├── postgres/
│   │       └── redis/
│   │
│   └── infrastructure/      # Внешние сервисы
│       ├── database/
│       ├── cache/
│       ├── queue/
│       └── push/
│
├── pkg/                     # Переиспользуемые пакеты
│   ├── logger/
│   ├── jwt/
│   └── elo/
│
├── migrations/              # SQL миграции
├── configs/                 # Конфигурации
└── deployments/             # Docker, K8s
```

---

## 🔌 Компоненты системы

### 1. REST API (Stateless)

**Назначение:** Обработка запросов, не требующих real-time

| Эндпоинт | Описание |
|----------|----------|
| `POST /auth/register` | Регистрация |
| `POST /auth/login` | Авторизация |
| `GET /users/me` | Профиль |
| `GET /leaderboard/daily` | Лидерборд дня |
| `POST /daily-challenge/submit` | Отправка результата Daily Challenge |
| `POST /marathon/start` | Начать марафон |
| `POST /marathon/answer` | Ответ в марафоне |

**Технологии:**
- `chi` или `gin` — HTTP router
- `sqlx` — работа с PostgreSQL
- JWT — аутентификация

---

### 2. WebSocket Hub (Stateful)

**Назначение:** Real-time режимы (Quick Duel, Party Mode)

```
                    ┌─────────────────────┐
                    │    WebSocket Hub    │
                    │                     │
                    │  ┌───────────────┐  │
   Client A ◄──────►│  │   Clients     │  │◄──────► Client B
                    │  │   Registry    │  │
                    │  └───────────────┘  │
                    │                     │
                    │  ┌───────────────┐  │
                    │  │    Rooms      │  │  (Party Mode)
                    │  │   Manager     │  │
                    │  └───────────────┘  │
                    │                     │
                    │  ┌───────────────┐  │
                    │  │    Games      │  │  (Quick Duel)
                    │  │   Sessions    │  │
                    │  └───────────────┘  │
                    │                     │
                    └─────────────────────┘
```

**WebSocket события:**

| Событие (Client → Server) | Описание |
|---------------------------|----------|
| `find_match` | Поиск соперника (Quick Duel) |
| `cancel_match` | Отмена поиска |
| `submit_answer` | Отправка ответа |
| `create_room` | Создать комнату (Party) |
| `join_room` | Войти в комнату |
| `player_ready` | Готовность к игре |
| `start_game` | Начать игру (хост) |

| Событие (Server → Client) | Описание |
|---------------------------|----------|
| `match_found` | Соперник найден |
| `game_started` | Игра началась |
| `new_question` | Новый вопрос |
| `player_answered` | Игрок ответил (без ответа) |
| `question_result` | Результат вопроса |
| `game_finished` | Игра завершена |
| `player_joined` | Игрок вошёл в комнату |
| `player_left` | Игрок вышел |

**Технологии:**
- `gorilla/websocket` — WebSocket
- `sync.Map` или mutex — управление состоянием

---

### 3. Matchmaking Service

**Назначение:** Поиск соперников по ELO-рейтингу

```
┌─────────────────────────────────────────────────────────────┐
│                    MATCHMAKING FLOW                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   Player A (ELO: 1200)                                      │
│         │                                                   │
│         ▼                                                   │
│   ┌─────────────────┐                                       │
│   │  Redis Sorted   │  Score = ELO                          │
│   │     Set         │                                       │
│   │                 │  ZRANGEBYSCORE queue 1150 1250        │
│   │  [1180] User_X  │  ──────────────────────────────────►  │
│   │  [1195] User_Y  │                                       │
│   │  [1200] User_A  │◄── Найден User_Y (ELO: 1195)          │
│   │  [1220] User_Z  │                                       │
│   └─────────────────┘                                       │
│                                                             │
│   Расширение диапазона:                                     │
│   • 0-5 сек:  ±50 ELO                                       │
│   • 5-10 сек: ±100 ELO                                      │
│   • 10-15 сек: ±200 ELO                                     │
│   • 15+ сек: любой игрок или бот                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### 4. Game Service

**Назначение:** Логика всех игровых режимов

| Режим | Тип | Особенности |
|-------|-----|-------------|
| **Quick Duel** | Real-time (WS) | 2 игрока, синхронные вопросы, ELO |
| **Daily Challenge** | Async (REST) | Один набор вопросов для всех, лидерборд |
| **Solo Marathon** | Async (REST) | Бесконечный режим, система жизней |
| **Party Mode** | Real-time (WS) | 2-8 игроков, комнаты, хост |

---

### 5. Leaderboard Service

**Назначение:** Рейтинги и лидерборды

**Хранение в Redis (Sorted Sets):**

```
leaderboard:daily:2026-01-24     → Score = очки дня
leaderboard:marathon:alltime    → Score = рекорд
leaderboard:elo                 → Score = ELO рейтинг
leaderboard:weekly:2026-W04     → Score = очки недели
```

**Операции:**
- `ZADD` — добавить/обновить результат
- `ZREVRANK` — получить позицию игрока
- `ZREVRANGE` — получить топ N игроков
- `ZCARD` — общее количество игроков

---

## 🗄️ Хранение данных

### PostgreSQL (основные данные)

| Таблица | Назначение |
|---------|------------|
| `users` | Профили, ELO, монеты, подписка |
| `games` | История игр (JSONB для игроков и ответов) |
| `questions` | База вопросов |
| `daily_results` | Результаты Daily Challenge |
| `user_stats` | Детальная статистика |
| `achievements` | Достижения пользователей |
| `transactions` | История покупок |

### Redis (кеш и real-time)

| Ключ | Назначение | TTL |
|------|------------|-----|
| `session:{user_id}` | Сессия пользователя | 24h |
| `matchmaking:queue` | Очередь матчмейкинга | — |
| `room:{code}` | Данные комнаты Party Mode | 1h |
| `game:{id}` | Активная игра | 30min |
| `daily:{date}` | Вопросы дня (кеш) | 24h |
| `leaderboard:*` | Лидерборды | — |
| `user:{id}:lives` | Жизни игрока | — |

### NATS/Kafka (очереди событий)

| Топик | Назначение |
|-------|------------|
| `game.created` | Игра создана |
| `game.finished` | Игра завершена |
| `user.levelup` | Повышение уровня |
| `push.send` | Отправка push-уведомлений |
| `analytics.event` | События для аналитики |

---

## 🔄 Масштабирование

### Горизонтальное масштабирование

```
                         Load Balancer
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐      ┌────────┐      ┌────────┐
         │ API #1 │      │ API #2 │      │ API #3 │
         └────────┘      └────────┘      └────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                    ┌─────────────────┐
                    │  PostgreSQL     │
                    │  (Primary)      │
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    ▼                 ▼
              ┌──────────┐      ┌──────────┐
              │ Replica 1│      │ Replica 2│
              │ (Read)   │      │ (Read)   │
              └──────────┘      └──────────┘
```

### WebSocket масштабирование

```
                         Load Balancer
                        (Sticky Sessions)
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐      ┌────────┐      ┌────────┐
         │ WS #1  │      │ WS #2  │      │ WS #3  │
         └────────┘      └────────┘      └────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                    ┌─────────────────┐
                    │   Redis PubSub  │  ← Синхронизация
                    │                 │    между нодами
                    └─────────────────┘

Когда игрок A на WS#1 отвечает, а игрок B на WS#2:
1. WS#1 публикует событие в Redis PubSub
2. WS#2 получает событие и отправляет игроку B
```

---

## 🛡️ Безопасность

| Аспект | Решение |
|--------|---------|
| Аутентификация | JWT токены (access + refresh) |
| Защита WS | Проверка JWT при handshake |
| Rate Limiting | Redis + sliding window |
| Античит | Валидация времени ответа на сервере |
| DDOS | Cloudflare / API Gateway |
| Данные | HTTPS, шифрование sensitive данных |

---

## 📊 Мониторинг

| Инструмент | Назначение |
|------------|------------|
| **Prometheus** | Метрики (RPS, latency, errors) |
| **Grafana** | Визуализация метрик |
| **Jaeger** | Distributed tracing |
| **ELK Stack** | Логи |
| **Sentry** | Error tracking |

**Ключевые метрики:**
- Время поиска матча (p50, p95, p99)
- Latency WebSocket сообщений
- Количество активных игр/комнат
- Completion rate игр
- Ошибки по типам

---

## 🚀 Деплой

### Рекомендуемый стек

| Компонент | Технология |
|-----------|------------|
| Оркестрация | Kubernetes |
| CI/CD | GitHub Actions / GitLab CI |
| Registry | Docker Hub / GCR / ECR |
| База данных | Managed PostgreSQL (RDS, Cloud SQL) |
| Redis | Managed Redis (ElastiCache, Memorystore) |
| CDN | Cloudflare |

### Минимальная инфраструктура (старт)

```
┌─────────────────────────────────────────┐
│            Docker Compose               │
├─────────────────────────────────────────┤
│  • api (2 replicas)                     │
│  • ws-hub (2 replicas)                  │
│  • worker (1 replica)                   │
│  • postgresql                           │
│  • redis                                │
│  • nginx (reverse proxy)                │
└─────────────────────────────────────────┘
```

---

## 📚 Рекомендуемые библиотеки Go

| Категория | Библиотека | Назначение |
|-----------|------------|------------|
| HTTP Router | `chi` или `gin` | REST API |
| WebSocket | `gorilla/websocket` | Real-time |
| Database | `sqlx` + `pgx` | PostgreSQL |
| Redis | `go-redis/redis` | Кеш, очереди |
| Config | `viper` | Конфигурация |
| Logger | `zap` или `zerolog` | Логирование |
| Validation | `validator` | Валидация |
| JWT | `golang-jwt/jwt` | Токены |
| Migration | `golang-migrate` | Миграции БД |
| Testing | `testify` | Тесты |

---

## ✅ Итого: почему эта архитектура

| Требование | Как решается |
|------------|--------------|
| Real-time PvP | WebSocket Hub с Redis PubSub для масштабирования |
| Низкая задержка | Redis для горячих данных, WS для мгновенной доставки |
| Масштабируемость | Stateless API, sticky sessions для WS, реплики БД |
| Надёжность | Graceful degradation (бот вместо игрока), reconnect logic |
| Простота разработки | Clean Architecture, чёткое разделение слоёв |
| Тестируемость | Интерфейсы на границах слоёв, DI |
