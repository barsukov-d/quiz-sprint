# Game Composables

Composables для управления игровой логикой Daily Challenge и Solo Marathon.

## 📦 Доступные Composables

### `useDailyChallenge(playerId: string)`

Управление Daily Challenge игрой.

**Возможности:**

- Старт игры (один раз в день)
- Отправка ответов (без показа правильности до конца)
- Получение статуса (сыграл ли сегодня)
- Локальное сохранение прогресса
- Система серий (streaks)
- Таймер до сброса квиза

**Пример использования:**

```vue
<script setup lang="ts">
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useGameTimer } from '@/composables/useGameTimer'

const playerId = 'user123'
const { state, isPlaying, isCompleted, canPlay, progress, startGame, submitAnswer, initialize } =
	useDailyChallenge(playerId)

// Инициализация при монтировании
onMounted(async () => {
	await initialize()
})

// Старт игры
const handleStart = async () => {
	await startGame()
}

// Отправка ответа
const handleAnswer = async (answerId: string) => {
	const timeTaken = timer.elapsedTime.value
	await submitAnswer(answerId, timeTaken)
}

// Таймер для вопроса
const timer = useGameTimer({
	initialTime: state.value.timeLimit,
	autoStart: true,
	onTimeout: () => {
		// Время вышло - автоматически отправить пустой ответ
		handleAnswer('')
	},
})
</script>
```

---

### `useMarathon(playerId: string)`

Управление Marathon игрой.

**Возможности:**

- Старт игры с выбором категории
- Отправка ответов с немедленным feedback (правильно/неправильно)
- Система жизней с восстановлением
- Подсказки: 50/50, +10сек, Skip, Hint
- Личные рекорды по категориям
- Адаптивная сложность
- Локальное сохранение незавершённой игры

**Пример использования:**

```vue
<script setup lang="ts">
import { useMarathon } from '@/composables/useMarathon'

const playerId = 'user123'
const {
	state,
	isPlaying,
	hasLives,
	canUseFiftyFifty,
	progressToRecord,
	startGame,
	submitAnswer,
	useHint,
	abandonGame,
	initialize,
} = useMarathon(playerId)

// Инициализация
onMounted(async () => {
	await initialize()
})

// Старт с категорией
const handleStart = async (categoryId: string) => {
	await startGame(categoryId)
}

// Отправка ответа
const handleAnswer = async (answerId: string) => {
	const result = await submitAnswer(answerId, timer.elapsedTime.value)

	if (result.isCorrect) {
		toast.success('Верно! 🎉')
	} else {
		toast.error('Неверно! ❌')
	}
}

// Использование подсказки 50/50
const handleFiftyFifty = async () => {
	if (canUseFiftyFifty.value) {
		const result = await useHint('fifty_fifty')
		// UI должен отфильтровать result.eliminatedAnswers
	}
}

// Досрочное завершение
const handleAbandon = async () => {
	await abandonGame()
}
</script>
```

---

### `useGameTimer(options: GameTimerOptions)`

Универсальный таймер для вопросов.

**Опции:**

- `initialTime` - начальное время в секундах
- `onTimeout` - callback при окончании времени
- `onTick` - callback каждую секунду
- `autoStart` - автоматический старт
- `soundWarning` - звук на последних секундах
- `warningThreshold` - за сколько секунд включить предупреждение

**Пример использования:**

```vue
<script setup lang="ts">
import { useGameTimer } from '@/composables/useGameTimer'

const timer = useGameTimer({
	initialTime: 15,
	autoStart: false,
	warningThreshold: 5,
	onTimeout: () => {
		console.log('Время вышло!')
		// Автоматически отправить ответ
	},
	onTick: (remaining) => {
		if (remaining === 10) {
			toast.info('Осталось 10 секунд!')
		}
	},
})

// Управление
const startTimer = () => timer.start()
const pauseTimer = () => timer.pause()
const resumeTimer = () => timer.resume()
const resetTimer = () => timer.reset()

// Добавить время (для подсказки +10сек)
const addExtraTime = () => timer.addTime(10)
</script>

<template>
	<div>
		<!-- Форматированное время -->
		<div>{{ timer.formattedTime }}</div>

		<!-- Прогресс-бар -->
		<UProgress :value="timer.progress" :color="timer.isWarning ? 'red' : 'green'" />

		<!-- Индикатор критического времени -->
		<div v-if="timer.isWarning" class="warning">⚠️ Осталось мало времени!</div>
	</div>
</template>
```

