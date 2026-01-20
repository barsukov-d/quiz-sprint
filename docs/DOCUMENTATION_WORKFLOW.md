# Documentation Workflow - Quiz Sprint TMA

## 📋 Когда и как обновлять документацию

Этот документ описывает правильный порядок обновления документации при внесении изменений в проект.

---

## Типы изменений и порядок документирования

### 1️⃣ **Backend-First Features** (новая бизнес-логика)

**Когда:** Добавляется новая доменная логика, агрегаты, use cases

**Порядок:**
```
1. DOMAIN.md → 2. API (backend) → 3. USER_FLOW.md → 4. Frontend
```

**Примеры:**
- Добавление категорий квизов
- Система достижений (achievements)
- Multiplayer режим
- Комментарии к квизам
- Система рейтинга (stars/likes)

**Workflow:**

```
┌─────────────────────────────────────────────────┐
│ STEP 1: Обновить DOMAIN.md                     │
├─────────────────────────────────────────────────┤
│ • Определить новые Aggregates/Entities          │
│ • Описать Ubiquitous Language                   │
│ • Определить бизнес-правила (Invariants)        │
│ • Описать Domain Events                         │
│ • Создать Use Cases                             │
│ • Определить Repository interfaces              │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 2: Реализовать Backend                    │
├─────────────────────────────────────────────────┤
│ • Создать Domain layer (entities, VOs)         │
│ • Реализовать Use Cases                         │
│ • Создать Repository implementations            │
│ • Добавить HTTP Handlers + Swagger              │
│ • Написать тесты                                │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 3: Обновить USER_FLOW.md                  │
├─────────────────────────────────────────────────┤
│ • Добавить новые экраны (wireframes)            │
│ • Описать User Journey с новой функцией         │
│ • Определить UI компоненты                      │
│ • Описать интерактивные механики                │
│ • Добавить Edge Cases                           │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 4: Реализовать Frontend                   │
├─────────────────────────────────────────────────┤
│ • Сгенерировать TypeScript types (Swagger)      │
│ • Создать UI компоненты                         │
│ • Реализовать экраны                            │
│ • Интегрировать с API                           │
│ • Написать тесты                                │
└─────────────────────────────────────────────────┘
```

**Пример: Добавление категорий квизов**

```diff
# 1. DOMAIN.md
+ ## Category Aggregate
+
+ Value Object: CategoryID, CategoryName, CategorySlug
+
+ Invariants:
+ - Уникальное название категории
+ - Slug автогенерируется из названия
+ - Категория может содержать 0+ квизов
+
+ Use Cases:
+ - ListCategoriesUseCase() → (categories[])
+ - GetQuizzesByCategoryUseCase(categoryID) → (quizzes[])

# 2. Backend Implementation
# internal/domain/category/category.go
# internal/application/category/list_categories.go
# internal/infrastructure/http/handlers/category_handler.go

# 3. USER_FLOW.md
+ ## Экран: Categories
+
+ [Wireframe категорий]
+
+ Navigation: Home → Categories → Quiz List (filtered)

# 4. Frontend
# tma/src/views/CategoriesView.vue
# tma/src/components/CategoryCard.vue
```

---

### 2️⃣ **UX-First Features** (улучшения интерфейса)

**Когда:** Изменения в UI/UX без новой бизнес-логики

**Порядок:**
```
1. USER_FLOW.md → 2. Frontend → 3. DOMAIN.md (если нужно)
```

**Примеры:**
- Изменение дизайна экрана Results
- Добавление анимаций
- Улучшение навигации
- Добавление фильтров/сортировки (frontend-only)
- Dark mode toggle

**Workflow:**

```
┌─────────────────────────────────────────────────┐
│ STEP 1: Обновить USER_FLOW.md                  │
├─────────────────────────────────────────────────┤
│ • Добавить/изменить wireframes                  │
│ • Обновить UI компоненты                        │
│ • Описать новые интерактивные механики          │
│ • Обновить animations/transitions               │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 2: Реализовать Frontend                   │
├─────────────────────────────────────────────────┤
│ • Создать/изменить компоненты                   │
│ • Добавить анимации                             │
│ • Обновить стили                                │
│ • Написать тесты                                │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 3: DOMAIN.md (если нужно)                 │
├─────────────────────────────────────────────────┤
│ • Обновить только если изменилась доменная      │
│   логика (например, добавили новый фильтр,      │
│   который требует новый Use Case)               │
└─────────────────────────────────────────────────┘
```

