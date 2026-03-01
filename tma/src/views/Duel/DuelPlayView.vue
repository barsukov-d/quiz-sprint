<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useDuelWebSocket } from '@/composables/useDuelWebSocket'
import { useGameTimer } from '@/composables/useGameTimer'
import GameTimer from '@/components/shared/GameTimer.vue'
import AnswerButton from '@/components/shared/AnswerButton.vue'
import QuestionCard from '@/components/shared/QuestionCard.vue'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()

const duelId = computed(() => route.params.duelId as string)
const playerId = computed(() => currentUser.value?.id ?? '')
const { t } = useI18n()

// ===========================
// WebSocket
// ===========================

const {
	isConnected,
	isReconnecting,
	error,
	connect,
	game,
	currentQuestion,
	currentRound,
	countdownSeconds,
	opponentAnswered,
	lastRoundResult,
	myPlayer,
	opponent,
	myScore,
	opponentScore,
	isWaiting,
	isCountdown,
	isPlaying,
	isFinished,
	didWin,
	myMmrChange,
	myAnswerCorrect,
	myAnswerTime,
	opponentAnswerTime,
	opponentReconnecting,
	opponentReconnectCountdown,
	emotesLeft,
	unlockedEmotes,
	sendAnswer,
	sendEmote,
} = useDuelWebSocket(duelId.value, playerId.value)

// ===========================
// Timer
// ===========================

const timer = useGameTimer({
	initialTime: 10,
	autoStart: false,
	onTimeout: () => handleTimeout(),
})

// ===========================
// UI State
// ===========================

const selectedAnswerId = ref<string | null>(null)
const showFeedback = ref(false)
const hasAnswered = ref(false)
const answerStartTime = ref(0)

// ===========================
// Computed
// ===========================

const totalRounds = computed(() => game.value?.totalRounds ?? 7)

const answerLabels = ['A', 'B', 'C', 'D']

// Cast question to expected DTO format
const formattedQuestion = computed(() => {
	if (!currentQuestion.value) return null
	return {
		id: currentQuestion.value.id,
		text: currentQuestion.value.text,
		answers: [],
		points: 100,
		position: currentRound.value,
	}
})

// Format answers for AnswerButton component
const formattedAnswers = computed(() => {
	if (!currentQuestion.value) return []
	return currentQuestion.value.answers.map((a, idx) => ({
		id: a.id,
		text: a.text,
		position: idx,
	}))
})

// ===========================
// Watch
// ===========================

// Watch for new round
watch(currentQuestion, (newQuestion) => {
	if (newQuestion) {
		// Reset UI state for new round
		selectedAnswerId.value = null
		showFeedback.value = false
		hasAnswered.value = false
		answerStartTime.value = Date.now()

		// Start timer
		timer.reset()
		timer.start()
	}
})

// Watch for round result — only show feedback after player has answered
watch(lastRoundResult, (result) => {
	if (result && hasAnswered.value) {
		showFeedback.value = true
		timer.stop()
	}
})

// Watch for match finished
watch(isFinished, (finished) => {
	if (finished) {
		timer.stop()
		// Navigate to results after short delay
		setTimeout(() => {
			router.push({ name: 'duel-results', params: { duelId: duelId.value } })
		}, 2000)
	}
})

// ===========================
// Actions
// ===========================

const handleAnswerSelect = async (answerId: string) => {
	if (hasAnswered.value || showFeedback.value) return

	selectedAnswerId.value = answerId
	hasAnswered.value = true

	const timeTaken = Date.now() - answerStartTime.value

	// Send answer via WebSocket
	sendAnswer(answerId, timeTaken)
}

const handleTimeout = () => {
	if (!hasAnswered.value) {
		// Auto-submit empty answer on timeout
		hasAnswered.value = true
		sendAnswer('', 10000)
	}
}

const isCorrectAnswer = (answerId: string) => {
	if (!lastRoundResult.value) return false
	return answerId === lastRoundResult.value.correctAnswerId
}

