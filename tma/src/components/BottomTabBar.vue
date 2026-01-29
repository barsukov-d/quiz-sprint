<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

interface Tab {
	name: string
	path: string
	icon: string
	label: string
}

const tabs: Tab[] = [
	{
		name: 'home',
		path: '/',
		icon: 'i-heroicons-home',
		label: 'Home',
	},
	{
		name: 'leaderboard',
		path: '/leaderboard',
		icon: 'i-heroicons-trophy',
		label: 'Leaderboard',
	},
	{
		name: 'profile',
		path: '/profile',
		icon: 'i-heroicons-user',
		label: 'Profile',
	},
]

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
	<div class="bottom-tab-bar">
		<div class="tab-bar-container">
			<button
				v-for="tab in tabs"
				:key="tab.name"
				:class="['tab-button', { active: currentTab === tab.name }]"
				@click="navigateTo(tab)"
			>
				<UIcon :name="tab.icon" class="tab-icon" />
				<span class="tab-label">{{ tab.label }}</span>
			</button>
		</div>
	</div>
</template>

<style scoped>
.bottom-tab-bar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	background: white;
	border-top: 1px solid #e5e7eb;
	padding-bottom: env(safe-area-inset-bottom);
	z-index: 50;
}

.tab-bar-container {
	display: flex;
	justify-content: space-around;
	align-items: center;
	height: 64px;
	max-width: 1200px;
	margin: 0 auto;
}

.tab-button {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 4px;
	padding: 8px;
	border: none;
	background: transparent;
	cursor: pointer;
	transition: all 0.2s;
	color: #6b7280;
}

.tab-button:hover {
	color: #374151;
}

.tab-button.active {
	color: #3b82f6;
}

.tab-icon {
	font-size: 24px;
	width: 24px;
	height: 24px;
}

.tab-label {
	font-size: 12px;
	font-weight: 500;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
	.bottom-tab-bar {
		background: #1f2937;
		border-top-color: #374151;
	}

	.tab-button {
		color: #9ca3af;
	}

	.tab-button:hover {
		color: #d1d5db;
	}

	.tab-button.active {
		color: #60a5fa;
	}
}
</style>
