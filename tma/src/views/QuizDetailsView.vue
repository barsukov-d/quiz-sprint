<script setup lang="ts">
import { useGetQuizId, useGetQuizIdLeaderboard } from '@/api'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()

// Получаем ID квиза из URL
const quizId = route.params.id as string

// Загружаем детали квиза
const { data: quizData, isLoading, isError, error } = useGetQuizId({ id: quizId })

// Топ-лидеры будут в quizData.data.topScores
// const { data: leaderboardData } = useGetQuizIdLeaderboard({ id: quizId })

// Начать квиз
const startQuiz = () => {
	router.push({
		name: 'quiz-play',
		params: { id: quizId }
	})
}

// Назад к списку квизов
const goBack = () => {
	router.back()
}

// Форматирование времени
const formatTime = (seconds: number) => {
	const minutes = Math.floor(seconds / 60)
	return `${minutes} min`
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
				@click="goBack"
			/>
		</div>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex justify-center items-center py-12">
			<UProgress animation="carousel" />
			<span class="ml-4">Loading quiz details...</span>
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="mb-4">
			<UAlert
				color="red"
				variant="soft"
				title="Error loading quiz"
				:description="error?.error.message || 'Failed to load quiz details'"
			/>
		</div>

		<!-- Success state -->
		<div v-else-if="quizData?.data?.quiz" class="max-w-2xl mx-auto">
			<!-- Quiz Header -->
			<div class="text-center mb-8">
				<div class="text-6xl mb-4">🧠</div>
				<h1 class="text-3xl font-bold mb-4">{{ quizData.data.quiz.title }}</h1>
				<p class="text-gray-600">{{ quizData.data.quiz.description }}</p>
			</div>

			<!-- Quiz Stats Card -->
			<UCard class="mb-6">
				<template #header>
					<h2 class="text-xl font-semibold">📊 Quiz Stats</h2>
				</template>

				<div class="space-y-3">
					<div class="flex justify-between items-center">
						<span class="text-gray-600">Questions:</span>
						<span class="font-semibold">{{ quizData.data.quiz.questions?.length || 0 }}</span>
					</div>
					<div class="flex justify-between items-center">
						<span class="text-gray-600">Time Limit:</span>
						<span class="font-semibold">{{
							formatTime(quizData.data.quiz.timeLimit || 0)
						}}</span>
					</div>
					<div class="flex justify-between items-center">
						<span class="text-gray-600">Passing Score:</span>
						<span class="font-semibold">{{ quizData.data.quiz.passingScore || 0 }}%</span>
					</div>
				</div>
			</UCard>

			<!-- Top 3 Leaderboard Card -->
			<UCard v-if="quizData?.data?.topScores && quizData.data.topScores.length > 0" class="mb-8">
				<template #header>
					<div class="flex justify-between items-center">
						<h2 class="text-xl font-semibold">👑 Top 3 Leaders</h2>
					</div>
				</template>

				<div class="space-y-3">
					<div
						v-for="(entry, index) in quizData.data.topScores.slice(0, 3)"
						:key="entry.userId"
						class="flex items-center justify-between p-3 rounded-lg"
						:class="{
							'bg-yellow-50': index === 0,
							'bg-gray-50': index === 1,
							'bg-orange-50': index === 2
						}"
					>
						<div class="flex items-center gap-3">
							<span class="text-2xl">
								{{ index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉' }}
							</span>
							<div>
								<div class="font-semibold">{{ entry.username || 'Anonymous' }}</div>
								<div class="text-sm text-gray-500">
									{{ new Date(entry.completedAt * 1000).toLocaleDateString() }}
								</div>
							</div>
						</div>
						<div class="font-bold text-lg">{{ entry.score }}</div>
					</div>
				</div>
			</UCard>

			<!-- Start Quiz Button -->
			<div class="text-center">
				<UButton size="xl" color="primary" class="px-12 py-4" @click="startQuiz">
					START QUIZ →
				</UButton>
			</div>
		</div>
	</div>
</template>

<style scoped>
.container {
	max-width: 1200px;
}
</style>
