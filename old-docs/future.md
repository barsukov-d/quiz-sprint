## Future Enhancements (Inspired by Trivia Crack)

Следующие механики рассматриваются для будущих версий приложения. Они отсортированы по приоритету и влиянию на engagement.

### Phase 1: 1v1 Асинхронные дуэли (Duel Mode) 🎯

**Бизнес-цель:**
- Увеличить retention через социальные взаимодействия
- Мотивировать пользователей возвращаться проверить результаты дуэлей
- Leveraging Telegram social graph для виральности

**New Bounded Context: Duel Context (Supporting)**

#### Aggregate: DuelSession

**Ответственность:**
- Управление асинхронным соревнованием между двумя игроками
- Отслеживание состояния дуэли (кто прошел, кто ждет)
- Определение победителя по сумме очков
- Отправка уведомлений через Telegram

**Entities внутри:**
- `ParticipantResult` - результат одного участника (sessionID, score, completedAt)

**Value Objects:**
- `DuelID` - уникальный идентификатор дуэли
- `ChallengerID` - пользователь, создавший вызов
- `OpponentID` - пользователь, принявший вызов
- `DuelStatus` - статус дуэли
- `WinnerID` - победитель (nullable)
- `QuizSnapshot` - набор вопросов (фиксирован при создании)

**DuelStatus States:**
```
waiting_for_opponent   → Challenger завершил, ждет opponent
both_completed         → Оба завершили, определен победитель
expired                → Opponent не ответил в течение 48 часов
```

**Бизнес-правила (Invariants):**
1. Оба участника проходят одинаковый набор вопросов
2. Вопросы фиксируются при создании дуэли (snapshot)
3. Нельзя создать дуэль с самим собой
4. Нельзя создать дуэль, если у opponent уже есть активная дуэль с вами
5. Дуэль истекает через 48 часов после создания
6. Победитель определяется по правилу: higher score wins, при равенстве - faster completion wins
7. Winner получает бонус +20% к очкам в leaderboard

**Domain Events:**
- `DuelCreatedEvent` - когда challenger создает дуэль
- `DuelAcceptedEvent` - когда opponent начинает квиз
- `DuelCompletedEvent` - когда оба завершили, есть победитель
- `DuelExpiredEvent` - когда opponent не ответил в срок

**Use Cases:**
```go
CreateDuelUseCase(challengerID, opponentID, quizID) → (duelID)
  • Проверяет отсутствие активной дуэли между игроками
  • Создает snapshot вопросов квиза
  • Challenger автоматически проходит квиз (или сразу после создания)
  • Отправляет Telegram notification opponent
  • Event: DuelCreatedEvent

AcceptDuelUseCase(duelID, opponentID) → (session)
  • Opponent начинает прохождение квиза
  • Event: DuelAcceptedEvent

OnQuizCompletedInDuel(duelID, participantID, score) → (winner?)
  • Обновляет результат участника
  • Если оба завершили, определяет победителя
  • Отправляет Telegram notifications обоим
  • Event: DuelCompletedEvent

GetUserDuelsUseCase(userID, status?) → (duels[])
  • Возвращает активные и завершенные дуэли
  • Фильтрация по статусу (waiting, completed)

ExpireDuelsUseCase() → void
  • Cron job каждые 30 минут
  • Помечает просроченные дуэли как expired
```

**Repository Interface:**
```go
type DuelRepository interface {
    Save(duel *DuelSession) error
    FindByID(duelID DuelID) (*DuelSession, error)
    FindActiveByParticipants(userID1, userID2 UserID) (*DuelSession, error)
    FindByUserID(userID UserID, status DuelStatus) ([]*DuelSession, error)
    FindExpired(olderThan timestamp) ([]*DuelSession, error)
}
```

**Leaderboard Integration:**
- Winner получает +20% bonus к очкам при записи в leaderboard
- Новое поле `LeaderboardEntry.DuelBonus` (boolean)

---

### Phase 2: Badge Collection (Category Mastery) 👑

**Бизнес-цель:**
- Мотивировать прохождение квизов в разных категориях
- Визуальная коллекция достижений в профиле
- Увеличить completion rate

**New Supporting Domain: Achievements Context**

#### Aggregate: Achievement

**Ответственность:**
- Определение условий получения badge
- Отслеживание прогресса пользователя к achievement
- Unlocking badge при выполнении условий

