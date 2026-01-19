# User Flow & UX Specification - Quiz Sprint TMA

## 📋 Содержание

1. [User Journey Overview](#user-journey-overview)
2. [Экраны приложения](#экраны-приложения)
   - [0. Выбор категории (Categories)](#0-выбор-категории-categories)
   - [1. Список квизов по категории (Quiz List)](#1-список-квизов-по-категории-quiz-list)
   - [2. Детали Quiz](#2-детали-quiz)
   - [3. Прохождение Quiz](#3-прохождение-quiz)
   - [4. Результаты](#4-результаты)
   - [5. Leaderboard](#5-leaderboard)
   - [6. Профиль пользователя](#6-профиль-пользователя)
3. [Компоненты UI](#компоненты-ui)
4. [Интерактивные механики](#интерактивные-механики)
5. [Edge Cases & Error Handling](#edge-cases--error-handling)

---

## User Journey Overview

### Основной флоу (Happy Path)

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER JOURNEY                              │
└─────────────────────────────────────────────────────────────────┘

1. Вход в TMA
   ↓
2. Авторегистрация (Telegram Auth)
   ↓
3. Экран выбора категории
   ↓
4. Клик на категорию → Список квизов по категории
   ↓
5. Клик на квиз → Детали квиза
   ↓
6. Клик "Start Quiz" → Начало игры
   ↓
7. Прохождение вопросов (Question by Question)
   │  • Выбор ответа
   │  • Мгновенная обратная связь
   │  • Переход к следующему
   ↓
8. Завершение квиза → Экран результатов
   ↓
9. Просмотр Leaderboard
   ↓
10. Возврат к списку квизов или выбор другой категории
```

### Альтернативные флоу

```
• Пользователь закрывает TMA во время квиза
  → Сессия сохраняется как "Abandoned"
  → При возврате: продолжить или начать заново?

• Пользователь исчерпывает время
  → Автоматическое завершение
  → Подсчет очков по отвеченным вопросам

• Пользователь уже проходил квиз ранее
  → Показать предыдущий результат на карточке
  → Кнопка "Try Again" вместо "Start"
```

---

## Экраны приложения

### 0. Выбор категории (Categories)

**Назначение:** Дать пользователю выбрать интересующую категорию квизов

**Wireframe:**
```
┌────────────────────────────────────────────┐
│  Quiz Sprint                         [👤] │  ← Header
├────────────────────────────────────────────┤
│                                            │
│  Choose a Category                         │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │                                      │ │
│  │  🧠  General Knowledge               │ │
│  │      Test your general knowledge     │ │
│  │                                      │ │
│  │                         12 quizzes → │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │                                      │ │
│  │  🌍  Geography                       │ │
│  │      Explore the world               │ │
│  │                                      │ │
│  │                          8 quizzes → │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │                                      │ │
│  │  💻  Technology                      │ │
│  │      Test your tech skills           │ │
│  │                                      │ │
│  │                         15 quizzes → │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │                                      │ │
│  │  🎬  Movies & TV                     │ │
│  │      Cinema and series trivia        │ │
│  │                                      │ │
│  │                          6 quizzes → │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │                                      │ │
│  │  📚  History                         │ │
│  │      Journey through time            │ │
│  │                                      │ │
│  │                         10 quizzes → │ │
│  └──────────────────────────────────────┘ │
│                                            │
└────────────────────────────────────────────┘
      ↑                ↑                ↑
   [Home]        [Leaderboard]      [Profile]
```

**Элементы карточки категории:**
- **Иконка категории** - крупная эмодзи для визуальной идентификации
- **Название категории** - понятное, запоминающееся
- **Описание** - короткая фраза, объясняющая тему
- **Количество квизов** - показывает объем контента
- **Стрелка вправо** - индикатор навигации

**Состояния:**
- **Loading** - Skeleton cards во время загрузки
- **Empty** - Заглушка "No categories available yet"
- **Error** - Сообщение с кнопкой retry

**UX детали:**
- Крупные, удобные для тапа карточки (min 64px высота)
- Плавная анимация при скролле
- Haptic feedback при тапе (если поддерживается)
- При клике → переход на список квизов этой категории

**Navigation:**
```
Categories (Home) → Quiz List (by category) → Quiz Details
     ↑                      ↓
     └──────── Back ────────┘
```

---

### 1. Список квизов по категории (Quiz List)

**Назначение:** Показать квизы выбранной категории, мотивировать к прохождению

**Wireframe:**
```
┌────────────────────────────────────────────┐
│  ← Back      🧠 General Knowledge    [👤] │  ← Header with category
├────────────────────────────────────────────┤
│                                            │
│  Quizzes in this category                  │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │ 🧠 General Knowledge                 │ │
│  │ 10 questions • 5 min                 │ │
│  │ Your best: 85/100 🏆 #12             │ │  ← Если уже проходил
│  │                                      │ │
│  │                    [Start Quiz] →    │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │ 🌍 Geography Challenge               │ │
│  │ 15 questions • 7 min                 │ │
│  │ 1,234 players • Top: 95/100          │ │
│  │                                      │ │
│  │                    [Start Quiz] →    │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │ 💻 Tech Trivia                       │ │
│  │ 20 questions • 10 min                │ │
│  │ Not attempted yet                    │ │
│  │                                      │ │
│  │                    [Start Quiz] →    │ │
│  └──────────────────────────────────────┘ │
│                                            │
└────────────────────────────────────────────┘
      ↑                ↑                ↑
   [Home]        [Leaderboard]      [Profile]
```

**Элементы карточки квиза:**
- **Иконка категории** - эмодзи для быстрой идентификации
- **Название квиза** - короткое, привлекающее внимание
- **Метаданные:**
  - Количество вопросов
  - Примерное время прохождения
  - Количество игроков (социальное доказательство)
- **Личная статистика** (если проходил):
  - Лучший результат
  - Позиция в leaderboard
  - Badge (если топ-10)
- **CTA кнопка:**
  - "Start Quiz" (первый раз)
  - "Try Again" (если уже проходил)

**Состояния:**
- **Loading** - Skeleton cards во время загрузки
- **Empty** - Заглушка "No quizzes available"
- **Error** - Сообщение с кнопкой retry

---

### 2. Детали Quiz

**Назначение:** Дать информацию о квизе перед началом, мотивировать старт

**Wireframe:**
```
┌────────────────────────────────────────────┐
│  ← Back                              [👤]  │
├────────────────────────────────────────────┤
│                                            │
│         🧠 General Knowledge               │
│                                            │
│  Test your knowledge across various       │
│  topics including history, science,       │
│  and pop culture!                         │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 📊 Quiz Stats                      │   │
│  │                                    │   │
│  │ Questions:        10               │   │
│  │ Time Limit:       5 minutes        │   │
│  │ Passing Score:    60%              │   │
│  │ Total Players:    1,234            │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 🏆 Your Best Result                │   │
│  │                                    │   │
│  │ Score:     85/100                  │   │
│  │ Rank:      #12 of 1,234            │   │
│  │ Date:      2 days ago              │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 👑 Top 3 Leaders                   │   │
│  │                                    │   │
│  │ 1. @username1     100/100 🥇       │   │
│  │ 2. @username2      98/100 🥈       │   │
│  │ 3. @username3      95/100 🥉       │   │
│  │                                    │   │
│  │            [View Full Leaderboard] │   │
│  └────────────────────────────────────┘   │
│                                            │
│                                            │
│        ┌──────────────────────┐           │
│        │   START QUIZ  →      │           │
│        └──────────────────────┘           │
│                                            │
└────────────────────────────────────────────┘
```

**Ключевые элементы:**
- **Описание квиза** - О чем квиз, что будет проверяться
- **Правила:**
  - Количество вопросов
  - Лимит времени
  - Минимальный проходной балл
- **Социальные элементы:**
  - Количество игроков
  - Топ-3 результата
  - Ваша позиция (если проходили)
- **CTA:** Большая яркая кнопка "Start Quiz"

**UX детали:**
- При повторном прохождении показать:
  - Предыдущий результат
  - Изменение позиции в рейтинге
  - Мотивационный текст: "Can you beat your score?"

---

### 3. Прохождение Quiz

**Назначение:** Интерактивное прохождение вопросов с мгновенной обратной связью

**Wireframe (во время ответа на вопрос):**
```
┌────────────────────────────────────────────┐
│  Question 3 of 10              ⏱ 2:34     │  ← Progress + Timer
├────────────────────────────────────────────┤
│  ▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░              │  ← Progress bar
│                                            │
│                                            │
│  Which planet is known as the Red Planet? │  ← Question text
│                                            │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  A) Venus                            │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  B) Mars                             │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  C) Jupiter                          │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  D) Saturn                           │ │
│  └──────────────────────────────────────┘ │
│                                            │
│                                            │
│                                            │
│  Score: 20/30                              │  ← Current score
└────────────────────────────────────────────┘
```

**После выбора ответа (feedback):**
```
┌────────────────────────────────────────────┐
│  Question 3 of 10              ⏱ 2:31     │
├────────────────────────────────────────────┤
│  ▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░              │
│                                            │
│                                            │
│  Which planet is known as the Red Planet? │
│                                            │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  A) Venus                            │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  B) Mars                          ✓  │ │  ← Правильный ответ
│  └──────────────────────────────────────┘ │  (зеленая обводка)
│  ↑ Selected                                │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  C) Jupiter                          │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  D) Saturn                           │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ ✓ Correct! +10 points              │   │  ← Feedback banner
│  └────────────────────────────────────┘   │
│                                            │
│        ┌──────────────────────┐           │
│        │   NEXT QUESTION  →   │           │  ← Next button
│        └──────────────────────┘           │
│                                            │
│  Score: 30/30                              │
└────────────────────────────────────────────┘
```

**Если ответ неправильный:**
```
│  ┌──────────────────────────────────────┐ │
│  │  A) Venus                            │ │
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  B) Mars                          ✓  │ │  ← Правильный (зеленая)
│  └──────────────────────────────────────┘ │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  C) Jupiter                       ✗  │ │  ← Ваш выбор (красная)
│  └──────────────────────────────────────┘ │
│  ↑ Your answer                             │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ ✗ Incorrect. No points.            │   │
│  │ The correct answer is Mars.        │   │
│  └────────────────────────────────────┘   │
```

**Ключевые элементы:**
1. **Header:**
   - Номер вопроса (Question X of Y)
   - Таймер обратного отсчета
   - Кнопка выхода (с подтверждением)

2. **Progress Bar:**
   - Визуальный прогресс прохождения
   - Заполняется по мере ответов

3. **Вопрос:**
   - Текст вопроса (крупный шрифт)
   - Опционально: картинка (если есть)

4. **Варианты ответов:**
   - 2-4 кнопки-варианта
   - Крупные, удобные для тапа
   - **Состояния:**
     - Default (не выбран)
     - Selected (выбран, до отправки)
     - Correct (зеленая обводка + ✓)
     - Incorrect (красная обводка + ✗)
     - Disabled (после выбора)

5. **Feedback:**
   - Баннер с результатом (Correct/Incorrect)
   - Начисленные очки
   - Если неправильно - показать верный ответ

6. **Кнопка "Next":**
   - Появляется после выбора ответа
   - Автоматический переход через 2 секунды (или по клику)

7. **Footer:**
   - Текущий счет
   - Прогресс (X/Y вопросов)

**UX Flow:**
```
1. Показать вопрос
2. Пользователь выбирает ответ (тап на кнопку)
3. Отправка на backend (POST /quiz/session/:id/answer)
4. Получение результата (isCorrect, points, nextQuestion)
5. Показать feedback (анимация: зеленая/красная обводка)
6. Обновить score
7. Показать кнопку "Next"
8. Переход к следующему вопросу
9. Repeat до последнего вопроса
10. Redirect на экран Results
```

---

### 4. Результаты

**Назначение:** Показать результат, мотивировать к повторному прохождению

**Wireframe:**
```
┌────────────────────────────────────────────┐
│                                            │
│              🎉 Quiz Completed!            │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │                                    │   │
│  │         Your Score                 │   │
│  │                                    │   │
│  │           85/100                   │   │  ← Крупный счет
│  │                                    │   │
│  │         🏆 Rank #12                │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 📊 Performance                     │   │
│  │                                    │   │
│  │ Correct Answers:   17/20           │   │
│  │ Accuracy:          85%             │   │
│  │ Time Taken:        4:23            │   │
│  │ Points Earned:     850             │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 🎯 Achievement                     │   │
│  │                                    │   │
│  │ ✓ First Completion                │   │
│  │ ✓ Top 20% Score                   │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│                                            │
│    ┌────────────────┐  ┌──────────────┐  │
│    │ Try Again  🔄  │  │ Leaderboard  │  │
│    └────────────────┘  └──────────────┘  │
│                                            │
│              [Back to Quizzes]             │
│                                            │
└────────────────────────────────────────────┘
```

**Вариации по результату:**

**Если прошел успешно (≥ passing score):**
```
│              🎉 Congratulations!           │
│                You Passed!                 │
│                                            │
│              85/100 (85%)                  │
│         Passing score: 60%                 │
```

**Если не прошел (< passing score):**
```
│              😔 Almost There!              │
│           Keep practicing!                 │
│                                            │
│              45/100 (45%)                  │
│         Passing score: 60%                 │
│                                            │
│   You need 15 more points to pass.        │
```

**Если установлен новый личный рекорд:**
```
│              🏆 NEW RECORD!                │
│         You beat your best score!          │
│                                            │
│    Previous: 75  →  Current: 85           │
│      Rank: #18  →  Rank: #12              │
```

**Элементы:**
- **Эмоциональный заголовок** (зависит от результата)
- **Крупный счет** с процентом
- **Позиция в leaderboard**
- **Детальная статистика:**
  - Правильных ответов
  - Точность (accuracy)
  - Затраченное время
  - Очки
- **Достижения** (если есть):
  - Первое прохождение
  - Топ-X%
  - Perfect score
  - Speed bonus
- **Действия:**
  - Try Again (повторить квиз)
  - View Leaderboard (посмотреть рейтинг)
  - Back to Quizzes (вернуться к списку)

**UX детали:**
- Анимация при появлении счета
- Конфетти при новом рекорде
- Share button (поделиться результатом в Telegram)

---

### 5. Leaderboard

**Назначение:** Показать рейтинг игроков, социальное сравнение

**Wireframe:**
```
┌────────────────────────────────────────────┐
│  ← Back        Leaderboard          [👤]   │
├────────────────────────────────────────────┤
│                                            │
│        🧠 General Knowledge                │
│                                            │
│  ┌─────┬────────────────┬────────┬──────┐ │
│  │Rank │ Player         │ Score  │ Date │ │
│  ├─────┼────────────────┼────────┼──────┤ │
│  │ 🥇  │ @username1     │ 100/100│ 1d   │ │
│  │ 🥈  │ @username2     │  98/100│ 2d   │ │
│  │ 🥉  │ @username3     │  95/100│ 5d   │ │
│  │  4  │ @username4     │  92/100│ 1d   │ │
│  │  5  │ @username5     │  90/100│ 3d   │ │
│  │ ... │ ...            │ ...    │ ...  │ │
│  │ 12  │ @you        ⭐ │  85/100│ now  │ │  ← Highlighted
│  │ ... │ ...            │ ...    │ ...  │ │
│  │ 50  │ @username50    │  60/100│ 10d  │ │
│  └─────┴────────────────┴────────┴──────┘ │
│                                            │
│  Showing top 50 of 1,234 players           │
│                                            │
│  [Load More]                               │
│                                            │
└────────────────────────────────────────────┘
```

**Ключевые элементы:**
1. **Топ-3 с медалями** (🥇🥈🥉)
2. **Ваша позиция** - всегда видна (highlight + scroll to)
3. **Колонки:**
   - Rank (позиция)
   - Player (username или display name)
   - Score (очки)
   - Date (время прохождения)
4. **Пагинация:**
   - Топ-50 by default
   - Load More для остальных
5. **Фильтры** (optional):
   - All Time
   - This Week
   - Today

**Состояния:**
- **Loading** - Skeleton rows
- **Empty** - "Be the first to complete this quiz!"
- **Real-time updates** - WebSocket connection для live updates

**UX детали:**
- При входе на экран - автоскролл к вашей позиции
- Анимация при изменении позиции (если WebSocket)
- Pull-to-refresh для обновления

---

### 6. Профиль пользователя

**Назначение:** Показать статистику пользователя, настройки

**Wireframe:**
```
┌────────────────────────────────────────────┐
│  ← Back           Profile                  │
├────────────────────────────────────────────┤
│                                            │
│  ┌────────────────────────────────────┐   │
│  │         👤 @username               │   │
│  │                                    │   │
│  │    John Doe                        │   │
│  │    Member since Jan 2026           │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 📊 Statistics                      │   │
│  │                                    │   │
│  │ Quizzes Completed:     12          │   │
│  │ Total Points:          1,250       │   │
│  │ Average Score:         78%         │   │
│  │ Best Rank:             #3          │   │
│  │ Time Spent:            2h 15m      │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 🏆 Achievements                    │   │
│  │                                    │   │
│  │ ✓ First Quiz Completed             │   │
│  │ ✓ 10 Quizzes Milestone             │   │
│  │ ✓ Top 10 in Any Quiz               │   │
│  │ ✓ Perfect Score                    │   │
│  │ ⏳ Speed Demon (locked)            │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ 📜 Recent Activity                 │   │
│  │                                    │   │
│  │ • General Knowledge  85/100  #12   │   │
│  │   2 hours ago                      │   │
│  │                                    │   │
│  │ • Geography Quiz     92/100  #5    │   │
│  │   1 day ago                        │   │
│  │                                    │   │
│  │ • Tech Trivia        78/100  #25   │   │
│  │   3 days ago                       │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│              [Edit Profile]                │
│              [Sign Out]                    │
│                                            │
└────────────────────────────────────────────┘
```

**Элементы:**
- **User Info:**
  - Avatar (Telegram photo)
  - Username
  - Display name
  - Member since
- **Statistics:**
  - Quizzes completed
  - Total points
  - Average score
  - Best rank achieved
  - Total time spent
- **Achievements:**
  - Unlocked badges
  - Progress к новым achievements
- **Recent Activity:**
  - Последние 5 квизов
  - Счет и позиция
  - Время прохождения
- **Actions:**
  - Edit Profile (имя, avatar)
  - Settings (язык, уведомления)
  - Sign Out

---

## Компоненты UI

### Reusable Components

#### 1. CategoryCard
```
Использование: Экран выбора категории
Props:
  - category: Category
  - quizCount: number
  - icon: string (emoji)
  - description: string

States:
  - default
  - hover/active
  - loading

Visual:
  - Крупная карточка с иконкой, названием, описанием
  - Счетчик квизов справа
  - Стрелка вправо как индикатор навигации
```

#### 2. QuizCard
```
Использование: Список квизов по категории
Props:
  - quiz: Quiz
  - userBestScore?: number
  - userRank?: number

States:
  - default
  - hover/active
  - completed (с результатом)
```

#### 3. ProgressBar
```
Использование: Quiz прохождение
Props:
  - current: number
  - total: number
  - color?: string

Visual: Заполненная полоса с процентом
```

#### 4. Timer
```
Использование: Quiz прохождение
Props:
  - timeLimit: number (секунды)
  - onTimeUp: callback

States:
  - normal (зеленый)
  - warning (< 60s, желтый)
  - critical (< 10s, красный, мигает)
```

#### 5. AnswerButton
```
Использование: Quiz вопросы
Props:
  - text: string
  - state: 'default' | 'selected' | 'correct' | 'incorrect'
  - disabled: boolean
  - onClick: callback

Visual:
  - default: серая обводка
  - selected: синяя заливка
  - correct: зеленая обводка + ✓
  - incorrect: красная обводка + ✗
```

#### 6. LeaderboardRow
```
Использование: Leaderboard
Props:
  - rank: number
  - username: string
  - score: number
  - date: string
  - isCurrentUser: boolean

Visual: Highlight если isCurrentUser
```

#### 7. FeedbackBanner
```
Использование: Quiz (после ответа)
Props:
  - type: 'correct' | 'incorrect'
  - points?: number
  - correctAnswer?: string

Auto-dismiss через 2 секунды
```

---

## Интерактивные механики

### 1. Выбор ответа

```
User taps answer
  ↓
Button state: default → selected (синяя)
  ↓
Отправка на backend (optimistic UI)
  ↓
Получение результата
  ↓
Анимация:
  - Если correct: selected → correct (зеленая) + ✓ + sound
  - Если incorrect:
      selected → incorrect (красная) + ✗
      + показать correct ответ (зеленая)
  ↓
Показать FeedbackBanner
  ↓
Обновить score (анимация +points)
  ↓
Показать кнопку "Next"
  ↓
Auto-redirect через 2s (или по клику)
```

### 2. Таймер

```
Квиз начинается с timeLimit секунд
  ↓
Каждую секунду: timeLeft--
  ↓
Если timeLeft < 60s: warning state (желтый)
  ↓
Если timeLeft < 10s: critical state (красный, мигает)
  ↓
Если timeLeft === 0:
  - Автозавершение квиза
  - Подсчет очков по отвеченным вопросам
  - Редирект на Results
```

### 3. Прогресс бар

```
При каждом ответе:
  progress = (currentQuestion / totalQuestions) * 100%

Анимация:
  - Плавное заполнение от старого к новому значению
  - Duration: 0.3s
```

### 4. Real-time Leaderboard

```
WebSocket connection к /ws/leaderboard/:quizId
  ↓
На сервере: QuizCompletedEvent
  ↓
Server broadcasts новый leaderboard
  ↓
Client получает update
  ↓
Анимация:
  - Если ваша позиция изменилась: highlight + sound
  - Если новый игрок в топ-10: появление с анимацией
  - Re-sort списка с transition
```

---

## Edge Cases & Error Handling

### 1. Пользователь закрывает TMA во время квиза

**Проблема:** Незавершенная сессия

**Решение:**
```
При возврате в TMA:
  1. Проверить наличие активной сессии (GET /quiz/session/:id)
  2. Если есть active session:
     - Показать modal: "You have unfinished quiz. Continue or start over?"
     - [Continue] → восстановить состояние (currentQuestion, score, timeLeft)
     - [Start Over] → abandon текущую сессию, создать новую
  3. Если нет active session → обычный флоу
```

### 2. Истекло время

**Проблема:** Таймер дошел до 0

**Решение:**
```
timeLeft === 0:
  1. Автоматически завершить квиз (POST /quiz/session/:id/complete)
  2. Подсчитать очки по отвеченным вопросам
  3. Показать Results с пометкой "Time's up!"
  4. Leaderboard обновляется как обычно
```

### 3. Потеря интернета во время квиза

**Проблема:** Нет связи с backend

**Решение:**
```
Network error:
  1. Показать toast: "Connection lost. Retrying..."
  2. Retry request (3 попытки с exponential backoff)
  3. Если все retry failed:
     - Показать modal: "Connection lost. Please check your internet."
     - Сохранить состояние локально (localStorage)
     - [Retry] button
     - При восстановлении связи → sync с backend
```

### 4. Backend вернул ошибку

**Проблема:** 500 Internal Server Error

**Решение:**
```
API Error:
  1. Rollback optimistic UI (если был)
  2. Показать error toast: "Something went wrong. Please try again."
  3. Log error to analytics
  4. Не менять state пользователя
  5. [Retry] option
```

### 5. Пользователь пытается ответить дважды на один вопрос

**Проблема:** Double submission

**Решение:**
```
Frontend:
  - Disable все answer buttons после первого клика
  - Ignore последующие клики

Backend:
  - Проверка в Session.SubmitAnswer():
    if questionAlreadyAnswered(questionID) {
      return ErrQuestionAlreadyAnswered
    }
```

### 6. Quiz был удален/изменен во время прохождения

**Проблема:** Quiz больше не существует или вопросы изменились

**Решение:**
```
Backend:
  - Сессия хранит snapshot вопросов при старте (не ссылку на Quiz)
  - Изменения в Quiz не влияют на активные сессии

Frontend:
  - Если QuizNotFound при старте:
    Show error: "This quiz is no longer available."
    Redirect to Quiz List
```

### 7. Пустой leaderboard

**Проблема:** Никто еще не прошел квиз

**Решение:**
```
Empty state:
  ┌────────────────────────────────────┐
  │                                    │
  │         🏆                         │
  │                                    │
  │   No scores yet!                   │
  │   Be the first to complete         │
  │   this quiz.                       │
  │                                    │
  │   [Start Quiz]                     │
  │                                    │
  └────────────────────────────────────┘
```

### 8. Validation errors

**Проблема:** Некорректные данные от пользователя

**Решение:**
```
Validation на frontend:
  - Quiz ID format (UUID)
  - Answer ID format
  - Non-empty fields

Validation на backend:
  - Повторная проверка всех данных
  - Return 400 Bad Request с понятным сообщением

Frontend обработка:
  Show inline error: "Invalid answer selected. Please try again."
```

---

## Navigation Flow

### Bottom Tab Bar (главная навигация)

```
┌────────────────────────────────────────────┐
│                                            │
│              [Content]                     │
│                                            │
└────────────────────────────────────────────┘
      ↑                ↑                ↑
   [🏠 Home]    [🏆 Leaderboard]   [👤 Profile]
```

**Routes:**
- `/` - Home (Quiz List)
- `/leaderboard/:quizId?` - Leaderboard (global или конкретного квиза)
- `/profile` - User Profile

### In-Quiz Navigation

**Во время прохождения квиза:**
- **Back button** - показать confirmation modal:
  ```
  Are you sure you want to quit?
  Your progress will be lost.

  [Cancel]  [Quit Quiz]
  ```
- **Home button (tab bar)** - то же самое
- **Minimize TMA** - сессия сохраняется как active

**После завершения квиза:**
- Back button → Quiz List
- Leaderboard button → Leaderboard этого квиза
- Try Again → новая сессия

---

## Animations & Transitions

### Page Transitions
- **Slide right** - при переходе вглубь (Quiz List → Quiz Details)
- **Slide left** - при возврате назад
- **Fade** - для модалов и overlays

### Micro-interactions
1. **Answer selection:**
   - Scale: 0.95 (tap feedback)
   - Color transition: 0.2s

2. **Correct/Incorrect feedback:**
   - Border color transition: 0.3s
   - Checkmark/Cross appearance: scale from 0 to 1
   - Sound effect (если Telegram позволяет)

3. **Score update:**
   - Count-up animation (от старого к новому)
   - Duration: 0.5s
   - Easing: ease-out

4. **Progress bar:**
   - Width transition: 0.3s ease-out

5. **Timer warning:**
   - Pulse animation (< 10s)
   - Color transition: green → yellow → red

6. **Leaderboard position change:**
   - Row highlight: fade in/out
   - Position swap: translate Y with spring physics

---

## Mobile-First Considerations

### Touch Targets
- **Minimum size:** 44x44px (Apple HIG)
- **Answer buttons:** Full-width, 56px height
- **Spacing:** 12px между кнопками

### Responsive Text
- **Quiz title:** 24px (heading)
- **Question:** 18px (body large)
- **Answers:** 16px (body)
- **Metadata:** 14px (caption)

### Safe Areas
- **Top:** Учитывать Telegram header
- **Bottom:** Учитывать tab bar (если есть)
- **Padding:** 16px по бокам

### Performance
- **Lazy loading:** Leaderboard pagination
- **Optimistic UI:** Мгновенный feedback на действия
- **Skeleton screens:** Во время загрузки
- **Image optimization:** WebP format, lazy load

---

## Theme Integration

### Telegram Theme Variables
```typescript
// Используем цвета из Telegram theme
const theme = {
  bg: window.Telegram.WebApp.themeParams.bg_color,
  text: window.Telegram.WebApp.themeParams.text_color,
  button: window.Telegram.WebApp.themeParams.button_color,
  buttonText: window.Telegram.WebApp.themeParams.button_text_color,
  // ...
}
```

### Dark/Light Mode
- **Автоматическое определение** через Telegram SDK
- **Динамические цвета** для всех компонентов
- **Тестирование** в обоих режимах

---

## Accessibility

### ARIA Labels
- Кнопки: `aria-label="Answer A: Venus"`
- Progress: `aria-label="Question 3 of 10"`
- Timer: `aria-live="polite"` для screen readers

### Keyboard Navigation
- Tab navigation по кнопкам
- Enter для выбора ответа
- ESC для отмены/выхода

### Color Contrast
- **WCAG AA compliance** (минимум 4.5:1)
- **Не только цвет** для feedback (иконки ✓/✗)

---

## Analytics Events

### Tracking Points

```typescript
// Quiz List
trackEvent('quiz_list_viewed')
trackEvent('quiz_card_clicked', { quizId })

// Quiz Details
trackEvent('quiz_details_viewed', { quizId })
trackEvent('quiz_started', { quizId, userId })

// Quiz Play
trackEvent('question_viewed', { quizId, questionId, questionNumber })
trackEvent('answer_submitted', { quizId, questionId, answerId, isCorrect, timeTaken })
trackEvent('quiz_completed', { quizId, score, rank, timeTaken })
trackEvent('quiz_abandoned', { quizId, questionNumber })

// Leaderboard
trackEvent('leaderboard_viewed', { quizId })
trackEvent('leaderboard_shared')

// Profile
trackEvent('profile_viewed')
trackEvent('achievement_unlocked', { achievementId })
```

---

## Implementation Checklist

### Phase 1: Core Flow (MVP)
- [ ] Quiz List screen
- [ ] Quiz Details screen
- [ ] Quiz Play (question flow)
- [ ] Results screen
- [ ] Basic Leaderboard
- [ ] Navigation (tab bar)

### Phase 2: Enhanced UX
- [ ] Animations & transitions
- [ ] Timer with visual feedback
- [ ] Sound effects
- [ ] Error handling
- [ ] Loading states
- [ ] Empty states

### Phase 3: Advanced Features
- [ ] Real-time leaderboard (WebSocket)
- [ ] User profile
- [ ] Achievements system
- [ ] Session recovery
- [ ] Offline support
- [ ] Share results

### Phase 4: Polish
- [ ] Accessibility
- [ ] Performance optimization
- [ ] Analytics integration
- [ ] A/B testing
- [ ] Localization

---

**Дата создания:** 2026-01-18
**Версия:** 1.0
**Проект:** Quiz Sprint TMA
