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
	<div class="mx-auto max-w-[800px] pb-8">
		<!-- Header -->
		<div class="flex items-center justify-between mb-4">
			<h1 class="text-xl font-bold text-(--ui-text-highlighted)">
				{{ t('profile.title') }}
			</h1>
			<UIcon name="i-heroicons-cog-6-tooth" class="size-5 text-(--ui-text-dimmed)" />
		</div>

		<!-- Cover Banner (compact) -->
		<div class="relative rounded-(--ui-radius) overflow-hidden mb-4">
			<div
				class="h-28 w-full bg-gradient-to-br from-primary-600 via-violet-500 to-primary-400"
			/>
		</div>

		<!-- Avatar + Name Row -->
		<div class="flex items-center gap-4 px-1 mb-5">
			<UAvatar
				:src="currentUser?.avatarUrl"
				:alt="currentUser?.username"
				size="xl"
				:ui="{ rounded: 'rounded-full' }"
			/>
			<div class="flex-1 min-w-0">
				<template v-if="isAuthenticated && currentUser">
					<h2 class="text-lg font-bold text-(--ui-text-highlighted)">
						{{ currentUser.username }}
					</h2>
					<p v-if="currentUser.telegramUsername" class="text-sm text-(--ui-text-muted)">
						@{{ currentUser.telegramUsername }}
					</p>
				</template>
				<template v-else>
					<h2 class="text-lg font-bold text-(--ui-text-highlighted)">Guest</h2>
				</template>
			</div>
		</div>

		<!-- Stats Grid (2 rows x 3 cols like Figma) -->
		<div class="rounded-(--ui-radius) border border-(--ui-border) overflow-hidden mb-5">
			<div class="grid grid-cols-3 divide-x divide-(--ui-border)">
				<div class="flex flex-col items-center py-3">
					<span class="text-lg font-bold text-(--ui-text-highlighted)">-</span>
					<span class="text-xs text-(--ui-text-muted)">{{
						t('profile.quizzesCompleted')
					}}</span>
				</div>
				<div class="flex flex-col items-center py-3">
					<span class="text-lg font-bold text-(--ui-text-highlighted)">-</span>
					<span class="text-xs text-(--ui-text-muted)">{{
						t('profile.totalPoints')
					}}</span>
				</div>
				<div class="flex flex-col items-center py-3">
					<span class="text-lg font-bold text-(--ui-text-highlighted)">-</span>
					<span class="text-xs text-(--ui-text-muted)">{{
						t('profile.averageScore')
					}}</span>
				</div>
			</div>
		</div>

		<!-- Language Selector -->
		<div class="mb-5">
			<p class="text-sm font-medium text-(--ui-text-muted) mb-2">
				{{ t('profile.language') }}
			</p>
			<div class="flex gap-2">
				<button
					v-for="opt in localeOptions"
					:key="opt.value"
					class="px-4 py-1.5 rounded-full text-sm font-medium border transition-colors"
					:class="
						locale === opt.value
							? 'bg-primary text-white border-primary'
							: 'border-(--ui-border) text-(--ui-text-muted) bg-(--ui-bg-elevated)'
					"
					@click="handleLocaleChange(opt.value)"
				>
					{{ opt.label }}
				</button>
			</div>
		</div>

		<!-- Recent Activity -->
		<div>
			<h3 class="text-base font-bold text-(--ui-text-highlighted) mb-3">
				{{ t('profile.recentActivity') }}
			</h3>
			<div
				class="flex flex-col items-center py-8 rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border) text-(--ui-text-dimmed)"
			>
				<UIcon name="i-heroicons-clock" class="size-10 mb-2" />
				<p class="text-sm">{{ t('profile.activitySoon') }}</p>
			</div>
		</div>
	</div>
</template>
