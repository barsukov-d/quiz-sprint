<script setup lang="ts">
import { useGetQuiz } from '@/api'

// Получаем квизы через сгенерированный hook
const { data: quizzes, isLoading, isError, error, refetch } = useGetQuiz()
</script>

<template>
	<div class="container mx-auto p-4">
		<h1 class="text-3xl font-bold mb-6">Quiz Sprint</h1>
		<p class="text-gray-600 mb-8">Выберите квиз для начала</p>

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
				:description="error?.message || 'Не удалось загрузить квизы'"
			/>
			<UButton color="red" class="mt-2" @click="() => refetch()"> Попробовать снова </UButton>
		</div>

		<!-- Success state with data -->
		<div
			v-else-if="quizzes?.data && Array.isArray(quizzes.data)"
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
					<span>{{ quiz.questionsCount || 0 }} вопросов</span>
					<UBadge>{{ quiz.difficulty || 'Medium' }}</UBadge>
				</div>

				<template #footer>
					<UButton block color="primary"> Начать квиз </UButton>
				</template>
			</UCard>
		</div>

		<!-- Empty state -->
		<div v-else class="text-center py-12 text-gray-500">
			<p>Квизы пока не доступны</p>
			<p class="text-sm mt-2">Убедитесь что backend запущен</p>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