// ===========================
// Lifecycle
// ===========================

onMounted(() => {
	connect()
})
</script>

<template>
	<div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex flex-col">
		<!-- Connection Status -->
		<div
			v-if="!isConnected || isReconnecting"
			class="bg-yellow-500 text-white text-center py-2 text-sm"
		>
			<UIcon name="i-heroicons-wifi" class="inline size-4 mr-1" />
			{{ isReconnecting ? t('duel.reconnecting') : t('duel.connecting') }}
		</div>

		<!-- Error -->
		<div v-if="error" class="bg-red-500 text-white text-center py-2 text-sm">
			{{ error }}
		</div>

		<!-- Initial connecting / loading state -->
		<div v-if="!game" class="flex-1 flex items-center justify-center p-4">
			<div class="text-center">
				<div class="animate-pulse mb-4">
					<UIcon name="i-heroicons-bolt" class="size-16 text-primary" />
				</div>
				<h2 class="text-xl font-bold mb-2">{{ t('duel.connecting') }}</h2>
			</div>
		</div>

		<!-- Waiting for Opponent -->
		<div v-else-if="isWaiting" class="flex-1 flex items-center justify-center p-4">
			<div class="text-center">
				<div class="animate-pulse mb-4">
					<UIcon name="i-heroicons-users" class="size-16 text-primary" />
				</div>
				<h2 class="text-xl font-bold mb-2">{{ t('duel.waitingOpponent') }}</h2>
				<p class="text-gray-500">{{ t('duel.waitingDesc') }}</p>
			</div>
		</div>

		<!-- Countdown -->
		<div v-else-if="isCountdown" class="flex-1 flex items-center justify-center p-4">
			<div class="text-center">
				<p class="text-gray-500 mb-2">{{ t('duel.getReady') }}</p>
				<p class="text-8xl font-bold text-primary animate-pulse">
					{{ countdownSeconds }}
				</p>
			</div>
		</div>

		<!-- Playing -->
		<template v-else-if="isPlaying && currentQuestion">
			<!-- Header: Players & Scores -->
			<div class="bg-white dark:bg-gray-800 shadow-sm px-4 py-3">
				<div class="flex items-center justify-between">
					<!-- My Player -->
					<div class="flex items-center gap-2">
						<div
							class="w-10 h-10 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center"
						>
							<span class="text-lg">{{ myPlayer?.leagueIcon }}</span>
						</div>
						<div>
							<p class="font-medium text-sm">{{ t('duel.you') }}</p>
							<p class="text-2xl font-bold text-primary">{{ myScore }}</p>
						</div>
					</div>

					<!-- VS / Round -->
					<div class="text-center">
						<p class="text-xs text-gray-500">{{ t('duel.round') }}</p>
						<p class="font-bold">
							{{ t('duel.roundOf', { current: currentRound, total: totalRounds }) }}
						</p>
					</div>

					<!-- Opponent -->
					<div class="flex items-center gap-2">
						<div class="text-right">
							<p class="font-medium text-sm">{{ opponent?.username }}</p>
							<p class="text-2xl font-bold text-orange-500">{{ opponentScore }}</p>
						</div>
						<div
							class="w-10 h-10 rounded-full bg-orange-100 dark:bg-orange-900 flex items-center justify-center relative"
						>
							<span class="text-lg">{{ opponent?.leagueIcon }}</span>
							<!-- Answered indicator -->
							<div
								v-if="opponentAnswered"
								class="absolute -bottom-1 -right-1 w-4 h-4 bg-green-500 rounded-full flex items-center justify-center"
							>
								<UIcon name="i-heroicons-check" class="size-3 text-white" />
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Timer -->
			<div class="px-4 py-3">
				<GameTimer
					ref="timerRef"
					:initial-time="10"
					:auto-start="false"
					:warning-threshold="3"
					show-progress
					size="lg"
				/>
			</div>

			<!-- Question -->
			<div class="flex-1 px-4 py-2">
				<QuestionCard
					:question="formattedQuestion!"
					:question-number="currentRound"
					:total-questions="totalRounds"
				/>

				<!-- Answers -->
				<div class="mt-6 space-y-3">
					<AnswerButton
						v-for="(answer, index) in formattedAnswers"
						:key="answer.id"
						:answer="answer"
						:label="answerLabels[index]"
						:selected="selectedAnswerId === answer.id"
						:disabled="hasAnswered"
						:show-feedback="showFeedback"
						:is-correct="isCorrectAnswer(answer.id)"
						@click="() => handleAnswerSelect(answer.id)"
					/>
				</div>

				<!-- Answer Feedback (1.5s overlay after answering) -->
				<Transition name="fade">
					<div
						v-if="showFeedback && myAnswerCorrect !== null"
						class="mt-4 rounded-xl p-4 text-center"
						:class="
							myAnswerCorrect
								? 'bg-green-100 dark:bg-green-900'
								: 'bg-red-100 dark:bg-red-900'
						"
					>
						<p class="text-lg font-bold mb-1">
							{{
								myAnswerCorrect
									? '✅ ' + t('duel.correct')
									: '❌ ' + t('duel.wrong')
							}}
						</p>
						<div
							class="flex justify-center gap-6 text-sm text-gray-600 dark:text-gray-300"
						>
							<span
								>{{ t('duel.yourTime') }}:
								{{ (myAnswerTime / 1000).toFixed(1) }}s</span
							>
							<span v-if="opponentAnswered">
								{{ opponent?.username }}:
								{{ (opponentAnswerTime / 1000).toFixed(1) }}s
							</span>
						</div>
					</div>
				</Transition>

				<!-- Emote Bar -->
				<div v-if="emotesLeft > 0" class="mt-3 flex items-center justify-center gap-2">
					<button
						v-for="emote in unlockedEmotes"
						:key="emote"
						class="text-2xl w-10 h-10 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center hover:scale-110 transition-transform"
						@click="() => sendEmote(emote)"
					>
						{{ emote }}
					</button>
					<span class="text-xs text-gray-400">{{ emotesLeft }}/3</span>
				</div>
			</div>

			<!-- Opponent Disconnect Overlay -->
			<Transition name="fade">
				<div
					v-if="opponentReconnecting"
					class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
				>
					<div
						class="bg-white dark:bg-gray-800 rounded-2xl p-6 mx-4 text-center max-w-sm w-full"
					>
						<div class="text-4xl mb-3">🔄</div>
						<h3 class="text-lg font-bold mb-1">
							{{ opponent?.username }} {{ t('duel.reconnecting') }}
						</h3>
						<p class="text-gray-500 text-sm mb-3">{{ t('duel.reconnectingDesc') }}</p>
						<div class="text-3xl font-bold text-primary">
							{{ opponentReconnectCountdown }}s
						</div>
					</div>
				</div>
			</Transition>
		</template>

		<!-- Match Finished (brief summary before redirect) -->
		<div v-else-if="isFinished" class="flex-1 flex items-center justify-center p-4">
			<div class="text-center">
				<div class="mb-4">
					<UIcon
						:name="didWin ? 'i-heroicons-trophy' : 'i-heroicons-x-circle'"
						:class="didWin ? 'text-yellow-500' : 'text-red-500'"
						class="size-20"
					/>
				</div>
				<h2 class="text-3xl font-bold mb-2">
					{{
						didWin === null
							? t('duel.draw')
							: didWin
								? t('duel.victory')
								: t('duel.defeat')
					}}
				</h2>
				<p class="text-xl">{{ myScore }} - {{ opponentScore }}</p>
				<p
					:class="myMmrChange >= 0 ? 'text-green-600' : 'text-red-600'"
					class="text-lg font-semibold mt-2"
				>
					{{
						t('duel.mmrChange', {
							sign: myMmrChange >= 0 ? '+' : '',
							amount: myMmrChange,
						})
					}}
				</p>
			</div>
		</div>
	</div>
</template>