**Value Objects:**
- `AchievementID` - уникальный идентификатор
- `AchievementType` - тип достижения
- `AchievementTitle` - название (например, "General Knowledge Master")
- `AchievementIcon` - эмодзи или иконка
- `UnlockCriteria` - условия получения

**AchievementType Enum:**
```go
category_master     // Пройти N квизов в категории с avg score >= X%
first_quiz          // Завершить первый квиз
speed_demon         // Пройти квиз с avg answer time < 5 секунд
perfectionist       // Пройти квиз со 100% правильных ответов
streak_champion     // Daily streak >= 7 дней
tournament_winner   // Победить в weekly tournament (будущая фича)
duel_champion       // Выиграть 10 дуэлей подряд
```

**UnlockCriteria Structure:**
```go
type UnlockCriteria struct {
    Type            string  // "category_quiz_count", "avg_score", "streak"
    CategoryID      *UUID   // Для category-specific achievements
    RequiredCount   int     // Например, 5 квизов
    RequiredScore   float64 // Например, 80%
}
```

**Бизнес-правила:**
1. Achievement нельзя забрать после получения (immutable unlock)
2. Прогресс отслеживается в реальном времени
3. Notification при unlock через Telegram
4. Badge отображается в профиле с датой получения

**Domain Events:**
- `AchievementUnlockedEvent` - когда пользователь получает badge

**Use Cases:**
```go
GetUserAchievementsUseCase(userID) → (achievements[], progress[])
  • Возвращает unlocked badges + прогресс к остальным
  • Сортировка: unlocked first, затем by progress

CheckAchievementProgressUseCase(userID, event) → (unlocked[])
  • Event handler для QuizCompletedEvent
  • Проверяет все условия achievements
  • Если условия выполнены, unlock badge
  • Event: AchievementUnlockedEvent

GetCategoryMasteryUseCase(userID, categoryID) → (progress)
  • Прогресс к Category Master badge
  • Например: 8/10 quizzes, avg score 82%
```

**Schema (PostgreSQL):**
```sql
CREATE TABLE achievements (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(10),
    criteria JSONB NOT NULL
);

CREATE TABLE user_achievements (
    user_id UUID NOT NULL,
    achievement_id UUID NOT NULL,
    unlocked_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, achievement_id)
);

CREATE TABLE achievement_progress (
    user_id UUID NOT NULL,
    achievement_id UUID NOT NULL,
    current_value INT NOT NULL,
    metadata JSONB, -- Доп. данные (например, category_id)
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, achievement_id)
);
```

---

### Phase 3: Power-Ups (Бустеры) 💪

**Бизнес-цель:**
- Добавить стратегический элемент в геймплей
- Потенциальная монетизация (покупка бустеров)
- Реварды за Daily Streak и achievements

**Extension of Quiz Taking Context**

#### Value Object: PowerUp

**PowerUpType Enum:**
```go
fifty_fifty     // Убирает 2 неправильных ответа
extra_time      // Добавляет +10 секунд к таймеру вопроса
skip_question   // Пропускает вопрос без штрафа
freeze_time     // Останавливает таймер на 5 секунд
```

**Structure:**
```go
type PowerUp struct {
    Type      PowerUpType
    Count     int       // Доступное количество
    UsedAt    *int      // Индекс вопроса, где использован (nullable)
}

type PowerUpInventory struct {
    UserID    UserID
    PowerUps  map[PowerUpType]int  // Type → Count
}
```

**Бизнес-правила:**
1. Можно использовать только 1 power-up на вопрос
2. Нельзя использовать power-up после выбора ответа
3. 50/50 нельзя использовать, если осталось 2 варианта ответа
4. Skip не влияет на streak (не сбрасывает и не увеличивает)
5. Power-ups не восстанавливаются между сессиями

**How to Earn PowerUps:**
- Daily Quiz completion: +1 random power-up
- Daily Streak milestone (7 дней): +3 power-ups (на выбор)
- Achievement unlock: +2 power-ups
- (Будущее) Покупка за виртуальную валюту

**Domain Events:**
- `PowerUpUsedEvent` - когда игрок использует бустер
- `PowerUpEarnedEvent` - когда игрок получает бустер

