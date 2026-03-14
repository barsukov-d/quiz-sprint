<script setup lang="ts">
import { useGetQuiz } from '@/api'
import { useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

// Получаем categoryId из query параметров
const categoryId = computed(() => route.query.categoryId as string | undefined)
const categoryName = computed(() => route.query.categoryName as string | undefined)

// Получаем квизы с фильтрацией по категории
const {
	data: quizzes,
	isLoading,
	isError,
	error,
	refetch,
} = useGetQuiz({
	categoryId: categoryId.value,
})

// Навигация назад к категориям
const goBackToCategories = () => {
	router.push({ name: 'categories' })
}

// Перейти к деталям квиза
const goToQuizDetails = (quizId: string) => {
	router.push({ name: 'quiz-details', params: { id: quizId } })
}
</script>

<template>
	<div class="container mx-auto p-4 pt-32">
		<!-- Header with back button -->
		<div class="flex items-center gap-3 mb-6">
			<UButton
				icon="i-heroicons-arrow-left"
				color="neutral"
				variant="ghost"
				size="lg"
				@click="goBackToCategories"
			/>
			<div>
				<h1 class="text-3xl font-bold text-(--ui-text-highlighted)">
					{{ t('quiz.title') }}
				</h1>
				<p v-if="categoryId" class="text-sm text-(--ui-text-dimmed)">
					{{ t('quiz.category', { name: categoryName }) }}
				</p>
			</div>
		</div>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex justify-center items-center py-12">
			<UProgress animation="carousel" />
			<span class="ml-4">{{ t('quiz.loading') }}</span>
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="mb-4">
			<UAlert
				color="red"
				variant="soft"
				:title="t('quiz.loadError')"
				:description="error?.error.message || t('quiz.loadFailed')"
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
				{{ t('quiz.tryAgain') }}
			</UButton>
		</div>

		<!-- Success state with data -->
		<div
			v-else-if="quizzes?.data && Array.isArray(quizzes.data) && quizzes.data.length > 0"
			class="grid gap-4 md:grid-cols-2 lg:grid-cols-3"
		>
			<UCard
				v-for="(quiz, index) in quizzes.data"
				:key="quiz.id || index"
				class="hover:shadow-lg transition-shadow"
			>
				<template #header>
					<h3 class="text-xl font-semibold">{{ quiz.title || t('quiz.unnamed') }}</h3>
				</template>

				<p class="text-(--ui-text-muted) text-sm mb-4">
					{{ quiz.description || t('quiz.noDescription') }}
				</p>

				<div class="flex items-center justify-between text-sm text-(--ui-text-dimmed) mb-4">
					<span
						>📝
						{{ t('quiz.questionsCount', { count: quiz.questionsCount || 0 }) }}</span
					>
					<span
						>⏱️
						{{
							quiz.timeLimit
								? t('quiz.timeLimit', { min: Math.floor(quiz.timeLimit / 60) })
								: 'N/A'
						}}</span
					>
				</div>

				<div class="flex items-center text-sm text-(--ui-text-dimmed) mb-4">
					<span>{{ t('quiz.passingScore', { score: quiz.passingScore || 0 }) }}</span>
				</div>

				<template #footer>
					<UButton block color="primary" @click="() => goToQuizDetails(quiz.id)">
						{{ t('quiz.viewQuiz') }}
					</UButton>
				</template>
			</UCard>
		</div>

		<!-- Empty state -->
		<div v-else class="text-center py-12 text-(--ui-text-dimmed)">
			<div class="text-6xl mb-4">📋</div>
			<p class="text-lg font-semibold mb-2">{{ t('quiz.notFound') }}</p>
			<p class="text-sm mb-4">{{ t('quiz.notFoundDesc') }}</p>
			<UButton @click="goBackToCategories"> {{ t('quiz.backToCategories') }} </UButton>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
