<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useGetLeaderboard } from '@/api/generated/hooks/leaderboardController/useGetLeaderboard'
import { useAuth } from '@/composables/useAuth'
import LeaderboardList from '@/components/Leaderboard/LeaderboardList.vue'
import LeaderboardPodium from '@/components/Leaderboard/LeaderboardPodium.vue'
import LeaderboardEntryRow from '@/components/Leaderboard/LeaderboardEntryRow.vue'

const { t } = useI18n()
const { currentUser } = useAuth()

const { data, isLoading, isError, error, refetch } = useGetLeaderboard({ limit: 50 })

const entries = computed(() => data.value?.data ?? [])

const podiumEntries = computed(() =>
	entries.value.slice(0, 3).map((e) => ({
		rank: e.rank,
		username: e.username,
		value: e.totalScore,
		label: t('leaderboard.pts'),
	})),
)

const listEntries = computed(() => entries.value.slice(3))

const isEmpty = computed(() => !isLoading.value && !isError.value && entries.value.length === 0)
const errorMessage = computed(() => (error.value as { message?: string })?.message)
</script>

<template>
	<LeaderboardList
		:is-loading="isLoading"
		:is-error="isError"
		:error-message="errorMessage"
		:is-empty="isEmpty"
		:empty-title="t('leaderboard.empty')"
		:empty-description="t('leaderboard.global.emptyDesc')"
		@retry="refetch"
		@refresh="refetch"
	>
		<template #podium>
			<LeaderboardPodium :entries="podiumEntries" />
		</template>

		<template #entries>
			<LeaderboardEntryRow
				v-for="entry in listEntries"
				:key="entry.userId"
				:rank="entry.rank"
				:username="entry.username"
				:is-current-user="entry.userId === currentUser?.id"
			>
				<template #stats>
					<p class="text-sm font-bold text-primary">{{ entry.totalScore }}</p>
					<p class="text-xs text-(--ui-text-dimmed)">
						{{ t('leaderboard.global.quizzes', { count: entry.quizzesCompleted }) }}
					</p>
				</template>
			</LeaderboardEntryRow>
		</template>
	</LeaderboardList>
</template>
