<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useDuelWebSocket } from '@/composables/useDuelWebSocket'
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
	opponent,
	myScore,
	opponentScore,
	isWaiting,
	isCountdown,
	isPlaying,
	isFinished,
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

const timerRef = ref<InstanceType<typeof GameTimer> | null>(null)

// ===========================
// UI State
// ===========================

const selectedAnswerId = ref<string | null>(null)
const showFeedback = ref(false)
const hasAnswered = ref(false)
const answerStartTime = ref(0)

// Show cancel button after 5s if no game state received
const isStuck = ref(false)
let stuckTimeout: ReturnType<typeof setTimeout> | null = null

// ===========================
// Computed
// ===========================

const totalRounds = computed(() => game.value?.totalRounds ?? 7)

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
		timerRef.value?.reset()
		timerRef.value?.start()
	}
})

// Watch for round result — only show feedback after player has answered
watch(lastRoundResult, (result) => {
	if (result && hasAnswered.value) {
		showFeedback.value = true
		timerRef.value?.stop()
	}
})

// Watch for match finished
watch(isFinished, (finished) => {
	if (finished) {
		timerRef.value?.stop()
		router.push({ name: 'duel-results', params: { duelId: duelId.value } })
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

const isCorrectAnswer = (answerId: string): boolean | null => {
	if (!showFeedback.value || !lastRoundResult.value) return null
	// Correct answer → green
	if (answerId === lastRoundResult.value.correctAnswerId) return true
	// Selected wrong answer → red
	if (answerId === selectedAnswerId.value) return false
	// Other answers → dimmed (null)
	return null
}

// ===========================
// Lifecycle
// ===========================

watch(game, (newGame) => {
	if (newGame && stuckTimeout) {
		clearTimeout(stuckTimeout)
		stuckTimeout = null
	}
})

onMounted(() => {
	connect()
	stuckTimeout = setTimeout(() => {
		if (!game.value) isStuck.value = true
	}, 5000)
})

onUnmounted(() => {
	if (stuckTimeout) clearTimeout(stuckTimeout)
})
</script>

<template>
	<div class="min-h-screen mx-auto max-w-[800px] flex flex-col">
		<!-- Connection Status -->
		<div
			v-if="!isConnected || isReconnecting"
			class="bg-yellow-500 text-white text-center py-2 text-sm"
		>
			<UIcon name="i-heroicons-wifi" class="inline size-4 mr-1" />
			{{ isReconnecting ? t('duel.reconnecting') : t('duel.connecting') }}
		</div>

		<!-- Error (не показываем timeout-ошибку — таймер уже сигнализирует) -->
		<div
			v-if="error && !error.toLowerCase().includes('invalid answer')"
			class="bg-red-500 text-white text-center py-2 text-sm"
		>
			{{ error }}
		</div>

		<!-- Initial connecting / loading state -->
		<div v-if="!game" class="flex-1 flex flex-col items-center justify-center gap-6 px-6">
			<div class="relative flex items-center justify-center">
				<div
					class="absolute size-24 rounded-full border-2 border-primary/20 animate-ping"
				/>
				<div
					class="absolute size-16 rounded-full border border-primary/10 animate-ping"
					style="animation-delay: 0.3s"
				/>
				<UIcon name="i-heroicons-bolt" class="size-14 text-primary relative z-10" />
			</div>
			<div class="text-center space-y-1">
				<p class="text-(--ui-text-highlighted) font-bold text-xl">
					{{ t('duel.connecting') }}
				</p>
				<p class="text-(--ui-text-muted) text-sm">{{ t('duel.waitingDesc') }}</p>
			</div>
			<UButton
				v-if="isStuck"
				color="neutral"
				variant="soft"
				icon="i-heroicons-arrow-left"
				@click="router.push({ name: 'duel-lobby', query: { skipRedirect: '1' } })"
			>
				{{ t('duel.cancelGame') }}
			</UButton>
		</div>

		<!-- Waiting for Opponent -->
		<div
			v-else-if="isWaiting"
			class="flex-1 flex flex-col items-center justify-center gap-8 px-6"
		>
			<div class="flex items-center gap-8">
				<!-- Me -->
				<div class="flex flex-col items-center gap-2">
					<UAvatar
						:src="currentUser?.avatarUrl"
						:alt="currentUser?.username"
						size="xl"
						class="ring-2 ring-primary/40"
					/>
					<p class="text-sm font-semibold text-(--ui-text-muted)">{{ t('duel.you') }}</p>
				</div>

				<span class="text-2xl font-black text-(--ui-text-dimmed)">VS</span>

				<!-- Unknown opponent -->
				<div class="flex flex-col items-center gap-2">
					<div
						class="size-16 rounded-full bg-(--ui-bg-accented) border-2 border-dashed border-(--ui-border) flex items-center justify-center"
					>
						<UIcon
							name="i-heroicons-question-mark-circle"
							class="size-8 text-(--ui-text-muted)"
						/>
					</div>
					<p class="text-sm font-semibold text-(--ui-text-dimmed)">?</p>
				</div>
			</div>

			<div class="text-center">
				<p class="text-(--ui-text-highlighted) font-semibold mb-1">
					{{ t('duel.waitingOpponent') }}
				</p>
				<p class="text-sm text-(--ui-text-muted)">{{ t('duel.waitingDesc') }}</p>
			</div>
		</div>

		<!-- Countdown -->
		<div
			v-else-if="isCountdown"
			class="flex-1 flex flex-col items-center justify-center gap-10 px-6"
		>
			<!-- Both players -->
			<div class="flex items-center gap-6 w-full max-w-xs">
				<div class="flex-1 flex flex-col items-center gap-2">
					<UAvatar
						:src="currentUser?.avatarUrl"
						:alt="currentUser?.username"
						size="xl"
						class="ring-2 ring-primary/40"
					/>
					<p
						class="text-sm font-semibold text-(--ui-text-muted) truncate max-w-[80px] text-center"
					>
						{{ t('duel.you') }}
					</p>
				</div>

				<span class="text-xl font-black text-(--ui-text-dimmed) shrink-0">VS</span>

				<div class="flex-1 flex flex-col items-center gap-2">
					<UAvatar
						:src="opponent?.avatar"
						:alt="opponent?.username"
						size="xl"
						class="ring-2 ring-orange-500/40"
					/>
					<p
						class="text-sm font-semibold text-(--ui-text-muted) truncate max-w-[80px] text-center"
					>
						{{ opponent?.username || '...' }}
					</p>
				</div>
			</div>

			<!-- Countdown number -->
			<div class="text-center">
				<p class="text-xs uppercase tracking-widest text-(--ui-text-muted) mb-2">
					{{ t('duel.getReady') }}
				</p>
				<p class="text-9xl font-black text-primary tabular-nums leading-none">
					{{ countdownSeconds }}
				</p>
			</div>
		</div>

		<!-- Playing -->
		<template v-else-if="isPlaying && currentQuestion">
			<!-- Header -->
			<div class="px-4 pt-2 pb-3 space-y-3">
				<!-- Row 1: round + title + menu -->
				<div class="flex items-center justify-between">
					<span class="text-xl font-bold text-(--ui-text-highlighted) tabular-nums">
						{{ currentRound }}/{{ totalRounds }}
					</span>
					<span class="text-sm font-semibold text-(--ui-text-muted)">{{
						t('duel.title')
					}}</span>
					<UIcon
						name="i-heroicons-ellipsis-horizontal-circle"
						class="size-6 text-(--ui-text-dimmed)"
					/>
				</div>

				<!-- Players row -->
				<div class="flex items-center gap-2">
					<!-- Me -->
					<div class="flex items-center gap-2 flex-1 min-w-0">
						<UAvatar
							:src="currentUser?.avatarUrl"
							:alt="currentUser?.username"
							size="sm"
							class="ring-2 ring-primary shrink-0"
						/>
						<div class="min-w-0">
							<p class="text-xs text-(--ui-text-muted) truncate">
								{{ t('duel.you') }}
							</p>
							<p class="text-xl font-black text-primary tabular-nums leading-tight">
								{{ myScore }}
							</p>
						</div>
					</div>

					<!-- VS -->
					<span class="text-[10px] font-bold text-(--ui-text-dimmed) px-1">VS</span>

					<!-- Opponent -->
					<div class="flex items-center gap-2 flex-row-reverse flex-1 min-w-0">
						<div class="relative shrink-0">
							<UAvatar
								:src="opponent?.avatar"
								:alt="opponent?.username"
								size="sm"
								class="ring-2 ring-orange-500"
							/>
							<div
								v-if="opponentAnswered"
								class="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 bg-green-500 rounded-full flex items-center justify-center"
							>
								<UIcon name="i-heroicons-check" class="size-2.5 text-white" />
							</div>
						</div>
						<div class="text-right min-w-0">
							<p class="text-xs text-(--ui-text-muted) truncate">
								{{ opponent?.username }}
							</p>
							<p
								class="text-xl font-black text-orange-500 tabular-nums leading-tight"
							>
								{{ opponentScore }}
							</p>
						</div>
					</div>
				</div>

				<!-- Timer bar -->
				<div class="relative h-6 rounded-full bg-(--ui-bg-accented) overflow-hidden">
					<GameTimer
						ref="timerRef"
						:initial-time="10"
						:auto-start="false"
						:warning-threshold="3"
						:on-timeout="handleTimeout"
						show-progress
						size="lg"
						class="sr-only"
					/>
					<div
						class="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-yellow-400 via-primary-500 to-primary-600 transition-all duration-1000"
						:style="{ width: `${((timerRef?.remainingTime ?? 10) / 10) * 100}%` }"
					/>
					<span
						class="absolute inset-0 flex items-center justify-center text-xs font-bold tabular-nums drop-shadow"
						:class="
							timerRef?.remainingTime !== undefined && timerRef.remainingTime <= 3
								? 'text-red-400'
								: 'text-white'
						"
					>
						{{ timerRef?.remainingTime ?? 10 }}s
					</span>
				</div>
			</div>

			<!-- Question -->
			<div class="flex-1 px-4 py-2">
				<QuestionCard
					:question="formattedQuestion!"
					:question-number="currentRound"
					:total-questions="totalRounds"
				/>

				<!-- Answers (colored bars) -->
				<div class="mt-6 flex flex-col gap-3">
					<AnswerButton
						v-for="(answer, index) in formattedAnswers"
						:key="answer.id"
						:answer="answer"
						:color-index="index % 4"
						:color-bar="true"
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
						<div class="flex justify-center gap-6 text-sm text-(--ui-text-muted)">
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
						class="text-2xl w-10 h-10 rounded-full bg-(--ui-bg-accented) flex items-center justify-center hover:scale-110 transition-transform"
						@click="() => sendEmote(emote)"
					>
						{{ emote }}
					</button>
					<span class="text-xs text-(--ui-text-dimmed)">{{ emotesLeft }}/3</span>
				</div>
			</div>

			<!-- Opponent Disconnect Overlay -->
			<Transition name="fade">
				<div
					v-if="opponentReconnecting"
					class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
				>
					<div
						class="bg-(--ui-bg-elevated) rounded-2xl p-6 mx-4 text-center max-w-sm w-full"
					>
						<div class="text-4xl mb-3">🔄</div>
						<h3 class="text-lg font-bold mb-1">
							{{ opponent?.username }} {{ t('duel.reconnecting') }}
						</h3>
						<p class="text-(--ui-text-muted) text-sm mb-3">
							{{ t('duel.reconnectingDesc') }}
						</p>
						<div class="text-3xl font-bold text-primary">
							{{ opponentReconnectCountdown }}s
						</div>
					</div>
				</div>
			</Transition>
		</template>
	</div>
</template>