---

### `useStreaks(streak: StreakDTO | null)`

Система серий (streaks) для Daily Challenge.

**Возможности:**

- Определение текущего и следующего milestone
- Прогресс до следующего milestone
- Milestone анимации (3, 7, 14, 30, 100 дней)
- Форматирование отображения серии

**Milestones:**

- 🔥 **3 дня** - Начинающий
- ⚡ **7 дней** - Недельник
- ✨ **14 дней** - Двухнедельник
- 💎 **30 дней** - Месячник
- 👑 **100 дней** - Легенда

**Пример использования:**

```vue
<script setup lang="ts">
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useStreaks } from '@/composables/useStreaks'

const { state } = useDailyChallenge('user123')

const streaks = useStreaks(state.value.streak)

// Показать уведомление при достижении milestone
watch(
	() => streaks.justReachedMilestone.value,
	(reached) => {
		if (reached) {
			const info = streaks.currentMilestoneInfo.value
			toast.success(`🎉 Достигнут ${info.label}! ${info.emoji}`)
		}
	},
)
</script>

<template>
	<div>
		<!-- Текущая серия -->
		<div class="streak-display">
			{{ streaks.formattedStreak }}
		</div>

		<!-- Прогресс до следующего milestone -->
		<div v-if="streaks.nextMilestone">
			<UProgress :value="streaks.progressToNextMilestone" />
			<p>{{ streaks.progressBarText }}</p>
		</div>

		<!-- Все достигнутые milestones -->
		<div class="achievements">
			<UBadge
				v-for="milestone in streaks.achievedMilestones"
				:key="milestone.value"
				:color="milestone.color"
			>
				{{ milestone.emoji }} {{ milestone.label }}
			</UBadge>
		</div>

		<!-- Новый рекорд -->
		<div v-if="streaks.isNewRecord">🏆 Новый рекорд!</div>
	</div>
</template>
```

---

## 🔄 Интеграция с API

Все composables используют сгенерированные Vue Query hooks из `@/api/generated`:

**Daily Challenge:**

- `usePostDailyChallengeStart` - старт игры
- `usePostDailyChallengeGameidAnswer` - отправка ответа
- `useGetDailyChallengeStatus` - статус игры
- `useGetDailyChallengeStreak` - серия игрока

**Marathon:**

- `usePostMarathonStart` - старт игры
- `usePostMarathonGameidAnswer` - отправка ответа
- `usePostMarathonGameidHint` - использование подсказки
- `useDeleteMarathonGameid` - завершение игры
- `useGetMarathonStatus` - статус (жизни)
- `useGetMarathonPersonalBests` - личные рекорды

---

## 💾 Локальное хранение

Оба игровых composables сохраняют прогресс в `localStorage`:

**Daily Challenge (`daily-challenge-state`):**

- Текущая игра
- Текущий вопрос
- Индекс вопроса
- TTL: 24 часа

**Marathon (`marathon-state`):**

- Текущая игра
- Текущий вопрос
- Жизни, подсказки
- Серия, очки
- TTL: 7 дней

Прогресс автоматически восстанавливается при вызове `initialize()`.

---

## 🎮 Workflow

### Daily Challenge

```
1. initialize() - загрузка статуса с сервера
2. startGame() - начать игру
3. submitAnswer() × 10 - отправить ответы на все 10 вопросов
4. results - показать результаты
```

### Marathon

```
1. initialize() - загрузка статуса (жизни, рекорды)
2. startGame(categoryId) - начать игру
3. submitAnswer() - отправить ответ → feedback (правильно/неправильно)
4. Повторять пункт 3 до:
   - Потери всех жизней → game over
   - Досрочного завершения → abandonGame()
```

---

## ⚙️ TypeScript

Все composables полностью типизированы:

```typescript
import type { DailyChallengeStatus, MarathonStatus, HintType, StreakMilestone } from '@/composables'
```

---

## 🧪 Тестирование

TODO: Добавить unit-тесты для composables

```bash
pnpm test:unit src/composables
```
