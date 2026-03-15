<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
	usePostQuizIdStart,
	usePostQuizSessionSessionidAnswer,
	useDeleteQuizSessionSessionid,
} from '@/api'
import { useAuth } from '@/composables/useAuth'
import type {
	InternalInfrastructureHttpHandlersQuestionDTO as QuestionDTO,
	InternalInfrastructureHttpHandlersSessionDTO as SessionDTO,
} from '@/api/generated/types/internalInfrastructureHttpHandlers'
import axios from 'axios'
import AnswerButton from '@/components/shared/AnswerButton.vue'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()
const { t } = useI18n()

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
	basePoints: number
	timeBonus: number
	streakBonus: number
	pointsEarned: number
	currentStreak: number
	correctAnswerId?: string
} | null>(null)

// Timer state
const questionStartTime = ref<number>(0)
const timeElapsed = ref<number>(0)
const timerInterval = ref<ReturnType<typeof setInterval> | null>(null)
const timeLimitPerQuestion = ref<number>(30)

// Error state
const showConflictModal = ref(false)
const errorMessage = ref<string | null>(null)
const isResuming = ref(false)
const isAbandoning = ref(false)
const isSessionResumed = ref(false) // Track if session was successfully resumed
const isInitializing = ref(true) // Track initial load

// Mutations для API
const { mutateAsync: startQuiz, isPending: isStarting } = usePostQuizIdStart()
const { mutateAsync: submitAnswer, isPending: isSubmitting } = usePostQuizSessionSessionidAnswer()
const { mutateAsync: abandonSession } = useDeleteQuizSessionSessionid()

// Timer functions
const startQuestionTimer = () => {
	questionStartTime.value = Date.now()
	timeElapsed.value = 0

	timerInterval.value = setInterval(() => {
		timeElapsed.value = Math.floor((Date.now() - questionStartTime.value) / 1000)
	}, 100) // Update every 100ms for smoothness
}

const stopQuestionTimer = () => {
	if (timerInterval.value) {
		clearInterval(timerInterval.value)
		timerInterval.value = null
	}
}

const timeRemaining = computed(() => {
	const remaining = timeLimitPerQuestion.value - timeElapsed.value
	return Math.max(0, remaining)
})

// Handle resuming existing session
const handleResumeSession = async () => {
	if (!currentUser.value) return

	// Close modal first, show loading screen
	showConflictModal.value = false
	isResuming.value = true
	errorMessage.value = null

	try {
		const response = await axios.get(
			`/api/v1/quiz/${quizId}/active-session?userId=${currentUser.value.id}`,
		)

		if (response.data?.data) {
			console.log('Resume session data:', response.data.data)

			session.value = response.data.data.session
			currentQuestion.value = response.data.data.currentQuestion
			totalQuestions.value = response.data.data.totalQuestions
			timeLimitPerQuestion.value = response.data.data.timeLimitPerQuestion || 30
			currentQuestionIndex.value = response.data.data.session.currentQuestion + 1
			isSessionResumed.value = true // Mark as successfully resumed

			console.log('Session resumed:', {
				sessionId: session.value?.id,
				questionIndex: currentQuestionIndex.value,
				hasQuestion: !!currentQuestion.value,
			})
		} else {
			throw new Error('Invalid response data')
		}
	} catch (error) {
		console.error('Failed to resume session:', error)
		errorMessage.value = 'Failed to resume session. Please try again.'
		isSessionResumed.value = false
		// Show modal again on error
		showConflictModal.value = true
	} finally {
		isResuming.value = false
	}
}

// Handle abandoning session and starting fresh
const handleStartFresh = async () => {
	if (!currentUser.value || !session.value) return

	isAbandoning.value = true
	errorMessage.value = null

	try {
		// Delete the existing session
		await abandonSession({
			sessionId: session.value.id,
			data: { userId: currentUser.value.id },
		})

		// Reset session resumed flag
		isSessionResumed.value = false

		// Close modal first
		showConflictModal.value = false

		// Clear old session data
		session.value = null
		currentQuestion.value = null

		// Start a new quiz (this will show loading state)
		await startNewQuiz()
	} catch (error) {
		console.error('Failed to abandon session:', error)
		errorMessage.value = 'Failed to start fresh. Please try again.'
		showConflictModal.value = true
	} finally {
		isAbandoning.value = false
	}
}

