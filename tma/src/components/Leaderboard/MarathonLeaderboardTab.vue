<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useGetMarathonLeaderboard } from '@/api/generated/hooks/marathonController/useGetMarathonLeaderboard'
import { useGetCategories } from '@/api/generated/hooks/categoryController/useGetCategories'
import { useAuth } from '@/composables/useAuth'
import LeaderboardList from '@/components/Leaderboard/LeaderboardList.vue'
import LeaderboardPodium from '@/components/Leaderboard/LeaderboardPodium.vue'
import LeaderboardEntryRow from '@/components/Leaderboard/LeaderboardEntryRow.vue'

const { t } = useI18n()
const { currentUser } = useAuth()

// ===========================
// Filter state
// ===========================

const selectedCategoryId = ref<string | undefined>(undefined)
const selectedTimeFrame = ref<'weekly' | 'all_time'>('weekly')

// ===========================
// Categories
// ===========================

const { data: categoriesData } = useGetCategories()

const categoryOptions = computed(() => {
	const all = [{ id: undefined, name: t('leaderboard.marathon.allCategories') }]
	const cats = (categoriesData.value?.data ?? []).map((c) => ({ id: c.id, name: c.name }))
	return [...all, ...cats]
})

const selectedCategoryOption = computed(
	() =>
		categoryOptions.value.find((o) => o.id === selectedCategoryId.value) ??
		categoryOptions.value[0] ?? {
			id: undefined,
			name: t('leaderboard.marathon.allCategories'),
		},
)

function onCategoryChange(option: unknown) {
	const o = option as { id: string | undefined; name: string }
	selectedCategoryId.value = o.id
}

// ===========================
// Leaderboard query
// ===========================

const queryParams = computed(() => ({
	categoryId: selectedCategoryId.value,
	timeFrame: selectedTimeFrame.value,
	limit: 50,
}))

const {
	data: leaderboardData,
	isLoading,
	isError,
	error,
	refetch,
} = useGetMarathonLeaderboard(queryParams)

// ===========================
// Derived data
// ===========================

const entries = computed(() => leaderboardData.value?.data?.entries ?? [])

const podiumEntries = computed(() =>
	entries.value
		.filter((e) => e.rank <= 3)
		.map((e) => ({
			rank: e.rank,
			username: e.username,
			value: e.bestScore,
			label: t('leaderboard.pts'),
		})),
)

const listEntries = computed(() => entries.value.filter((e) => e.rank > 3))

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
		:empty-description="t('leaderboard.marathon.emptyDesc')"
		:player-rank="playerRank"
		@retry="refetch"
		@refresh="refetch"
	>
		<!-- Header: category + timeFrame filters -->
		<template #header>
			<div class="flex items-center gap-2">
				<!-- Category selector -->
				<USelectMenu
					:model-value="selectedCategoryOption"
					:options="categoryOptions"
					option-attribute="name"
					value-attribute="id"
					class="flex-1 min-w-0"
					@update:model-value="onCategoryChange"
				/>

				<!-- TimeFrame toggle -->
				<UButtonGroup size="sm">
					<UButton
						:variant="selectedTimeFrame === 'weekly' ? 'solid' : 'ghost'"
						:color="selectedTimeFrame === 'weekly' ? 'primary' : 'neutral'"
						@click="selectedTimeFrame = 'weekly'"
					>
						{{ t('leaderboard.marathon.weekly') }}
					</UButton>
					<UButton
						:variant="selectedTimeFrame === 'all_time' ? 'solid' : 'ghost'"
						:color="selectedTimeFrame === 'all_time' ? 'primary' : 'neutral'"
						@click="selectedTimeFrame = 'all_time'"
					>
						{{ t('leaderboard.marathon.allTime') }}
					</UButton>
				</UButtonGroup>
			</div>
		</template>

		<!-- Podium: top 3 -->
		<template v-if="podiumEntries.length > 0" #podium>
			<LeaderboardPodium :entries="podiumEntries" />
		</template>

		<!-- Entries: rank 4+ -->
		<template #entries>
			<LeaderboardEntryRow
				v-for="entry in listEntries"
				:key="entry.playerId"
				:rank="entry.rank"
				:username="entry.username"
				:is-current-user="entry.playerId === currentUser?.id"
			>
				<template #stats>
					<div class="flex flex-col items-end gap-0.5">
						<p class="text-sm font-bold text-(--ui-text-highlighted)">
							{{ entry.bestScore }}
						</p>
						<p class="text-xs text-(--ui-text-muted)">
							<UIcon name="i-heroicons-fire" class="size-3 inline text-orange-400" />
							{{ entry.bestStreak }}
						</p>
					</div>
				</template>
			</LeaderboardEntryRow>
		</template>
	</LeaderboardList>
</template>
