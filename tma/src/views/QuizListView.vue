<script setup lang="ts">
import { useGetQuiz } from '@/api'
import { useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const router = useRouter()

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
				color="gray"
				variant="ghost"
				size="lg"
				@click="goBackToCategories"
			/>
			<div>
				<h1 class="text-3xl font-bold">Квизы</h1>
				<p v-if="categoryId" class="text-sm text-gray-500">Категория: {{ categoryName }}</p>
			</div>
		</div>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex justify-center items-center py-12">
			<UProgress animation="carousel" />
			<span class="ml-4">Загрузка квизов...</span>
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="mb-4">
			<UAlert
				color="red"
				variant="soft"
				title="Ошибка загрузки"
				:description="error?.error.message || 'Не удалось загрузить квизы'"
			/>
			<UButton color="red" class="mt-2" @click="refetch()"> Попробовать снова </UButton>
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
					<h3 class="text-xl font-semibold">{{ quiz.title || 'Unnamed Quiz' }}</h3>
				</template>

				<p class="text-gray-600 text-sm mb-4">{{ quiz.description || 'No description' }}</p>

				<div class="flex items-center justify-between text-sm text-gray-500 mb-4">
					<span>📝 {{ quiz.questionsCount || 0 }} вопросов</span>
					<span
						>⏱️
						{{
							quiz.timeLimit ? `${Math.floor(quiz.timeLimit / 60)} мин` : 'N/A'
						}}</span
					>
				</div>

				<div class="flex items-center text-sm text-gray-500 mb-4">
					<span>✅ Проходной балл: {{ quiz.passingScore || 0 }}%</span>
				</div>

				<template #footer>
					<UButton block color="primary" @click="goToQuizDetails(quiz.id)">
						View Quiz
					</UButton>
				</template>
			</UCard>
		</div>

		<!-- Empty state -->
		<div v-else class="text-center py-12 text-gray-500">
			<div class="text-6xl mb-4">📋</div>
			<p class="text-lg font-semibold mb-2">Квизы не найдены</p>
			<p class="text-sm mb-4">В этой категории пока нет квизов</p>
			<UButton @click="goBackToCategories"> Вернуться к категориям </UButton>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