// Check for existing session first, then start or resume
const initializeQuiz = async () => {
	if (!currentUser.value) {
		console.error('User not authenticated')
		isInitializing.value = false
		return
	}

	isInitializing.value = true
	showConflictModal.value = false // Explicitly hide modal during init

	try {
		// First, check if there's an active session
		const activeSessionResponse = await axios.get(
			`/api/v1/quiz/${quizId}/active-session?userId=${currentUser.value.id}`,
		)

		if (activeSessionResponse.data?.data) {
			// Active session exists - show modal
			console.log('Active session found, showing modal')
			session.value = activeSessionResponse.data.data.session
			isInitializing.value = false
			// Show modal AFTER initialization is complete
			showConflictModal.value = true
		}
	} catch (error) {
		// No active session (404) - start a new one
		if (axios.isAxiosError(error) && error.response?.status === 404) {
			console.log('No active session, starting new quiz')
			showConflictModal.value = false // Ensure modal stays hidden
			await startNewQuiz()
			isInitializing.value = false
		} else {
			console.error('Error checking for active session:', error)
			errorMessage.value = 'Failed to initialize quiz. Please try again.'
			showConflictModal.value = false
			isInitializing.value = false
		}
	}
}

// Start a completely new quiz
const startNewQuiz = async () => {
	if (!currentUser.value) return

	try {
		const result = await startQuiz({
			id: quizId,
			data: {
				userId: currentUser.value.id,
			},
		})

		if (result?.data) {
			session.value = result.data.session
			currentQuestion.value = result.data.firstQuestion
			totalQuestions.value = result.data.totalQuestions
			timeLimitPerQuestion.value = result.data.timeLimitPerQuestion || 30
			currentQuestionIndex.value = 1
			isSessionResumed.value = false
		}
	} catch (error) {
		console.error('Failed to start quiz:', error)
		errorMessage.value = 'Failed to start quiz. Please try again.'
	}
}

// Начать квиз при монтировании
onMounted(async () => {
	await initializeQuiz()
})

// Cleanup timer on unmount
onUnmounted(() => {
	stopQuestionTimer()
})

// Start timer when question changes
watch(currentQuestion, (newQuestion) => {
	if (newQuestion && !isAnswerSubmitted.value) {
		startQuestionTimer()
	}
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

		// Stop timer and calculate time taken
		stopQuestionTimer()
		const timeTaken = Date.now() - questionStartTime.value // milliseconds

		const result = await submitAnswer({
			sessionId: session.value.id,
			data: {
				questionId: currentQuestion.value.id,
				answerId: selectedAnswerId.value,
				userId: currentUser.value.id,
				timeTaken: timeTaken,
			},
		})

		if (result?.data) {
			answerResult.value = {
				isCorrect: result.data.isCorrect,
				basePoints: result.data.basePoints || 0,
				timeBonus: result.data.timeBonus || 0,
				streakBonus: result.data.streakBonus || 0,
				pointsEarned: result.data.pointsEarned || 0,
				currentStreak: result.data.currentStreak || 0,
				correctAnswerId: result.data.correctAnswerId,
			}

			// Обновить счет сессии
			if (session.value) {
				session.value.score = result.data.totalScore || 0
			}

			// Показать результат на 3 секунды (дать время увидеть бонусы), затем переход
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
						params: { sessionId: session.value!.id },
					})
				}
			}, 3000)
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
</script>

