import { computed, type Ref } from 'vue'
import type { InternalInfrastructureHttpHandlersStreakDTO } from '@/api/generated'

/**
 * Milestone значения для серий (дни подряд)
 */
export const STREAK_MILESTONES = [3, 7, 14, 30] as const

export type StreakMilestone = (typeof STREAK_MILESTONES)[number]

/**
 * Данные о milestone
 */
export interface MilestoneInfo {
	value: StreakMilestone
	label: string
	emoji: string
	description: string
	color: string // Tailwind color class
}

/**
 * Информация о milestone'ах
 */
export const MILESTONES_INFO: Record<StreakMilestone, MilestoneInfo> = {
	3: {
		value: 3,
		label: 'Начинающий',
		emoji: '🔥',
		description: '3 дня подряд',
		color: 'text-orange-500',
	},
	7: {
		value: 7,
		label: 'Недельник',
		emoji: '⚡',
		description: 'Неделя подряд',
		color: 'text-yellow-500',
	},
	14: {
		value: 14,
		label: 'Двухнедельник',
		emoji: '✨',
		description: '2 недели подряд',
		color: 'text-blue-500',
	},
	30: {
		value: 30,
		label: 'Месячник',
		emoji: '💎',
		description: 'Месяц подряд',
		color: 'text-purple-500',
	},
}

/**
 * Composable для работы с системой серий (streaks)
 *
 * Возможности:
 * - Определение текущего и следующего milestone
 * - Прогресс до следующего milestone
 * - Milestone анимации и уведомления
 * - Форматирование отображения серии
 */
export function useStreaks(streakRef: Ref<InternalInfrastructureHttpHandlersStreakDTO | null>) {
	// ===========================
	// Computed Properties
	// ===========================

	/**
	 * Текущая серия
	 */
	const currentStreak = computed(() => streakRef.value?.currentStreak ?? 0)

	/**
	 * Самая длинная серия (best streak)
	 */
	const longestStreak = computed(() => streakRef.value?.bestStreak ?? 0)

	/**
	 * Текущий достигнутый milestone
	 */
	const currentMilestone = computed((): StreakMilestone | null => {
		const current = currentStreak.value

		// Находим максимальный достигнутый milestone
		for (let i = STREAK_MILESTONES.length - 1; i >= 0; i--) {
			const milestone = STREAK_MILESTONES[i]
			if (milestone !== undefined && current >= milestone) {
				return milestone
			}
		}

		return null
	})

	/**
	 * Следующий milestone для достижения
	 */
	const nextMilestone = computed((): StreakMilestone | null => {
		const current = currentStreak.value

		// Находим следующий недостигнутый milestone
		for (const milestone of STREAK_MILESTONES) {
			if (current < milestone) {
				return milestone
			}
		}

		// Все milestone достигнуты
		return null
	})

	/**
	 * Дней до следующего milestone
	 */
	const daysToNextMilestone = computed(() => {
		const next = nextMilestone.value
		if (!next) return 0

		return next - currentStreak.value
	})

	/**
	 * Прогресс до следующего milestone (0-100%)
	 */
	const progressToNextMilestone = computed(() => {
		const next = nextMilestone.value
		if (!next) return 100 // Все milestone достигнуты

		const current = currentStreak.value
		const previous = currentMilestone.value ?? 0

		const range = next - previous
		const progress = current - previous

		return Math.round((progress / range) * 100)
	})

	/**
	 * Информация о текущем milestone
	 */
	const currentMilestoneInfo = computed(() => {
		const milestone = currentMilestone.value
		return milestone ? MILESTONES_INFO[milestone] : null
	})

	/**
	 * Информация о следующем milestone
	 */
	const nextMilestoneInfo = computed(() => {
		const milestone = nextMilestone.value
		return milestone ? MILESTONES_INFO[milestone] : null
	})

	/**
	 * Достиг ли игрок нового рекорда
	 */
	const isNewRecord = computed(() => {
		return currentStreak.value > 0 && currentStreak.value >= longestStreak.value
	})

	/**
	 * Только что достиг milestone
	 */
	const justReachedMilestone = computed(() => {
		const current = currentStreak.value
		return STREAK_MILESTONES.includes(current as StreakMilestone)
	})

	/**
	 * Форматированное отображение серии для UI
	 */
	const formattedStreak = computed(() => {
		const current = currentStreak.value

		if (current === 0) {
			return 'Начните серию!'
		}

		if (current === 1) {
			return '🔥 1 день'
		}

		if (current < 3) {
			return `🔥 ${current} дня`
		}

		// С milestone emoji
		const info = currentMilestoneInfo.value
		if (info) {
			return `${info.emoji} ${current} ${getDaysWord(current)}`
		}

		return `🔥 ${current} ${getDaysWord(current)}`
	})

	/**
	 * Прогресс-бар для UI (текст)
	 */
	const progressBarText = computed(() => {
		const next = nextMilestone.value
		if (!next) {
			return `👑 Максимум достигнут: ${currentStreak.value}`
		}

		const current = currentStreak.value
		const remaining = next - current

		return `До ${next}: осталось ${remaining} ${getDaysWord(remaining)}`
	})

	// ===========================
	// Helper Functions
	// ===========================

	/**
	 * Получить правильное склонение слова "день"
	 */
	function getDaysWord(count: number): string {
		const lastDigit = count % 10
		const lastTwoDigits = count % 100

		if (lastTwoDigits >= 11 && lastTwoDigits <= 19) {
			return 'дней'
		}

		if (lastDigit === 1) {
			return 'день'
		}

		if (lastDigit >= 2 && lastDigit <= 4) {
			return 'дня'
		}

		return 'дней'
	}

	/**
	 * Получить цвет для прогресс-бара
	 */
	const getProgressColor = computed(() => {
		const progress = progressToNextMilestone.value

		if (progress < 25) return 'bg-red-500'
		if (progress < 50) return 'bg-orange-500'
		if (progress < 75) return 'bg-yellow-500'
		return 'bg-green-500'
	})

	/**
	 * Получить emoji для отображения
	 */
	const getStreakEmoji = computed(() => {
		const info = currentMilestoneInfo.value
		return info?.emoji ?? '🔥'
	})

	// ===========================
	// Milestone Achievements
	// ===========================

	/**
	 * Получить все достигнутые milestones
	 */
	const achievedMilestones = computed(() => {
		const current = currentStreak.value
		return STREAK_MILESTONES.filter((m) => current >= m).map((m) => MILESTONES_INFO[m])
	})

	/**
	 * Получить все недостигнутые milestones
	 */
	const upcomingMilestones = computed(() => {
		const current = currentStreak.value
		return STREAK_MILESTONES.filter((m) => current < m).map((m) => MILESTONES_INFO[m])
	})

	// ===========================
	// Return
	// ===========================

	return {
		// Basic info
		currentStreak,
		longestStreak,

		// Milestones
		currentMilestone,
		nextMilestone,
		currentMilestoneInfo,
		nextMilestoneInfo,
		daysToNextMilestone,
		progressToNextMilestone,

		// Status
		isNewRecord,
		justReachedMilestone,

		// Formatting
		formattedStreak,
		progressBarText,
		getProgressColor,
		getStreakEmoji,

		// Achievements
		achievedMilestones,
		upcomingMilestones,

		// Constants
		STREAK_MILESTONES,
		MILESTONES_INFO,
	}
}
