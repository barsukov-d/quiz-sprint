<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

const {
	state,
	isLoading,
	canContinue,
	continueOffer,
	continueGame,
	reset,
	initialize,
} = useMarathon(playerId)

const gameOverResult = computed(() => state.value.gameOverResult)

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

const handleBackToHome = () => {
	reset()
	router.push({ name: 'home' })
}

const handlePlayAgain = () => {
	reset()
	router.push({ name: 'home' })
}

onMounted(async () => {
	await initialize()

	// If there's no game over result and no active game, redirect home
	if (!state.value.gameOverResult && state.value.status !== 'game-over') {
		router.push({ name: 'home' })
	}
})
</script>

<template>
	<div class="min-h-screen mx-auto max-w-[800px] px-4 pt-14 pb-8 sm:px-3 sm:pt-12">
		<div class="flex flex-col items-center gap-6">
			<!-- Game Over Header -->
			<div class="text-center">
				<UIcon name="i-heroicons-trophy" class="size-16 text-yellow-500 mb-4" />
				<h1 class="text-2xl font-bold">Game Over</h1>
				<p class="text-gray-500 dark:text-gray-400 mt-1">Marathon Complete</p>
			</div>

			<!-- Score Card -->
			<UCard class="w-full">
				<div class="text-center space-y-4">
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">Final Score</p>
						<p class="text-4xl font-bold text-primary">
							{{ gameOverResult?.finalScore ?? state.score }}
						</p>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Questions</p>
							<p class="text-lg font-semibold">
								{{ gameOverResult?.totalQuestions ?? state.totalQuestions }}
							</p>
						</div>
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Personal Best</p>
							<p class="text-lg font-semibold">
								<template v-if="gameOverResult?.isNewPersonalBest">
									<span class="text-green-500">New Record!</span>
								</template>
								<template v-else>
									{{ gameOverResult?.previousRecord ?? '-' }}
								</template>
							</p>
						</div>
					</div>
				</div>
			</UCard>

			<!-- Continue Offer -->
			<UCard v-if="canContinue && continueOffer" class="w-full">
				<div class="text-center space-y-3">
					<h3 class="font-semibold">Continue Playing?</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">
						Pick up where you left off
					</p>

					<div class="flex flex-col gap-2">
						<UButton
							color="primary"
							block
							size="lg"
							:loading="isLoading"
							icon="i-heroicons-currency-dollar"
							@click="handleContinueWithCoins"
						>
							Continue ({{ continueOffer.costCoins }} coins)
						</UButton>

						<UButton
							v-if="continueOffer.hasAd"
							color="gray"
							variant="soft"
							block
							size="lg"
							:loading="isLoading"
							icon="i-heroicons-play"
							@click="handleContinueWithAd"
						>
							Watch Ad to Continue
						</UButton>
					</div>
				</div>
			</UCard>

			<!-- Actions -->
			<div class="w-full flex flex-col gap-2">
				<UButton
					color="primary"
					variant="soft"
					block
					size="lg"
					icon="i-heroicons-arrow-path"
					@click="handlePlayAgain"
				>
					Play Again
				</UButton>

				<UButton
					color="gray"
					variant="ghost"
					block
					size="lg"
					icon="i-heroicons-home"
					@click="handleBackToHome"
				>
					Back to Home
				</UButton>
			</div>
		</div>
	</div>
</template>