**Пример: Добавление фильтра по сложности**

```diff
# 1. USER_FLOW.md
+ ## Фильтры на главной странице
+
+ [Wireframe с фильтрами]
+
+ Dropdown: All | Easy | Medium | Hard

# 2. Frontend
# tma/src/views/HomeView.vue (добавить фильтр)
# tma/src/composables/useQuizFilters.ts

# 3. DOMAIN.md (опционально, если сложность уже есть в Quiz)
# Если сложность - новое поле, добавить в Quiz Value Object
```

---

### 3️⃣ **Full-Stack Features** (одновременно domain + UX)

**Когда:** Крупные фичи, затрагивающие и backend, и frontend

**Порядок:**
```
1. DOMAIN.md + USER_FLOW.md (параллельно) → 2. Backend → 3. Frontend
```

**Примеры:**
- Система друзей/подписок
- Multiplayer квизы
- Чат/комментарии
- Уведомления

**Workflow:**

```
┌─────────────────────────────────────────────────┐
│ STEP 1a: DOMAIN.md                             │
│ (Domain design)                                 │
│                                                 │
│ • Aggregates                                    │
│ • Use Cases                                     │
│ • Domain Events                                 │
└─────────────────────────────────────────────────┘
                    ↓
                   AND  (parallel)
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 1b: USER_FLOW.md                          │
│ (UX design)                                     │
│                                                 │
│ • Wireframes                                    │
│ • User Journey                                  │
│ • Interactions                                  │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 2: Review & Align                         │
├─────────────────────────────────────────────────┤
│ • Проверить, что domain поддерживает UX         │
│ • Проверить, что UX отражает domain правила     │
│ • Согласовать API contract                      │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ STEP 3: Implementation                          │
├─────────────────────────────────────────────────┤
│ Backend → Frontend (или параллельно)            │
└─────────────────────────────────────────────────┘
```

---

## Правила поддержания документации в актуальном состоянии

### ✅ **DO's (делать)**

1. **Обновлять документацию ДО написания кода**
   ```
   Documentation-First → Code → Tests
   ```

2. **Коммитить документацию вместе с кодом**
   ```bash
   git add docs/DOMAIN.md backend/internal/domain/category/
   git commit -m "feat: Add Category aggregate"
   ```

3. **Ссылаться на документацию в PR**
   ```markdown
   ## Changes
   - Added Category feature (see `docs/DOMAIN.md` line 500)
   - Updated Quiz List UI (see `docs/USER_FLOW.md` line 150)
   ```

4. **Использовать версионирование для крупных изменений**
   ```markdown
   <!-- В начале файла -->
   **Last Updated:** 2026-01-18
   **Version:** 1.1
   **Changelog:**
   - v1.1 (2026-01-18): Added Category aggregate
   - v1.0 (2026-01-15): Initial version
   ```

5. **Отмечать TODO и устаревшие секции**
   ```markdown
   ## Quiz Recommendations

   > ⚠️ **TODO:** Design needed (scheduled for v2.0)

   ## Old Leaderboard Design

   > ⚠️ **DEPRECATED:** Replaced by real-time WebSocket version (see v1.2)
   ```

### ❌ **DON'Ts (не делать)**

1. **НЕ писать код без документации**
   - Если фича не задокументирована, значит она не продумана

2. **НЕ обновлять документацию "потом"**
   - "Потом" = никогда
   - Документация стареет быстрее молока

3. **НЕ дублировать информацию**
   - Используйте ссылки между документами
   - Если что-то уже описано в DOMAIN.md, не копируйте в USER_FLOW.md

4. **НЕ делать документы слишком длинными**
   - Разбивайте на логические секции
   - Используйте table of contents
   - Создавайте отдельные файлы для крупных тем

---

## Связи между документами

