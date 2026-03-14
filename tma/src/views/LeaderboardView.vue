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


// Check if entry is current user
const isCurrentUser = (userId: string) => {
	return currentUser.value?.id === userId
}
</script>

<template>
	<div class="mx-auto max-w-[800px] pb-8">
		<!-- Header -->
		<div class="text-center mb-6">
			<div class="text-4xl mb-2">🏆</div>
			<h1 class="text-2xl font-bold text-(--ui-text-highlighted)">
				{{ t('leaderboard.title') }}
			</h1>
			<p class="text-sm text-(--ui-text-muted) mt-1">{{ t('leaderboard.subtitle') }}</p>
		</div>

		<!-- Loading -->
		<div v-if="isLoading" class="flex flex-col items-center justify-center py-12">
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-(--ui-text-muted) mt-4">{{ t('leaderboard.loading') }}</p>
		</div>

		<!-- Error -->
		<div v-else-if="isError" class="text-center py-12">
			<UAlert
				color="red"
				variant="soft"
				:title="t('leaderboard.loadFailed')"
				:description="error?.error?.message || t('quiz.tryAgain2')"
			/>
			<UButton
				color="red"
				class="mt-3"
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
		<div
			v-else-if="entries.length === 0"
			class="flex flex-col items-center justify-center py-16 rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border)"
		>
			<div class="text-5xl mb-3">🏆</div>
			<h2 class="text-xl font-bold text-(--ui-text-highlighted) mb-1">
				{{ t('leaderboard.empty') }}
			</h2>
			<p class="text-sm text-(--ui-text-muted)">{{ t('leaderboard.emptyDesc') }}</p>
		</div>

		<!-- Leaderboard content -->
		<div v-else>
			<!-- Current user rank badge (if not in top 10) -->
			<div
				v-if="currentUserRank && currentUserRank > 10"
				class="flex items-center justify-between px-4 py-3 mb-3 rounded-(--ui-radius) bg-primary-500/10 border border-primary-500/20"
			>
				<p class="text-sm text-(--ui-text-muted)">{{ t('leaderboard.yourRank') }}</p>
				<p class="text-xl font-bold text-primary">#{{ currentUserRank }}</p>
			</div>

			<!-- Entries list -->
			<div class="rounded-(--ui-radius) border border-(--ui-border) overflow-hidden">
				<div
					v-for="entry in entries"
					:key="entry.userId"
					class="flex items-center gap-3 px-4 py-3 border-b border-(--ui-border-muted) last:border-b-0 transition-colors"
					:class="
						isCurrentUser(entry.userId) ? 'bg-primary-500/10' : 'bg-(--ui-bg-elevated)'
					"
				>
					<!-- Rank -->
					<span v-if="entry.rank <= 3" class="text-xl w-7 text-center shrink-0">{{
						getMedalEmoji(entry.rank)
					}}</span>
					<span
						v-else
						class="text-sm font-semibold text-(--ui-text-dimmed) w-7 text-center shrink-0"
						>{{ entry.rank }}</span
					>

					<!-- Avatar initial -->
					<div
						class="size-10 rounded-full flex items-center justify-center text-sm font-bold shrink-0"
						:class="
							isCurrentUser(entry.userId)
								? 'bg-primary-500 text-white'
								: 'bg-(--ui-bg-accented) text-(--ui-text-highlighted)'
						"
					>
						{{ (entry.username || '?')[0]?.toUpperCase() }}
					</div>

					<!-- Name -->
					<div class="flex-1 min-w-0">
						<span
							class="font-medium truncate block"
							:class="
								isCurrentUser(entry.userId)
									? 'text-(--ui-text-highlighted)'
									: 'text-(--ui-text)'
							"
						>
							{{ entry.username || t('leaderboard.anonymous') }}
						</span>
						<span v-if="isCurrentUser(entry.userId)" class="text-xs text-primary"
							>You</span
						>
					</div>

					<!-- Score -->
					<span
						class="font-bold tabular-nums shrink-0"
						:class="
							entry.rank <= 3
								? 'text-primary text-lg'
								: 'text-(--ui-text-highlighted)'
						"
					>
						{{ entry.totalScore }}
					</span>
				</div>
			</div>

			<!-- Footer -->
			<p class="text-center text-xs text-(--ui-text-dimmed) mt-3">
				{{ t('leaderboard.showingTop', { count: entries.length }) }}
			</p>
		</div>
	</div>
</template>
