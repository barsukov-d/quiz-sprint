import { ref, computed } from 'vue'

// Module-level singletons — session persists across component remounts
const runCount = ref(0)
const sessionBest = ref(0)

export function useMarathonSession() {
	const recordRunResult = (score: number, _streak: number) => {
		runCount.value++
		if (score > sessionBest.value) {
			sessionBest.value = score
		}
	}

	const getMotivationalPrompt = (currentScore: number, personalBest: number | null): string => {
		if (personalBest && personalBest > currentScore) {
			const deficit = personalBest - currentScore
			return `До рекорда ${deficit} ответов. Ещё один забег?`
		}
		if (currentScore > 0 && (!personalBest || currentScore >= personalBest)) {
			return 'Новый рекорд! Сможешь побить его снова?'
		}
		if (runCount.value >= 2) {
			return `Забег #${runCount.value} — лучший в сессии: ${sessionBest.value}`
		}
		return 'Ещё один забег?'
	}

	const resetSession = () => {
		runCount.value = 0
		sessionBest.value = 0
	}

	const sessionLabel = computed(() =>
		runCount.value > 0 ? `Забег #${runCount.value} | Лучший: ${sessionBest.value}` : null,
	)

	return {
		runCount,
		sessionBest,
		sessionLabel,
		recordRunResult,
		getMotivationalPrompt,
		resetSession,
	}
}
