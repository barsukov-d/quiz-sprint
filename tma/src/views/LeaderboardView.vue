<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useGetQuiz, useGetLeaderboard } from '@/api'
import { useAuth } from '@/composables/useAuth'
import { useLastQuiz } from '@/composables/useLastQuiz'

const route = useRoute()
const { currentUser } = useAuth()
const { lastQuizId, saveLastQuizId } = useLastQuiz()

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
				<h1 class="text-3xl font-bold mb-2">Leaderboard</h1>
				<p class="text-gray-600">Top players by score</p>
			</div>

			<!-- Loading -->
			<div v-if="isLoading" class="flex justify-center items-center py-12">
				<UProgress animation="carousel" />
				<span class="ml-4">Loading leaderboard...</span>
			</div>

			<!-- Error -->
			<div v-else-if="isError" class="mb-4">
				<UAlert
					color="red"
					variant="soft"
					title="Failed to load leaderboard"
					:description="error?.error?.message || 'Please try again'"
				/>
				<UButton color="red" class="mt-2" @click="refetch()"> Retry </UButton>
			</div>

			<!-- Empty state -->
			<div v-else-if="entries.length === 0" class="text-center py-12">
				<div class="text-6xl mb-4">🏆</div>
				<h2 class="text-2xl font-bold mb-2">No scores yet!</h2>
				<p class="text-gray-600">Be the first to complete this quiz.</p>
			</div>

			<!-- Leaderboard Table -->
			<div v-else>
				<!-- Current user rank badge (if not in top 10) -->
				<UCard v-if="currentUserRank && currentUserRank > 10" class="mb-4 bg-blue-50">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-gray-600">Your Rank</p>
							<p class="text-2xl font-bold text-blue-600">#{{ currentUserRank }}</p>
						</div>
						<UIcon name="i-heroicons-star" class="text-4xl text-blue-600" />
					</div>
				</UCard>

				<!-- Leaderboard entries -->
				<UCard>
					<div class="overflow-x-auto">
						<table class="w-full">
							<thead>
								<tr class="border-b border-gray-200">
									<th class="text-left py-3 px-4 font-semibold text-gray-700">
										Rank
									</th>
									<th class="text-left py-3 px-4 font-semibold text-gray-700">
										Player
									</th>
									<th class="text-right py-3 px-4 font-semibold text-gray-700">
										Score
									</th>
									<th class="text-right py-3 px-4 font-semibold text-gray-700">
										Date
									</th>
								</tr>
							</thead>
							<tbody>
								<tr
									v-for="entry in entries"
									:key="entry.userId"
									:class="[
										'border-b border-gray-100 transition-colors',
										isCurrentUser(entry.userId)
											? 'bg-blue-50 hover:bg-blue-100'
											: 'hover:bg-gray-50',
									]"
								>
									<!-- Rank -->
									<td class="py-3 px-4">
										<div class="flex items-center gap-2">
											<span v-if="entry.rank <= 3" class="text-2xl">
												{{ getMedalEmoji(entry.rank) }}
											</span>
											<span class="font-semibold text-gray-700">
												{{ entry.rank }}
											</span>
										</div>
									</td>

									<!-- Player -->
									<td class="py-3 px-4">
										<div class="flex items-center gap-2">
											<span class="font-medium">{{
												entry.username || 'Anonymous'
											}}</span>
											<UIcon
												v-if="isCurrentUser(entry.userId)"
												name="i-heroicons-star-solid"
												class="text-blue-600"
											/>
										</div>
									</td>

									<!-- Score -->
									<td class="py-3 px-4 text-right">
										<span class="font-bold text-gray-900">{{
											entry.totalScore
										}}</span>
									</td>

									Date
									<td class="py-3 px-4 text-right">
										<span class="text-sm text-gray-500">
											{{ formatDate(entry.rank) }}
										</span>
									</td>
								</tr>
							</tbody>
						</table>
					</div>
				</UCard>

				<!-- Stats footer -->
				<div class="mt-4 text-center text-sm text-gray-500">
					Showing top {{ entries.length }} players
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
		background-color: rgb(239 246 255);
	}
	50% {
		background-color: rgb(219 234 254);
	}
}

.bg-blue-50 {
	animation: highlight-pulse 2s ease-in-out infinite;
}
</style>
