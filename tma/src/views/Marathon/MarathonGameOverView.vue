<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useMarathonSession } from '@/composables/useMarathonSession'
import { useAuth } from '@/composables/useAuth'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { currentUser } = useAuth()
const { t } = useI18n()
const playerId = currentUser.value?.id || 'guest'

const { state, isLoading, canContinue, continueOffer, continueGame, reset, initialize } =
	useMarathon(playerId)

const session = useMarathonSession()
const resultRecorded = ref(false)

const gameOverResult = computed(() => state.value.gameOverResult)

const motivationalPrompt = computed(() =>
	session.getMotivationalPrompt(
		gameOverResult.value?.finalScore ?? state.value.score,
		state.value.personalBest,
	),
)

const handleContinueWithCoins = async () => {
	try {
		await continueGame('coins')
	} catch (error) {
		console.error('Failed to continue with coins:', error)
	}
}

const handleContinueWithAd = async () => {
	try {
		await continueGame('ad')
	} catch (error) {
		console.error('Failed to continue with ad:', error)
	}
}

const handleStartNewRun = () => {
	reset()
	router.push({ name: 'marathon-category' })
}

const handleBackToHome = () => {
	reset()
	session.resetSession()
	router.push({ name: 'home' })
}

onMounted(async () => {
	await initialize()

	if (!state.value.gameOverResult && state.value.status !== 'game-over') {
		router.push({ name: 'home' })
		return
	}

	if (!resultRecorded.value) {
		session.recordRunResult(
			gameOverResult.value?.finalScore ?? state.value.score,
			state.value.streakCount,
		)
		resultRecorded.value = true
	}
})
</script>

<template>
	<div
		class="min-h-screen bg-gradient-to-b from-primary-600 to-primary-900 flex flex-col px-4 pb-8"
	>
		<!-- Close button -->
		<div class="pt-4 pb-2">
			<UButton
				color="neutral"
				variant="ghost"
				icon="i-heroicons-x-mark"
				size="sm"
				class="text-white/70 hover:text-white"
				@click="handleBackToHome"
			/>
		</div>

		<!-- Score section -->
		<div class="flex flex-col items-center gap-2 py-8">
			<UIcon name="i-heroicons-trophy" class="size-14 text-yellow-300 mb-2" />
			<p class="text-white/60 text-sm uppercase tracking-wide font-medium">
				{{ t('marathon.correctAnswers') }}
			</p>
			<p class="text-6xl font-bold text-white tabular-nums">
				{{ gameOverResult?.finalScore ?? state.score }}
			</p>
			<span
				v-if="gameOverResult?.isNewPersonalBest"
				class="mt-1 px-3 py-1 rounded-full bg-yellow-400/20 text-yellow-300 text-xs font-bold uppercase tracking-wide"
			>
				{{ t('marathon.newRecord') }}
			</span>
		</div>

		<!-- Stats summary -->
		<div class="grid grid-cols-2 gap-3 mb-6">
			<div class="rounded-(--ui-radius) bg-white/10 p-4 text-center">
				<p class="text-white/60 text-xs mb-1">{{ t('marathon.questions') }}</p>
				<p class="text-xl font-bold text-white tabular-nums">
					{{ gameOverResult?.totalQuestions ?? state.totalQuestions }}
				</p>
			</div>
			<div class="rounded-(--ui-radius) bg-white/10 p-4 text-center">
				<p class="text-white/60 text-xs mb-1">{{ t('marathon.personalBest') }}</p>
				<p class="text-xl font-bold text-white tabular-nums">
					{{ gameOverResult?.previousRecord ?? state.personalBest ?? '-' }}
				</p>
			</div>
		</div>

		<!-- Continue Offer -->
		<div
			v-if="canContinue && continueOffer"
			class="rounded-(--ui-radius) bg-white/10 p-4 mb-6 space-y-3"
		>
			<h3 class="font-semibold text-white text-center">{{ t('marathon.continueRun') }}</h3>
			<p class="text-sm text-white/60 text-center">{{ t('marathon.continueRunDesc') }}</p>
			<div class="flex flex-col gap-2">
				<UButton
					color="primary"
					block
					size="lg"
					:loading="isLoading"
					icon="i-heroicons-currency-dollar"
					@click="handleContinueWithCoins"
				>
					{{ t('marathon.continueWithCoins', { coins: continueOffer.costCoins }) }}
				</UButton>
				<UButton
					v-if="continueOffer.hasAd"
					color="neutral"
					variant="soft"
					block
					size="lg"
					:loading="isLoading"
					icon="i-heroicons-play"
					@click="handleContinueWithAd"
				>
					{{ t('marathon.watchAd') }}
				</UButton>
			</div>
		</div>

		<!-- Motivational Prompt -->
		<p class="text-center text-sm text-white/70 font-medium mb-2">{{ motivationalPrompt }}</p>

		<!-- Session Stats -->
		<p v-if="session.runCount.value >= 2" class="text-center text-xs text-white/50 mb-6">
			{{ session.sessionLabel.value }}
		</p>

		<!-- Actions as text links -->
		<div class="flex flex-col items-center gap-4 mt-auto pt-4">
			<button
				class="text-white font-semibold text-base hover:opacity-80 active:opacity-60 transition-opacity"
				@click="handleStartNewRun"
			>
				{{ t('marathon.newRun') }}
			</button>
			<button
				class="text-white/50 text-sm hover:opacity-80 active:opacity-60 transition-opacity"
				@click="handleBackToHome"
			>
				{{ t('marathon.home') }}
			</button>
		</div>
	</div>
</template>