<template>
	<div class="min-h-screen mx-auto max-w-[800px] flex flex-col">
		<!-- Loading -->
		<div
			v-if="isInitializing || isStarting || isResuming"
			class="flex-1 flex flex-col justify-center items-center gap-4 p-8"
		>
			<UProgress animation="carousel" class="w-48" />
			<span class="text-(--ui-text-muted) text-sm">
				{{
					isResuming
						? t('quiz.resumingSession')
						: isInitializing
							? t('quiz.loadingQuiz')
							: t('quiz.startingQuiz')
				}}
			</span>
		</div>

		<!-- Quiz Interface -->
		<template v-else-if="session && currentQuestion">
			<!-- Feedback overlay (correct/incorrect) — replaces question UI -->
			<template v-if="answerResult">
				<!-- Correct feedback -->
				<div v-if="answerResult.isCorrect" class="flex flex-col min-h-screen">
					<!-- Green gradient header -->
					<div
						class="px-6 pt-12 pb-8 text-center"
						style="background: linear-gradient(180deg, #2ecc71 0%, #27ae60 100%)"
					>
						<h2 class="text-white text-3xl font-bold mb-3">{{ t('quiz.correct') }}</h2>
						<span
							class="inline-block bg-white text-green-600 font-bold text-lg px-6 py-2 rounded-full"
						>
							+{{ answerResult.pointsEarned }}
						</span>
					</div>

					<!-- Question review content -->
					<div class="flex-1 bg-(--ui-bg) px-4 py-6 space-y-5">
						<!-- Question text -->
						<p class="text-(--ui-text-highlighted) text-xl font-bold text-center">
							{{ currentQuestion.text }}
						</p>

						<div class="border-t border-(--ui-border)" />

						<!-- Your answer -->
						<div class="text-center">
							<span class="text-(--ui-text-muted) text-base">You answered: </span>
							<span class="text-purple-500 font-semibold text-base">
								{{
									currentQuestion.answers.find((a) => a.id === selectedAnswerId)
										?.text
								}}
							</span>
						</div>

						<div class="border-t border-(--ui-border)" />

						<!-- Streak bonus if any -->
						<div
							v-if="answerResult.streakBonus > 0"
							class="text-center text-sm text-orange-500 font-semibold"
						>
							{{ t('quiz.streakBonus', { points: answerResult.streakBonus }) }}
						</div>
					</div>

					<!-- Next button -->
					<div class="px-4 pb-8 bg-(--ui-bg)">
						<UProgress animation="carousel" size="xs" class="mb-4" />
						<p class="text-center text-(--ui-text-dimmed) text-sm mb-3">
							{{ t('quiz.loadingNext') }}
						</p>
					</div>
				</div>

				<!-- Incorrect feedback -->
				<div v-else class="flex flex-col min-h-screen">
					<!-- Red gradient header -->
					<div
						class="px-6 pt-12 pb-8 text-center"
						style="background: linear-gradient(180deg, #e74c3c 0%, #c0392b 100%)"
					>
						<h2 class="text-white text-3xl font-bold mb-3">
							{{ t('quiz.incorrect') || 'Incorrect!' }}
						</h2>
						<span
							class="inline-block bg-white text-red-500 font-semibold text-base px-5 py-2 rounded-full"
						>
							{{ t('quiz.wrongAnswer') }}
						</span>
					</div>

					<!-- Review content -->
					<div class="flex-1 bg-(--ui-bg) px-4 py-6 space-y-5">
						<!-- Question text -->
						<p class="text-(--ui-text-highlighted) text-xl font-bold text-center">
							{{ currentQuestion.text }}
						</p>

						<div class="border-t border-(--ui-border)" />

						<!-- Correct answer -->
						<div class="space-y-3">
							<p class="text-(--ui-text-muted) text-center text-base font-semibold">
								Correct answer:
							</p>
							<div class="grid grid-cols-2 gap-3">
								<div
									v-for="(answer, index) in currentQuestion.answers"
									:key="answer.id"
									class="rounded-2xl min-h-[80px] flex items-center justify-center px-3"
									:style="{
										backgroundColor:
											answer.id === answerResult.correctAnswerId
												? ['#4A90D9', '#E74C3C', '#F39C12', '#2ECC71'][
														index % 4
													]
												: 'var(--ui-bg-elevated)',
										opacity:
											answer.id === answerResult.correctAnswerId ? 1 : 0.35,
									}"
								>
									<span
										class="text-center text-sm font-bold leading-snug"
										:style="{
											color:
												answer.id === answerResult.correctAnswerId
													? '#fff'
													: 'var(--ui-text)',
										}"
									>
										{{ answer.text }}
									</span>
								</div>
							</div>
						</div>
					</div>

					<!-- Next indicator -->
					<div class="px-4 pb-8 bg-(--ui-bg)">
						<UProgress animation="carousel" size="xs" class="mb-4" />
						<p class="text-center text-(--ui-text-dimmed) text-sm mb-3">
							{{ t('quiz.loadingNext') }}
						</p>
					</div>
				</div>
			</template>

			<!-- Question UI -->
			<template v-else>
				<!-- Header: counter + title + menu -->
				<div class="flex items-center justify-between px-4 pt-14 pb-3">
					<span class="text-xl font-bold text-(--ui-text-highlighted) tabular-nums">
						{{ currentQuestionIndex }}/{{ totalQuestions }}
					</span>
					<span class="text-sm font-semibold text-(--ui-text-muted)">Quiz</span>
					<UIcon
						name="i-heroicons-ellipsis-horizontal-circle"
						class="size-6 text-(--ui-text-dimmed)"
					/>
				</div>

				<!-- Timer bar -->
				<div class="px-4 pb-4">
					<div class="relative h-6 rounded-full overflow-hidden bg-(--ui-bg-accented)">
						<div
							class="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-yellow-400 via-primary-500 to-primary-600 transition-all duration-1000"
							:style="{ width: `${(timeRemaining / timeLimitPerQuestion) * 100}%` }"
						/>
						<span
							class="absolute inset-0 flex items-center justify-center text-xs font-bold tabular-nums drop-shadow"
							:class="
								timeRemaining <= timeLimitPerQuestion * 0.25
									? 'text-red-400'
									: 'text-white'
							"
						>
							{{ timeRemaining }}s
						</span>
					</div>
				</div>

				<!-- Question card -->
				<div class="px-4 pb-5">
					<div class="bg-(--ui-bg-elevated) rounded-2xl p-6 text-center">
						<p class="text-(--ui-text-highlighted) text-xl font-bold leading-snug">
							{{ currentQuestion.text }}
						</p>
					</div>
				</div>

				<!-- Answer buttons (colored bars) -->
				<div class="px-4 flex flex-col gap-3 pb-6">
					<AnswerButton
						v-for="(answer, index) in currentQuestion.answers"
						:key="answer.id"
						:answer="answer"
						:color-index="index % 4"
						:color-bar="true"
						:disabled="isAnswerSubmitted || isSubmitting"
						:selected="selectedAnswerId === answer.id"
						@click="selectAnswer"
					/>
				</div>

				<!-- Submit button (shown when answer selected, before submission) -->
				<div v-if="selectedAnswerId && !isAnswerSubmitted" class="px-4 pb-6">
					<UButton
						:loading="isSubmitting"
						size="xl"
						block
						class="rounded-full"
						style="background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%)"
						@click="confirmAnswer"
					>
						{{ isSubmitting ? t('quiz.submitting') : t('quiz.submit') }}
					</UButton>
				</div>
			</template>
		</template>

		<!-- Error state -->
		<div v-else-if="!showConflictModal" class="flex-1 flex items-center justify-center p-6">
			<UAlert
				color="red"
				:title="t('quiz.loadFailed2')"
				:description="errorMessage || t('quiz.tryAgain2')"
			/>
		</div>

		<!-- Active Session Conflict Modal -->
		<UModal
			v-if="showConflictModal && !isInitializing && (!session || !currentQuestion)"
			v-model="showConflictModal"
			:prevent-close="isAbandoning"
		>
			<UCard>
				<template #header>
					<h3 class="text-xl font-bold">{{ t('quiz.activeSession') }}</h3>
				</template>

				<div class="space-y-4">
					<p class="text-(--ui-text)">{{ t('quiz.activeSessionDesc') }}</p>

					<!-- Error message -->
					<UAlert v-if="errorMessage" color="red" :title="errorMessage" />

					<div class="flex flex-col gap-3">
						<UButton
							size="lg"
							color="primary"
							block
							:disabled="isAbandoning"
							@click="handleResumeSession"
						>
							{{ t('quiz.continueSession') }}
						</UButton>
						<UButton
							size="lg"
							color="neutral"
							variant="outline"
							block
							:loading="isAbandoning"
							@click="handleStartFresh"
						>
							{{ isAbandoning ? t('quiz.startingFresh') : t('quiz.startFresh') }}
						</UButton>
					</div>
				</div>
			</UCard>
		</UModal>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