```
┌──────────────────────────────────────────────────────┐
│                  CLAUDE.md                           │
│  (Project overview, commands, tech stack)            │
│                                                      │
│  • References: все остальные docs                    │
│  • Update when: tech stack changes, new commands     │
└──────────────────────────────────────────────────────┘
                         ↓
        ┌────────────────┴────────────────┐
        ↓                                  ↓
┌──────────────────┐              ┌──────────────────┐
│   DOMAIN.md      │              │  USER_FLOW.md    │
│                  │              │                  │
│  Domain model    │◄────────────►│  UX/UI spec      │
│  DDD patterns    │   Aligned    │  Wireframes      │
│  Use Cases       │              │  User Journey    │
│                  │              │                  │
│  Update when:    │              │  Update when:    │
│  • New domain    │              │  • New screens   │
│  • Business      │              │  • UX changes    │
│    rules change  │              │  • New flows     │
└──────────────────┘              └──────────────────┘
        ↓                                  ↓
        └────────────────┬─────────────────┘
                         ↓
        ┌────────────────────────────────┐
        │    API_CONTRACT.md             │
        │    (optional)                  │
        │                                │
        │  • Swagger spec summary        │
        │  • Request/Response examples   │
        │  • Error codes                 │
        └────────────────────────────────┘
                         ↓
        ┌────────────────────────────────┐
        │    DEPLOYMENT.md               │
        │    (infrastructure)            │
        │                                │
        │  • Environments                │
        │  • CI/CD pipelines             │
        │  • Database migrations         │
        └────────────────────────────────┘
```

### Перекрестные ссылки

**В DOMAIN.md:**
```markdown
## Quiz Taking Use Case

Полный User Journey см. в [USER_FLOW.md](./USER_FLOW.md#3-прохождение-quiz)

API endpoints см. в [CLAUDE.md](../CLAUDE.md#backend-api-structure)
```

**В USER_FLOW.md:**
```markdown
## Экран: Quiz Details

Доменная модель Quiz описана в [DOMAIN.md](./DOMAIN.md#quiz-aggregate)

Backend endpoint: `GET /api/v1/quiz/:id` (см. [CLAUDE.md](../CLAUDE.md))
```

---

## Checklist: Перед началом новой фичи

### Backend-First Feature

- [ ] 1. Открыть `DOMAIN.md`
- [ ] 2. Описать новый Aggregate / Value Object
- [ ] 3. Определить Invariants (бизнес-правила)
- [ ] 4. Описать Use Cases
- [ ] 5. Определить Domain Events (если нужно)
- [ ] 6. Описать Repository interface
- [ ] 7. Реализовать backend (domain → application → infrastructure)
- [ ] 8. Обновить `USER_FLOW.md` (wireframes + journey)
- [ ] 9. Реализовать frontend
- [ ] 10. Обновить changelog в обоих документах

### UX-First Feature

- [ ] 1. Открыть `USER_FLOW.md`
- [ ] 2. Создать wireframe нового экрана
- [ ] 3. Описать User Journey
- [ ] 4. Определить UI компоненты
- [ ] 5. Описать интерактивные механики
- [ ] 6. Добавить Edge Cases
- [ ] 7. Реализовать frontend
- [ ] 8. Проверить: нужны ли изменения в `DOMAIN.md`?
- [ ] 9. Если да → обновить domain документацию
- [ ] 10. Обновить changelog

### Full-Stack Feature

- [ ] 1. Создать branch: `feature/название-фичи`
- [ ] 2. Открыть `DOMAIN.md` + `USER_FLOW.md` одновременно
- [ ] 3. Описать domain model (DOMAIN.md)
- [ ] 4. Описать UX flow (USER_FLOW.md)
- [ ] 5. Проверить alignment (domain поддерживает UX?)
- [ ] 6. Согласовать API contract (request/response DTOs)
- [ ] 7. Реализовать backend
- [ ] 8. Реализовать frontend
- [ ] 9. Написать тесты (domain + E2E)
- [ ] 10. Обновить changelog в обоих документах
- [ ] 11. Создать PR с ссылками на документацию

---

## Примеры: Типовые изменения

### Пример 1: Добавление Quiz Categories

**Тип:** Backend-First

**Шаги:**

