# Quick Start - Swagger Generation

## 🚀 Самые частые команды

### Вариант 1: Из папки `tma/` (Frontend)

```bash
cd tma

# Сгенерировать только Swagger (backend)
pnpm run generate:swagger

# Сгенерировать только TypeScript (frontend)
pnpm run generate:api

# Сгенерировать всё (Swagger + TypeScript)
pnpm run generate:all
```

### Вариант 2: Из папки `backend/`

```bash
cd backend

# Сгенерировать Swagger
make swagger

# Показать все команды
make help

# Сгенерировать Swagger + запустить сервер
make dev
```

---

## 📋 Когда что запускать?

| Что изменили | Команда | Где запускать |
|-------------|---------|---------------|
| Go handlers (аннотации `@Summary`, `@Param`, etc.) | `make swagger` | `backend/` |
| Go handlers | `pnpm run generate:swagger` | `tma/` |
| Swagger.json уже готов, нужен TypeScript | `pnpm run generate:api` | `tma/` |
| Изменили Go + нужен TypeScript | `pnpm run generate:all` | `tma/` |

---

## 🎯 Типичный workflow

### 1. Добавили новый endpoint в Go

```go
// @Summary Get quiz by ID
// @Tags quiz
// @Param id path string true "Quiz ID"
// @Success 200 {object} handlers.QuizDTO
// @Router /quiz/{id} [get]
func (h *QuizHandler) GetQuiz(c *fiber.Ctx) error {
    // ...
}
```

### 2. Генерируем Swagger

```bash
cd backend
make swagger
```

### 3. Генерируем TypeScript

```bash
cd ../tma
pnpm run generate:api
```

### 4. Используем в Vue

```typescript
import { useGetQuizId } from '@/api/generated/hooks/quizController'

const { data: quiz, isLoading } = useGetQuizId({ id: '123' })
```

---

## ✨ Одной командой

```bash
# Из tma/
pnpm run generate:all
```

Эта команда:
1. ✅ Генерирует Swagger из Go кода (`backend/docs/swagger.json`)
2. ✅ Генерирует TypeScript типы (`tma/src/api/generated/`)
3. ✅ Генерирует Vue Query хуки
4. ✅ Генерирует Zod схемы

---

## 🔧 Первая настройка

### Установить swag глобально (опционально)

```bash
cd backend
make swagger-install

# Теперь можно использовать swag напрямую
swag init --generalInfo cmd/api/main.go --output docs
```

---

## 📚 Подробная документация

Смотрите `backend/SWAGGER.md` для:
- Примеры Swagger аннотаций
- Troubleshooting
- CI/CD интеграция
- Required fields
