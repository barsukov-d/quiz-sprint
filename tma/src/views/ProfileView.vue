<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { useI18n } from 'vue-i18n'
import { setLocale, type Locale } from '@/i18n/index'

const { currentUser, isAuthenticated } = useAuth()
const { t, locale } = useI18n()

const localeOptions = [
	{ label: 'English', value: 'en' },
	{ label: 'Русский', value: 'ru' },
]

function handleLocaleChange(val: string) {
	setLocale(val as Locale)
}
</script>

<template>
	<div class="container mx-auto p-4 pt-20 pb-24">
		<div class="max-w-2xl mx-auto">
			<!-- Header -->
			<h1 class="text-3xl font-bold mb-6">{{ t('profile.title') }}</h1>

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
							<p class="text-sm text-gray-600">{{ t('profile.email') }}</p>
							<p class="font-medium">
								{{ currentUser.email || t('profile.notSet') }}
							</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">{{ t('profile.language') }}</p>
							<USelect
								:model-value="locale"
								:items="localeOptions"
								value-key="value"
								@update:model-value="handleLocaleChange"
							/>
						</div>
					</div>
				</div>
			</UCard>

			<!-- Statistics Card (TODO) -->
			<UCard class="mb-6">
				<h3 class="text-xl font-bold mb-4">{{ t('profile.statistics') }}</h3>
				<div class="grid grid-cols-2 gap-4">
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-blue-600">-</div>
						<div class="text-sm text-gray-600 mt-1">
							{{ t('profile.quizzesCompleted') }}
						</div>
					</div>
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-green-600">-</div>
						<div class="text-sm text-gray-600 mt-1">{{ t('profile.totalPoints') }}</div>
					</div>
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-purple-600">-</div>
						<div class="text-sm text-gray-600 mt-1">
							{{ t('profile.averageScore') }}
						</div>
					</div>
					<div class="text-center p-4 bg-gray-50 rounded-lg">
						<div class="text-3xl font-bold text-orange-600">-</div>
						<div class="text-sm text-gray-600 mt-1">{{ t('profile.bestRank') }}</div>
					</div>
				</div>
			</UCard>

			<!-- Achievements Card (TODO) -->
			<UCard class="mb-6">
				<h3 class="text-xl font-bold mb-4">{{ t('profile.achievements') }}</h3>
				<div class="text-center py-8 text-gray-500">
					<UIcon name="i-heroicons-trophy" class="text-6xl mb-2" />
					<p>{{ t('profile.achievementsSoon') }}</p>
				</div>
			</UCard>

			<!-- Recent Activity (TODO) -->
			<UCard>
				<h3 class="text-xl font-bold mb-4">{{ t('profile.recentActivity') }}</h3>
				<div class="text-center py-8 text-gray-500">
					<UIcon name="i-heroicons-clock" class="text-6xl mb-2" />
					<p>{{ t('profile.activitySoon') }}</p>
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