1. **DOMAIN.md** (15 минут)
   ```markdown
   ## Category Aggregate

   ### Value Objects
   - CategoryID (UUID)
   - CategoryName (string, 1-50 chars)
   - CategorySlug (string, lowercase, hyphenated)

   ### Invariants
   - Уникальное имя категории (case-insensitive)
   - Slug автогенерируется: "Tech Trivia" → "tech-trivia"
   - Категория может быть пустой (0 квизов)

   ### Use Cases
   - ListCategoriesUseCase() → []Category
   - GetCategoryUseCase(id) → Category
   - CreateCategoryUseCase(name) → CategoryID
   - GetQuizzesByCategoryUseCase(categoryID) → []Quiz

   ### Repository
   interface CategoryRepository {
     FindAll() ([]Category, error)
     FindByID(id CategoryID) (*Category, error)
     FindBySlug(slug string) (*Category, error)
     Save(category *Category) error
   }
   ```

2. **Backend Implementation** (2 hours)
   - `internal/domain/category/` - entities, VOs
   - `internal/application/category/` - use cases
   - `internal/infrastructure/persistence/postgres/category_repository.go`
   - `internal/infrastructure/http/handlers/category_handler.go`
   - Migration: `003_create_categories_table.sql`

3. **USER_FLOW.md** (30 минут)
   ```markdown
   ## Экран: Categories (новый)

   [ASCII wireframe]

   Navigation:
   Home → Categories → Quiz List (filtered by category)

   ## Изменения в Quiz List

   - Добавить filter dropdown "All Categories"
   - Показывать category badge на Quiz Card
   ```

4. **Frontend** (3 hours)
   - Regenerate API: `pnpm run generate:api`
   - `tma/src/views/CategoriesView.vue`
   - `tma/src/components/CategoryFilter.vue`
   - Update `QuizCard.vue` (show category badge)

---

### Пример 2: Улучшение Results Screen

**Тип:** UX-First

**Шаги:**

1. **USER_FLOW.md** (20 минут)
   ```markdown
   ## 4. Результаты (обновлено)

   [Новый wireframe с графиком прогресса]

   Новые элементы:
   - График правильных/неправильных ответов
   - Breakdown по темам (если есть categories)
   - Comparison с вашим предыдущим результатом
   - Social share button
   ```

2. **Frontend** (4 hours)
   - `tma/src/views/ResultsView.vue` - редизайн
   - `tma/src/components/ProgressChart.vue` - новый компонент
   - `tma/src/composables/useShareResult.ts` - share функционал
   - Add CSS animations

3. **DOMAIN.md** (если нужно)
   - Если breakdown по темам требует новых данных → обновить Session aggregate
   - Если нет → DOMAIN.md не трогаем

---

### Пример 3: Система Achievements

**Тип:** Full-Stack

**Шаги:**

1. **DOMAIN.md** (30 минут)
   ```markdown
   ## Achievement Context (новый Bounded Context)

   ### Aggregate: UserAchievement

   Entities:
   - Achievement (определение достижения)
   - UserAchievement (progress пользователя)

   Value Objects:
   - AchievementID, AchievementType (enum)
   - Progress (current, target)

   Invariants:
   - Achievement можно разблокировать только один раз
   - Progress не может быть > target

   Domain Events:
   - AchievementUnlockedEvent

   Use Cases:
   - CheckAchievementsUseCase(userID) → []UnlockedAchievement
   - GetUserAchievementsUseCase(userID) → AchievementProgress
   ```

2. **USER_FLOW.md** (30 минут)
   ```markdown
   ## Экран: Profile (обновлено)

   [Wireframe с секцией Achievements]

   ## Экран: Achievement Modal (новый)

   [Popup при разблокировке]

   Триггеры:
   - После завершения квиза
   - Real-time при достижении milestone
   ```

3. **Backend** (4 hours)
   - Domain: `internal/domain/achievement/`
   - Use Cases: `internal/application/achievement/`
   - Event Handler: subscribe to QuizCompletedEvent
   - API endpoints + Swagger

4. **Frontend** (5 hours)
   - `AchievementsSection.vue`
   - `AchievementModal.vue` (popup)
   - `useAchievements.ts` composable
   - Integrate with Profile view

---

## Инструменты для синхронизации документов

### VS Code Extensions (рекомендуемые)

