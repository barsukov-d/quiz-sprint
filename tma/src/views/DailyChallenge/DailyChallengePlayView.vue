<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useAuth } from '@/composables/useAuth'
import GameTimer from '@/components/shared/GameTimer.vue'
import QuestionCard from '@/components/shared/QuestionCard.vue'
import AnswerButton from '@/components/shared/AnswerButton.vue'
import { useI18n } from 'vue-i18n'

// ===========================
// Auth & Router
// ===========================

const router = useRouter()
const { currentUser } = useAuth()
const { t } = useI18n()
const playerId = currentUser.value?.id || 'guest'

// ===========================
// Daily Challenge Composable
// ===========================

const {
	currentQuestion,
	questionIndex,
	totalQuestions,
	timeLimit,
	timeRemaining,
	isPlaying,
	submitAnswer,
	initialize,
	refetchStatus,
	refetchStreak,
} = useDailyChallenge(playerId)

// ===========================
// Local State (UI only)
// ===========================

const selectedAnswerId = ref<string | null>(null)
const isSubmitting = ref(false)
const timerRef = ref<InstanceType<typeof GameTimer> | null>(null)
const questionStartTime = ref(Date.now())

// Feedback state (from backend response)
const showFeedback = ref(false)
const feedbackIsCorrect = ref<boolean | null>(null)
const feedbackCorrectAnswerId = ref<string | null>(null)
const feedbackGameCompleted = ref(false)

// ===========================
// Computed
// ===========================

const questionProgress = computed(() =>
	Math.round(((questionIndex.value + 1) / totalQuestions.value) * 100),
)

const canSubmit = computed(() => {
	return selectedAnswerId.value !== null && !isSubmitting.value && !showFeedback.value
})

// ===========================
// Methods
// ===========================

const handleAnswerSelect = async (answerId: string) => {
	if (isSubmitting.value || showFeedback.value || timerRef.value?.remainingTime === 0) return

	selectedAnswerId.value = answerId

	// Auto-submit immediately after selection
	await handleSubmit()
}

const handleSubmit = async () => {
	if (!canSubmit.value || !selectedAnswerId.value) return

	isSubmitting.value = true

	try {
		// Calculate time taken in milliseconds using precise Date.now()
		const timeTaken = Date.now() - questionStartTime.value

		// Submit answer - backend returns isCorrect + correctAnswerId
		const answerData = await submitAnswer(selectedAnswerId.value, timeTaken)

		// Pause timer during feedback
		timerRef.value?.pause()

		// Show instant feedback from backend
		feedbackIsCorrect.value = answerData.isCorrect
		feedbackCorrectAnswerId.value = answerData.correctAnswerId
		feedbackGameCompleted.value = answerData.isGameCompleted ?? false
		showFeedback.value = true
		isSubmitting.value = false

		// Wait 1.5s then move to next question
		setTimeout(() => {
			handleNextStep()
		}, 1500)
	} catch (error) {
		console.error('Failed to submit answer:', error)
		isSubmitting.value = false
	}
}

const handleTimeout = async () => {
	// Prevent multiple timeout calls
	if (showFeedback.value || isSubmitting.value) return

	// Auto-submit when timer expires
	if (selectedAnswerId.value) {
		// User selected answer — submit it
		await handleSubmit()
	} else {
		// No answer selected — submit first answer as wrong (backend needs an answer to progress)
		if (!currentQuestion.value?.answers?.[0]) return

		selectedAnswerId.value = currentQuestion.value.answers[0].id
		isSubmitting.value = true

		try {
			// Submit with full time taken (timeout) in milliseconds
			const answerData = await submitAnswer(selectedAnswerId.value, timeLimit.value * 1000)

			// Show brief feedback (will be wrong)
			feedbackIsCorrect.value = answerData.isCorrect
			feedbackCorrectAnswerId.value = answerData.correctAnswerId
			feedbackGameCompleted.value = answerData.isGameCompleted ?? false
			showFeedback.value = true
			isSubmitting.value = false

			// Move to next question after 1s
			setTimeout(() => {
				handleNextStep()
			}, 1000)
		} catch (error) {
			console.error('Failed to submit timeout answer:', error)
			isSubmitting.value = false
		}
	}
}

const handleNextStep = async () => {
	const gameCompleted = feedbackGameCompleted.value

	// Reset feedback state
	showFeedback.value = false
	selectedAnswerId.value = null
	isSubmitting.value = false
	feedbackIsCorrect.value = null
	feedbackCorrectAnswerId.value = null
	feedbackGameCompleted.value = false

	if (gameCompleted) {
		// Game completed — fetch results from server, then navigate
		await refetchStatus()
		await refetchStreak()
		router.push({ name: 'daily-challenge-results' })
	} else {
		// Game continues — fetch next question from server, then reset timer
		await refetchStatus()
		questionStartTime.value = Date.now()
		timerRef.value?.reset(timeLimit.value)
		timerRef.value?.start()
	}
}

/**
 * Per-answer feedback state for AnswerButton.
 * 4 states per docs/game_modes/daily_challenge/02_gameplay.md:
 * - Correct answer → green (checkmark), full opacity
 * - Selected + wrong → red (cross), full opacity
 * - Not selected + not correct → muted (opacity-40)
 * - Selected + correct → green (checkmark), full opacity
 */
