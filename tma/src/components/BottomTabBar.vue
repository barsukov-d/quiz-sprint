<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

interface Tab {
	name: string
	path: string
	icon: string
	label: string
}

const tabs = computed<Tab[]>(() => [
	{
		name: 'home',
		path: '/',
		icon: 'i-heroicons-home',
		label: t('nav.home'),
	},
	{
		name: 'leaderboard',
		path: '/leaderboard',
		icon: 'i-heroicons-trophy',
		label: t('nav.leaderboard'),
	},
	{
		name: 'profile',
		path: '/profile',
		icon: 'i-heroicons-user',
		label: t('nav.profile'),
	},
])

const currentTab = computed(() => {
	const path = route.path
	if (path === '/') return 'home'
	if (path.startsWith('/leaderboard')) return 'leaderboard'
	if (path.startsWith('/profile')) return 'profile'
	return ''
})

const navigateTo = (tab: Tab) => {
	router.push(tab.path)
}
</script>

<template>
	<div
		class="fixed bottom-0 inset-x-0 z-50 bg-(--ui-bg-elevated) border-t border-(--ui-border) pb-[env(safe-area-inset-bottom)]"
	>
		<div class="flex justify-around items-center h-16 max-w-[1200px] mx-auto">
			<button
				v-for="tab in tabs"
				:key="tab.name"
				class="flex-1 flex flex-col items-center justify-center gap-1 py-2 transition-colors"
				:class="
					currentTab === tab.name
						? 'text-(--ui-primary)'
						: 'text-(--ui-text-dimmed) hover:text-(--ui-text-muted)'
				"
				@click="() => navigateTo(tab)"
			>
				<UIcon :name="tab.icon" class="size-6" />
				<span class="text-xs font-medium">{{ tab.label }}</span>
			</button>
		</div>
	</div>
</template>