**Use Cases:**
```go
GetUserPowerUpsUseCase(userID) → (inventory)
  • Возвращает доступные power-ups пользователя

UsePowerUpInSessionUseCase(sessionID, questionIndex, powerUpType) → (result)
  • Применяет эффект бустера к текущему вопросу
  • Уменьшает count в inventory
  • Для 50/50: возвращает список оставшихся вариантов
  • Для Extra Time: обновляет timeLimit вопроса
  • Для Skip: переходит к следующему вопросу
  • Event: PowerUpUsedEvent

AwardPowerUpUseCase(userID, powerUpType, count, reason) → void
  • Добавляет power-ups в inventory
  • Event: PowerUpEarnedEvent
  • Reasons: "daily_quiz", "streak_milestone", "achievement"
```

**Schema Extension:**
```sql
CREATE TABLE user_power_ups (
    user_id UUID NOT NULL,
    power_up_type VARCHAR(50) NOT NULL,
    count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, power_up_type)
);

CREATE TABLE power_up_usage_log (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    user_id UUID NOT NULL,
    power_up_type VARCHAR(50) NOT NULL,
    question_index INT NOT NULL,
    used_at TIMESTAMP NOT NULL
);
```

**QuizSession Changes:**
```go
type QuizSession struct {
    // ... existing fields
    UsedPowerUps []PowerUpUsage  // Лог использованных бустеров
}

type PowerUpUsage struct {
    Type          PowerUpType
    QuestionIndex int
    UsedAt        int64
}

// New method
func (s *QuizSession) UsePowerUp(questionIndex int, powerUpType PowerUpType) error {
    // Validate: нет уже использованного бустера на этом вопросе
    // Validate: вопрос еще не отвечен
    // Apply effect, log usage
}
```

---

### Phase 4: Weekly Mini-Tournaments 🏆

**Бизнес-цель:**
- FOMO механика (fear of missing out)
- Увеличить weekly active users
- Community building
- Retention через recurring events

**Extension of Leaderboard Context**

#### Aggregate: Tournament

**Ответственность:**
- Определение правил турнира
- Отслеживание участников
- Подсчет результатов
- Награждение победителей

**Value Objects:**
- `TournamentID` - уникальный идентификатор
- `TournamentTitle` - название (например, "Programming Week")
- `CategoryID` - категория турнира
- `StartDate`, `EndDate` - временные рамки
- `TournamentStatus` - статус
- `EligibleQuizIDs` - список квизов, входящих в турнир

**TournamentStatus States:**
```
upcoming     → Анонсирован, но еще не начался
active       → Идет прием результатов
completed    → Завершен, определены победители
archived     → Старый турнир (> 30 дней)
```

**Бизнес-правила:**
1. Турнир длится ровно 7 дней (Monday 00:00 → Sunday 23:59 UTC)
2. В турнире участвуют только квизы определенной категории
3. Учитывается лучший score каждого пользователя по каждому квизу
4. Итоговый tournament score = сумма лучших scores по всем квизам
5. Минимум 3 завершенных квиза для попадания в leaderboard
6. Top 3 получают special badge
7. Можно переигрывать квизы для улучшения score

**Domain Events:**
- `TournamentStartedEvent` - когда турнир начинается
- `TournamentCompletedEvent` - когда турнир завершается
- `TournamentParticipantJoinedEvent` - когда пользователь завершает первый квиз турнира

**Use Cases:**
```go
GetActiveTournamentUseCase() → (tournament?)
  • Возвращает текущий активный турнир
  • Null если нет активного

GetTournamentLeaderboardUseCase(tournamentID, limit) → (entries[])
  • Leaderboard конкретного турнира
  • Sorted by tournament_score DESC

GetUserTournamentProgressUseCase(userID, tournamentID) → (progress)
  • Сколько квизов завершено из eligible
  • Текущий tournament score
  • Позиция в leaderboard

CreateWeeklyTournamentUseCase(categoryID, startDate) → (tournamentID)
  • Admin use case
  • Создает турнир на неделю
  • Автоматически выбирает квизы категории

FinalizeTournamentUseCase(tournamentID) → (winners[])
  • Cron job (запускается в Monday 00:05 UTC)
  • Определяет Top 3
  • Награждает badge "Tournament Winner"
  • Event: TournamentCompletedEvent
```

**Schema:**
```sql
CREATE TABLE tournaments (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    category_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE tournament_quizzes (
    tournament_id UUID NOT NULL,
    quiz_id UUID NOT NULL,
    PRIMARY KEY (tournament_id, quiz_id)
);

CREATE TABLE tournament_leaderboard (
    tournament_id UUID NOT NULL,
    user_id UUID NOT NULL,
    tournament_score INT NOT NULL,  -- Сумма лучших scores
    quizzes_completed INT NOT NULL, -- Сколько квизов завершено
    rank INT NOT NULL,
    PRIMARY KEY (tournament_id, user_id)
);

CREATE INDEX idx_tournament_leaderboard_rank
ON tournament_leaderboard(tournament_id, rank);
```

