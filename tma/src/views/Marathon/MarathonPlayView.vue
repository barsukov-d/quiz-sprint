<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import type { BonusType } from '@/composables/useMarathon'
import { useAuth } from '@/composables/useAuth'
import GameTimer from '@/components/shared/GameTimer.vue'
import QuestionCard from '@/components/shared/QuestionCard.vue'
import AnswerButton from '@/components/shared/AnswerButton.vue'

// ===========================
// Auth & Router
// ===========================

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

// ===========================
// Marathon Composable
// ===========================

const {
	state,
	isPlaying,
	lives,
	canUseShield,
	canUseFiftyFifty,
	canUseSkip,
	canUseFreeze,
	submitAnswer,
	applyAnswerResult,
	useBonus,
	initialize,
} = useMarathon(playerId)

// ===========================
// Local State (UI only)
// ===========================

const selectedAnswerId = ref<string | null>(null)
const isSubmitting = ref(false)
const timerRef = ref<InstanceType<typeof GameTimer> | null>(null)
const questionStartTime = ref(Date.now())

const showFeedback = ref(false)
const feedbackIsCorrect = ref<boolean | null>(null)
const feedbackCorrectAnswerId = ref<string | null>(null)
const feedbackIsGameOver = ref(false)
const feedbackLifeLost = ref(false)
const feedbackShieldConsumed = ref(false)

// ===========================
// Computed
// ===========================

const answerLabels = ['A', 'B', 'C', 'D']

const currentQuestion = computed(() => state.value.currentQuestion)

const visibleAnswers = computed(() => {
	if (!currentQuestion.value?.answers) return []
	return currentQuestion.value.answers.map((answer) => ({
		...answer,
		hidden: state.value.hiddenAnswerIds.includes(answer.id),
	}))
})

const livesDisplay = computed(() => {
	const hearts = []
	for (let i = 0; i < lives.value.maxLives; i++) {
		hearts.push(i < lives.value.currentLives)
	}
	return hearts
})

const canSubmit = computed(() => {
	return selectedAnswerId.value !== null && !isSubmitting.value && !showFeedback.value
})

// ===========================
// Methods
// ===========================

const handleAnswerSelect = async (answerId: string) => {
	if (isSubmitting.value || showFeedback.value || timerRef.value?.remainingTime === 0) return

	selectedAnswerId.value = answerId
	await handleSubmit()
}

const handleSubmit = async () => {
	if (!canSubmit.value || !selectedAnswerId.value) return

	isSubmitting.value = true

	try {
		// Calculate time taken in milliseconds using precise Date.now()
		const timeTaken = Date.now() - questionStartTime.value
		const answerData = await submitAnswer(selectedAnswerId.value, timeTaken)

		timerRef.value?.pause()

		feedbackIsCorrect.value = answerData.isCorrect
		feedbackCorrectAnswerId.value = answerData.correctAnswerId
		feedbackIsGameOver.value = answerData.isGameOver
		feedbackLifeLost.value = answerData.lifeLost
		feedbackShieldConsumed.value = answerData.shieldConsumed
		showFeedback.value = true
		isSubmitting.value = false

		const delay = answerData.isGameOver ? 2000 : answerData.lifeLost ? 1800 : 1500
		setTimeout(() => {
			handleNextStep()
		}, delay)
	} catch (error) {
		console.error('Failed to submit answer:', error)
		isSubmitting.value = false
	}
}

