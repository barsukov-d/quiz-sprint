<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useAuth } from '@/composables/useAuth'
import DailyChallengeReviewAnswer from '@/components/DailyChallenge/DailyChallengeReviewAnswer.vue'

// ===========================
// Auth & Router
// ===========================

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

// ===========================
// Daily Challenge Composable
// ===========================

const { results, isCompleted, initialize } = useDailyChallenge(playerId)

// ===========================
// Computed
// ===========================

const reviewAnswers = computed(() => {
	return results.value?.answeredQuestions || []
})

const correctCount = computed(() => {
	return reviewAnswers.value.filter((a) => a.isCorrect).length
})

const wrongCount = computed(() => {
	return reviewAnswers.value.length - correctCount.value
})

// ===========================
// Methods
// ===========================

const handleBackToResults = () => {
	router.push({ name: 'daily-challenge-results' })
}

const handleGoHome = () => {
	router.push({ name: 'home' })
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	await initialize()

	// Redirect if game is not completed
	if (!isCompleted.value || !results.value) {
		router.push({ name: 'home' })
	}
})
</script>

<template>
	<div class="review-container">
		<!-- Loading State -->
		<div v-if="!results" class="loading-container">
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-gray-500 dark:text-gray-400 mt-4">Loading review...</p>
		</div>

		<!-- Review View -->
		<div v-else class="review-content">
			<!-- Header -->
			<div class="review-header-section">
				<div class="header-title">
					<UIcon name="i-heroicons-document-text" class="size-7 text-primary" />
					<h1 class="text-2xl font-bold">Answer Review</h1>
				</div>
				<p class="header-subtitle">Review your answers from today's challenge</p>

				<!-- Summary Stats -->
				<div class="summary-stats">
					<div class="stat-item correct">
						<UIcon name="i-heroicons-check-circle" class="size-5" />
						<span class="stat-value">{{ correctCount }}</span>
						<span class="stat-label">Correct</span>
					</div>
					<div class="stat-item wrong">
						<UIcon name="i-heroicons-x-circle" class="size-5" />
						<span class="stat-value">{{ wrongCount }}</span>
						<span class="stat-label">Wrong</span>
					</div>
				</div>
			</div>

			<!-- Review Answers List -->
			<div class="review-list">
				<DailyChallengeReviewAnswer
					v-for="(answeredQuestion, index) in reviewAnswers"
					:key="answeredQuestion.questionId"
					:answered-question="answeredQuestion"
					:question-number="index + 1"
					:total-questions="reviewAnswers.length"
				/>
			</div>

			<!-- Action Buttons -->
			<div class="actions">
				<UButton
					color="primary"
					size="xl"
					icon="i-heroicons-chart-bar"
					variant="outline"
					block
					@click="handleBackToResults"
				>
					Back to Results
				</UButton>

				<UButton
					color="gray"
					size="xl"
					icon="i-heroicons-home"
					variant="outline"
					block
					@click="handleGoHome"
				>
					Back to Home
				</UButton>
			</div>
		</div>
	</div>
</template>

<style scoped>
.review-container {
	min-height: 100vh;
	padding: 1rem;
	padding-top: 6rem;
	padding-bottom: 2rem;
	max-width: 800px;
	margin: 0 auto;
}

.loading-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-height: 50vh;
}

.review-content {
	display: flex;
	flex-direction: column;
	gap: 1.5rem;
}

/* Header Section */
.review-header-section {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	padding: 1.5rem;
	background: rgb(var(--color-gray-50));
	border-radius: 0.75rem;
}

.header-title {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.header-subtitle {
	color: rgb(var(--color-gray-600));
	font-size: 0.875rem;
}

.summary-stats {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 1rem;
	margin-top: 0.5rem;
}

.stat-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.25rem;
	padding: 1rem;
	border-radius: 0.5rem;
	background: white;
}

.stat-item.correct {
	color: rgb(var(--color-green-600));
}

.stat-item.wrong {
	color: rgb(var(--color-red-600));
}

.stat-value {
	font-size: 1.5rem;
	font-weight: 700;
}

.stat-label {
	font-size: 0.75rem;
	text-transform: uppercase;
	letter-spacing: 0.05em;
	color: rgb(var(--color-gray-500));
}

/* Review List */
.review-list {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

/* Actions */
.actions {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
	margin-top: 1rem;
	padding-top: 1.5rem;
	border-top: 2px solid rgb(var(--color-gray-200));
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
	.review-header-section {
		background: rgb(var(--color-gray-800));
	}

	.header-subtitle {
		color: rgb(var(--color-gray-400));
	}

	.stat-item {
		background: rgb(var(--color-gray-900));
	}

	.stat-item.correct {
		color: rgb(var(--color-green-400));
	}

	.stat-item.wrong {
		color: rgb(var(--color-red-400));
	}

	.actions {
		border-top-color: rgb(var(--color-gray-700));
	}
}

/* Mobile optimizations */
@media (max-width: 640px) {
	.review-container {
		padding: 0.75rem;
		padding-top: 5rem;
	}

	.review-header-section {
		padding: 1rem;
	}

	.header-title {
		font-size: 1.25rem;
	}

	.summary-stats {
		gap: 0.75rem;
	}

	.stat-item {
		padding: 0.75rem;
	}

	.stat-value {
		font-size: 1.25rem;
	}
}
</style>
