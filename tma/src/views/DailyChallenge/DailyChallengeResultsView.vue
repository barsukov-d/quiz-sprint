<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useAuth } from '@/composables/useAuth'
import { useStreaks } from '@/composables/useStreaks'
import DailyChallengeLeaderboard from '@/components/DailyChallenge/DailyChallengeLeaderboard.vue'
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
	results,
	game,
	streak,
	isCompleted,
	timeToExpireFormatted,
	initialize,
	retryChallenge,
	isLoading,
} = useDailyChallenge(playerId)

const streaks = useStreaks(streak)

const scorePercentage = computed(() => {
	if (!results.value) return 0
	return Math.round((results.value.correctAnswers / results.value.totalQuestions) * 100)
})

const performanceLevel = computed(() => {
	const pct = scorePercentage.value
	if (pct >= 90) return { label: 'Excellent!', color: 'success' as const, emoji: '🌟' }
	if (pct >= 70) return { label: 'Great!', color: 'info' as const, emoji: '👏' }
	if (pct >= 50) return { label: 'Good!', color: 'warning' as const, emoji: '👍' }
	return { label: 'Keep trying!', color: 'neutral' as const, emoji: '💪' }
})

const hasNewStreakRecord = computed(() => {
	if (!streak.value) return false
	return streak.value.currentStreak > streak.value.bestStreak
})

const chestReward = computed(() => results.value?.chestReward || null)

const chestEmoji = computed(() => {
	if (!chestReward.value) return '📦'
	switch (chestReward.value.chestType) {
		case 'golden':
			return '🏆'
		case 'silver':
			return '🥈'
		case 'wooden':
			return '📦'
		default:
			return '📦'
	}
})

const chestLabel = computed(() => {
	if (!chestReward.value) return 'Chest'
	switch (chestReward.value.chestType) {
		case 'golden':
			return 'Golden Chest'
		case 'silver':
			return 'Silver Chest'
		case 'wooden':
			return 'Wooden Chest'
		default:
			return 'Chest'
	}
})


// Bonus display mapping
const bonusInfo: Record<
	string,
	{ get label(): string; icon: string; color: string; get description(): string }
> = {
	shield: {
		get label() {
			return t('daily.shieldName')
		},
		icon: 'i-heroicons-shield-check',
		color: 'text-blue-500',
		get description() {
			return t('daily.shieldDesc')
		},
	},
	fifty_fifty: {
		get label() {
			return t('daily.fiftyfiftyName')
		},
		icon: 'i-heroicons-scissors',
		color: 'text-yellow-500',
		get description() {
			return t('daily.fiftyfiftyDesc')
		},
	},
	skip: {
		get label() {
			return t('daily.skipName')
		},
		icon: 'i-heroicons-forward',
		color: 'text-green-500',
		get description() {
			return t('daily.skipDesc')
		},
	},
	freeze: {
		get label() {
			return t('daily.freezeName')
		},
		icon: 'i-heroicons-clock',
		color: 'text-cyan-500',
		get description() {
			return t('daily.freezeDesc')
		},
	},
}

const getBonusInfo = (bonus: string) =>
	bonusInfo[bonus] ?? {
		label: bonus,
		icon: 'i-heroicons-gift',
		color: 'text-(--ui-text-muted)',
		description: '',
	}

// Score breakdown
const streakMultiplier = computed(() => {
	if (!results.value || results.value.baseScore === 0) return 1
	return results.value.streakBonus > 0
		? Number((results.value.finalScore / results.value.baseScore).toFixed(2))
		: 1
})

const hasStreakBonus = computed(() => results.value && results.value.streakBonus > 0)

// ===========================
// Methods
// ===========================

const handleGoHome = () => {
	router.push({ name: 'home' })
}

const handleRetryWithCoins = async () => {
	try {
		await retryChallenge('coins')
	} catch (error: unknown) {
		console.error('Failed to retry with coins:', error)
		// TODO: Show error toast
	}
}

const handleRetryWithAd = async () => {
	try {
		await retryChallenge('ad')
	} catch (error: unknown) {
		console.error('Failed to retry with ad:', error)
		// TODO: Show error toast
	}
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	console.log('[DailyChallengeResults] onMounted', {
		isCompleted: isCompleted.value,
		hasResults: !!results.value,
	})

	// Only initialize if we don't have results yet
	// (e.g., user refreshed the page)
	if (!results.value) {
		console.log('[DailyChallengeResults] No results in state, calling initialize...')
		await initialize()
	}

	// Redirect if game is not completed
	if (!isCompleted.value || !results.value) {
		console.log('[DailyChallengeResults] Redirecting to home - missing results')
		router.push({ name: 'home' })
	}
})
</script>

