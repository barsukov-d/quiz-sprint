<script setup lang="ts">
import { useGetCategories } from '@/api'
import { useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// Получаем категории через сгенерированный hook
const { data: categories, isLoading, isError, error, refetch } = useGetCategories()

// Router для навигации
const router = useRouter()

// Получаем данные авторизованного пользователя
const { currentUser, isAuthenticated } = useAuth()

// Навигация к списку квизов по категории
const navigateToQuizzes = (categoryId: string, categoryName: string) => {
	router.push({
		name: 'quizzes',
		query: { categoryId, categoryName },
	})
}

// Маппинг категорий к иконкам (можно вынести в конфиг)
const categoryIcons: Record<string, string> = {
	'general-knowledge': '🧠',
	geography: '🌍',
	technology: '💻',
	'movies-tv': '🎬',
	history: '📚',
	science: '🔬',
	sports: '⚽',
	music: '🎵',
	art: '🎨',
	food: '🍕',
}

// Получить иконку для категории (по slug или дефолтную)
const getCategoryIcon = (categoryName: string): string => {
	const slug = categoryName.toLowerCase().replace(/\s+/g, '-')
	return categoryIcons[slug] || '📋'
}
</script>

<template>
	<div class="container mx-auto p-4 pt-32">
		<!-- User Info Card -->
		<UCard v-if="isAuthenticated && currentUser" class="mb-6">
			<div class="flex items-center gap-4">
				<UAvatar
					:src="currentUser.avatarUrl"
					:alt="currentUser.username"
					size="lg"
					:ui="{ rounded: 'rounded-full' }"
				/>
				<div>
					<h2 class="text-xl font-semibold">{{ currentUser.username }}</h2>
					<p v-if="currentUser.telegramUsername" class="text-sm text-(--ui-text-dimmed)">
						{{ currentUser.telegramUsername }}
					</p>
				</div>
			</div>
		</UCard>

		<h1 class="text-3xl font-bold mb-2 text-(--ui-text-highlighted)">
			{{ t('categories.title') }}
		</h1>
		<p class="text-(--ui-text-muted) mb-8">{{ t('categories.subtitle') }}</p>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex justify-center items-center py-12">
			<UProgress animation="carousel" />
			<span class="ml-4">{{ t('categories.loading') }}</span>
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="mb-4">
			<UAlert
				color="red"
				variant="soft"
				:title="t('categories.loadError')"
				:description="error?.error.message || t('categories.loadFailed')"
			/>
			<UButton
				color="red"
				class="mt-2"
				@click="
					() => {
						refetch()
					}
				"
			>
				{{ t('categories.tryAgain') }}
			</UButton>
		</div>

		<!-- Success state with data -->
		<div v-else-if="categories?.data && Array.isArray(categories.data)" class="space-y-3">
			<UCard
				v-for="category in categories.data"
				:key="category.id"
				class="hover:shadow-lg transition-all cursor-pointer hover:scale-[1.02]"
				@click="() => navigateToQuizzes(category.id, category.name)"
			>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4 flex-1">
						<!-- Icon -->
						<div class="text-4xl">{{ getCategoryIcon(category.name) }}</div>

						<!-- Category Info -->
						<div class="flex-1">
							<h3 class="text-lg font-semibold mb-1">{{ category.name }}</h3>
							<p class="text-sm text-(--ui-text-dimmed)">
								{{
									t('categories.exploreDesc', {
										name: category.name.toLowerCase(),
									})
								}}
							</p>
						</div>
					</div>

					<!-- Arrow indicator -->
					<div class="text-(--ui-text-dimmed)">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-6 w-6"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 5l7 7-7 7"
							/>
						</svg>
					</div>
				</div>
			</UCard>
		</div>

		<!-- Empty state -->
		<div v-else class="text-center py-12 text-(--ui-text-dimmed)">
			<div class="text-6xl mb-4">📂</div>
			<p class="text-lg font-semibold mb-2">{{ t('categories.empty') }}</p>
			<p class="text-sm">{{ t('categories.emptyDesc') }}</p>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 800px;
}
</style>
