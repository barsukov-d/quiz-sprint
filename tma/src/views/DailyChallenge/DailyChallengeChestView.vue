<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useAuth } from '@/composables/useAuth'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { currentUser } = useAuth()
const { t } = useI18n()
const playerId = currentUser.value?.id || 'guest'

const { game, openChest } = useDailyChallenge(playerId)

// Animation states
const phase = ref<'idle' | 'shaking' | 'opening' | 'rewards'>('idle')
const chestData = ref<{
	chestType: string
	rewards: { coins: number; pvpTickets: number; marathonBonuses: string[] }
} | null>(null)

const chestEmoji = computed(() => {
	const type = chestData.value?.chestType || ''
	if (type.includes('golden')) return '🏆'
	if (type.includes('silver')) return '🥈'
	return '📦'
})

const chestLabel = computed(() => {
	const type = chestData.value?.chestType || ''
	if (type.includes('golden')) return t('daily.goldenChest')
	if (type.includes('silver')) return t('daily.silverChest')
	return t('daily.woodenChest')
})

const handleOpenChest = async () => {
	phase.value = 'shaking'

	// Shake animation for 1.5s
	await new Promise((r) => setTimeout(r, 1500))

	phase.value = 'opening'

	try {
		const result = await openChest()
		chestData.value = result || null
	} catch {
		// Fallback — chest already opened, go to results
		router.replace({ name: 'daily-challenge-results' })
		return
	}

	// Brief pause before showing rewards
	await new Promise((r) => setTimeout(r, 500))
	phase.value = 'rewards'
}

const goToResults = () => {
	router.replace({ name: 'daily-challenge-results' })
}

onMounted(() => {
	if (!game.value?.gameId) {
		router.replace({ name: 'daily-challenge-results' })
	}
})
</script>

<template>
	<div class="min-h-screen flex flex-col items-center justify-center px-6 bg-(--ui-bg)">
		<!-- Phase: Idle — tap to open -->
		<div v-if="phase === 'idle'" class="flex flex-col items-center gap-6 text-center">
			<div class="text-sm font-medium text-(--ui-text-muted) uppercase tracking-wider">
				{{ t('daily.yourReward') }}
			</div>

			<div class="text-8xl animate-bounce">{{ chestEmoji }}</div>

			<button
				class="px-8 py-4 bg-primary-500 hover:bg-primary-600 text-white font-bold text-lg rounded-(--ui-radius) transition-all active:scale-95 shadow-lg shadow-primary-500/25"
				@click="handleOpenChest"
			>
				{{ t('daily.openChest') }}
			</button>
		</div>

		<!-- Phase: Shaking -->
		<div v-if="phase === 'shaking'" class="flex flex-col items-center gap-6 text-center">
			<div class="text-sm font-medium text-(--ui-text-muted) uppercase tracking-wider">
				{{ t('daily.opening') }}...
			</div>

			<div class="text-8xl animate-chest-shake">{{ chestEmoji }}</div>
		</div>

		<!-- Phase: Opening (brief flash) -->
		<div v-if="phase === 'opening'" class="flex flex-col items-center gap-6 text-center">
			<div class="text-9xl animate-pulse">✨</div>
		</div>

		<!-- Phase: Rewards revealed -->
		<div
			v-if="phase === 'rewards' && chestData"
			class="flex flex-col items-center gap-6 w-full max-w-sm"
		>
			<div class="text-center">
				<div class="text-5xl mb-2">{{ chestEmoji }}</div>
				<h2 class="text-xl font-bold text-(--ui-text-highlighted)">{{ chestLabel }}</h2>
			</div>

			<!-- Reward cards -->
			<div class="w-full flex flex-col gap-3">
				<!-- Coins -->
				<div
					v-if="chestData.rewards.coins > 0"
					class="flex items-center gap-4 px-4 py-3 bg-(--ui-bg-elevated) rounded-(--ui-radius) border border-(--ui-border) animate-fade-in-up"
				>
					<span class="text-3xl">🪙</span>
					<div class="flex-1">
						<div class="text-lg font-bold text-(--ui-text-highlighted)">
							+{{ chestData.rewards.coins }}
						</div>
						<div class="text-xs text-(--ui-text-muted)">{{ t('daily.coins') }}</div>
					</div>
				</div>

				<!-- PvP Tickets -->
				<div
					v-if="chestData.rewards.pvpTickets > 0"
					class="flex items-center gap-4 px-4 py-3 bg-(--ui-bg-elevated) rounded-(--ui-radius) border border-(--ui-border) animate-fade-in-up"
					style="animation-delay: 0.15s"
				>
					<span class="text-3xl">🎟️</span>
					<div class="flex-1">
						<div class="text-lg font-bold text-(--ui-text-highlighted)">
							+{{ chestData.rewards.pvpTickets }}
						</div>
						<div class="text-xs text-(--ui-text-muted)">
							{{ t('daily.pvpTickets') }}
						</div>
					</div>
				</div>

				<!-- Marathon Bonuses -->
				<div
					v-for="(bonus, index) in chestData.rewards.marathonBonuses"
					:key="bonus"
					class="flex items-center gap-4 px-4 py-3 bg-(--ui-bg-elevated) rounded-(--ui-radius) border border-(--ui-border) animate-fade-in-up"
					:style="{ animationDelay: `${0.3 + index * 0.15}s` }"
				>
					<span class="text-3xl">{{
						bonus === 'shield'
							? '🛡️'
							: bonus === 'freeze'
								? '❄️'
								: bonus === 'fifty_fifty'
									? '5️⃣'
									: '⏭️'
					}}</span>
					<div class="flex-1">
						<div class="text-lg font-bold text-(--ui-text-highlighted)">+1</div>
						<div class="text-xs text-(--ui-text-muted)">{{ bonus }}</div>
					</div>
				</div>
			</div>

			<!-- Continue button -->
			<button
				class="w-full px-6 py-3 bg-(--ui-bg-elevated) hover:bg-(--ui-bg-accented) text-(--ui-text-highlighted) font-semibold rounded-(--ui-radius) border border-(--ui-border) transition-all mt-4"
				@click="goToResults"
			>
				{{ t('common.continue') }}
			</button>
		</div>
	</div>
</template>

<style scoped>
@keyframes chest-shake {
	0%,
	100% {
		transform: rotate(0deg);
	}
	10%,
	30%,
	50%,
	70%,
	90% {
		transform: rotate(-8deg);
	}
	20%,
	40%,
	60%,
	80% {
		transform: rotate(8deg);
	}
}

@keyframes fade-in-up {
	from {
		opacity: 0;
		transform: translateY(20px);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}

.animate-chest-shake {
	animation: chest-shake 0.8s ease-in-out infinite;
}

.animate-fade-in-up {
	opacity: 0;
	animation: fade-in-up 0.4s ease-out forwards;
}
</style>
