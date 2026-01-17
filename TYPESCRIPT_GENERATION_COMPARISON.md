# Сравнение подходов к генерации TypeScript типов из Go Backend

Анализ 4 основных подходов к синхронизации типов между Go API и Vue 3 фронтенда.

## Сравнительная таблица

| Критерий | swaggo+kubb (текущий) | ogen-go | oapi-codegen+tygo | tygo |
|----------|----------------------|---------|-------------------|------|
| **Источник истины** | Go code (свагг-аннотации) → OpenAPI → TS | OpenAPI spec | OpenAPI spec + Go code | Go code напрямую |
| **Где определяются типы** | Go handlers (swagger_models.go) | OpenAPI YAML/JSON | OpenAPI YAML/JSON + Go structs | Go structs с комментариями |
| **Количество шагов** | 3 этапа: Go → Swagger → OpenAPI → TS | 2 этапа: OpenAPI → TS | 3 этапа: Go + OpenAPI → Merged → TS | 1 этап: Go → TS |
| **Генератор типов** | kubb/plugin-ts | ogen-go (встроенный) | oapi-codegen (REST) + tygo (types) | tygo (standalone) |
| **Поддержка Fiber** | ✅ Да (swaggo/swag) | ✅ Да (gen.Do() in handlers) | ⚠️ Partial (требует адаптер) | ❌ Нет |
| **Имена типов** | `InternalInfrastructureHttpHandlersQuizDetailDTO` (очень длинные) | `QuizDetailDTO` (чистые) | `QuizDetailDTO` (чистые) | `QuizDetailDTO` (чистые) |
| **Namespace в TS** | Структурирован по Go пакетам | Структурирован по операциям | Структурирован по операциям | Прямое маппирование |
| **Вложенные типы** | ✅ Да | ✅ Да | ✅ Да | ✅ Да |
| **Swagger UI** | ✅ Auto-generated | ✅ Auto-generated | ✅ Auto-generated | ❌ Нет (need OpenAPI manually) |
| **REST Client (TS)** | ✅ Да (kubb/plugin-oas) | ✅ Да (ogen-go generated) | ✅ Да (oapi-codegen generated) | ❌ Нет |
| **Vue Query hooks** | ✅ Да (kubb/plugin-vue-query) | ❌ Нет (нужна интеграция) | ❌ Нет (нужна интеграция) | ❌ Нет |
| **Zod schemas** | ✅ Да (kubb/plugin-zod) | ⚠️ Отдельно (need goenv) | ⚠️ Отдельно | ❌ Нет |
| **GraphQL поддержка** | ⚠️ Limited (OpenAPI only) | ❌ Нет | ❌ Нет | ❌ Нет |
| **WebSocket поддержка** | ❌ Нет (OpenAPI limitation) | ❌ Нет | ❌ Нет | ❌ Нет |
| **Синхронизация типов** | ⚠️ Manual (swaggo + kubb) | ⚠️ Manual (OpenAPI spec update) | ⚠️ Manual (OpenAPI + Go) | ✅ Auto (из Go кода) |
| **Вероятность расхождений** | 🔴 Высокая (3 источника) | 🟠 Средняя (OpenAPI + TS) | 🟠 Средняя (OpenAPI + Go + TS) | 🟢 Низкая (1 источник) |
| **Цена синхронизации** | 💰 Высокая (duplicate types) | 💰 Средняя | 💰 Средняя | 🟢 Низкая |
| **Конфигурация (⭐)** | ⭐⭐ (2) - просто, но много файлов | ⭐⭐⭐ (3) - чуть сложнее | ⭐⭐⭐⭐ (4) - гибридный подход | ⭐⭐⭐⭐⭐ (5) - просто, один конфиг |
| **Время разработки** | 📊 Быстро для REST | 📊 Очень быстро (все в Go) | 📊 Медленно (два конфига) | 📊 Самое быстрое |
| **Performance (генерация)** | 🟢 Быстро (~2-3 сек) | 🟢 Быстро (~1-2 сек) | 🟠 Медленно (~5-10 сек) | 🟢 Быстро (~1 сек) |
| **Изучение кривой** | 📈 Средняя (swaggo + kubb) | 📈 Средняя (ogen concepts) | 📈 Высокая (2 tool combo) | 📈 Низкая (просто Go) |
| **Документация** | ✅ Хорошая (swaggo + kubb) | ⚠️ Хорошая (ogen) | ⚠️ Хорошая (oapi-codegen) | ⚠️ Хорошая (tygo) |
| **Best for** | REST API + Swagger UI + Vue Query | Все Go + Fiber без OpenAPI | Hybrid: OpenAPI contracts + Go impl | Чистый Go → TS синтаксис |
| **Не рекомендуется для** | Комплексные типы (имена!) | Нужна OpenAPI документация | Микросервисы (много конфигов) | Нужен Swagger UI или REST client |

