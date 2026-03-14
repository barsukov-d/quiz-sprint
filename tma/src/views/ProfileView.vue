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
	<div class="mx-auto max-w-2xl">
		<!-- Header -->
		<h1 class="text-2xl font-bold mb-6 text-(--ui-text-highlighted)">
			{{ t('profile.title') }}
		</h1>

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
					<h2 class="text-xl font-bold text-(--ui-text-highlighted)">
						{{ currentUser.username }}
					</h2>
					<p v-if="currentUser.telegramUsername" class="text-sm text-(--ui-text-muted)">
						@{{ currentUser.telegramUsername }}
					</p>
					<p class="text-xs text-(--ui-text-dimmed) mt-1">ID: {{ currentUser.id }}</p>
				</div>
			</div>

			<div class="border-t border-(--ui-border) pt-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<p class="text-sm text-(--ui-text-muted)">{{ t('profile.email') }}</p>
						<p class="font-medium text-(--ui-text)">
							{{ currentUser.email || t('profile.notSet') }}
						</p>
					</div>
					<div>
						<p class="text-sm text-(--ui-text-muted)">{{ t('profile.language') }}</p>
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

		<!-- Statistics Card -->
		<UCard class="mb-6">
			<h3 class="text-lg font-bold mb-4 text-(--ui-text-highlighted)">
				{{ t('profile.statistics') }}
			</h3>
			<div class="grid grid-cols-2 gap-3">
				<div class="text-center p-4 rounded-xl bg-(--ui-bg-muted)">
					<div class="text-3xl font-bold text-primary">-</div>
					<div class="text-sm text-(--ui-text-muted) mt-1">
						{{ t('profile.quizzesCompleted') }}
					</div>
				</div>
				<div class="text-center p-4 rounded-xl bg-(--ui-bg-muted)">
					<div class="text-3xl font-bold text-primary">-</div>
					<div class="text-sm text-(--ui-text-muted) mt-1">
						{{ t('profile.totalPoints') }}
					</div>
				</div>
				<div class="text-center p-4 rounded-xl bg-(--ui-bg-muted)">
					<div class="text-3xl font-bold text-primary">-</div>
					<div class="text-sm text-(--ui-text-muted) mt-1">
						{{ t('profile.averageScore') }}
					</div>
				</div>
				<div class="text-center p-4 rounded-xl bg-(--ui-bg-muted)">
					<div class="text-3xl font-bold text-primary">-</div>
					<div class="text-sm text-(--ui-text-muted) mt-1">
						{{ t('profile.bestRank') }}
					</div>
				</div>
			</div>
		</UCard>

		<!-- Achievements Card -->
		<UCard class="mb-6">
			<h3 class="text-lg font-bold mb-4 text-(--ui-text-highlighted)">
				{{ t('profile.achievements') }}
			</h3>
			<div class="text-center py-8">
				<UIcon
					name="i-heroicons-trophy"
					class="size-14 text-(--ui-text-dimmed) mx-auto mb-3"
				/>
				<p class="text-(--ui-text-muted)">{{ t('profile.achievementsSoon') }}</p>
			</div>
		</UCard>

		<!-- Recent Activity -->
		<UCard>
			<h3 class="text-lg font-bold mb-4 text-(--ui-text-highlighted)">
				{{ t('profile.recentActivity') }}
			</h3>
			<div class="text-center py-8">
				<UIcon
					name="i-heroicons-clock"
					class="size-14 text-(--ui-text-dimmed) mx-auto mb-3"
				/>
				<p class="text-(--ui-text-muted)">{{ t('profile.activitySoon') }}</p>
			</div>
		</UCard>
	</div>
</template>