```json
{
  "recommendations": [
    "yzhang.markdown-all-in-one",     // Table of Contents auto-update
    "davidanson.vscode-markdownlint", // Linting
    "shd101wyy.markdown-preview-enhanced", // Preview
    "bierner.markdown-mermaid"        // Diagrams
  ]
}
```

### Git Hooks (автоматическая проверка)

**`.git/hooks/pre-commit`** (создать скрипт):
```bash
#!/bin/bash

# Проверка: если изменен backend domain код, напомнить про DOMAIN.md
if git diff --cached --name-only | grep -q "internal/domain/"; then
  if ! git diff --cached --name-only | grep -q "docs/DOMAIN.md"; then
    echo "⚠️  WARNING: You modified domain code but didn't update DOMAIN.md"
    echo "   Consider updating documentation before committing."
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
    fi
  fi
fi

# Проверка: если изменены views, напомнить про USER_FLOW.md
if git diff --cached --name-only | grep -q "tma/src/views/"; then
  if ! git diff --cached --name-only | grep -q "docs/USER_FLOW.md"; then
    echo "⚠️  WARNING: You modified views but didn't update USER_FLOW.md"
    echo "   Consider updating documentation before committing."
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
    fi
  fi
fi
```

### Makefile команды

```makefile
# В backend/Makefile
.PHONY: docs
docs:
	@echo "📚 Generating documentation..."
	@make swagger
	@echo "✅ Remember to update docs/DOMAIN.md if domain changed"
	@echo "✅ Remember to update docs/USER_FLOW.md if API changed"

.PHONY: check-docs
check-docs:
	@echo "🔍 Checking documentation..."
	@git diff --name-only origin/main | grep "internal/domain/" && \
	  echo "⚠️  Domain changed - update DOMAIN.md" || true
	@git diff --name-only origin/main | grep "handlers/" && \
	  echo "⚠️  Handlers changed - update USER_FLOW.md" || true
```

---

## FAQ: Документация

### Q: Что делать, если изменение слишком маленькое?

**A:** Даже маленькие изменения стоит документировать, если они:
- Меняют бизнес-правила
- Добавляют новые API endpoints
- Меняют User Journey

Для очень мелких (typo, рефакторинг без изменения поведения) - можно пропустить.

### Q: Как синхронизировать DOMAIN.md и USER_FLOW.md?

**A:** Используйте cross-references:
```markdown
# В DOMAIN.md
See User Journey: [USER_FLOW.md](./USER_FLOW.md#quiz-taking)

# В USER_FLOW.md
Domain model: [DOMAIN.md](./DOMAIN.md#quiz-session-aggregate)
```

### Q: Что делать, если документация устарела?

**A:**
1. Пометить секцию как `DEPRECATED`
2. Создать issue: "Update documentation for X"
3. Постепенно обновить (не всё сразу)

### Q: Нужно ли документировать эксперименты?

**A:** Да, но в отдельной секции:
```markdown
## Experimental Features

> ⚠️ **EXPERIMENTAL:** This feature is under development and may change.

### Quiz Hints System

[Description]
```

### Q: Как часто делать review документации?

**A:**
- **После каждого спринта** - быстрый check
- **После мажорного релиза** - полный audit
- **При онбординге** - проверка актуальности

---

## Итоговая шпаргалка

| Тип изменения | Начать с | Порядок |
|---------------|----------|---------|
| 🔧 Новая доменная логика | `DOMAIN.md` | DOMAIN → Backend → USER_FLOW → Frontend |
| 🎨 UI/UX улучшения | `USER_FLOW.md` | USER_FLOW → Frontend → DOMAIN (если нужно) |
| 🚀 Крупная фича | `DOMAIN.md` + `USER_FLOW.md` | Оба параллельно → Backend → Frontend |
| 🐛 Bugfix (domain) | `DOMAIN.md` | Исправить документацию + код |
| 🐛 Bugfix (UI) | `USER_FLOW.md` | Исправить документацию + код |
| 📝 Рефакторинг | Не требуется | Только если меняется публичное API |

---

**Дата создания:** 2026-01-18
**Проект:** Quiz Sprint TMA
**Принцип:** Documentation-Driven Development
