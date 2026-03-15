<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DailyLeaderboardTab from '@/components/Leaderboard/DailyLeaderboardTab.vue'
import MarathonLeaderboardTab from '@/components/Leaderboard/MarathonLeaderboardTab.vue'
import DuelLeaderboardTab from '@/components/Leaderboard/DuelLeaderboardTab.vue'
import GlobalLeaderboardTab from '@/components/Leaderboard/GlobalLeaderboardTab.vue'

const { t } = useI18n()

const activeTab = ref('daily')

const tabs = [
	{ value: 'daily', label: () => t('leaderboard.tabs.daily'), icon: 'i-heroicons-calendar' },
	{
		value: 'marathon',
		label: () => t('leaderboard.tabs.marathon'),
		icon: 'i-heroicons-fire',
	},
	{ value: 'duel', label: () => t('leaderboard.tabs.duel'), icon: 'i-heroicons-bolt' },
	{
		value: 'global',
		label: () => t('leaderboard.tabs.global'),
		icon: 'i-heroicons-globe-alt',
	},
]
</script>

<template>
	<div class="mx-auto max-w-[800px] pb-4">
		<!-- Header -->
		<div class="text-center mb-4">
			<h1 class="text-xl font-bold text-(--ui-text-highlighted)">
				{{ t('leaderboard.title') }}
			</h1>
		</div>

		<!-- Tab Bar -->
		<div
			class="flex gap-1 mb-4 p-1 rounded-xl bg-(--ui-bg-elevated) border border-(--ui-border)"
		>
			<button
				v-for="tab in tabs"
				:key="tab.value"
				class="flex-1 flex items-center justify-center gap-1.5 px-2 py-2 rounded-lg text-xs font-semibold transition-all duration-200"
				:class="
					activeTab === tab.value
						? 'bg-primary text-white shadow-sm'
						: 'text-(--ui-text-muted) hover:text-(--ui-text) hover:bg-(--ui-bg-accented)'
				"
				@click="activeTab = tab.value"
			>
				<UIcon :name="tab.icon" class="size-4" />
				<span>{{ tab.label() }}</span>
			</button>
		</div>

		<!-- Tab Content -->
		<KeepAlive>
			<DailyLeaderboardTab v-if="activeTab === 'daily'" />
			<MarathonLeaderboardTab v-else-if="activeTab === 'marathon'" />
			<DuelLeaderboardTab v-else-if="activeTab === 'duel'" />
			<GlobalLeaderboardTab v-else-if="activeTab === 'global'" />
		</KeepAlive>
	</div>
</template>
