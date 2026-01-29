<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'

const { currentUser, isAuthenticated } = useAuth()
</script>

<template>
	<div class="container mx-auto p-4 pt-20 pb-24">
		<div class="max-w-2xl mx-auto">
			<!-- Header -->
			<h1 class="text-3xl font-bold mb-6">Profile</h1>

			<!-- User Info Card -->
			<UCard v-if="isAuthenticated && currentUser" class="mb-6">
				<div class="flex items-center gap-4 mb-6">
					<UAvatar
						:src="currentUser.avatarUrl"
						:alt="currentUser.username"
						size="xl"
						:ui="{ rounded: 'rounded-full' }"
					/>
					<div>
						<h2 class="text-2xl font-bold">{{ currentUser.username }}</h2>
						<p v-if="currentUser.telegramUsername" class="text-gray-600">
							@{{ currentUser.telegramUsername }}
						</p>
						<p class="text-sm text-gray-400 mt-1">ID: {{ currentUser.id }}</p>
					</div>
				</div>

				<div class="border-t border-gray-200 pt-4">
					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-sm text-gray-600">Email</p>
							<p class="font-medium">{{ currentUser.email || 'Not set' }}</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">Language</p>
							<p class="font-medium">{{ currentUser.languageCode || 'en' }}</p>
						</div>
					</div>
				</div>
			</UCard>

			<!-- Statistics Card (TODO) -->
			<UCard class="mb-6">
				<h3 class="text-xl font-bold mb-4">Statistics</h3>
				<div class="grid grid-cols-2 gap-4">
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-blue-600">-</div>
						<div class="text-sm text-gray-600 mt-1">Quizzes Completed</div>
					</div>
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-green-600">-</div>
						<div class="text-sm text-gray-600 mt-1">Total Points</div>
					</div>
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-purple-600">-</div>
						<div class="text-sm text-gray-600 mt-1">Average Score</div>
					</div>
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-orange-600">-</div>
						<div class="text-sm text-gray-600 mt-1">Best Rank</div>
					</div>
				</div>
			</UCard>

			<!-- Achievements Card (TODO) -->
			<UCard class="mb-6">
				<h3 class="text-xl font-bold mb-4">Achievements</h3>
				<div class="text-center py-8 text-gray-500">
					<UIcon name="i-heroicons-trophy" class="text-6xl mb-2" />
					<p>Achievement system coming soon!</p>
				</div>
			</UCard>

			<!-- Recent Activity (TODO) -->
			<UCard>
				<h3 class="text-xl font-bold mb-4">Recent Activity</h3>
				<div class="text-center py-8 text-gray-500">
					<UIcon name="i-heroicons-clock" class="text-6xl mb-2" />
					<p>Activity history coming soon!</p>
				</div>
			</UCard>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
