 📋 План реализации UI для Daily Challenge и Solo Marathon

  Фаза 1: Подготовка инфраструктуры (1-2 дня)

  1.1 Генерация TypeScript типов из Swagger

  cd tma
  pnpm run generate:all
  - ✅ Проверить, что типы для Daily Challenge сгенерированы
  - ✅ Проверить, что типы для Marathon сгенерированы
  - ✅ Проверить Vue Query hooks (useStartDailyChallenge, useSubmitDailyAnswer, etc.)

  1.2 Создание Composables

  Файлы:
  - tma/src/composables/useDailyChallenge.ts - логика Daily Challenge
  - tma/src/composables/useMarathon.ts - логика Marathon
  - tma/src/composables/useGameTimer.ts - общий таймер вопросов
  - tma/src/composables/useStreaks.ts - система серий

  Задачи:
  - Управление состоянием игры
  - Таймеры
  - Локальное хранение прогресса (localStorage)
  - Обработка событий (ответы, паузы, завершение)

  ---
  Фаза 2: Daily Challenge UI (3-4 дня)

  2.1 Переработка главного экрана (HomeView.vue)

  Текущее состояние: Есть базовая зона Daily Challenge

  Что добавить:
  <DailyChallengeCard
    :status="dailyStatus"
    :quiz="dailyQuiz"
    :streak="playerStreak"
    @start="startDaily"
  />

  Состояния:
  - available - можно начать (показать кнопку START)
  - completed - уже прошёл (показать результат + лидерборд)
  - in_progress - есть незавершенная игра (продолжить)
  - loading / error

  2.2 Создание новых View компонентов

  Файлы:
  tma/src/views/
  ├── DailyChallenge/
  │   ├── DailyChallengeIntroView.vue      # Этап 2: Информация перед стартом
  │   ├── DailyChallengePlayView.vue       # Этап 3: Игровой процесс
  │   ├── DailyChallengeResultsView.vue    # Этап 5: Финальный экран
  │   └── DailyChallengeReviewView.vue     # Этап 6: Разбор ошибок

  2.3 Создание UI компонентов

  Файлы:
  tma/src/components/DailyChallenge/
  ├── DailyChallengeCard.vue           # Карточка на главной
  ├── DailyChallengeTimer.vue          # Таймер до сброса (⏰ До обновления: 14:32:08)
  ├── DailyChallengeStreak.vue         # 🔥 Серия: 5 дней подряд
  ├── DailyChallengeQuestion.vue       # Вопрос БЕЗ индикации правильности
  ├── DailyChallengeAnswerFeedback.vue # ✓ Ответ принят (без правильности)
  ├── DailyChallengeLeaderboard.vue    # Глобальный лидерборд
  └── DailyChallengeReviewAnswer.vue   # Разбор одного вопроса

  2.4 Роутинг

  // tma/src/router/index.ts
  {
    path: '/daily-challenge',
    children: [
      { path: '', name: 'daily-challenge-intro', component: DailyChallengeIntroView },
      { path: 'play', name: 'daily-challenge-play', component: DailyChallengePlayView },
      { path: 'results', name: 'daily-challenge-results', component: DailyChallengeResultsView },
      { path: 'review', name: 'daily-challenge-review', component: DailyChallengeReviewView }
    ]
  }

  2.5 Ключевые особенности реализации

  Важно:
  - ❌ НЕ показывать правильность ответа до конца всех 10 вопросов
  - ✅ Сохранять прогресс локально (на случай закрытия приложения)
  - ✅ Показывать таймер обратного отсчёта до сброса (00:00 UTC)
  - ✅ Система серий с milestone анимациями
  - ✅ Real-time обновление лидерборда после завершения

  API последовательность:
  // 1. Проверка статуса
  GET /api/v1/daily-challenge/status?playerId={id}

  // 2. Старт
  POST /api/v1/daily-challenge/start { playerId }
  // Ответ: { gameId, questions: [{ id, text, answers }], ... }

  // 3. Ответы (10 раз)
  POST /api/v1/daily-challenge/{gameId}/answer {
    questionId, answerId, playerId, timeTaken
  }
  // Ответ: { isLastQuestion, nextQuestion?, results? }

  // 4. Просмотр результатов
  GET /api/v1/daily-challenge/status?playerId={id}

  // 5. Лидерборд
  GET /api/v1/daily-challenge/leaderboard?date={date}&limit=100

  ---
  Фаза 3: Solo Marathon UI (3-4 дня)

  3.1 Создание View компонентов

  Файлы:
  tma/src/views/Marathon/
  ├── MarathonHomeView.vue             # Этап 1: Главный экран
  ├── MarathonCategorySelectView.vue   # Этап 2: Выбор категории
  ├── MarathonPlayView.vue             # Этап 3: Игровой процесс
  ├── MarathonGameOverView.vue         # Этап 6: Game Over
  └── MarathonLeaderboardView.vue      # Лидерборд

  3.2 Создание UI компонентов

  Файлы:
  tma/src/components/Marathon/
  ├── MarathonLivesIndicator.vue       # ❤️ ❤️ ❤️ (жизни)
  ├── MarathonStreakCounter.vue        # 🎯 Вопросов подряд: 23
  ├── MarathonQuestion.vue             # Вопрос с таймером
  ├── MarathonAnswerFeedback.vue       # ✅ ВЕРНО! или ❌ НЕВЕРНО!
  ├── MarathonHintsPanel.vue           # 💡 50/50, +10сек, Skip
  ├── MarathonRecordProgress.vue       # Прогресс до рекорда
  ├── MarathonCategoryCard.vue         # Карточка категории
  └── MarathonNewRecordCelebration.vue # 🎉 Новый рекорд!

  3.3 Роутинг

  {
    path: '/marathon',
    children: [
      { path: '', name: 'marathon-home', component: MarathonHomeView },
      { path: 'category', name: 'marathon-category', component: MarathonCategorySelectView },
      { path: 'play', name: 'marathon-play', component: MarathonPlayView },
      { path: 'game-over', name: 'marathon-gameover', component: MarathonGameOverView },
      { path: 'leaderboard', name: 'marathon-leaderboard', component: MarathonLeaderboardView }
    ]
  }

  3.4 Ключевые особенности реализации

  Важно:
  - ✅ Система жизней с восстановлением (⏰ +1 жизнь через: 2:34:12)
  - ✅ Адаптивная сложность (растёт с количеством правильных ответов)
  - ✅ Система подсказок (50/50, +10сек, Skip, Hint)
  - ✅ Прогресс-бар до рекорда (Текущий: 38 │████████████░░░░│ Рекорд: 47)
  - ✅ Milestone анимации (каждые 10, 25 вопросов)
  - ✅ Сохранение незавершённой игры

  API последовательность:
  // 1. Проверка статуса (жизни, рекорды)
  GET /api/v1/marathon/status?playerId={id}

  // 2. Старт
  POST /api/v1/marathon/start {
    playerId, categoryId
  }
  // Ответ: { gameId, currentQuestion, lives, hints, personalBest }

  // 3. Ответы (до ошибки или выхода)
  POST /api/v1/marathon/{gameId}/answer {
    questionId, answerId, playerId, timeTaken
  }
  // Ответ: {
  //   isCorrect,
  //   nextQuestion?,
  //   gameOver?,
  //   currentStreak,
  //   score
  // }

  // 4. Использование подсказки
  POST /api/v1/marathon/{gameId}/hint {
    hintType: "fifty_fifty" | "extra_time" | "skip"
  }

  // 5. Завершение / Game Over
  DELETE /api/v1/marathon/{gameId}
  // Или автоматически при потере всех жизней

  // 6. Личные рекорды
  GET /api/v1/marathon/personal-bests?playerId={id}

  // 7. Лидерборд
  GET /api/v1/marathon/leaderboard?categoryId={id}&limit=100

  ---
  Фаза 4: Обновление главного экрана (HomeView) (1 день)

  4.1 Добавление карточек режимов

  <template>
    <div class="home-container">
      <!-- Zone 1: Daily Challenge -->
      <DailyChallengeCard />

      <!-- Zone 2: Game Modes -->
      <section class="game-modes">
        <h3>🎮 Game Modes</h3>

        <!-- Solo Marathon -->
        <GameModeCard
          title="Solo Marathon"
          icon="🏃"
          description="Answer until first mistake"
          :lives="marathonLives"
          @click="goToMarathon"
        />

        <!-- Coming Soon modes -->
        <GameModeCard
          title="Quick Duel"
          icon="⚔️"
          description="1v1 real-time battle"
          :disabled="true"
          badge="Coming Soon"
        />

        <GameModeCard
          title="Party Mode"
          icon="🎉"
          description="Multiplayer quiz party"
          :disabled="true"
          badge="Coming Soon"
        />
      </section>

      <!-- Zone 3: Categories (existing) -->
    </div>
  </template>

  Компонент:
  <!-- tma/src/components/GameModeCard.vue -->
  <template>
    <div
      class="game-mode-card"
      :class="{ disabled }"
      @click="!disabled && $emit('click')"
    >
      <div class="mode-header">
        <span class="mode-icon">{{ icon }}</span>
        <div>
          <h4>{{ title }}</h4>
          <p>{{ description }}</p>
        </div>
      </div>

      <div v-if="lives !== undefined" class="mode-meta">
        <span>❤️ {{ lives }} lives</span>
      </div>

      <span v-if="badge" class="mode-badge">{{ badge }}</span>
    </div>
  </template>

  ---
  Фаза 5: Общие компоненты и утилиты (1-2 дня)

  5.1 Создание переиспользуемых компонентов

  Файлы:
  tma/src/components/shared/
  ├── QuestionCard.vue           # Общий компонент вопроса
  ├── AnswerButton.vue           # Кнопка ответа
  ├── GameTimer.vue              # Таймер (⏱️ 15 сек)
  ├── ScoreDisplay.vue           # Счёт очков
  ├── ProgressBar.vue            # Прогресс-бар
  ├── StreakBadge.vue            # 🔥 Streak indicator
  ├── LeaderboardTable.vue       # Таблица лидерборда
  └── CelebrationAnimation.vue   # Конфетти, анимации

  5.2 Утилиты

  Файлы:
  tma/src/utils/
  ├── gameUtils.ts               # Подсчёт очков, бонусов
  ├── timeUtils.ts               # Форматирование времени
  ├── streakUtils.ts             # Логика серий
  ├── storageUtils.ts            # LocalStorage для прогресса
  └── animationUtils.ts          # Trigger анимаций

  ---
  Фаза 6: Тестирование и полировка (2-3 дня)

  6.1 Функциональное тестирование

  - ✅ Daily Challenge полный цикл (start → play → results)
  - ✅ Marathon полный цикл (start → play → game over)
  - ✅ Сохранение прогресса при закрытии
  - ✅ Таймеры работают корректно
  - ✅ Серии считаются правильно
  - ✅ Лидерборды обновляются

  6.2 UX полировка

  - Анимации переходов между экранами
  - Haptic feedback (вибрация при ответах)
  - Звуковые эффекты (опционально)
  - Optimistic UI updates
  - Loading states и error handling
  - Telegram Mini App интеграция (BackButton, MainButton)

  6.3 Responsive design

  - Тестирование на разных размерах экранов
  - Поддержка landscape режима
  - Dark/Light theme

  ---
  📁 Итоговая структура файлов

  tma/src/
  ├── views/
  │   ├── HomeView.vue (обновлён)
  │   ├── DailyChallenge/
  │   │   ├── DailyChallengeIntroView.vue
  │   │   ├── DailyChallengePlayView.vue
  │   │   ├── DailyChallengeResultsView.vue
  │   │   └── DailyChallengeReviewView.vue
  │   └── Marathon/
  │       ├── MarathonHomeView.vue
  │       ├── MarathonCategorySelectView.vue
  │       ├── MarathonPlayView.vue
  │       ├── MarathonGameOverView.vue
  │       └── MarathonLeaderboardView.vue
  │
  ├── components/
  │   ├── GameModeCard.vue
  │   ├── DailyChallenge/
  │   │   ├── DailyChallengeCard.vue
  │   │   ├── DailyChallengeTimer.vue
  │   │   ├── DailyChallengeStreak.vue
  │   │   ├── DailyChallengeQuestion.vue
  │   │   ├── DailyChallengeAnswerFeedback.vue
  │   │   ├── DailyChallengeLeaderboard.vue
  │   │   └── DailyChallengeReviewAnswer.vue
  │   ├── Marathon/
  │   │   ├── MarathonLivesIndicator.vue
  │   │   ├── MarathonStreakCounter.vue
  │   │   ├── MarathonQuestion.vue
  │   │   ├── MarathonAnswerFeedback.vue
  │   │   ├── MarathonHintsPanel.vue
  │   │   ├── MarathonRecordProgress.vue
  │   │   ├── MarathonCategoryCard.vue
  │   │   └── MarathonNewRecordCelebration.vue
  │   └── shared/
  │       ├── QuestionCard.vue
  │       ├── AnswerButton.vue
  │       ├── GameTimer.vue
  │       ├── ScoreDisplay.vue
  │       ├── ProgressBar.vue
  │       ├── StreakBadge.vue
  │       ├── LeaderboardTable.vue
  │       └── CelebrationAnimation.vue
  │
  ├── composables/
  │   ├── useDailyChallenge.ts
  │   ├── useMarathon.ts
  │   ├── useGameTimer.ts
  │   ├── useStreaks.ts
  │   └── useGameState.ts
  │
  ├── utils/
  │   ├── gameUtils.ts
  │   ├── timeUtils.ts
  │   ├── streakUtils.ts
  │   ├── storageUtils.ts
  │   └── animationUtils.ts
  │
  └── router/index.ts (обновлён)

  ---
  ⏱️ Оценка времени

  | Фаза  | Задача                     | Время      |
  |-------|----------------------------|------------|
  | 1     | Подготовка инфраструктуры  | 1-2 дня    |
  | 2     | Daily Challenge UI         | 3-4 дня    |
  | 3     | Solo Marathon UI           | 3-4 дня    |
  | 4     | Обновление главного экрана | 1 день     |
  | 5     | Общие компоненты           | 1-2 дня    |
  | 6     | Тестирование и полировка   | 2-3 дня    |
  | Итого |                            | 11-16 дней |

  ---
  🚀 Порядок реализации (рекомендация)

  Sprint 1: Daily Challenge (5-6 дней)

  1. Генерация API типов
  2. Composables для Daily Challenge
  3. Все View компоненты Daily Challenge
  4. Все UI компоненты Daily Challenge
  5. Интеграция с API
  6. Тестирование

  Sprint 2: Solo Marathon (5-6 дней)

  7. Composables для Marathon
  8. Все View компоненты Marathon
  9. Все UI компоненты Marathon
  10. Система жизней и подсказок
  11. Интеграция с API
  12. Тестирование

  Sprint 3: Интеграция и полировка (2-3 дня)

  13. Обновление HomeView с карточками режимов
  14. Общие компоненты
  15. Анимации и UX полировка
  16. Финальное тестирование
  17. Deploy

  ---
  Готов начинать реализацию? Могу помочь с любой фазой - от генерации типов до создания конкретных компонентов! 🚀