const handleTimeout = async () => {
	if (showFeedback.value || isSubmitting.value) return

	if (selectedAnswerId.value) {
		await handleSubmit()
	} else {
		if (!currentQuestion.value?.answers?.[0]) return

		selectedAnswerId.value = currentQuestion.value.answers[0].id
		isSubmitting.value = true

		try {
			const answerData = await submitAnswer(selectedAnswerId.value, state.value.timeLimit * 1000)

			feedbackIsCorrect.value = answerData.isCorrect
			feedbackCorrectAnswerId.value = answerData.correctAnswerId
			feedbackIsGameOver.value = answerData.isGameOver
			feedbackLifeLost.value = answerData.lifeLost
			feedbackShieldConsumed.value = answerData.shieldConsumed
			showFeedback.value = true
			isSubmitting.value = false

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
	const isGameOver = feedbackIsGameOver.value

	// Reset feedback state
	showFeedback.value = false
	selectedAnswerId.value = null
	isSubmitting.value = false
	feedbackIsCorrect.value = null
	feedbackCorrectAnswerId.value = null
	feedbackIsGameOver.value = false
	feedbackLifeLost.value = false
	feedbackShieldConsumed.value = false

	// Apply deferred state transition (next question or game over navigation)
	await applyAnswerResult()

	if (isGameOver) {
		// Navigation handled by applyAnswerResult (routes to marathon-gameover)
		return
	}

	// Next question - reset timer with server-provided time limit
	await nextTick()
	questionStartTime.value = Date.now()
	timerRef.value?.reset(state.value.timeLimit)
	timerRef.value?.start()
}

const handleUseBonus = async (bonusType: BonusType) => {
	if (isSubmitting.value || showFeedback.value) return

	try {
		await useBonus(bonusType)

		// If skip, reset timer for new question
		if (bonusType === 'skip') {
			await nextTick()
			questionStartTime.value = Date.now()
			timerRef.value?.reset(state.value.timeLimit)
			timerRef.value?.start()
		}

		// If freeze, update timer with new time limit
		if (bonusType === 'freeze') {
			timerRef.value?.addTime(5)
		}
	} catch (error) {
		console.error('Failed to use bonus:', error)
	}
}

const getAnswerFeedback = (answerId: string) => {
	if (!showFeedback.value || feedbackIsCorrect.value === null) {
		return { showFeedback: false, isCorrect: null as boolean | null }
	}

	if (answerId === feedbackCorrectAnswerId.value) {
		return { showFeedback: true, isCorrect: true }
	}

	if (answerId === selectedAnswerId.value && !feedbackIsCorrect.value) {
		return { showFeedback: true, isCorrect: false }
	}

	return { showFeedback: true, isCorrect: null as boolean | null }
}

const startTimer = () => {
	if (!timerRef.value) return
	questionStartTime.value = Date.now()
	timerRef.value.reset(state.value.timeLimit)
	timerRef.value.start()
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	try {
		await initialize()
	} catch (error: unknown) {
		if (error instanceof Error && error.message === 'canceled') return
		console.error('[MarathonPlayView] Failed to initialize:', error)
	}

	if (!isPlaying.value) {
		router.push({ name: 'home' })
		return
	}

	await nextTick()
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
			<p class="text-gray-500 dark:text-gray-400 mt-4">Loading question...</p>
		</div>

		<!-- Game View -->
		<div v-else class="flex flex-col gap-4">
			<!-- Header: lives + score + timer -->
			<div class="flex items-center gap-3">
				<!-- Lives -->
				<div class="flex gap-0.5 shrink-0">
					<UIcon
						v-for="(filled, index) in livesDisplay"
						:key="index"
						:name="filled ? 'i-heroicons-heart-solid' : 'i-heroicons-heart'"
						:class="filled ? 'text-red-500' : 'text-gray-300 dark:text-gray-600'"
						class="size-4"
					/>
				</div>

				<!-- Score -->
				<span class="shrink-0 text-sm font-semibold text-primary tabular-nums">
					{{ state.score }}
				</span>

				<!-- Shield indicator -->
				<UIcon
					v-if="state.shieldActive"
					name="i-heroicons-shield-check"
					class="size-4 text-blue-500 shrink-0"
				/>

				<div class="flex-1" />

				<!-- Timer -->
				<GameTimer
					ref="timerRef"
					:initial-time="state.timeLimit"
					:auto-start="false"
					:warning-threshold="5"
					:show-progress="false"
					size="sm"
					:on-timeout="handleTimeout"
					class="shrink-0"
				/>
			</div>

			<!-- Milestone progress -->
			<div
				v-if="state.milestone && state.milestone.next > 0"
				class="text-xs text-center text-gray-500 dark:text-gray-400"
			>
				Next milestone: {{ state.milestone.next }} ({{ state.milestone.remaining }} to go)
			</div>

			<!-- Question -->
			<QuestionCard
				:question="currentQuestion"
				:question-number="state.totalQuestions"
				:show-badge="false"
			/>

			<!-- Answer Buttons -->
			<div class="flex flex-col gap-3">
				<AnswerButton
					v-for="(answer, index) in visibleAnswers"
					:key="answer.id"
					:answer="answer"
					:selected="selectedAnswerId === answer.id"
					:disabled="isSubmitting || showFeedback || timerRef?.remainingTime === 0 || answer.hidden"
					:show-feedback="getAnswerFeedback(answer.id).showFeedback"
					:is-correct="getAnswerFeedback(answer.id).isCorrect"
					:label="answerLabels[index]"
					:class="{ 'opacity-0 pointer-events-none': answer.hidden }"
					@click="handleAnswerSelect"
				/>
			</div>

			<!-- Bonus Bar -->
			<div class="flex gap-2 justify-center pt-2">
				<UButton
					size="sm"
					:color="canUseShield ? 'blue' : 'gray'"
					variant="soft"
					:disabled="!canUseShield || isSubmitting || showFeedback"
					icon="i-heroicons-shield-check"
					@click="handleUseBonus('shield')"
				>
					{{ state.bonusInventory.shield }}
				</UButton>
				<UButton
					size="sm"
					:color="canUseFiftyFifty ? 'yellow' : 'gray'"
					variant="soft"
					:disabled="!canUseFiftyFifty || isSubmitting || showFeedback"
					icon="i-heroicons-scissors"
					@click="handleUseBonus('fifty_fifty')"
				>
					{{ state.bonusInventory.fiftyFifty }}
				</UButton>
				<UButton
					size="sm"
					:color="canUseSkip ? 'green' : 'gray'"
					variant="soft"
					:disabled="!canUseSkip || isSubmitting || showFeedback"
					icon="i-heroicons-forward"
					@click="handleUseBonus('skip')"
				>
					{{ state.bonusInventory.skip }}
				</UButton>
				<UButton
					size="sm"
					:color="canUseFreeze ? 'cyan' : 'gray'"
					variant="soft"
					:disabled="!canUseFreeze || isSubmitting || showFeedback"
					icon="i-heroicons-clock"
					@click="handleUseBonus('freeze')"
				>
					{{ state.bonusInventory.freeze }}
				</UButton>
			</div>

			<!-- Feedback alerts -->
			<div v-if="isSubmitting || showFeedback" class="mt-2">
				<UAlert
					v-if="isSubmitting"
					color="gray"
					variant="soft"
					title="Submitting answer..."
				>
					<template #icon>
						<UIcon name="i-heroicons-arrow-path" class="animate-spin" />
					</template>
				</UAlert>

				<UAlert
					v-else-if="showFeedback && feedbackShieldConsumed"
					color="blue"
					variant="soft"
					title="Shield absorbed the hit!"
					icon="i-heroicons-shield-check"
				/>

				<UAlert
					v-else-if="showFeedback && feedbackIsCorrect"
					color="green"
					variant="soft"
					title="Correct!"
					icon="i-heroicons-check-circle"
				/>

				<UAlert
					v-else-if="showFeedback && feedbackLifeLost"
					color="red"
					variant="soft"
					title="Wrong! Life lost"
					icon="i-heroicons-heart"
				/>

				<UAlert
					v-else-if="showFeedback && feedbackIsCorrect === false"
					color="red"
					variant="soft"
					title="Incorrect"
					icon="i-heroicons-x-circle"
				/>
			</div>
		</div>
	</div>
</template>
