# Quiz Sprint TMA - Documentation

**Telegram Mini App for interactive quizzes with real-time leaderboards**

---

## 🗺️ Documentation Navigation

### Quick Start

**Для нового разработчика:**
1. Начни с [`ARCHITECTURE.md`](./ARCHITECTURE.md) - понять систему в целом
2. Затем [`UBIQUITOUS_LANGUAGE.md`](./UBIQUITOUS_LANGUAGE.md) - выучить словарь терминов
3. Затем [`current/domain.md`](./current/domain.md) - изучить бизнес-логику

**Для AI code generation:**
- **Bugfix существующей фичи:** Читай `current/domain.md` → `current/api.md`
- **Новая фича:** Читай `future/ROADMAP.md` → `future/{feature}.md`
- **UI/UX вопросы:** Читай `current/user-flows.md`
- **API integration:** Читай `current/api.md`

---

## 📚 Documentation Structure

```
docs/
├── README.md                          ← Ты здесь!
├── ARCHITECTURE.md                    ← System overview, Bounded Contexts
├── UBIQUITOUS_LANGUAGE.md             ← Словарь терминов (quick reference)
│
├── current/                           ← Текущая реализация
│   ├── domain.md                      ← Aggregates, Use Cases, Events
│   ├── api.md                         ← REST & WebSocket endpoints catalog
│   └── user-flows.md                  ← Экраны, UI flows, wireframes
│
└── future/                            ← Планируемые фичи (roadmap)
    ├── ROADMAP.md                     ← Priority matrix, dependencies
    └── {feature}.md                   ← See old DOMAIN.md & USER_FLOW.md
```

---

## 📖 Core Documentation Files

### System Overview

#### [`ARCHITECTURE.md`](./ARCHITECTURE.md)
**Что внутри:**
- Bounded Contexts (Quiz Taking, Quiz Catalog, Leaderboard, Identity, User Stats)
- Context Map (как контексты взаимодействуют)
- Tech Stack (Vue 3, Go, PostgreSQL, Redis, Docker)
- DDD Layer Responsibilities
- Database Schema
- API Structure (Swagger → TypeScript codegen)

**Читай когда:**
- Нужно понять общую архитектуру
- Выбираешь куда добавить новую фичу
- Планируешь изменения в инфраструктуре

---

#### [`UBIQUITOUS_LANGUAGE.md`](./UBIQUITOUS_LANGUAGE.md)
**Что внутри:**
- Словарь терминов (Quiz Session, Score, Streak, etc.)
- Value Objects (QuizID, Points, Timestamp)
- Domain Events каталог
- Scoring Formula (как рассчитываются очки)
- Category vs Tag (различие)

**Читай когда:**
- Нужно понять значение термина
- Пишешь код и нужно использовать правильные названия
- Проверяешь бизнес-правила

---

### Current Implementation

#### [`current/domain.md`](./current/domain.md)
**Что внутри:**
- Core Domain: QuizSession aggregate
- Supporting Domains: Quiz, Category, Tag, Leaderboard, UserStats
- Use Cases с детальной бизнес-логикой
- Domain Events flow
- Repository Interfaces

**Читай когда:**
- Фиксишь баг в существующей фиче
- Добавляешь новый use case
- Нужно понять как работает подсчет очков
- Изменяешь бизнес-правила

**Размер:** ~400 строк (компактно!)

---

#### [`current/api.md`](./current/api.md)
**Что внутри:**
- Все REST endpoints с примерами
- Request/Response форматы
- Authentication (Telegram Mini App)
- WebSocket (Leaderboard real-time)
- Error responses
- Rate limiting

**Читай когда:**
- Интегрируешь frontend с backend
- Нужно узнать формат API запроса
- Добавляешь новый endpoint
- Проверяешь error handling

**Размер:** ~550 строк

**Альтернатива:** Swagger UI http://localhost:3000/swagger/index.html

---

#### [`current/user-flows.md`](./current/user-flows.md)
**Что внутри:**
- User Journey (main flow)
- Экраны приложения (wireframes)
- UI компоненты (reusable)
- Интерактивные механики
- Edge cases & error handling

**Читай когда:**
- Работаешь над UI
- Нужно понять UX flow
- Добавляешь новый экран
- Фиксишь UI bug

**Примечание:** См. также старый `USER_FLOW.md` для детальных wireframes (будет refactored)

---

### Future Plans

#### [`future/ROADMAP.md`](./future/ROADMAP.md)
**Что внутри:**
- Implementation Priority Matrix
- Dependencies между фичами
- 6 фаз развития:
  - Phase 1: 1v1 Asynchronous Duels
  - Phase 2: Badge Collection
  - Phase 3: Power-Ups
  - Phase 4: Weekly Tournaments
  - Phase 5: Category Roulette
  - Phase 6: Random Matchmaking