**Event Handler:**
```go
OnQuizCompleted(event QuizCompletedEvent) {
    // Проверить, входит ли quiz в активный турнир
    tournament := GetActiveTournament()
    if tournament == nil || !tournament.ContainsQuiz(event.QuizID) {
        return
    }

    // Обновить tournament leaderboard
    UpdateTournamentLeaderboard(tournament.ID, event.UserID, event.Score)
    RecalculateRanks(tournament.ID)
}
```

---

### Phase 5: Category Roulette (Mixed Quiz Mode) 🎰

**Бизнес-цель:**
- Добавить разнообразие в геймплей
- Тестировать широту знаний пользователей
- Вариация Daily Challenge для advanced users

**Extension of Quiz Catalog Context**

#### Special Quiz Type: Mixed Quiz

**Характеристики:**
- Каждый вопрос из случайной категории
- Всего 10 вопросов (по 1-2 из разных категорий)
- Повышенный бонус: +50% к очкам
- Сложность: только для пользователей с 10+ завершенными квизами

**Generation Algorithm:**
```go
func GenerateMixedQuiz(userID UserID) *Quiz {
    // 1. Выбрать 5 случайных категорий
    categories := SelectRandomCategories(5)

    // 2. Из каждой категории взять по 2 случайных вопроса
    questions := []Question{}
    for _, cat := range categories {
        qs := GetRandomQuestionsFromCategory(cat.ID, 2)
        questions = append(questions, qs...)
    }

    // 3. Перемешать вопросы
    Shuffle(questions)

    // 4. Создать quiz с special flag
    quiz := NewQuiz("Mixed Quiz", "Test your knowledge across categories!", questions)
    quiz.IsMixed = true
    quiz.ScoreMultiplier = 1.5

    return quiz
}
```

**Бизнес-правила:**
1. Mixed Quiz генерируется on-demand (не хранится в БД)
2. Score учитывается в общем leaderboard с multiplier
3. Нельзя начать Mixed Quiz, пока не завершено 10+ обычных квизов
4. Mixed Quiz доступен 1 раз в день (как Daily Challenge)

**Use Cases:**
```go
GenerateMixedQuizUseCase(userID) → (quiz)
  • Проверяет eligibility (10+ completed quizzes)
  • Проверяет daily limit (1 per day)
  • Генерирует случайный набор вопросов
  • Возвращает ephemeral quiz (не сохраняется в quiz catalog)

StartMixedQuizSessionUseCase(userID, quizSnapshot) → (session)
  • Создает сессию с special flag: isMixed = true
  • Применяет score multiplier при подсчете очков
```

---

### Phase 6 (Low Priority): Random Opponent Matchmaking ⚔️

**Бизнес-цель:**
- Расширение Duel Mode
- Подбор соперника без необходимости знать друзей
- Fairness через skill-based matchmaking

**Extension of Duel Context**

#### Matchmaking Service

**Algorithm:**
- Rating calculation: `rating = (total_score / quizzes_completed)`
- Match users within ±15% rating range
- Matchmaking queue with 30-second timeout
- Fallback: если нет match, предложить Random Quiz

**Бизнес-правила:**
1. Matchmaking учитывает только пользователей онлайн (last_active < 5 min)
2. Нельзя играть с одним и тем же opponent чаще 1 раза в час
3. При отсутствии match в течение 30s → fallback на Random Quiz

**Use Cases:**
```go
JoinMatchmakingQueueUseCase(userID) → (matchID | timeout)
  • Добавляет пользователя в очередь
  • Ищет match в течение 30 секунд
  • Если найден → создает DuelSession
  • Если timeout → возвращает null

LeaveMatchmakingQueueUseCase(userID) → void
  • Удаляет из очереди
```

---

### Implementation Priority Matrix

| Feature | Impact (Engagement) | Complexity | Priority |
|---------|-------------------|------------|----------|
| **1v1 Duels** | ⭐⭐⭐⭐⭐ (Very High) | Medium | **P0** |
| **Badge Collection** | ⭐⭐⭐⭐ (High) | Low | **P1** |
| **Power-Ups** | ⭐⭐⭐⭐ (High) | Medium | **P2** |
| **Weekly Tournaments** | ⭐⭐⭐⭐ (High) | Medium | **P3** |
| **Category Roulette** | ⭐⭐⭐ (Medium) | Low | **P4** |
| **Random Matchmaking** | ⭐⭐ (Low) | High | **P5** |

