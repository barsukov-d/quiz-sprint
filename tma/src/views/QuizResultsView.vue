<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed, watch } from 'vue'
import { useGetQuizSessionSessionid } from '@/api'
import { useLastQuiz } from '@/composables/useLastQuiz'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { saveLastQuizId } = useLastQuiz()
const { t } = useI18n()

// Get sessionId from route params
const sessionId = route.params.sessionId as string

// Fetch session results from API
const {
	data: resultsResponse,
	isLoading,
	isError,
	error,
} = useGetQuizSessionSessionid({
	sessionId,
})

// Extract data from response
const results = computed(() => resultsResponse.value?.data)

// Save quiz ID when results are loaded
watch(
	results,
	(newResults) => {
		if (newResults?.quiz?.id) {
			saveLastQuizId(newResults.quiz.id)
		}
	},
	{ immediate: true },
)

// Computed properties
const scorePercentage = computed(() => {
	return results.value?.scorePercentage || 0
})

const formatTime = computed(() => {
	const timeSpent = results.value?.timeSpent || 0
	const minutes = Math.floor(timeSpent / 60)
	const seconds = timeSpent % 60
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
	const quizId = results.value?.quiz.id
	if (quizId) {
		router.push({ name: 'quiz-details', params: { id: quizId } })
	}
}

const goHome = () => {
	router.push({ name: 'categories' })
}
</script>

<template>
	<div class="min-h-screen bg-gradient-to-b from-primary-600 to-primary-900 flex flex-col">
		<!-- Loading -->
		<div v-if="isLoading" class="flex flex-1 justify-center items-center py-12">
			<UProgress animation="carousel" class="w-48" />
			<span class="ml-4 text-white">{{ t('quiz.loadingResults') }}</span>
		</div>

		<!-- Error -->
		<div v-else-if="isError" class="flex flex-1 items-center justify-center p-6">
			<UAlert
				color="red"
				:title="t('quiz.loadResultsFailed')"
				:description="error?.error?.message || t('quiz.tryAgain2')"
			/>
		</div>

		<!-- Results -->
		<template v-else-if="results">
			<!-- Header -->
			<div class="relative flex items-center justify-center pt-4 pb-2 px-4">
				<button class="absolute left-4 text-white p-2" @click="goHome">
					<UIcon name="i-heroicons-x-mark" class="text-2xl" />
				</button>
				<h1 class="text-xl font-bold text-white">{{ t('quiz.finalScoreboard') }}</h1>
			</div>

			<!-- Podium (single player — show as 1st place) -->
			<div class="flex justify-center items-end gap-6 px-8 pt-6 pb-2">
				<!-- 1st place -->
				<div class="flex flex-col items-center gap-2">
					<div
						class="w-20 h-20 rounded-full bg-white/20 border-4 border-yellow-400 flex items-center justify-center text-3xl font-bold text-white"
					>
						{{ results.quiz.title?.[0]?.toUpperCase() || 'U' }}
					</div>
					<span class="text-white font-semibold text-sm">{{
						results.quiz.title || 'You'
					}}</span>
					<div class="bg-gray-900 rounded-full px-3 py-1 text-white text-sm font-bold">
						{{ results.session.score }}
					</div>
					<!-- Podium block -->
					<div
						class="w-24 h-20 bg-white/20 rounded-t-lg flex items-center justify-center"
					>
						<span class="text-4xl">🥇</span>
					</div>
				</div>
			</div>

			<!-- Stats Summary -->
			<div class="mx-4 mt-4 bg-black/20 rounded-2xl px-4 py-4">
				<div class="grid grid-cols-3 gap-4 text-center">
					<div>
						<div class="text-2xl font-bold text-white">
							{{ results.correctAnswers }}/{{ results.totalQuestions }}
						</div>
						<div class="text-xs text-white/70 mt-1">{{ t('quiz.correctAnswers') }}</div>
					</div>
					<div>
						<div class="text-2xl font-bold text-white">{{ formatTime }}</div>
						<div class="text-xs text-white/70 mt-1">{{ t('quiz.timeSpent') }}</div>
					</div>
					<div>
						<div class="text-2xl font-bold text-white">{{ scorePercentage }}%</div>
						<div class="text-xs text-white/70 mt-1">{{ t('quiz.accuracy') }}</div>
					</div>
				</div>
			</div>

			<!-- Performance message -->
			<div class="text-center mt-4 px-4">
				<p class="text-white/90 text-lg font-medium">
					{{ performanceEmoji }} {{ performanceMessage }}
				</p>
				<p class="text-white/60 text-sm mt-1">{{ results.quiz.title }}</p>
			</div>

			<!-- Spacer -->
			<div class="flex-1" />

			<!-- Bottom Buttons -->
			<div class="flex gap-3 px-4 pb-8 pt-4">
				<UButton
					size="xl"
					class="flex-1 border border-white/60 text-white bg-transparent hover:bg-white/10"
					variant="outline"
					color="neutral"
					@click="tryAgain"
				>
					{{ t('quiz.tryAgainBtn') }}
				</UButton>
				<UButton
					size="xl"
					class="flex-1 border border-white/60 text-white bg-transparent hover:bg-white/10"
					variant="outline"
					color="neutral"
					icon="i-heroicons-share"
				>
					{{ t('quiz.share') }}
				</UButton>
			</div>
		</template>
	</div>
</template>
