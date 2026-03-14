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
	<div class="flex flex-col gap-3">
		<div class="flex items-center gap-2 mb-1">
			<span class="text-lg">🏆</span>
			<div>
				<h3 class="text-sm font-semibold text-(--ui-text-highlighted)">
					{{ t('daily.topPlayers') }}
				</h3>
				<p class="text-xs text-(--ui-text-muted)">{{ t('daily.todaysChallenge') }}</p>
			</div>
		</div>

		<div class="flex flex-col gap-1.5">
			<div
				v-for="entry in displayedLeaderboard"
				:key="entry.userId"
				class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200"
				:class="
					isCurrentPlayer(entry.userId)
						? 'bg-primary/10 border border-primary/30'
						: 'bg-(--ui-bg) border border-(--ui-border)'
				"
			>
				<!-- Rank -->
				<div class="shrink-0 w-8 text-center">
					<span v-if="showRank && entry.rank <= 3" class="text-xl">{{
						getRankEmoji(entry.rank)
					}}</span>
					<span v-else-if="showRank" class="text-xs font-bold text-(--ui-text-dimmed)"
						>#{{ entry.rank }}</span
					>
				</div>

				<!-- Avatar & Name -->
				<div class="flex-1 flex items-center gap-2.5 min-w-0">
					<UAvatar :alt="entry.username" size="sm" />
					<div class="flex-1 min-w-0">
						<p
							class="text-sm font-semibold text-(--ui-text-highlighted) truncate flex items-center gap-1.5"
						>
							{{ entry.username }}
							<span
								v-if="isCurrentPlayer(entry.userId)"
								class="text-xs font-medium text-primary bg-primary/10 px-1.5 py-0.5 rounded-full"
								>{{ t('daily.youBadge') }}</span
							>
						</p>
					</div>
				</div>

				<!-- Score -->
				<div class="shrink-0 text-right">
					<div class="text-sm font-bold text-(--ui-primary)">{{ entry.score }}</div>
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
