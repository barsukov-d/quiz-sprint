<script setup lang="ts">
import { computed } from 'vue'
import type { InternalInfrastructureHttpHandlersLeaderboardEntryDTO } from '@/api/generated'
import { useI18n } from 'vue-i18n'

interface Props {
	leaderboard: InternalInfrastructureHttpHandlersLeaderboardEntryDTO[] | null | undefined
	currentPlayerId?: string
	showRank?: boolean
	maxEntries?: number
}

const props = withDefaults(defineProps<Props>(), {
	showRank: true,
	maxEntries: 10,
})
const { t } = useI18n()

// ===========================
// Computed
// ===========================

const displayedLeaderboard = computed(() => {
	if (!props.leaderboard) return []
	return props.leaderboard.slice(0, props.maxEntries)
})

const getRankBadgeColor = (rank: number) => {
	if (rank === 1) return 'yellow'
	if (rank === 2) return 'gray'
	if (rank === 3) return 'orange'
	return 'blue'
}

const getRankEmoji = (rank: number) => {
	if (rank === 1) return '🥇'
	if (rank === 2) return '🥈'
	if (rank === 3) return '🥉'
	return ''
}

const isCurrentPlayer = (playerId: string) => {
	return playerId === props.currentPlayerId
}
</script>

<template>
	<div class="flex flex-col gap-4">
		<div class="pb-2 border-b border-(--ui-border)">
			<h3 class="text-lg font-semibold flex items-center gap-2 text-(--ui-text-highlighted)">
				<UIcon name="i-heroicons-trophy" class="size-5 text-yellow-500" />
				{{ t('daily.topPlayers') }}
			</h3>
			<p class="text-sm text-(--ui-text-muted)">{{ t('daily.todaysChallenge') }}</p>
		</div>

		<div class="flex flex-col gap-2">
			<div
				v-for="entry in displayedLeaderboard"
				:key="entry.userId"
				class="flex items-center gap-4 px-3 py-3 rounded-lg transition-all duration-200"
				:class="
					isCurrentPlayer(entry.userId)
						? 'bg-(--ui-primary)/10 border-2 border-(--ui-primary)/40'
						: 'bg-(--ui-bg-muted) hover:bg-(--ui-bg-elevated) border-2 border-transparent'
				"
			>
				<!-- Rank -->
				<div class="shrink-0 min-w-12 flex justify-center">
					<UBadge
						v-if="showRank"
						:color="getRankBadgeColor(entry.rank)"
						size="lg"
						variant="soft"
					>
						<span v-if="entry.rank <= 3" class="text-lg">{{
							getRankEmoji(entry.rank)
						}}</span>
						<span v-else>#{{ entry.rank }}</span>
					</UBadge>
				</div>

				<!-- Avatar & Name -->
				<div class="flex-1 flex items-center gap-3 min-w-0">
					<UAvatar :alt="entry.username" size="md" />
					<div class="flex-1 min-w-0">
						<p
							class="font-semibold text-(--ui-text-highlighted) overflow-hidden text-ellipsis whitespace-nowrap flex items-center gap-2"
						>
							{{ entry.username }}
							<UBadge v-if="isCurrentPlayer(entry.userId)" color="primary" size="xs">
								{{ t('daily.youBadge') }}
							</UBadge>
						</p>
					</div>
				</div>

				<!-- Score -->
				<div class="shrink-0 text-right">
					<div class="text-lg font-bold text-(--ui-primary) sm:text-base">
						{{ entry.score }}
					</div>
					<div class="text-xs text-(--ui-text-dimmed)">{{ t('daily.points') }}</div>
				</div>
			</div>

			<!-- Empty State -->
			<UEmpty
				v-if="displayedLeaderboard.length === 0"
				:title="t('daily.noPlayers')"
				:description="t('daily.noPlayersDesc')"
				icon="i-heroicons-user-group"
			/>
		</div>
	</div>
</template>