## Детальное описание каждого подхода

### 1. swaggo/swag + kubb (ТЕКУЩИЙ)

**Текущее состояние в Quiz Sprint:**

```
Go Code (handlers/swagger_models.go)
  ↓ (swaggo annotations)
Swagger JSON (docs/swagger.json)
  ↓ (OpenAPI format)
Kubb Parser
  ↓ (TS generation)
TypeScript Types + Vue Query Hooks + Zod Schemas
```

**Плюсы:**
- ✅ Swagger UI автоматически
- ✅ Vue Query hooks встроены в kubb
- ✅ Zod validation schemas
- ✅ Хорошо документировано (swaggo community)
- ✅ Быстрая генерация

**Минусы:**
- ❌ Очень длинные имена типов: `InternalInfrastructureHttpHandlersQuizDetailDTO`
- ❌ 3 этапа синхронизации = больше точек отказа
- ❌ Дублирование типов в `swagger_models.go`
- ❌ Нет WebSocket поддержки (OpenAPI limitation)
- ❌ Сложно переиспользовать внутренние типы

**Пример текущего состояния:**
```go
// backend/internal/infrastructure/http/handlers/swagger_models.go
type QuizDetailDTO struct {
    ID        string        `json:"id"`
    Title     string        `json:"title"`
    Questions []QuestionDTO `json:"questions"`
}
```

```typescript
// tma/src/api/generated/types/internalInfrastructureHttpHandlers/QuizDetailDTO.ts
export type InternalInfrastructureHttpHandlersQuizDetailDTO = {
    id?: string
    title?: string
    questions?: InternalInfrastructureHttpHandlersQuestionDTO[]
}
```

**Миграция сложность:** ⭐⭐ (низкая) - просто убрать swaggo

---

### 2. ogen-go (Schema-first)

**Архитектура:**

```
OpenAPI Spec (openapi.yaml)
  ↓ (ogen-go generator)
Typed Go Server + Types + Handlers
  ↓ (handlers implement generated interface)
Fiber Adapter
  ↓ + (TypeScript types exported)
TypeScript Types (via reflection/export)
```

**Плюсы:**
- ✅ Чистые имена типов: `QuizDTO`
- ✅ Type-safe handlers (реализуют интерфейс)
- ✅ OpenAPI spec как источник истины
- ✅ Меньше дублирования кода
- ✅ Быстрая генерация

**Минусы:**
- ❌ Требует переписания всех handlers
- ❌ Fiber не полностью поддерживается (нужен адаптер)
- ❌ Нет встроенной интеграции с Vue Query
- ❌ TypeScript экспорт требует доп. утилит
- ❌ OpenAPI spec всё еще нужен = дублирование

**Пример использования:**

```go
// Сгенерировано ogen-go из OpenAPI
type GetQuizIDRes interface {
    getQuizIDRes()
}

// Вы реализуете
func (h *Handler) GetQuizID(ctx context.Context, params GetQuizIDParams) (GetQuizIDRes, error) {
    // business logic
}
```

**Миграция сложность:** ⭐⭐⭐⭐ (очень высокая) - полная переработка handlers

---

### 3. oapi-codegen + tygo (Hybrid)

**Архитектура:**

```
OpenAPI Spec
  ├─→ oapi-codegen (REST client + types)
  │   ↓
  │   TypeScript types + HTTP client
  │
└─→ Go handlers (manual)
    ↓ (tygo)
    Go → TypeScript types (direct conversion)
    ↓
Merged TypeScript types + Client
```

**Плюсы:**
- ✅ Чистые имена типов
- ✅ OpenAPI contract guarantees
- ✅ Два независимых генератора
- ✅ REST client полностью типизирован
- ✅ Гибкая архитектура

**Минусы:**
- ❌ Нужно поддерживать ДВА конфига (OpenAPI + Go)
- ❌ Медленная генерация (5-10 сек)
- ❌ tygo требует правильных struct tags
- ❌ Можно в итоге иметь конфликты типов
- ❌ Нет Vue Query встроенно

