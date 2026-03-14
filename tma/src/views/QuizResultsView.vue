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

const avgAnswerTime = computed(() => {
	const avgTime = results.value?.avgAnswerTime || 0
	return avgTime.toFixed(1)
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

const viewLeaderboard = () => {
	const quizId = results.value?.quiz.id
	if (quizId) {
		router.push({ name: 'leaderboard', params: { quizId } })
	}
}

const goHome = () => {
	router.push({ name: 'categories' })
}
</script>

<template>
	<div class="mx-auto max-w-2xl">
		<!-- Loading -->
		<div v-if="isLoading" class="flex justify-center items-center py-12">
			<UProgress animation="carousel" />
			<span class="ml-4">{{ t('quiz.loadingResults') }}</span>
		</div>

		<!-- Error -->
		<div v-else-if="isError" class="text-center py-12">
			<UAlert
				color="red"
				:title="t('quiz.loadResultsFailed')"
				:description="error?.error?.message || t('quiz.tryAgain2')"
			/>
		</div>

		<!-- Results -->
		<div v-else-if="results">
			<!-- Main Result Card -->
			<UCard class="mb-6 text-center">
				<div class="py-8">
					<div class="text-8xl mb-4">{{ performanceEmoji }}</div>
					<h1 class="text-3xl font-bold mb-2">{{ t('quiz.completed') }}</h1>
					<p class="text-xl text-(--ui-text-muted) mb-6">{{ performanceMessage }}</p>

					<!-- Score Circle -->
					<div class="flex justify-center mb-6">
						<div
							class="w-40 h-40 rounded-full border-8 flex items-center justify-center"
							:class="{
								'border-emerald-500 bg-emerald-50 dark:bg-emerald-950/30':
									results.passed,
								'border-rose-500 bg-rose-50 dark:bg-rose-950/30': !results.passed,
							}"
						>
							<div>
								<div class="text-5xl font-bold">{{ scorePercentage }}%</div>
								<div class="text-sm text-(--ui-text-muted)">
									{{ results.session.score }}/{{
										results.quiz.passingScore * results.totalQuestions
									}}
								</div>
							</div>
						</div>
					</div>

					<!-- Pass/Fail Badge -->
					<UBadge
						:color="results.passed ? 'green' : 'red'"
						variant="solid"
						size="lg"
						class="mb-4"
					>
						{{ results.passed ? t('quiz.passed') : t('quiz.notPassed') }}
					</UBadge>

					<!-- Quiz Title -->
					<p class="text-sm text-(--ui-text-dimmed) mt-4">{{ results.quiz.title }}</p>
				</div>
			</UCard>

			<!-- Stats Grid -->
			<div class="grid grid-cols-2 gap-4 mb-6">
				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">
							{{ results.correctAnswers }}/{{ results.totalQuestions }}
						</div>
						<div class="text-sm text-(--ui-text-muted)">
							{{ t('quiz.correctAnswers') }}
						</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ formatTime }}</div>
						<div class="text-sm text-(--ui-text-muted)">{{ t('quiz.timeSpent') }}</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ results.session.score }}</div>
						<div class="text-sm text-(--ui-text-muted)">
							{{ t('quiz.pointsEarnedLabel') }}
						</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ results.scorePercentage }}%</div>
						<div class="text-sm text-(--ui-text-muted)">{{ t('quiz.accuracy') }}</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2 text-orange-500">
							🔥 {{ results.longestStreak }}
						</div>
						<div class="text-sm text-(--ui-text-muted)">
							{{ t('quiz.longestStreak') }}
						</div>
					</div>
				</UCard>

				<UCard>
					<div class="text-center py-4">
						<div class="text-3xl font-bold mb-2">{{ avgAnswerTime }}s</div>
						<div class="text-sm text-(--ui-text-muted)">
							{{ t('quiz.avgResponse') }}
						</div>
					</div>
				</UCard>
			</div>

			<!-- Actions -->
			<div class="space-y-3">
				<UButton size="xl" color="primary" block @click="tryAgain">
					{{ t('quiz.tryAgainBtn') }}
				</UButton>
				<UButton size="xl" color="neutral" variant="outline" block @click="viewLeaderboard">
					{{ t('quiz.viewLeaderboard') }}
				</UButton>
				<UButton size="xl" color="neutral" variant="ghost" block @click="goHome">
					{{ t('quiz.backToHome') }}
				</UButton>
			</div>

			<!-- Share Section (Optional) -->
			<UCard class="mt-6">
				<div class="text-center py-4">
					<p class="text-sm text-(--ui-text-muted) mb-3">
						{{ t('quiz.shareAchievement') }}
					</p>
					<div class="flex justify-center gap-3">
						<UButton icon="i-heroicons-share" color="neutral" variant="outline">
							{{ t('quiz.share') }}
						</UButton>
					</div>
				</div>
			</UCard>
		</div>
	</div>
</template>
