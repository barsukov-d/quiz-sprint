<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useGetQuiz, useGetLeaderboard } from '@/api'
import { useAuth } from '@/composables/useAuth'
import { useLastQuiz } from '@/composables/useLastQuiz'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { currentUser } = useAuth()
const { lastQuizId, saveLastQuizId } = useLastQuiz()
const { t } = useI18n()

// Get quizId from route params or localStorage
const routeQuizId = computed(() => route.params.quizId as string | undefined)

// Determine which quiz ID to use
// Priority: route param > localStorage > fetch first available quiz
const activeQuizId = computed(() => {
	return routeQuizId.value || lastQuizId.value || null
})

// Fetch first available quiz if no ID is available
const { data: quizzesResponse, isLoading: isLoadingQuizzes } = useGetQuiz(undefined, {
	query: {
		enabled: computed(() => !activeQuizId.value),
	},
})

// Get first quiz ID from the list if needed
const firstQuizId = computed(() => {
	return quizzesResponse.value?.data?.[0]?.id
})

// Final quiz ID to use (route > localStorage > first available)
const quizId = computed(() => {
	return activeQuizId.value || firstQuizId.value || null
})

// Save quiz ID to localStorage when it changes
watch(
	quizId,
	(newId) => {
		if (newId) {
			saveLastQuizId(newId)
		}
	},
	{ immediate: true },
)

// Fetch leaderboard data
const {
	data: leaderboardResponse,
	isLoading: isLoadingLeaderboard,
	isError,
	error,
	refetch,
} = useGetLeaderboard({ limit: 50 })

// Combined loading state
const isLoading = computed(() => isLoadingQuizzes.value || isLoadingLeaderboard.value)

// Extract entries
const entries = computed(() => leaderboardResponse.value?.data || [])

// Find current user's rank
const currentUserRank = computed(() => {
	if (!currentUser.value) return null
	const entry = entries.value.find((e) => e.userId === currentUser.value?.id)
	return entry?.rank || null
})

// Medal emoji for top 3
const getMedalEmoji = (rank: number) => {
	switch (rank) {
		case 1:
			return '🥇'
		case 2:
			return '🥈'
		case 3:
			return '🥉'
		default:
			return ''
	}
}

// Format date
const formatDate = (timestamp: number) => {
	const now = Date.now()
	const diff = now - timestamp * 1000
	const days = Math.floor(diff / (1000 * 60 * 60 * 24))

	if (days === 0) return 'today'
	if (days === 1) return '1d'
	if (days < 30) return `${days}d`
	return new Date(timestamp * 1000).toLocaleDateString()
}

// Check if entry is current user
const isCurrentUser = (userId: string) => {
	return currentUser.value?.id === userId
}
</script>

<template>
	<div class="container mx-auto">
		<div class="max-w-4xl mx-auto">
			<!-- Header -->
			<div class="mb-6 mx-auto text-center">
				<h1 class="text-3xl font-bold mb-2">{{ t('leaderboard.title') }}</h1>
				<p class="text-(--ui-text-muted)">{{ t('leaderboard.subtitle') }}</p>
			</div>

			<!-- Loading -->
			<div v-if="isLoading" class="flex justify-center items-center py-12">
				<UProgress animation="carousel" />
				<span class="ml-4">{{ t('leaderboard.loading') }}</span>
			</div>

			<!-- Error -->
			<div v-else-if="isError" class="mb-4">
				<UAlert
					color="red"
					variant="soft"
					:title="t('leaderboard.loadFailed')"
					:description="error?.error?.message || t('quiz.tryAgain2')"
				/>
				<UButton
					color="red"
					class="mt-2"
					@click="
						() => {
							refetch()
						}
					"
				>
					{{ t('leaderboard.retry') }}
				</UButton>
			</div>

			<!-- Empty state -->
			<div v-else-if="entries.length === 0" class="text-center py-12">
				<div class="text-6xl mb-4">🏆</div>
				<h2 class="text-2xl font-bold mb-2">{{ t('leaderboard.empty') }}</h2>
				<p class="text-(--ui-text-muted)">{{ t('leaderboard.emptyDesc') }}</p>
			</div>

			<!-- Leaderboard Table -->
			<div v-else>
				<!-- Current user rank badge (if not in top 10) -->
				<UCard
					v-if="currentUserRank && currentUserRank > 10"
					class="mb-4 bg-(--ui-bg-muted)"
				>
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-(--ui-text-muted)">
								{{ t('leaderboard.yourRank') }}
							</p>
							<p class="text-2xl font-bold text-primary">
								#{{ currentUserRank }}
							</p>
						</div>
						<UIcon
							name="i-heroicons-star"
							class="text-4xl text-primary"
						/>
					</div>
				</UCard>

				<!-- Leaderboard entries -->
				<UCard>
					<div class="overflow-x-auto">
						<table class="w-full">
							<thead>
								<tr class="border-b border-(--ui-border)">
									<th
										class="text-left py-3 px-4 font-semibold text-(--ui-text-toned)"
									>
										{{ t('leaderboard.colRank') }}
									</th>
									<th
										class="text-left py-3 px-4 font-semibold text-(--ui-text-toned)"
									>
										{{ t('leaderboard.colPlayer') }}
									</th>
									<th
										class="text-right py-3 px-4 font-semibold text-(--ui-text-toned)"
									>
										{{ t('leaderboard.colScore') }}
									</th>
									<th
										class="text-right py-3 px-4 font-semibold text-(--ui-text-toned)"
									>
										{{ t('leaderboard.colDate') }}
									</th>
								</tr>
							</thead>
							<tbody>
								<tr
									v-for="entry in entries"
									:key="entry.userId"
									:class="[
										'border-b border-(--ui-border-muted) transition-colors',
										isCurrentUser(entry.userId)
											? 'current-user-row bg-(--ui-bg-muted) hover:bg-(--ui-bg-elevated)'
											: 'hover:bg-(--ui-bg-muted)',
									]"
								>
									<!-- Rank -->
									<td class="py-3 px-4">
										<div class="flex items-center gap-2">
											<span v-if="entry.rank <= 3" class="text-2xl">
												{{ getMedalEmoji(entry.rank) }}
											</span>
											<span class="font-semibold text-(--ui-text-toned)">
												{{ entry.rank }}
											</span>
										</div>
									</td>

									<!-- Player -->
									<td class="py-3 px-4">
										<div class="flex items-center gap-2">
											<span class="font-medium">{{
												entry.username || t('leaderboard.anonymous')
											}}</span>
											<UIcon
												v-if="isCurrentUser(entry.userId)"
												name="i-heroicons-star-solid"
												class="text-primary"
											/>
										</div>
									</td>

									<!-- Score -->
									<td class="py-3 px-4 text-right">
										<span class="font-bold text-(--ui-text-highlighted)">{{
											entry.totalScore
										}}</span>
									</td>

									<!-- Date -->
									<td class="py-3 px-4 text-right">
										<span class="text-sm text-(--ui-text-dimmed)">
											{{ formatDate(entry.rank) }}
										</span>
									</td>
								</tr>
							</tbody>
						</table>
					</div>
				</UCard>

				<!-- Stats footer -->
				<div class="mt-4 text-center text-sm text-(--ui-text-dimmed)">
					{{ t('leaderboard.showingTop', { count: entries.length }) }}
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}

/* Highlight animation for current user */
@keyframes highlight-pulse {
	0%,
	100% {
		background-color: var(--ui-bg-muted);
	}
	50% {
		background-color: var(--ui-bg-elevated);
	}
}

.current-user-row {
	animation: highlight-pulse 2s ease-in-out infinite;
}
</style>