---

### Dependencies Between Features

```
Phase 1: Duels
    ↓ (requires social mechanics)
Phase 2: Badges
    ↓ (badges можно давать за tournament wins)
Phase 4: Tournaments
    ↓ (power-ups как tournament rewards)
Phase 3: Power-Ups
    ↓ (power-ups в matchmaking для fairness)
Phase 6: Random Matchmaking
```

---

### Excluded Mechanics (Why NOT)

**User-Generated Questions:**
- ❌ Требует модерацию (spam, offensive content)
- ❌ Качество контента непредсказуемо
- ❌ Юридические риски (copyright)

**Real-Time Multiplayer (синхронный):**
- ❌ Высокая latency в TMA (WebSocket через Telegram unreliable)
- ❌ Требует оба игрока онлайн одновременно (плохо для retention)
- ❌ Сложная инфраструктура (WebSocket scaling)

**Paid Tournaments с денежными призами:**
- ❌ Юридические сложности (gambling laws в разных странах)
- ❌ Налогообложение выигрышей
- ❌ KYC/AML compliance
- ❌ Риск fraud

**Complex Progression Systems (уровни, XP, skill trees):**
- ❌ Может overwhelm casual пользователей
- ❌ Долгий onboarding
- ❌ Не подходит для casual TMA experience

---

## Changelog

**v1.5 (2026-01-21):**
- 🚀 **Добавлен раздел Future Enhancements!**
  - Описаны 6 фаз будущих фич, вдохновленных Trivia Crack
  - **Phase 1: 1v1 Асинхронные дуэли** - новый DuelSession aggregate, Telegram integration
  - **Phase 2: Badge Collection** - новый Achievements Context, прогресс-трекинг
  - **Phase 3: Power-Ups** - система бустеров с inventory и earning mechanics
  - **Phase 4: Weekly Tournaments** - расширение Leaderboard с FOMO механикой
  - **Phase 5: Category Roulette** - Mixed Quiz mode с multiplier +50%
  - **Phase 6: Random Matchmaking** - skill-based подбор соперников
  - Добавлены детальные DDD модели: aggregates, value objects, use cases, repositories
  - Описаны бизнес-правила и инварианты для каждой фичи
  - Implementation Priority Matrix и Dependencies между фичами
  - Excluded Mechanics с обоснованием почему не подходят

**v1.4 (2026-01-21):**
- 🚀 **Добавлена система Discovery и User Engagement!**
  - Добавлен новый **User Stats Domain** (Supporting) для отслеживания прогресса
  - **Daily Quiz**: новый use case `GetDailyQuizUseCase` для квиза дня
  - **Random Quiz**: `GetRandomQuizUseCase` с опциональной фильтрацией по категории
  - **Active Sessions**: `GetUserActiveSessionsUseCase` для восстановления прогресса
  - **Streak Tracking**: бизнес-логика для серий ежедневных активностей
  - Новые Value Objects: `CurrentStreak`, `LongestStreak`, `LastDailyQuizDate`
  - Поддержка мотивационной механики на главном экране (3 зоны: Daily, Quick Actions, Categories)

**v1.3 (2026-01-20):**
- 🚀 **Введена новая система начисления очков!**
  - Добавлен **бонус за скорость ответа** (Time Bonus).
  - Добавлен **бонус за серию правильных ответов** (Streak Bonus).
  - Обновлена логика `SubmitAnswerUseCase` и затронутые агрегаты (`Quiz`, `QuizSession`).
  - Обновлен Ubiquitous Language для отражения новых концепций.

**v1.2 (2026-01-20):**
- ✅ Добавлен `GetSessionResultsUseCase` для получения детальных результатов сессии
- Обновлен список Use Cases с реализованными фичами

**v1.1 (2026-01-18):**
- Добавлен Category aggregate и Use Cases

**v1.0 (2026-01-15):**
- Первоначальная версия документа

---

**Дата создания:** 2026-01-15
**Последнее обновление:** 2026-01-21
**Версия:** 1.5
**Методология:** Pragmatic DDD (по мотивам Vernon Vaughn IDDD)
**Проект:** Quiz Sprint TMA
