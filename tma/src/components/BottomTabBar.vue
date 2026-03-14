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
	<div class="tab-bar-wrapper">
		<div class="tab-bar-container">
			<button
				v-for="tab in tabs"
				:key="tab.name"
				class="tab-item"
				:class="currentTab === tab.name ? 'tab-item--active' : 'tab-item--inactive'"
				@click="() => navigateTo(tab)"
			>
				<div
					class="tab-icon-wrap"
					:class="currentTab === tab.name ? 'tab-icon-wrap--active' : ''"
				>
					<UIcon :name="tab.icon" class="size-5" />
				</div>
				<span class="tab-label">{{ tab.label }}</span>
			</button>
		</div>
	</div>
</template>

<style scoped>
.tab-bar-wrapper {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	z-index: 50;
	background-color: var(--ui-bg-elevated);
	border-top: 1px solid var(--ui-border);
	padding-bottom: env(safe-area-inset-bottom);
	/* Subtle shadow for elevated feel */
	box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.08);
}

.dark .tab-bar-wrapper {
	box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.4);
}

.tab-bar-container {
	display: flex;
	justify-content: space-around;
	align-items: center;
	height: 4rem;
	max-width: 1200px;
	margin: 0 auto;
	padding: 0 0.5rem;
}

.tab-item {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 2px;
	padding: 0.375rem 0;
	border-radius: 0.5rem;
	transition: color 150ms ease;
	cursor: pointer;
	background: none;
	border: none;
	outline: none;
}

.tab-item--active {
	color: var(--ui-primary);
}

.tab-item--inactive {
	color: var(--ui-text-dimmed);
}

.tab-item--inactive:hover {
	color: var(--ui-text-muted);
}

.tab-icon-wrap {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 1.75rem;
	border-radius: 1rem;
	transition: background-color 150ms ease;
}

.tab-icon-wrap--active {
	background-color: var(--ui-bg-accented);
}

.tab-label {
	font-size: 0.6875rem;
	font-weight: 500;
	line-height: 1;
}
</style>