**Пример конфигурации:**

```yaml
# openapi.yaml → oapi-codegen
openapi: 3.0.0
paths:
  /quiz/{id}:
    get:
      operationId: getQuizById
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QuizDetailDTO'
```

```toml
# tygo.toml
[[packages]]
path = "github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/handlers"
type_defs = true
```

**Миграция сложность:** ⭐⭐⭐⭐ (высокая) - нужно добавить оба tool

---

### 4. tygo (Code-first Direct)

**Архитектура:**

```
Go Code (handlers with proper tags)
  ↓ (tygo parser)
Direct Go → TypeScript conversion
  ↓
Clean TypeScript types
  ✗ Нет OpenAPI/Swagger
  ✗ Нет REST client
  ✗ Требует manual Vue Query hooks
```

**Плюсы:**
- ✅ Самый простой setup
- ✅ Чистые имена типов: `QuizDTO`
- ✅ Один источник истины (Go)
- ✅ Самая низкая вероятность расхождений
- ✅ Быстрая генерация (~1 сек)
- ✅ Идеально для внутренних типов

**Минусы:**
- ❌ Нет Swagger UI (пришлось бы писать OpenAPI самостоятельно)
- ❌ Нет REST client (нужно писать вручную или использовать отдельно)
- ❌ Нет автогенерации Vue Query hooks
- ❌ Не поддерживает Fiber аннотации
- ❌ Нужно поддерживать клиент отдельно

**Пример использования:**

```go
// backend/internal/infrastructure/http/handlers/types.go
//ts:type QuizDetailDTO
type QuizDetailDTO struct {
    ID        string         `json:"id"`
    Title     string         `json:"title"`
    Questions []QuestionDTO  `json:"questions"`
}
```

```typescript
// Generated: tma/src/api/types.ts
export type QuizDetailDTO = {
    id: string
    title: string
    questions: QuestionDTO[]
}
```

**Миграция сложность:** ⭐⭐ (низкая) - просто заменить kubb на tygo

---

## Рекомендация для Quiz Sprint TMA

### 🎯 РЕКОМЕНДУЕМОЕ РЕШЕНИЕ: **tygo** + **Manual OpenAPI** (миграция с текущего swaggo)

**Почему:**

1. **Проблема с текущим решением:**
   - Имена типов невероятно длинные: `InternalInfrastructureHttpHandlersQuizDetailDTO`
   - Это усложняет код фронтенда и делает его нечитаемым
   - Дублирование в `swagger_models.go` = два источника истины

2. **Почему tygo лучше:**
   - ✅ Чистые имена: `QuizDTO`
   - ✅ Один источник истины (Go structs)
   - ✅ Минимальная конфигурация
   - ✅ Очень быстро генерируется
   - ✅ Структуры уже есть в коде

3. **Что теряем:**
   - ❌ Swagger UI (но можно добавить отдельно, если нужен)
   - ❌ REST client (но это легко написать вручную для 5-6 эндпоинтов)
   - ❌ Vue Query hooks (но можно написать вручную, это просто обёртка)

### 🔧 План миграции (2-3 часа работы)

**Шаг 1: Установка tygo**
```bash
cd /Users/barsukov/projects/quiz-sprint/tma
npm install --save-dev tygo
```

**Шаг 2: Удалить дублирующие типы из Go**
```bash
# Удалить файл
rm /Users/barsukov/projects/quiz-sprint/backend/internal/infrastructure/http/handlers/swagger_models.go

# Переместить типы DTO в отдельный файл для экспорта
mkdir -p /Users/barsukov/projects/quiz-sprint/backend/internal/infrastructure/http/dto
# Создать types.go с типами DTO без swaggo аннотаций
```

**Шаг 3: Конфигурация tygo**
```yaml
# tygo.yaml
packages:
  - path: "github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/dto"
    output_file: "tma/src/api/generated/types.ts"
    # Экспортировать только public типы (QuizDTO, SessionDTO и т.д.)
```

**Шаг 4: Обновить kubb конфиг**
```typescript
// kubb.config.ts - убрать plugin-ts (теперь используем tygo)
// Оставить только plugin-oas + plugin-vue-query
```

