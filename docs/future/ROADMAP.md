# Product Roadmap - Future Enhancements

> **Для текущей реализации см.:** [`../current/domain.md`](../current/domain.md)
> **Для детальных спецификаций см.:** Старые `DOMAIN.md` и `USER_FLOW.md` (sections "Future Enhancements")

---

## 📋 Содержание

1. [Overview](#overview)
2. [Implementation Priority Matrix](#implementation-priority-matrix)
3. [Phase Summaries](#phase-summaries)
4. [Dependencies](#dependencies-between-features)
5. [Excluded Mechanics](#excluded-mechanics)

---

## Overview

Эти фичи вдохновлены успешной механикой **Trivia Crack** и адаптированы для Quiz Sprint TMA.

**Цель:** Увеличить engagement, retention и social interaction через:
- Асинхронные PvP дуэли
- Геймификацию (badges, power-ups)
- FOMO механики (tournaments)
- Разнообразие контента (mixed quizzes)

---

## Implementation Priority Matrix

| Feature | Impact (Engagement) | Complexity | Priority | Timeline |
|---------|-------------------|------------|----------|----------|
| **1v1 Duels** | ⭐⭐⭐⭐⭐ Very High | Medium | **P0** | 3-4 weeks |
| **Badge Collection** | ⭐⭐⭐⭐ High | Low | **P1** | 2-3 weeks |
| **Power-Ups** | ⭐⭐⭐⭐ High | Medium | **P2** | 3-4 weeks |
| **Weekly Tournaments** | ⭐⭐⭐⭐ High | Medium | **P3** | 2-3 weeks |
| **Category Roulette** | ⭐⭐⭐ Medium | Low | **P4** | 1-2 weeks |
| **Random Matchmaking** | ⭐⭐ Low | High | **P5** | 3-4 weeks |

**Критерии оценки:**
- **Impact:** Влияние на retention и daily active users
- **Complexity:** Инфраструктура, backend, frontend work
- **Priority:** Очередность реализации

---

## Phase Summaries

### Phase 1: 1v1 Асинхронные дуэли 🎯 (P0)

**Бизнес-цель:**
- Увеличить retention через social gameplay
- Мотивировать пользователей возвращаться проверить результаты

**Ключевые фичи:**
- Challenge friend или accept challenge
- Оба проходят один и тот же набор вопросов (snapshot)
- Победитель определяется по score (при равенстве - по времени)
- Winner получает +20% bonus к очкам в leaderboard
- Telegram notifications

**Новые Aggregates:**
- `DuelSession` (extends QuizSession concept)

**Domain Events:**
- `DuelCreatedEvent`, `DuelCompletedEvent`

**API Endpoints:**
- `POST /api/v1/duels`
- `POST /api/v1/duels/:id/accept`
- `GET /api/v1/duels?status=waiting|completed`

**UI Screens:**
- Duel Challenge screen
- Active Duels List
- Duel Results comparison

**Детали:** См. старый DOMAIN.md секция "Future Enhancements / Phase 1"

---

### Phase 2: Badge Collection 👑 (P1)

**Бизнес-цель:**
- Мотивировать прохождение квизов в разных категориях
- Визуальная коллекция достижений

**Ключевые фичи:**
- Achievement types: category_master, first_quiz, speed_demon, perfectionist, etc.
- Progress tracking к каждому badge
- Unlock notification
- Отображение в профиле

**Новый Supporting Domain:**
- `Achievements Context`

**Aggregate:**
- `Achievement` с UnlockCriteria

**Event Handler:**
- `OnQuizCompleted` → check achievement progress

**UI Screens:**
- Achievements screen в Profile
- Progress bars для locked badges

**Детали:** См. старый DOMAIN.md секция "Future Enhancements / Phase 2"

---

### Phase 3: Power-Ups 💪 (P2)

**Бизнес-цель:**
- Добавить стратегический элемент
- Потенциальная монетизация

**Ключевые фичи:**
- Power-Up types: 50/50, Extra Time, Skip Question, Freeze Time
- Inventory management
- Earning через Daily Quiz, Streaks, Achievements
- Использование во время квиза (1 per question)

**Extension of Quiz Taking Context:**
- `PowerUp` Value Object
- `PowerUpInventory` aggregate

**Бизнес-правила:**
- Только 1 power-up на вопрос
- Нельзя использовать после выбора ответа
- Skip не сбрасывает streak

**UI:**
- Power-up toolbar во время quiz
- Inventory в Profile

**Детали:** См. старый DOMAIN.md секция "Future Enhancements / Phase 3"

---

### Phase 4: Weekly Tournaments 🏆 (P3)

**Бизнес-цель:**
- FOMO механика для weekly active users
- Community building

**Ключевые фичи:**
- Еженедельный турнир (Monday-Sunday)
- Фиксированная категория + список eligible quizzes
- Tournament leaderboard (сумма лучших scores)
- Top 3 получают badge
- Минимум 3 квиза для попадания в leaderboard

**Extension of Leaderboard Context:**
- `Tournament` aggregate

**Бизнес-правила:**
- Длительность: ровно 7 дней
- Можно переигрывать для улучшения score
- Cron job для finalization

**UI:**
- Tournament banner на главной
- Tournament Hub (progress, leaderboard)

**Детали:** См. старый DOMAIN.md секция "Future Enhancements / Phase 4"

---

### Phase 5: Category Roulette 🎰 (P4)

**Бизнес-цель:**
- Добавить разнообразие
- Тестировать широту знаний

**Ключевые фичи:**
- Special quiz type: 10 вопросов из 5 случайных категорий
- Score multiplier +50%
- Доступен 1 раз в день
- Eligibility: 10+ completed quizzes

**Extension of Quiz Catalog:**
- Ephemeral quiz generation (не сохраняется в БД)

**UI:**
- Mixed Quiz card на главной
- Показ категории каждого вопроса

**Детали:** См. старый DOMAIN.md секция "Future Enhancements / Phase 5"

---

### Phase 6: Random Matchmaking ⚔️ (P5 - Low Priority)

**Бизнес-цель:**
- Расширение Duel Mode
- Не нужно знать друзей

**Ключевые фичи:**
- Skill-based matchmaking (±15% rating)
- 30-second queue timeout
- Cooldown: 1 час с одним opponent
- Fallback на Random Quiz

**Новый Service:**
- `MatchmakingService`

**Сложность:**
- Требует WebSocket для queue management
- Rating calculation system
- Edge cases (timeout, cancel)

**Детали:** См. старый DOMAIN.md секция "Future Enhancements / Phase 6"

---

## Dependencies Between Features

```
Phase 1: Duels (P0)
    ↓ (social mechanics foundation)
Phase 2: Badges (P1)
    ↓ (можно давать badges за tournament wins)
Phase 4: Tournaments (P3)
    ↓ (power-ups как rewards за tournaments)
Phase 3: Power-Ups (P2)
    ↓ (power-ups в matchmaking для balance)
Phase 6: Matchmaking (P5)
```

**Recommended Order:**
1. Duels (P0) - наибольший impact, foundation для PvP
2. Badges (P1) - простая геймификация
3. Tournaments (P3) - FOMO, можно давать badges
4. Power-Ups (P2) - глубина gameplay, rewards за tournaments
5. Mixed Quiz (P4) - легкий content variation
6. Matchmaking (P5) - extension of Duels

---

## Excluded Mechanics

### ❌ User-Generated Questions

**Почему НЕ подходит:**
- Требует модерацию (spam, offensive content)
- Качество контента непредсказуемо
- Юридические риски (copyright infringement)

---

### ❌ Real-Time Multiplayer (синхронный)

**Почему НЕ подходит:**
- Высокая latency в TMA (WebSocket через Telegram unreliable)
- Требует оба игрока онлайн одновременно (плохо для retention)
- Сложная инфраструктура (WebSocket scaling)

**Альтернатива:** Асинхронные дуэли (Phase 1)

---

### ❌ Paid Tournaments с денежными призами

**Почему НЕ подходит:**
- Юридические сложности (gambling laws)
- Налогообложение выигрышей
- KYC/AML compliance
- Риск fraud

**Альтернатива:** Free tournaments с badge rewards

---

### ❌ Complex Progression Systems (уровни, XP, skill trees)

**Почему НЕ подходит:**
- Может overwhelm casual пользователей
- Долгий onboarding
- Не подходит для casual TMA experience

**Альтернатива:** Простая badge collection

---

## Next Steps

**Для реализации Phase 1 (Duels):**
1. Прочитай детальную спецификацию в старом DOMAIN.md
2. Создай DDD модель: aggregate, value objects, events
3. Напиши use cases
4. Имплементируй backend (domain → application → infrastructure)
5. Добавь Swagger endpoints
6. Генерируй TypeScript types
7. Имплементируй UI screens
8. Тестирование
9. Обнови документацию

**Не забудь:**
- Обновить `current/domain.md` после реализации
- Удалить Phase 1 из `future/ROADMAP.md`
- Коммитить docs вместе с кодом

---

## Metrics для оценки успеха

**Phase 1 (Duels):**
- % пользователей создавших хотя бы 1 дуэль
- Avg. дуэлей на активного пользователя в неделю
- Retention rate (Day 7, Day 30)

**Phase 2 (Badges):**
- % пользователей с хотя бы 1 unlocked badge
- Avg. badges per user
- Completion rate по категориям

**Phase 3 (Power-Ups):**
- % квизов с использованием power-ups
- Avg. power-ups used per quiz
- Conversion rate (free → paid power-ups, если монетизация)

**Phase 4 (Tournaments):**
- % пользователей участвующих в tournament
- Weekly active users growth
- Avg. quizzes completed per tournament

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 1.0
**Проект:** Quiz Sprint TMA