const getAnswerFeedback = (answerId: string) => {
	if (!showFeedback.value || feedbackIsCorrect.value === null) {
		return { showFeedback: false, isCorrect: null as boolean | null }
	}

	// This is the correct answer → green
	if (answerId === feedbackCorrectAnswerId.value) {
		return { showFeedback: true, isCorrect: true }
	}

	// This is the selected wrong answer → red
	if (answerId === selectedAnswerId.value && !feedbackIsCorrect.value) {
		return { showFeedback: true, isCorrect: false }
	}

	// Other answers → muted (showFeedback=true but isCorrect=null triggers opacity-40)
	return { showFeedback: true, isCorrect: null as boolean | null }
}

// ===========================
// Timer initialization helper
// ===========================

const startTimer = () => {
	if (!timerRef.value) return

	// Use server-provided timeRemaining, fallback to full timeLimit
	const initialTime =
		timeRemaining.value !== null && timeRemaining.value !== undefined
			? timeRemaining.value
			: timeLimit.value

	timerRef.value.reset(initialTime)
	timerRef.value.start()
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	try {
		await initialize()
	} catch (error: unknown) {
		// Ignore canceled requests (component unmounting during navigation)
		if (error instanceof Error && error.message === 'canceled') return
		console.error('[DailyChallengePlayView] Failed to initialize:', error)
	}

	if (!isPlaying.value) {
		router.push({ name: 'home' })
		return
	}

	// Wait for DOM update so timerRef is available, then start timer
	await nextTick()
	questionStartTime.value = Date.now()
	startTimer()
})

onUnmounted(() => {
	timerRef.value?.stop()
})
</script>

<template>
	<div class="min-h-screen mx-auto max-w-[800px] px-4 pt-14 pb-8 sm:px-3 sm:pt-12">
		<!-- Loading State -->
		<div v-if="!currentQuestion" class="flex flex-col items-center justify-center min-h-[50vh]">
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-(--ui-text-muted) mt-4">{{ t('daily.loadingQuestion') }}</p>
		</div>

		<!-- Game View -->
		<div v-else class="flex flex-col gap-5">
			<!-- Header row: counter + title -->
			<div class="flex items-center justify-between">
				<span class="text-xl font-bold text-(--ui-text-highlighted) tabular-nums">
					{{ questionIndex + 1 }}/{{ totalQuestions }}
				</span>
				<span class="text-base font-semibold text-(--ui-text-muted)">
					{{ t('daily.title') }}
				</span>
				<UIcon
					name="i-heroicons-ellipsis-horizontal-circle"
					class="size-6 text-(--ui-text-dimmed)"
				/>
			</div>

			<!-- Progress bar with timer inside -->
			<div class="relative">
				<div class="h-7 w-full rounded-full bg-(--ui-bg-accented) overflow-hidden">
					<div
						class="h-full rounded-full bg-gradient-to-r from-yellow-400 via-primary-500 to-primary-600 transition-all duration-300"
						:style="{ width: `${questionProgress}%` }"
					/>
				</div>
				<span
					class="absolute inset-0 flex items-center justify-center text-xs font-bold tabular-nums"
					:class="
						timerRef?.remainingTime !== undefined && timerRef.remainingTime <= 5
							? 'text-red-400'
							: 'text-white'
					"
				>
					{{ timerRef?.remainingTime ?? timeLimit }}
				</span>
				<!-- Hidden GameTimer for logic -->
				<GameTimer
					ref="timerRef"
					:initial-time="timeLimit"
					:auto-start="false"
					:warning-threshold="5"
					:show-progress="false"
					size="sm"
					:on-timeout="handleTimeout"
					class="sr-only"
				/>
			</div>

			<!-- Question text (primary focus) -->
			<QuestionCard :question="currentQuestion" :show-badge="false" />

			<!-- Answer Buttons (colored bars) -->
			<div class="flex flex-col gap-3">
				<AnswerButton
					v-for="(answer, index) in currentQuestion.answers"
					:key="answer.id"
					:answer="answer"
					:selected="selectedAnswerId === answer.id"
					:disabled="isSubmitting || showFeedback || timerRef?.remainingTime === 0"
					:show-feedback="getAnswerFeedback(answer.id).showFeedback"
					:is-correct="getAnswerFeedback(answer.id).isCorrect"
					:color-index="index % 4"
					:color-bar="true"
					@click="handleAnswerSelect"
				/>
			</div>

			<!-- Feedback alerts (only when feedback is active) -->
			<div v-if="isSubmitting || showFeedback" class="mt-2">
				<UAlert
					v-if="isSubmitting"
					color="neutral"
					variant="soft"
					:title="t('daily.submitting')"
				>
					<template #icon>
						<UIcon name="i-heroicons-arrow-path" class="animate-spin" />
					</template>
				</UAlert>

				<UAlert
					v-else-if="showFeedback && feedbackIsCorrect"
					color="green"
					variant="soft"
					:title="t('daily.correct')"
					icon="i-heroicons-check-circle"
				/>

				<UAlert
					v-else-if="showFeedback && feedbackIsCorrect === false"
					color="red"
					variant="soft"
					:title="t('daily.incorrect')"
					icon="i-heroicons-x-circle"
				/>
			</div>
		</div>
	</div>
</template>
