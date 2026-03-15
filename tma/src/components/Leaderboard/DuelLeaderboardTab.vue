<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useGetDuelLeaderboard } from '@/api/generated/hooks/duelController/useGetDuelLeaderboard'
import { useAuth } from '@/composables/useAuth'
import LeaderboardList from '@/components/Leaderboard/LeaderboardList.vue'
import LeaderboardPodium from '@/components/Leaderboard/LeaderboardPodium.vue'
import LeaderboardEntryRow from '@/components/Leaderboard/LeaderboardEntryRow.vue'

const { t } = useI18n()
const { currentUser } = useAuth()

// ===========================
// Leaderboard query
// ===========================

const queryParams = computed(() => ({
	playerId: currentUser.value?.id ?? '',
	type: 'seasonal',
	limit: 50,
}))

const {
	data: leaderboardData,
	isLoading,
	isError,
	error,
	refetch,
} = useGetDuelLeaderboard(queryParams)

// ===========================
// Derived data
// ===========================

const entries = computed(() => leaderboardData.value?.data?.entries ?? [])

const podiumEntries = computed(() =>
	entries.value
		.filter((e) => (e.rank ?? 0) <= 3)
		.map((e) => ({
			rank: e.rank ?? 0,
			username: e.username ?? t('leaderboard.anonymous'),
			avatarUrl: e.avatar,
			value: e.mmr ?? 0,
			label: e.league ?? undefined,
		})),
)

const listEntries = computed(() => entries.value.filter((e) => (e.rank ?? 0) > 3))

const playerRank = computed(() => leaderboardData.value?.data?.playerRank ?? null)

const isEmpty = computed(() => !isLoading.value && !isError.value && entries.value.length === 0)

const errorMessage = computed(() => {
	if (!error.value) return undefined
	const e = error.value as { message?: string }
	return e?.message
})
</script>

<template>
	<LeaderboardList
		:is-loading="isLoading"
		:is-error="isError"
		:error-message="errorMessage"
		:is-empty="isEmpty"
		:empty-description="t('leaderboard.duel.emptyDesc')"
		:player-rank="playerRank"
		@retry="refetch"
		@refresh="refetch"
	>
		<!-- Podium: top 3 -->
		<template v-if="podiumEntries.length > 0" #podium>
			<LeaderboardPodium :entries="podiumEntries" />
		</template>

		<!-- Entries: rank 4+ -->
		<template #entries>
			<LeaderboardEntryRow
				v-for="entry in listEntries"
				:key="entry.playerId"
				:rank="entry.rank ?? 0"
				:username="entry.username ?? t('leaderboard.anonymous')"
				:avatar-url="entry.avatar"
				:is-current-user="entry.playerId === currentUser?.id"
			>
				<template #stats>
					<div class="flex flex-col items-end gap-0.5">
						<div class="flex items-center justify-end gap-1.5">
							<UIcon
								v-if="entry.leagueIcon"
								:name="entry.leagueIcon"
								class="size-4 shrink-0"
							/>
							<p class="text-sm font-bold text-(--ui-text-highlighted)">
								{{ entry.mmr ?? 0 }}
							</p>
						</div>
						<p class="text-xs text-(--ui-text-muted)">
							{{
								t('leaderboard.duel.record', {
									wins: entry.wins ?? 0,
									losses: entry.losses ?? 0,
								})
							}}
						</p>
					</div>
				</template>
			</LeaderboardEntryRow>
		</template>
	</LeaderboardList>
</template>
