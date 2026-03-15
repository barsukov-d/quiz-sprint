<script setup lang="ts">
import { useI18n } from 'vue-i18n'

interface Props {
	isLoading: boolean
	isError: boolean
	errorMessage?: string
	isEmpty: boolean
	emptyTitle?: string
	emptyDescription?: string
	playerRank?: number | null
	totalPlayers?: number | null
}

withDefaults(defineProps<Props>(), {
	emptyTitle: '',
	emptyDescription: '',
})

const emit = defineEmits<{
	retry: []
	refresh: []
}>()

const { t } = useI18n()
</script>

<template>
	<div class="flex flex-col gap-3">
		<!-- Header slot (for filters) -->
		<slot name="header" />

		<!-- Loading -->
		<div v-if="isLoading" class="flex flex-col items-center justify-center py-12">
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-(--ui-text-muted) mt-4">{{ t('leaderboard.loading') }}</p>
		</div>

		<!-- Error -->
		<div v-else-if="isError" class="text-center py-12">
			<UAlert
				color="error"
				variant="soft"
				:title="t('leaderboard.loadFailed')"
				:description="errorMessage || t('leaderboard.loadFailed')"
			/>
			<UButton color="error" class="mt-3" @click="emit('retry')">
				{{ t('leaderboard.retry') }}
			</UButton>
		</div>

		<!-- Empty -->
		<div v-else-if="isEmpty" class="py-12">
			<UEmpty
				:title="emptyTitle || t('leaderboard.empty')"
				:description="emptyDescription || t('leaderboard.emptyDesc')"
				icon="i-heroicons-trophy"
			/>
		</div>

		<!-- Content -->
		<template v-else>
			<!-- Player rank badge (when outside visible list) -->
			<div
				v-if="playerRank && playerRank > 50"
				class="flex items-center justify-between px-3 py-2.5 rounded-lg bg-primary/10 border border-primary/30"
			>
				<p class="text-sm text-(--ui-text-muted)">{{ t('leaderboard.yourRank') }}</p>
				<div class="text-right">
					<p class="text-xl font-bold text-primary">#{{ playerRank }}</p>
					<p v-if="totalPlayers" class="text-xs text-(--ui-text-dimmed)">
						{{ t('leaderboard.outOfPlayers', { total: totalPlayers }) }}
					</p>
				</div>
			</div>

			<!-- Podium slot -->
			<slot name="podium" />

			<!-- Entries -->
			<div class="flex flex-col gap-2">
				<slot name="entries" />
			</div>
		</template>
	</div>
</template>
