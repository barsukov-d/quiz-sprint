<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePostQuizIdStart, usePostQuizSessionSessionidAnswer } from '@/api'
import { useAuth } from '@/composables/useAuth'
import type {
	InternalInfrastructureHttpHandlersQuestionDTO as QuestionDTO,
	InternalInfrastructureHttpHandlersSessionDTO as SessionDTO
} from '@/api/generated/types/internalInfrastructureHttpHandlers'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()

// Quiz state
const quizId = route.params.id as string
const session = ref<SessionDTO | null>(null)
const currentQuestion = ref<QuestionDTO | null>(null)
const totalQuestions = ref(0)
const currentQuestionIndex = ref(0)
const selectedAnswerId = ref<string | null>(null)
const isAnswerSubmitted = ref(false)
const answerResult = ref<{
	isCorrect: boolean
	points: number
	correctAnswerId?: string
} | null>(null)

// Mutations для API
const { mutateAsync: startQuiz, isPending: isStarting } = usePostQuizIdStart()
const { mutateAsync: submitAnswer, isPending: isSubmitting } = usePostQuizSessionSessionidAnswer()

// Начать квиз при монтировании
onMounted(async () => {
	if (!currentUser.value) {
		console.error('User not authenticated')
		return
	}

	try {
		const result = await startQuiz({
			id: quizId,
			data: {
				userId: currentUser.value.id
			}
		})

		if (result?.data) {
			session.value = result.data.session
			currentQuestion.value = result.data.firstQuestion
			totalQuestions.value = result.data.totalQuestions
			currentQuestionIndex.value = 1
		}
	} catch (error) {
		console.error('Failed to start quiz:', error)
	}
})

// Progress percentage
const progress = computed(() => {
	if (!totalQuestions.value) return 0
	return (currentQuestionIndex.value / totalQuestions.value) * 100
})

// Выбрать ответ
const selectAnswer = (answerId: string) => {
	if (isAnswerSubmitted.value || isSubmitting.value) return
	selectedAnswerId.value = answerId
}

// Отправить ответ
const confirmAnswer = async () => {
	if (!selectedAnswerId.value || !session.value || !currentQuestion.value || !currentUser.value)
		return

	try {
		isAnswerSubmitted.value = true

		const result = await submitAnswer({
			sessionId: session.value.id,
			data: {
				questionId: currentQuestion.value.id,
				answerId: selectedAnswerId.value,
				userId: currentUser.value.id
			}
		})

		if (result?.data) {
			answerResult.value = {
				isCorrect: result.data.isCorrect,
				points: result.data.pointsEarned || 0,
				correctAnswerId: result.data.correctAnswerId
			}

			// Обновить счет сессии
			if (session.value) {
				session.value.score += result.data.pointsEarned || 0
			}

			// Показать результат на 2 секунды, затем переход
			setTimeout(() => {
				if (result.data.nextQuestion) {
					// Есть следующий вопрос
					currentQuestion.value = result.data.nextQuestion
					currentQuestionIndex.value++
					resetQuestionState()
				} else if (result.data.finalResult) {
					// Квиз завершен
					router.push({
						name: 'quiz-results',
						params: { sessionId: session.value!.id }
					})
				}
			}, 2000)
		}
	} catch (error) {
		console.error('Failed to submit answer:', error)
		isAnswerSubmitted.value = false
	}
}

// Сбросить состояние вопроса
const resetQuestionState = () => {
	selectedAnswerId.value = null
	isAnswerSubmitted.value = false
	answerResult.value = null
}

// Стили для кнопки ответа
const getAnswerButtonClass = (answerId: string) => {
	if (!isAnswerSubmitted.value) {
		return selectedAnswerId.value === answerId ? 'ring-2 ring-primary bg-primary-50' : ''
	}

	// После отправки показываем правильный/неправильный
	if (answerResult.value?.correctAnswerId === answerId) {
		return 'ring-2 ring-green-500 bg-green-50'
	}
	if (selectedAnswerId.value === answerId && !answerResult.value?.isCorrect) {
		return 'ring-2 ring-red-500 bg-red-50'
	}
	return 'opacity-50'
}
</script>

<template>
	<div class="container mx-auto p-4 pt-20">
		<!-- Loading -->
		<div v-if="isStarting" class="flex justify-center items-center py-12">
			<UProgress animation="carousel" />
			<span class="ml-4">Starting quiz...</span>
		</div>

		<!-- Quiz Interface -->
		<div v-else-if="session && currentQuestion" class="max-w-2xl mx-auto">
			<!-- Header with progress -->
			<div class="mb-6">
				<div class="flex justify-between items-center mb-2">
					<span class="text-sm font-semibold text-gray-600"
						>Question {{ currentQuestionIndex }} of {{ totalQuestions }}</span
					>
					<span class="text-sm font-semibold text-gray-600">Score: {{ session.score }}</span>
				</div>
				<UProgress :value="progress" color="primary" />
			</div>

			<!-- Question Card -->
			<UCard class="mb-6">
				<div class="text-center py-6">
					<h2 class="text-2xl font-bold mb-4">{{ currentQuestion.text }}</h2>
				</div>
			</UCard>

			<!-- Answers -->
			<div class="space-y-3 mb-6">
				<button
					v-for="answer in currentQuestion.answers"
					:key="answer.id"
					:disabled="isAnswerSubmitted || isSubmitting"
					class="w-full p-4 text-left border-2 rounded-lg transition-all hover:border-primary disabled:cursor-not-allowed"
					:class="getAnswerButtonClass(answer.id)"
					@click="selectAnswer(answer.id)"
				>
					<div class="flex items-center justify-between">
						<span class="font-medium">{{ answer.text }}</span>
						<span v-if="isAnswerSubmitted && answerResult?.correctAnswerId === answer.id">
							✓
						</span>
						<span
							v-if="
								isAnswerSubmitted &&
								selectedAnswerId === answer.id &&
								!answerResult?.isCorrect
							"
						>
							✗
						</span>
					</div>
				</button>
			</div>

			<!-- Answer feedback banner -->
			<div v-if="answerResult" class="mb-6">
				<UAlert
					:color="answerResult.isCorrect ? 'green' : 'red'"
					:title="answerResult.isCorrect ? '✓ Correct!' : '✗ Incorrect'"
					:description="
						answerResult.isCorrect
							? `+${answerResult.points} points`
							: 'Better luck next time'
					"
				/>
			</div>

			<!-- Submit button -->
			<UButton
				v-if="!isAnswerSubmitted"
				:disabled="!selectedAnswerId || isSubmitting"
				:loading="isSubmitting"
				size="xl"
				color="primary"
				block
				@click="confirmAnswer"
			>
				{{ isSubmitting ? 'Submitting...' : 'Submit Answer' }}
			</UButton>

			<!-- Next question indicator -->
			<div v-else class="text-center text-gray-500">
				<UProgress animation="carousel" size="sm" />
				<p class="mt-2">Loading next question...</p>
			</div>
		</div>

		<!-- Error state -->
		<div v-else class="text-center py-12">
			<UAlert color="red" title="Failed to load quiz" description="Please try again" />
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
