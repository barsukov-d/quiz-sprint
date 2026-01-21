# User Flows - Current Implementation

> **TEMPORARY NOTE:** Детальные user flows и wireframes пока находятся в:
> - Старый файл: `docs/USER_FLOW.md` (sections 0-6, excluding "Future User Flows")
> - Этот файл будет refactored в будущем для компактности

---

## Quick Reference

**Для текущих экранов см.:**

### Главная страница (3 зоны)
- Zone 1: Daily Challenge - [`USER_FLOW.md:76-230`](../USER_FLOW.md#zone-1-daily-challenge-)
- Zone 2: Quick Actions - [`USER_FLOW.md:232-272`](../USER_FLOW.md#zone-2-quick-actions-)
- Zone 3: Categories - [`USER_FLOW.md:274-303`](../USER_FLOW.md#zone-3-browse-by-category-)

### Quiz List по категории
- Wireframe и элементы - [`USER_FLOW.md:306-418`](../USER_FLOW.md#1-список-квизов-по-категории-quiz-list)

### Quiz Details
- Wireframe - [`USER_FLOW.md:420-492`](../USER_FLOW.md#2-детали-quiz)

### Quiz Play (прохождение)
- Процесс ответа - [`USER_FLOW.md:494-630`](../USER_FLOW.md#3-прохождение-quiz)

### Results
- Экран результатов - [`USER_FLOW.md:632-747`](../USER_FLOW.md#4-результаты)

### Leaderboard
- Таблица лидеров - [`USER_FLOW.md:749-818`](../USER_FLOW.md#5-leaderboard)

### Profile
- Профиль пользователя - [`USER_FLOW.md:820-904`](../USER_FLOW.md#6-профиль-пользователя)

---

## Navigation Flow

```
Home (3 zones)
  ├─ Daily Challenge → Quiz Details → Quiz Play → Results → Leaderboard
  ├─ Random Quiz → Quiz Play → Results
  ├─ Continue Playing → Quiz Play (resume)
  └─ Category → Quiz List → Quiz Details → Quiz Play → Results

Bottom Tab Bar:
  [Home] [Leaderboard] [Profile]
```

---

## UI Components (Reusable)

См. [`USER_FLOW.md:906-1009`](../USER_FLOW.md#компоненты-ui) для детального описания:
- CategoryCard
- QuizCard
- ProgressBar
- Timer
- AnswerButton
- LeaderboardRow
- FeedbackBanner

---

## Интерактивные механики

См. [`USER_FLOW.md:1011-1088`](../USER_FLOW.md#интерактивные-механики) для:
- Выбор ответа (анимации, feedback)
- Таймер (warning states)
- Прогресс бар
- Real-time Leaderboard (WebSocket)

---

## Edge Cases

См. [`USER_FLOW.md:1090-1240`](../USER_FLOW.md#edge-cases--error-handling) для handling:
- Закрытие TMA во время квиза (session recovery)
- Истекшее время
- Потеря интернета
- Backend errors
- Double submission
- Quiz deleted/changed
- Empty leaderboard
- Validation errors

---

**TODO:** Refactor USER_FLOW.md → extract current flows в этот файл (компактный формат)

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 1.0 (temporary - ссылки на старый файл)
**Проект:** Quiz Sprint TMA