**Шаг 5: Написать простой REST client**
```typescript
// tma/src/api/client.ts
export const quizApi = {
  listQuizzes: async () => fetch('/api/v1/quiz'),
  getQuiz: async (id: string) => fetch(`/api/v1/quiz/${id}`),
  // ...
}
```

### 📊 Сравнение затрат (для Quiz Sprint)

| Метрика | Текущее (swaggo) | Рекомендуемое (tygo) |
|---------|-----------------|-------------------|
| Время setup | 30 мин | 15 мин |
| Длина имён типов | 🔴 **60+ chars** | 🟢 **10-20 chars** |
| Точки синхронизации | 🔴 **3** (Go → Swagger → TS) | 🟢 **1** (Go → TS) |
| Вероятность ошибок | 🔴 **Высокая** | 🟢 **Низкая** |
| Swagger UI | 🟢 Auto | 🟠 Manual (опционально) |
| REST client | 🟢 Auto | 🟠 Manual (быстро писать) |
| Vue Query | 🟢 Auto | 🟠 Manual (шаблон простой) |
| **Общая оценка** | ⭐⭐ | ⭐⭐⭐⭐ |

### 🚫 Альтернативы (почему не выбирали)

| Вариант | Причина отказа |
|---------|----------------|
| **Оставить swaggo** | Имена типов слишком длинные, дублирование кода |
| **ogen-go** | Требует полной переписи всех handlers (очень дорого) |
| **oapi-codegen+tygo** | Сложная гибридная архитектура для малого проекта |

### ⚠️ Важные замечания

1. **WebSocket:** Никакой из инструментов не поддерживает WebSocket автогенерацию. Для leaderboard можно:
   - Писать типы вручную (простые)
   - Или использовать tygo для базовых типов + manual messages

2. **Валидация:** Zod schemas от kubb потеряются. Можно:
   - Использовать встроенные TypeScript checks
   - Или написать простые валидаторы вручную

3. **Swagger UI:** Если критична для API документирования:
   - Оставить swaggo в Go для документации
   - Но НЕ использовать его для TS generation
   - Только для `GET /api/docs`

---

## Примеры реализации для Quiz Sprint

### Вариант A: Чистый tygo (рекомендуемый)

```go
// backend/internal/infrastructure/http/types/quiz.go
package types

type QuizDTO struct {
    ID             string `json:"id"`
    Title          string `json:"title"`
    Description    string `json:"description"`
    QuestionsCount int    `json:"questionsCount"`
    TimeLimit      int    `json:"timeLimit"`
    PassingScore   int    `json:"passingScore"`
    CreatedAt      int64  `json:"createdAt"`
}

type QuestionDTO struct {
    ID       string      `json:"id"`
    Text     string      `json:"text"`
    Answers  []AnswerDTO `json:"answers"`
    Points   int         `json:"points"`
    Position int         `json:"position"`
}
```

```typescript
// Generated: tma/src/api/types.ts
export type QuizDTO = {
    id: string
    title: string
    description: string
    questionsCount: number
    timeLimit: number
    passingScore: number
    createdAt: number
}

export type QuestionDTO = {
    id: string
    text: string
    answers: AnswerDTO[]
    points: number
    position: number
}
```

```typescript
// tma/src/api/hooks.ts (manual, но очень простой)
import { useQuery } from '@tanstack/vue-query'
import type { QuizDTO } from './types'

export const useQuizzes = () => {
  return useQuery({
    queryKey: ['quizzes'],
    queryFn: async (): Promise<QuizDTO[]> => {
      const res = await fetch('/api/v1/quiz')
      return res.json().then(d => d.data)
    }
  })
}
```

### Вариант B: Гибридный (если нужен Swagger UI)

```bash
# Оставить swaggo ТОЛЬКО для документации
swag init --output ./docs --parseInternal

# Использовать tygo для TS типов (игнорируя длинные имена из swagger_models.go)
# swaggo аннотации на handlers остаются для /api/docs
```

---

## Ссылки и ресурсы

- **tygo**: https://github.com/gzuidhof/tygo
- **swaggo**: https://github.com/swaggo/swag
- **kubb**: https://kubb.dev/
- **ogen**: https://ogen.sh/
- **oapi-codegen**: https://github.com/deepmap/oapi-codegen

---

## Заключение

**TL;DR:** Переходите на **tygo** для чистоты типов и минимизации sync issues. Текущее решение (swaggo+kubb) работает, но создаёт ненужную сложность через дублирование кода и невероятно длинные имена типов.
