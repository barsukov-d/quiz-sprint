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

const chestColor = computed(() => {
	if (!chestReward.value) return 'gray'
	switch (chestReward.value.chestType) {
		case 'golden':
			return 'yellow'
		case 'silver':
			return 'gray'
		case 'wooden':
			return 'amber'
		default:
			return 'gray'
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
	<div class="min-h-screen mx-auto max-w-[800px] p-4 pt-24 pb-8 sm:p-3 sm:pt-20">
		<!-- Loading State -->
		<div v-if="!results" class="flex flex-col items-center justify-center min-h-[50vh]">
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-(--ui-text-muted) mt-4">{{ t('daily.loadingResults') }}</p>
		</div>

		<!-- Results View -->
		<div v-else class="flex flex-col gap-6">
			<!-- Header: Score Card -->
			<UCard
				class="bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/30 dark:to-primary-800/30"
			>
				<div class="flex flex-col items-center gap-6 p-4">
					<!-- Performance Level -->
					<div class="text-center">
						<span class="block text-5xl mb-2">{{ performanceLevel.emoji }}</span>
						<h2 class="text-2xl font-bold text-(--ui-text-highlighted)">
							{{ performanceLevel.label }}
						</h2>
					</div>

					<!-- Score -->
					<div class="text-center">
						<div
							class="text-6xl font-black text-primary-600 dark:text-primary-400 leading-none"
						>
							{{ game?.finalScore || 0 }}
						</div>
						<div class="text-sm text-(--ui-text-muted) uppercase tracking-wider mt-1">
							{{ t('daily.points') }}
						</div>
						<!-- Score breakdown -->
						<div
							v-if="results && results.baseScore > 0"
							class="flex items-center justify-center gap-2 mt-2 text-xs text-(--ui-text-muted)"
						>
							<span>{{ t('daily.baseScore', { score: results.baseScore }) }}</span>
							<span v-if="hasStreakBonus" class="text-yellow-500 font-semibold">
								{{
									t('daily.streakBonus', {
										bonus: results.streakBonus,
										multiplier: streakMultiplier,
									})
								}}
							</span>
						</div>
					</div>

					<!-- Accuracy -->
					<div class="w-full flex flex-col gap-2">
						<UProgress
							v-model="scorePercentage"
							:color="performanceLevel.color"
							size="lg"
						/>
						<p class="text-center text-sm text-(--ui-text) font-semibold">
							{{
								t('daily.correctOf', {
									correct: results.correctAnswers,
									total: results.totalQuestions,
								})
							}}
							<span class="text-(--ui-text-muted)">{{
								t('daily.scorePercent', { percent: scorePercentage })
							}}</span>
						</p>
					</div>
				</div>
			</UCard>

			<!-- Streak Info (if new record) -->
			<UAlert
				v-if="hasNewStreakRecord"
				color="yellow"
				variant="soft"
				:title="t('daily.newStreakRecord')"
				icon="i-heroicons-fire"
			>
				<template #description>
					<p>
						{{ t('daily.streakRecordDesc', { days: streak!.currentStreak }) }}
						{{ streaks.getStreakEmoji.value }}
					</p>
				</template>
			</UAlert>

			<!-- Chest Reward -->
			<UCard
				v-if="chestReward"
				:class="chestColor === 'yellow' ? 'bg-yellow-50 dark:bg-yellow-950' : ''"
			>
				<div class="flex flex-col gap-6">
					<div class="flex items-center gap-4 pb-4 border-b border-(--ui-border)">
						<span class="text-5xl">{{ chestEmoji }}</span>
						<div>
							<h3 class="text-xl font-bold">{{ chestLabel }}</h3>
							<p class="text-sm text-(--ui-text-muted)">
								{{ t('daily.yourRewards') }}
							</p>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div class="flex items-center gap-3 p-4 bg-(--ui-bg-muted) rounded-xl">
							<UIcon
								name="i-heroicons-currency-dollar"
								class="w-6 h-6 text-yellow-500"
							/>
							<div>
								<div class="text-2xl font-bold text-(--ui-text-highlighted)">
									{{ chestReward.coins }}
								</div>
								<div
									class="text-xs text-(--ui-text-muted) uppercase tracking-wider"
								>
									{{ t('daily.coins') }}
								</div>
							</div>
						</div>

						<div class="flex items-center gap-3 p-4 bg-(--ui-bg-muted) rounded-xl">
							<UIcon name="i-heroicons-ticket" class="w-6 h-6 text-blue-500" />
							<div>
								<div class="text-2xl font-bold text-(--ui-text-highlighted)">
									{{ chestReward.pvpTickets }}
								</div>
								<div
									class="text-xs text-(--ui-text-muted) uppercase tracking-wider"
								>
									{{ t('daily.pvpTickets') }}
								</div>
							</div>
						</div>
					</div>

					<!-- Streak multiplier applied -->
					<div
						v-if="hasStreakBonus"
						class="flex items-center gap-2 px-3 py-2 bg-yellow-50 dark:bg-yellow-950/30 rounded-lg text-xs"
					>
						<UIcon name="i-heroicons-fire" class="w-4 h-4 text-yellow-500" />
						<span class="text-yellow-700 dark:text-yellow-400">
							{{ t('daily.streakApplied', { multiplier: streakMultiplier }) }}
						</span>
					</div>

					<div
						v-if="chestReward.marathonBonuses && chestReward.marathonBonuses.length > 0"
						class="pt-4 border-t border-(--ui-border)"
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
								class="flex items-center gap-3 p-3 bg-(--ui-bg-muted) rounded-xl"
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
			</UCard>

			<!-- Rank Card -->
			<UCard>
				<div class="flex flex-col gap-4">
					<div class="flex items-center gap-3 pb-4 border-b border-(--ui-border)">
						<UIcon name="i-heroicons-chart-bar" class="w-6 h-6 text-primary" />
						<h3 class="text-lg font-semibold">{{ t('daily.yourRanking') }}</h3>
					</div>
					<div class="grid grid-cols-2 gap-8 py-4">
						<div class="text-center">
							<div class="mb-2">
								<UBadge color="primary" size="xl" variant="soft">
									#{{ results.rank }}
								</UBadge>
							</div>
							<div class="text-sm text-(--ui-text-muted)">
								{{ t('daily.yourRank') }}
							</div>
						</div>
						<div class="text-center">
							<div class="text-2xl font-bold text-(--ui-text-highlighted) mb-2">
								{{ results.totalPlayers }}
							</div>
							<div class="text-sm text-(--ui-text-muted)">
								{{ t('daily.totalPlayers') }}
							</div>
						</div>
					</div>
				</div>
			</UCard>

			<!-- Leaderboard -->
			<UCard>
				<DailyChallengeLeaderboard
					:leaderboard="results.leaderboard"
					:current-player-id="playerId"
					:max-entries="10"
				/>
			</UCard>

			<!-- Action Buttons -->
			<div class="flex flex-col gap-3">
				<!-- Retry Section -->
				<UCard class="bg-(--ui-bg-muted)">
					<div class="flex flex-col gap-3">
						<div class="flex items-center gap-2 pb-3 border-b border-(--ui-border)">
							<UIcon name="i-heroicons-arrow-path" class="w-5 h-5 text-primary" />
							<h3 class="text-lg font-semibold">{{ t('daily.tryAgain') }}</h3>
						</div>

						<p class="text-sm text-(--ui-text-muted)">
							{{ t('daily.tryAgainDesc') }}
						</p>

						<div class="grid grid-cols-2 gap-3">
							<UButton
								color="yellow"
								size="lg"
								icon="i-heroicons-currency-dollar"
								block
								:loading="isLoading"
								@click="handleRetryWithCoins"
							>
								<div class="flex flex-col items-center gap-1">
									<span class="font-bold">{{ t('daily.retryCoins') }}</span>
									<span class="text-xs opacity-80">{{ t('daily.retry') }}</span>
								</div>
							</UButton>

							<UButton
								color="blue"
								size="lg"
								icon="i-heroicons-play"
								block
								:loading="isLoading"
								@click="handleRetryWithAd"
							>
								<div class="flex flex-col items-center gap-1">
									<span class="font-bold">{{ t('daily.watchAd') }}</span>
									<span class="text-xs opacity-80">{{
										t('daily.freeRetry')
									}}</span>
								</div>
							</UButton>
						</div>
					</div>
				</UCard>

				<UButton
					color="neutral"
					size="xl"
					icon="i-heroicons-home"
					variant="outline"
					block
					@click="handleGoHome"
				>
					{{ t('daily.backToHome') }}
				</UButton>
			</div>

			<!-- Next Challenge Info -->
			<div class="flex items-center justify-center gap-2 p-4 text-center">
				<UIcon name="i-heroicons-calendar-days" class="w-5 h-5 text-(--ui-text-dimmed)" />
				<p class="text-sm text-(--ui-text-muted)">
					{{ t('daily.nextChallenge') }}
					<span class="font-semibold">{{ timeToExpireFormatted }}</span>
				</p>
			</div>
		</div>
	</div>
</template>
