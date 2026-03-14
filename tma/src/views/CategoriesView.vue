<script setup lang="ts">
import { useGetCategories } from '@/api'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// Получаем категории через сгенерированный hook
const { data: categories, isLoading, isError, error, refetch } = useGetCategories()

// Router для навигации
const router = useRouter()

// Получаем данные авторизованного пользователя

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
	<div class="min-h-screen bg-(--ui-bg)">
		<!-- Header -->
		<div class="flex items-center gap-3 px-4 py-4 border-b border-(--ui-border)">
			<h1 class="text-xl font-bold text-(--ui-text-highlighted)">
				{{ t('categories.title') }}
			</h1>
		</div>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex justify-center items-center py-16">
			<UProgress animation="carousel" class="w-32" />
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="px-4 pt-6">
			<UAlert
				color="red"
				variant="soft"
				:title="t('categories.loadError')"
				:description="error?.error.message || t('categories.loadFailed')"
			/>
			<UButton
				color="red"
				class="mt-3"
				@click="
					() => {
						refetch()
					}
				"
			>
				{{ t('categories.tryAgain') }}
			</UButton>
		</div>

		<!-- Success state — 2-column grid -->
		<div
			v-else-if="categories?.data && Array.isArray(categories.data)"
			class="grid grid-cols-2 gap-3 p-4"
		>
			<button
				v-for="category in categories.data"
				:key="category.id"
				class="flex flex-col items-start gap-2 p-4 rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border) text-left transition-transform active:scale-95 hover:border-(--ui-border-accented)"
				@click="() => navigateToQuizzes(category.id, category.name)"
			>
				<span class="text-3xl">{{ getCategoryIcon(category.name) }}</span>
				<span class="text-sm font-semibold text-(--ui-text-highlighted) leading-tight">
					{{ category.name }}
				</span>
			</button>
		</div>

		<!-- Empty state -->
		<div v-else class="flex flex-col items-center py-16 text-(--ui-text-dimmed) px-4">
			<span class="text-5xl mb-4">📂</span>
			<p class="text-base font-semibold mb-1">{{ t('categories.empty') }}</p>
			<p class="text-sm text-center">{{ t('categories.emptyDesc') }}</p>
		</div>
	</div>
</template>
