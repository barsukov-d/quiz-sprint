<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const router = useRouter()

// Временно используем моковые данные, так как нет API для получения результатов сессии
// TODO: Добавить API endpoint для получения финальных результатов
const sessionId = route.params.sessionId as string

// Мок данные (в реальности должны приходить из API)
const mockResults = {
	score: 85,
	totalPoints: 100,
	correctAnswers: 8,
	totalQuestions: 10,
	timeSpent: 245, // секунды
	passed: true,
	rank: 12,
	totalPlayers: 1234
}

// Computed properties
const scorePercentage = computed(() => {
	return Math.round((mockResults.score / mockResults.totalPoints) * 100)
})

const formatTime = computed(() => {
	const minutes = Math.floor(mockResults.timeSpent / 60)
	const seconds = mockResults.timeSpent % 60
	return `${minutes}:${seconds.toString().padStart(2, '0')}`
})

const performanceEmoji = computed(() => {
	const percentage = scorePercentage.value
	if (percentage >= 90) return '🏆'
	if (percentage >= 75) return '🎉'
	if (percentage >= 60) return '👍'
	return '💪'
})

const performanceMessage = computed(() => {
	const percentage = scorePercentage.value
	if (percentage >= 90) return 'Outstanding!'
	if (percentage >= 75) return 'Great job!'
	if (percentage >= 60) return 'Good effort!'
	return 'Keep practicing!'
})

// Navigation
const tryAgain = () => {
	// Вернуться к деталям квиза
	router.push({ name: 'quiz-details', params: { id: 'quiz-id-here' } })
}

const viewLeaderboard = () => {
	router.push({ name: 'categories' })
}

const goHome = () => {
	router.push({ name: 'categories' })
}
</script>

<template>
	<div class="container mx-auto p-4 pt-20">
		<div class="max-w-2xl mx-auto">
			<!-- Main Result Card -->
			<UCard class="mb-6 text-center">
				<div class="py-8">
					<div class="text-8xl mb-4">{{ performanceEmoji }}</div>
					<h1 class="text-3xl font-bold mb-2">Quiz Completed!</h1>
					<p class="text-xl text-gray-600 mb-6">{{ performanceMessage }}</p>

					<!-- Score Circle -->
					<div class="flex justify-center mb-6">
						<div
							class="w-40 h-40 rounded-full border-8 flex items-center justify-center"
							:class="{
								'border-green-500 bg-green-50': mockResults.passed,
								'border-red-500 bg-red-50': !mockResults.passed
							}"
						>
							<div>
								<div class="text-5xl font-bold">{{ scorePercentage }}%</div>
								<div class="text-sm text-gray-600">{{ mockResults.score }}/{{ mockResults.totalPoints }}</div>
							</div>
						</div>
					</div>

					<!-- Pass/Fail Badge -->
					<UBadge
						:color="mockResults.passed ? 'green' : 'red'"
						variant="solid"
						size="lg"
						class="mb-4"
					>
						{{ mockResults.passed ? '✓ Passed' : '✗ Not Passed' }}
					</UBadge>
				</div>
			</UCard>

			<!-- Stats Grid -->
			<div class="grid grid-cols-2 gap-4 mb-6">
				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ mockResults.correctAnswers }}/{{ mockResults.totalQuestions }}</div>
						<div class="text-sm text-gray-600">Correct Answers</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ formatTime }}</div>
						<div class="text-sm text-gray-600">Time Spent</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">#{{ mockResults.rank }}</div>
						<div class="text-sm text-gray-600">Your Rank</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ mockResults.totalPlayers }}</div>
						<div class="text-sm text-gray-600">Total Players</div>
					</div>
				</UCard>
			</div>

			<!-- Actions -->
			<div class="space-y-3">
				<UButton size="xl" color="primary" block @click="tryAgain">
					Try Again
				</UButton>
				<UButton size="xl" color="gray" variant="outline" block @click="viewLeaderboard">
					View Leaderboard
				</UButton>
				<UButton size="xl" color="gray" variant="ghost" block @click="goHome">
					Back to Home
				</UButton>
			</div>

			<!-- Share Section (Optional) -->
			<UCard class="mt-6">
				<div class="text-center py-4">
					<p class="text-sm text-gray-600 mb-3">Share your achievement!</p>
					<div class="flex justify-center gap-3">
						<UButton icon="i-heroicons-share" color="gray" variant="outline">
							Share
						</UButton>
					</div>
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