- Excluded mechanics (why NOT)

**Читай когда:**
- Планируешь новую фичу
- Нужно выбрать приоритет
- Хочешь понять видение продукта

**Примечание:** Детальные спецификации пока в старых DOMAIN.md & USER_FLOW.md

---

## 🔍 Quick Reference

### "Мне нужно..."

**...понять как работает Quiz Session:**
→ [`current/domain.md`](./current/domain.md#aggregate-quizsession)

**...добавить новый API endpoint:**
1. Читай [`current/api.md`](./current/api.md) - понять формат
2. Читай [`current/domain.md`](./current/domain.md) - найти use case
3. Добавь handler в `backend/internal/infrastructure/http/handlers/`
4. Добавь Swagger annotations
5. Запусти `pnpm run generate:all`

**...изменить scoring formula:**
1. Читай [`UBIQUITOUS_LANGUAGE.md`](./UBIQUITOUS_LANGUAGE.md#scoring-formula)
2. Редактируй `internal/domain/quiz/session.go`
3. Обнови `current/domain.md` + `UBIQUITOUS_LANGUAGE.md`

**...создать новый экран в UI:**
1. Читай [`current/user-flows.md`](./current/user-flows.md)
2. Добавь wireframe
3. Используй generated API hooks из `@/api/generated/hooks`

**...посмотреть что планируется в будущем:**
→ [`future/ROADMAP.md`](./future/ROADMAP.md)

---

## 📦 Related Files

**Backend:**
- `backend/IMPORT.md` - Как импортировать квизы
- `backend/data/quizzes/SCHEMA.md` - Формат JSON для квизов
- `backend/internal/domain/` - Domain layer код

**Frontend:**
- `tma/src/api/generated/` - Auto-generated API client
- `tma/src/views/` - Vue components для экранов

**Infrastructure:**
- `CLAUDE.md` - Инструкции для Claude Code (workflow)
- `DOCUMENTATION_WORKFLOW.md` - Когда/как обновлять docs

---

## ✏️ Documentation Workflow

**Когда обновлять документацию:**
1. **ПЕРЕД** реализацией фичи - обнови domain model & user flows
2. **ПОСЛЕ** изменения API - регенери Swagger, обнови `current/api.md`
3. **ПОСЛЕ** рефакторинга - проверь что docs актуальны

**Процесс:**
1. Внеси изменения в соответствующий `.md` файл
2. Коммит документации ВМЕСТЕ с кодом
3. Пример: `git commit -m "Add streak bonus logic" docs/current/domain.md backend/...`

**Не дублируй:**
- Ubiquitous Language → уже в отдельном файле
- API specs → генерируются из Swagger
- Architecture overview → в ARCHITECTURE.md

---

## 🚀 Getting Started (для разработчиков)

### 1. Локальная разработка

```bash
# Backend
cd backend
docker compose -f docker-compose.dev.yml up

# Frontend
cd tma
pnpm install
pnpm dev

# Dev tunnels (для HTTPS через Telegram)
./dev-tunnel/start-backend-tunnel.sh
./dev-tunnel/start-frontend-tunnel.sh
```

**URL:** `https://dev.quiz-sprint-tma.online`

### 2. Генерация API types

```bash
cd tma
pnpm run generate:all   # Swagger → TypeScript hooks
```

### 3. Документация

**Swagger UI:** http://localhost:3000/swagger/index.html

**База данных:**
- Web UI: http://localhost:8080 (Adminer)
- CLI: `docker compose -f docker-compose.dev.yml exec postgres psql -U quiz_user -d quiz_sprint_dev`

---

## 📝 Changelog

**v2.0 (2026-01-21) - Hybrid Documentation Structure**
- ✅ Разделена документация на `current/` и `future/`
- ✅ Создан ARCHITECTURE.md (system overview)
- ✅ Создан UBIQUITOUS_LANGUAGE.md (словарь терминов)
- ✅ current/domain.md - только текущая реализация (~400 строк)
- ✅ current/api.md - каталог всех endpoints (~550 строк)
- ✅ future/ROADMAP.md - priority matrix для будущих фич
- ✅ README.md - navigation hub (ты здесь)

**v1.5 (2026-01-21) - Future Enhancements Added**
- Added 6 phases of future features (Duels, Achievements, Power-Ups, etc.)
- Priority matrix and dependencies

**v1.4 (2026-01-21) - Daily Challenge & User Stats**
- Daily Quiz механика
- Streak tracking
- User Stats domain

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 2.0
**Проект:** Quiz Sprint TMA