<template>
	<div class="min-h-screen mx-auto max-w-[800px] pb-8">
		<!-- Loading State -->
		<div v-if="!results" class="flex flex-col items-center justify-center min-h-[50vh]">
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-(--ui-text-muted) mt-4">{{ t('daily.loadingResults') }}</p>
		</div>

		<!-- Results View -->
		<div v-else class="flex flex-col">
			<!-- Hero: Purple gradient card -->
			<div
				class="bg-gradient-to-b from-primary-600 to-primary-800 rounded-(--ui-radius) px-5 pt-8 pb-6 text-white text-center"
			>
				<!-- Performance emoji + label -->
				<div class="text-5xl mb-2">{{ performanceLevel.emoji }}</div>
				<h2 class="text-xl font-bold text-white">{{ performanceLevel.label }}</h2>
				<p class="text-primary-200 text-xs mb-4">{{ t('daily.title') }}</p>

				<!-- Score -->
				<div class="text-5xl font-black leading-none mb-0.5 text-white">
					{{ game?.finalScore || 0 }}
				</div>
				<div class="text-primary-200 text-[10px] uppercase tracking-widest mb-2">
					{{ t('daily.points') }}
				</div>

				<!-- Score breakdown -->
				<div v-if="results && results.baseScore > 0" class="text-xs text-primary-200">
					<span>{{ t('daily.baseScore', { score: results.baseScore }) }}</span>
					<span v-if="hasStreakBonus" class="text-yellow-300 font-semibold ml-1">
						{{
							t('daily.streakBonus', {
								bonus: results.streakBonus,
								multiplier: streakMultiplier,
							})
						}}
					</span>
				</div>

				<!-- Stats row -->
				<div class="flex justify-center gap-6 mt-4 pt-4 border-t border-white/20">
					<div class="text-center">
						<div class="text-lg font-bold text-white">
							{{ results.correctAnswers }}/{{ results.totalQuestions }}
						</div>
						<div class="text-[10px] text-primary-200">{{ t('daily.correct') }}</div>
					</div>
					<div class="text-center">
						<div class="text-lg font-bold text-white">{{ scorePercentage }}%</div>
						<div class="text-[10px] text-primary-200">{{ t('quiz.accuracy') }}</div>
					</div>
					<div v-if="results.rank" class="text-center">
						<div class="text-lg font-bold text-white">#{{ results.rank }}</div>
						<div class="text-[10px] text-primary-200">{{ t('daily.yourRank') }}</div>
					</div>
				</div>
			</div>

			<!-- Content area -->
			<div class="flex flex-col gap-4 pt-4">
				<!-- New streak record alert -->
				<div
					v-if="hasNewStreakRecord"
					class="flex items-center gap-3 px-4 py-3 rounded-(--ui-radius) bg-yellow-500/10 border border-yellow-500/30"
				>
					<span class="text-xl">🔥</span>
					<div>
						<div class="text-sm font-semibold text-(--ui-text-highlighted)">
							{{ t('daily.newStreakRecord') }}
						</div>
						<div class="text-xs text-(--ui-text-muted)">
							{{ t('daily.streakRecordDesc', { days: streak!.currentStreak }) }}
							{{ streaks.getStreakEmoji.value }}
						</div>
					</div>
				</div>

				<!-- Chest Reward -->
				<div
					v-if="chestReward"
					class="bg-(--ui-bg-elevated) rounded-(--ui-radius) border border-(--ui-border) p-4"
				>
					<div class="flex items-center gap-3 mb-4">
						<span class="text-4xl">{{ chestEmoji }}</span>
						<div>
							<h3 class="font-bold text-(--ui-text-highlighted)">{{ chestLabel }}</h3>
							<p class="text-xs text-(--ui-text-muted)">
								{{ t('daily.yourRewards') }}
							</p>
						</div>
					</div>

					<!-- Reward row -->
					<div class="flex gap-3">
						<div
							class="flex-1 flex items-center gap-2 px-3 py-2.5 bg-(--ui-bg) rounded-lg border border-(--ui-border)"
						>
							<span class="text-lg">🪙</span>
							<div>
								<div class="text-xl font-bold text-(--ui-text-highlighted)">
									{{ chestReward.coins }}
								</div>
								<div class="text-xs text-(--ui-text-muted)">
									{{ t('daily.coins') }}
								</div>
							</div>
						</div>
						<div
							class="flex-1 flex items-center gap-2 px-3 py-2.5 bg-(--ui-bg) rounded-lg border border-(--ui-border)"
						>
							<span class="text-lg">🎟️</span>
							<div>
								<div class="text-xl font-bold text-(--ui-text-highlighted)">
									{{ chestReward.pvpTickets }}
								</div>
								<div class="text-xs text-(--ui-text-muted)">
									{{ t('daily.pvpTickets') }}
								</div>
							</div>
						</div>
					</div>

					<!-- Streak multiplier -->
					<div
						v-if="hasStreakBonus"
						class="flex items-center gap-2 mt-3 px-3 py-2 bg-yellow-500/10 rounded-lg text-xs"
					>
						<UIcon name="i-heroicons-fire" class="w-4 h-4 text-yellow-500" />
						<span class="text-yellow-600 dark:text-yellow-400">
							{{ t('daily.streakApplied', { multiplier: streakMultiplier }) }}
						</span>
					</div>

					<!-- Marathon bonuses -->
					<div
						v-if="chestReward.marathonBonuses && chestReward.marathonBonuses.length > 0"
						class="mt-4 pt-4 border-t border-(--ui-border)"
					>
						<div class="flex items-center gap-2 mb-3">
							<UIcon name="i-heroicons-bolt" class="w-4 h-4 text-primary" />
							<h4 class="text-sm font-semibold text-(--ui-text)">
								{{ t('daily.marathonBonuses') }}
							</h4>
						</div>
						<div class="flex flex-col gap-2">
							<div
								v-for="bonus in chestReward.marathonBonuses"
								:key="bonus"
								class="flex items-center gap-3 p-3 bg-(--ui-bg) rounded-lg border border-(--ui-border)"
							>
								<UIcon
									:name="getBonusInfo(bonus).icon"
									:class="getBonusInfo(bonus).color"
									class="w-5 h-5 shrink-0"
								/>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-semibold text-(--ui-text-highlighted)">
										{{ getBonusInfo(bonus).label }}
									</div>
									<div class="text-xs text-(--ui-text-muted)">
										{{ getBonusInfo(bonus).description }}
									</div>
								</div>
								<UBadge color="primary" variant="soft" size="xs">+1</UBadge>
							</div>
						</div>
						<p class="text-xs text-(--ui-text-dimmed) mt-2">
							{{ t('daily.bonusUseDesc') }}
						</p>
					</div>
				</div>

				<!-- Leaderboard -->
				<div
					class="bg-(--ui-bg-elevated) rounded-(--ui-radius) border border-(--ui-border) p-4"
				>
					<DailyChallengeLeaderboard
						:leaderboard="results.leaderboard"
						:current-player-id="playerId"
						:max-entries="10"
					/>
				</div>

				<!-- Action buttons -->
				<div class="flex flex-col gap-3 pt-2">
					<!-- Retry section -->
					<div
						class="bg-(--ui-bg-elevated) rounded-(--ui-radius) border border-(--ui-border) p-4"
					>
						<div class="flex items-center gap-2 mb-1">
							<UIcon name="i-heroicons-arrow-path" class="w-4 h-4 text-primary" />
							<h3 class="text-sm font-semibold text-(--ui-text-highlighted)">
								{{ t('daily.tryAgain') }}
							</h3>
						</div>
						<p class="text-xs text-(--ui-text-muted) mb-4">
							{{ t('daily.tryAgainDesc') }}
						</p>
						<div class="flex gap-3">
							<button
								class="flex-1 flex items-center justify-center gap-2 py-3 rounded-lg bg-yellow-500/10 border border-yellow-500/30 text-yellow-600 dark:text-yellow-400 text-sm font-semibold transition-colors hover:bg-yellow-500/20 disabled:opacity-50"
								:disabled="isLoading"
								@click="handleRetryWithCoins"
							>
								<UIcon name="i-heroicons-currency-dollar" class="w-4 h-4" />
								{{ t('daily.retryCoins') }}
							</button>
							<button
								class="flex-1 flex items-center justify-center gap-2 py-3 rounded-lg bg-blue-500/10 border border-blue-500/30 text-blue-600 dark:text-blue-400 text-sm font-semibold transition-colors hover:bg-blue-500/20 disabled:opacity-50"
								:disabled="isLoading"
								@click="handleRetryWithAd"
							>
								<UIcon name="i-heroicons-play" class="w-4 h-4" />
								{{ t('daily.watchAd') }}
							</button>
						</div>
					</div>

					<!-- Home link -->
					<button
						class="flex items-center justify-center gap-2 py-3 text-sm text-(--ui-text-muted) hover:text-(--ui-text) transition-colors"
						@click="handleGoHome"
					>
						<UIcon name="i-heroicons-home" class="w-4 h-4" />
						{{ t('daily.backToHome') }}
					</button>
				</div>

				<!-- Next challenge info -->
				<div class="flex items-center justify-center gap-2 py-3 text-center">
					<UIcon
						name="i-heroicons-calendar-days"
						class="w-4 h-4 text-(--ui-text-dimmed)"
					/>
					<p class="text-xs text-(--ui-text-muted)">
						{{ t('daily.nextChallenge') }}
						<span class="font-semibold">{{ timeToExpireFormatted }}</span>
					</p>
				</div>
			</div>
		</div>
	</div>
</template>
