import { ref, watch } from 'vue'
import { useLocalStorage } from '@vueuse/core'

/**
 * Composable to track the last viewed quiz ID
 * Useful for restoring context when navigating to leaderboard or other pages
 */
export function useLastQuiz() {
	// Store last viewed quiz ID in localStorage
	const lastQuizId = useLocalStorage<string | null>('quiz-sprint:last-quiz-id', null)

	/**
	 * Save a quiz ID as the last viewed quiz
	 */
	const saveLastQuizId = (quizId: string) => {
		lastQuizId.value = quizId
	}

	/**
	 * Get the last viewed quiz ID
	 */
	const getLastQuizId = () => {
		return lastQuizId.value
	}

	/**
	 * Clear the last viewed quiz ID
	 */
	const clearLastQuizId = () => {
		lastQuizId.value = null
	}

	return {
		lastQuizId,
		saveLastQuizId,
		getLastQuizId,
		clearLastQuizId,
	}
}
