<script setup lang="ts">
import { useI18n } from 'vue-i18n'

interface Props {
	rank: number
	username: string
	avatarUrl?: string
	isCurrentUser: boolean
}

defineProps<Props>()
const { t } = useI18n()

const getRankEmoji = (rank: number) => {
	if (rank === 1) return '🥇'
	if (rank === 2) return '🥈'
	if (rank === 3) return '🥉'
	return ''
}
</script>

<template>
	<div
		class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors"
		:class="
			isCurrentUser
				? 'bg-primary/10 border border-primary/30'
				: 'bg-(--ui-bg) border border-(--ui-border)'
		"
	>
		<!-- Rank -->
		<div class="shrink-0 w-8 text-center">
			<span v-if="rank <= 3" class="text-xl">{{ getRankEmoji(rank) }}</span>
			<span v-else class="text-xs font-bold text-(--ui-text-dimmed)">#{{ rank }}</span>
		</div>

		<!-- Avatar & Name -->
		<div class="flex-1 flex items-center gap-2.5 min-w-0">
			<UAvatar :src="avatarUrl" :alt="username" size="sm" />
			<div class="flex-1 min-w-0">
				<p
					class="text-sm font-semibold truncate flex items-center gap-1.5"
					:class="isCurrentUser ? 'text-(--ui-text-highlighted)' : 'text-(--ui-text)'"
				>
					{{ username || t('leaderboard.anonymous') }}
					<span
						v-if="isCurrentUser"
						class="text-xs font-medium text-primary bg-primary/10 px-1.5 py-0.5 rounded-full"
						>{{ t('leaderboard.you') }}</span
					>
				</p>
			</div>
		</div>

		<!-- Stats slot -->
		<div class="shrink-0 text-right">
			<slot name="stats" />
		</div>
	</div>
</template>
