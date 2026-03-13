<script setup lang="ts">
import { computed, onMounted, ref, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useStreaks } from '@/composables/useStreaks'
import { useI18n } from 'vue-i18n'

interface Props {
	playerId: string
}

const props = defineProps<Props>()
const router = useRouter()
const { t } = useI18n()

// ===========================
// Composables
// ===========================

const {
	game,
	streak,
	totalPlayers,
	questionIndex,
	totalQuestions,
	isPlaying,
	isLoading,
	hasPlayed,
	canPlay,
	timeToExpireFormatted,
	startGame,
	checkStatus,
	initialize,
} = useDailyChallenge(props.playerId)

const streaks = useStreaks(streak)

// ===========================
// Countdown Timer
// ===========================

const countdownInterval = ref<number | null>(null)

const startCountdown = () => {
	if (countdownInterval.value) return

	countdownInterval.value = window.setInterval(() => {
		checkStatus()
	}, 60000) // Refresh every minute
}

const stopCountdown = () => {
	if (countdownInterval.value) {
		clearInterval(countdownInterval.value)
		countdownInterval.value = null
	}
}

// ===========================
// Computed
// ===========================

const buttonText = computed(() => {
	if (isPlaying.value) return t('daily.continue')
	if (hasPlayed.value) return t('daily.viewResults')
	return t('daily.startChallenge')
})

const buttonIcon = computed(() => {
	if (isPlaying.value) return 'i-heroicons-play'
	if (hasPlayed.value) return 'i-heroicons-chart-bar'
	return 'i-heroicons-play'
})

const buttonColor = computed(() => {
	if (hasPlayed.value) return 'gray'
	return 'primary'
})

const questionProgress = computed(() =>
	Math.round((questionIndex.value / totalQuestions.value) * 100),
)

// ===========================
// Actions
// ===========================

const handleClick = async () => {
	if (isLoading.value) return

	if (hasPlayed.value) {
		router.push({ name: 'daily-challenge-results' })
	} else if (isPlaying.value) {
		router.push({ name: 'daily-challenge-play' })
	} else if (canPlay.value) {
		try {
			await startGame()
			router.push({ name: 'daily-challenge-play' })
		} catch (error) {
			console.error('Failed to start game:', error)
		}
	}
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	try {
		await initialize()
		startCountdown()
	} catch (error) {
		console.error('Failed to initialize Daily Challenge:', error)
	}
})

onBeforeUnmount(() => {
	stopCountdown()
})
</script>

<template>
	<UCard class="overflow-hidden">
		<!-- Gradient accent bar -->
		<template #header>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2.5">
					<div
						class="flex items-center justify-center size-9 rounded-lg bg-indigo-100 dark:bg-indigo-900/30"
					>
						<UIcon
							name="i-heroicons-calendar-days"
							class="size-5 text-indigo-600 dark:text-indigo-400"
						/>
					</div>
					<div>
						<h3 class="text-base font-bold text-(--ui-text-highlighted)">
							{{ t('daily.title') }}
						</h3>
						<p class="text-xs text-(--ui-text-dimmed)">
							{{ t('daily.questionsInfo') }}
						</p>
					</div>
				</div>
				<UBadge v-if="hasPlayed" color="green" variant="subtle" size="sm">
					<UIcon name="i-heroicons-check-circle" class="size-3.5 mr-0.5" />
					{{ t('daily.completedBadge') }}
				</UBadge>
				<UBadge v-else-if="isPlaying" color="blue" variant="subtle" size="sm">
					<UIcon name="i-heroicons-play-circle" class="size-3.5 mr-0.5" />
					{{ t('daily.inProgressBadge') }}
				</UBadge>
			</div>
		</template>

		<div class="space-y-4">
			<!-- Completed State -->
			<div v-if="hasPlayed && game" class="flex items-center justify-between">
				<div>
					<p class="text-2xl font-black text-emerald-600 dark:text-emerald-400">
						{{ game.finalScore || 0 }}
						<span class="text-sm font-medium text-(--ui-text-dimmed)">{{
							t('daily.points')
						}}</span>
					</p>
					<p v-if="streak" class="text-sm text-(--ui-text-muted) mt-0.5">
						{{ streaks.formattedStreak.value }}
					</p>
				</div>
				<div class="text-right">
					<p class="text-xs text-(--ui-text-dimmed)">{{ t('daily.resetsIn') }}</p>
					<p class="text-sm font-mono font-semibold tabular-nums">
						{{ timeToExpireFormatted }}
					</p>
				</div>
			</div>

			<!-- In Progress State -->
			<div v-else-if="isPlaying" class="space-y-3">
				<div class="flex justify-between text-sm">
					<span class="font-medium text-(--ui-text-highlighted)">
						{{
							t('shared.questionOf', {
								current: questionIndex + 1,
								total: totalQuestions,
							})
						}}
					</span>
					<span class="text-(--ui-text-dimmed)">{{ questionProgress }}%</span>
				</div>
				<UProgress v-model="questionProgress" color="primary" size="sm" />
			</div>

			<!-- Not Played State -->
			<div v-else>
				<div class="flex items-center justify-between">
					<div>
						<p v-if="streak" class="text-sm font-medium text-(--ui-text)">
							{{ streaks.formattedStreak.value }}
						</p>
						<p v-else class="text-sm text-(--ui-text-muted)">
							{{ t('daily.startStreakHint') }}
						</p>
					</div>
					<div class="text-right">
						<p class="text-xs text-(--ui-text-dimmed)">{{ t('daily.resetsIn') }}</p>
						<p class="text-sm font-mono font-semibold tabular-nums">
							{{ timeToExpireFormatted }}
						</p>
					</div>
				</div>
				<div v-if="totalPlayers > 0" class="mt-3 pt-3 border-t border-(--ui-border-muted)">
					<p class="text-xs text-(--ui-text-dimmed)">
						<UIcon name="i-heroicons-user-group" class="inline size-3.5" />
						{{ t('daily.playersCount', { count: totalPlayers }) }}
					</p>
				</div>
			</div>
		</div>

		<template #footer>
			<UButton
				:icon="buttonIcon"
				:color="buttonColor"
				:loading="isLoading"
				:disabled="!canPlay && !hasPlayed && !isPlaying"
				block
				size="lg"
				@click="handleClick"
			>
				{{ buttonText }}
			</UButton>
		</template>
	</UCard>
</template>
