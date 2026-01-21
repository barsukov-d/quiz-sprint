<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useGetQuizDaily, useGetQuizRandom, useGetCategories } from '@/api'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { currentUser, isAuthenticated } = useAuth()

// ========================================
// Zone 1: Daily Quiz
// ========================================
const {
	data: dailyQuiz,
	isLoading: isDailyLoading,
	isError: isDailyError,
	refetch: refetchDaily
} = useGetQuizDaily()

// ========================================
// Zone 2: Quick Actions
// ========================================
// Active sessions - TODO: implement when user stats is ready
// For now, we'll just have Random Quiz button
const hasActiveSessions = computed(() => false) // TODO: load from API

// Random quiz query (must be called at top level)
const { refetch: refetchRandomQuiz } = useGetQuizRandom()

// ========================================
// Zone 3: Categories
// ========================================
const { data: categoriesData, isLoading: isCategoriesLoading } = useGetCategories()

// Helper function to get category icon based on name
const getCategoryIcon = (categoryName: string): string => {
	const name = categoryName.toLowerCase()
	if (name.includes('general') || name.includes('knowledge')) return '🧠'
	if (name.includes('geography') || name.includes('world')) return '🌍'
	if (name.includes('technology') || name.includes('tech') || name.includes('it')) return '💻'
	if (name.includes('science')) return '🔬'
	if (name.includes('history')) return '📜'
	if (name.includes('sports')) return '⚽'
	if (name.includes('music')) return '🎵'
	if (name.includes('art')) return '🎨'
	if (name.includes('movie') || name.includes('film')) return '🎬'
	if (name.includes('literature') || name.includes('book')) return '📚'
	if (name.includes('nature') || name.includes('animal')) return '🌿'
	if (name.includes('food') || name.includes('cooking')) return '🍳'
	return '📚' // default icon
}

// ========================================
// Actions
// ========================================
const startDailyQuiz = () => {
	if (dailyQuiz.value?.data?.quiz?.id) {
		router.push(`/quiz/${dailyQuiz.value.data.quiz.id}`)
	}
}

const goToRandomQuiz = async () => {
	// Get random quiz and redirect
	const result = await refetchRandomQuiz()
	if (result?.data?.data?.quiz?.id) {
		router.push(`/quiz/${result.data.data.quiz.id}`)
	}
}

const goToCategory = (categoryId: string) => {
	router.push({ name: 'quizzes', query: { categoryId } })
}
</script>

<template>
	<div class="home-container">
		<!-- User Info (optional) -->
		<div v-if="isAuthenticated && currentUser" class="user-card">
			<UAvatar :src="currentUser.avatarUrl" :alt="currentUser.username" size="md" />
			<div class="user-info">
				<h3>{{ currentUser.username }}</h3>
				<p v-if="currentUser.telegramUsername">@{{ currentUser.telegramUsername }}</p>
			</div>
		</div>

		<!-- ========================================
		     ZONE 1: Daily Challenge 🌟
		     ======================================== -->
		<section class="daily-challenge">
			<!-- Loading state -->
			<div v-if="isDailyLoading" class="daily-card loading">
				<UProgress animation="carousel" />
				<span>Loading Daily Challenge...</span>
			</div>

			<!-- Error state -->
			<div v-else-if="isDailyError" class="daily-card error">
				<UAlert
					color="red"
					variant="soft"
					title="Failed to load Daily Quiz"
					description="Please try again"
				/>
				<UButton color="red" size="sm" @click="refetchDaily()"> Retry </UButton>
			</div>

			<!-- Not attempted - show call to action -->
			<div
				v-else-if="dailyQuiz?.data?.quiz && dailyQuiz.data.completionStatus === 'not_attempted'"
				class="daily-card"
			>
				<div class="daily-header">
					<span class="daily-icon">🌟</span>
					<h2 class="daily-title">Daily Challenge</h2>
				</div>

				<h3 class="quiz-title">{{ dailyQuiz.data.quiz.title }}</h3>

				<div class="quiz-meta">
					<span>{{ dailyQuiz.data.quiz.questions?.length || 0 }} questions</span>
					<span>•</span>
					<span>{{ Math.ceil((dailyQuiz.data.quiz.timeLimit || 0) / 60) }} min</span>
				</div>

				<div class="daily-motivation">
					<p class="bonus">+50% bonus points!</p>
					<!-- TODO: Add streak when user stats are ready -->
					<!-- <p class="streak">🔥 Streak: 3</p> -->
				</div>

				<UButton block color="primary" size="lg" @click="startDailyQuiz">
					Start Daily Quiz →
				</UButton>
			</div>

			<!-- Completed - show result -->
			<div
				v-else-if="dailyQuiz?.data?.quiz && dailyQuiz.data.completionStatus === 'completed'"
				class="daily-card completed"
			>
				<div class="daily-header">
					<span class="daily-icon">✅</span>
					<h2 class="daily-title">Daily Challenge Completed!</h2>
				</div>

				<h3 class="quiz-title">{{ dailyQuiz.data.quiz.title }}</h3>

				<div v-if="dailyQuiz.data.userResult" class="result-summary">
					<div class="result-item">
						<span class="result-label">Score</span>
						<span class="result-value">{{ dailyQuiz.data.userResult.score }} points 🏆</span>
					</div>
					<div class="result-item">
						<span class="result-label">Rank</span>
						<span class="result-value">#{{ dailyQuiz.data.userResult.rank }}</span>
					</div>
				</div>

				<p class="encouragement">Keep it going tomorrow!</p>

				<UButton
					block
					color="gray"
					variant="soft"
					size="lg"
					@click="router.push(`/quiz/${dailyQuiz.data.quiz.id}/leaderboard`)"
				>
					View Leaderboard
				</UButton>
			</div>
		</section>

		<!-- ========================================
		     ZONE 2: Quick Actions ⚡
		     ======================================== -->
		<section class="quick-actions">
			<h3 class="section-title">⚡ Quick Actions</h3>

			<!-- Active Sessions (if any) -->
			<div v-if="hasActiveSessions" class="action-card continue-card">
				<!-- TODO: Implement Continue Playing -->
				<div class="card-header">
					<span class="icon">▶️</span>
					<h4>Continue Playing</h4>
				</div>
				<!-- Session details here -->
			</div>

			<!-- Random Quiz -->
			<div class="action-card random-card">
				<div class="card-header">
					<span class="icon">🎲</span>
					<h4>Random Quiz</h4>
				</div>
				<p class="card-description">Surprise me!</p>
				<UButton block color="gray" @click="goToRandomQuiz"> Play Random → </UButton>
			</div>
		</section>

		<!-- ========================================
		     ZONE 3: Browse by Category 📚
		     ======================================== -->
		<section class="categories">
			<h3 class="section-title">📚 Browse by Category</h3>

			<!-- Loading state -->
			<div v-if="isCategoriesLoading" class="categories-grid">
				<UProgress animation="carousel" />
			</div>

			<!-- Categories list -->
			<div v-else-if="categoriesData?.data" class="categories-grid">
				<div
					v-for="category in categoriesData.data"
					:key="category.id"
					class="category-card"
					@click="goToCategory(category.id)"
				>
					<div class="category-info">
						<span class="category-icon">{{ getCategoryIcon(category.name) }}</span>
						<div>
							<h4 class="category-name">{{ category.name }}</h4>
							<p class="category-description">Explore quizzes</p>
						</div>
					</div>
					<span class="category-count">View →</span>
				</div>
			</div>
		</section>
	</div>
</template>

<style scoped>
.home-container {
	max-width: 800px;
	margin: 0 auto;
	padding: 1rem;
	padding-top: 6rem;
	padding-bottom: 2rem;
}

/* User Info Card */
.user-card {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: 1rem;
	background: var(--color-background-soft);
	border-radius: 12px;
	margin-bottom: 1.5rem;
}

.user-info h3 {
	font-size: 1.125rem;
	font-weight: 600;
	margin: 0;
}

.user-info p {
	font-size: 0.875rem;
	color: var(--color-text-secondary);
	margin: 0.25rem 0 0;
}

/* ========================================
   ZONE 1: Daily Challenge
   ======================================== */
.daily-challenge {
	margin-bottom: 2rem;
}

.daily-card {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: white;
	padding: 1.5rem;
	border-radius: 16px;
	box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.daily-card.loading,
.daily-card.error {
	background: var(--color-background-soft);
	color: var(--color-text);
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 1rem;
}

.daily-header {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-bottom: 0.75rem;
}

.daily-icon {
	font-size: 1.5rem;
}

.daily-title {
	font-size: 1.25rem;
	font-weight: 700;
	margin: 0;
}

.quiz-title {
	font-size: 1.5rem;
	font-weight: 600;
	margin: 0.5rem 0;
}

.quiz-meta {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	font-size: 0.875rem;
	opacity: 0.9;
	margin-bottom: 1rem;
}

.daily-motivation {
	margin-bottom: 1.5rem;
}

.bonus {
	font-size: 1rem;
	font-weight: 600;
	margin: 0;
}

/* Completed state */
.daily-card.completed {
	background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}

.result-summary {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
	margin: 1rem 0;
	padding: 1rem;
	background: rgba(255, 255, 255, 0.1);
	border-radius: 12px;
}

.result-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.result-label {
	font-size: 0.875rem;
	opacity: 0.9;
}

.result-value {
	font-size: 1.125rem;
	font-weight: 700;
}

.encouragement {
	font-size: 0.875rem;
	opacity: 0.9;
	margin: 0.5rem 0 1.5rem;
	text-align: center;
}

.streak {
	font-size: 0.875rem;
	margin: 0.25rem 0 0;
}

/* ========================================
   ZONE 2: Quick Actions
   ======================================== */
.quick-actions {
	margin-bottom: 2rem;
}

.section-title {
	font-size: 1.125rem;
	font-weight: 600;
	margin: 0 0 1rem;
	color: var(--color-text);
}

.action-card {
	background: var(--color-background-soft);
	padding: 1.25rem;
	border-radius: 12px;
	margin-bottom: 1rem;
	border: 1px solid var(--color-border);
}

.card-header {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-bottom: 0.5rem;
}

.card-header .icon {
	font-size: 1.25rem;
}

.card-header h4 {
	font-size: 1rem;
	font-weight: 600;
	margin: 0;
}

.card-description {
	font-size: 0.875rem;
	color: var(--color-text-secondary);
	margin: 0 0 1rem;
}

/* ========================================
   ZONE 3: Categories
   ======================================== */
.categories {
	margin-bottom: 2rem;
}

.categories-grid {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.category-card {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1.25rem;
	background: var(--color-background-soft);
	border-radius: 12px;
	border: 1px solid var(--color-border);
	cursor: pointer;
	transition: all 0.2s;
}

.category-card:hover {
	background: var(--color-background-mute);
	transform: translateY(-2px);
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.category-info {
	display: flex;
	align-items: center;
	gap: 1rem;
}

.category-icon {
	font-size: 2rem;
}

.category-name {
	font-size: 1rem;
	font-weight: 600;
	margin: 0 0 0.25rem;
}

.category-description {
	font-size: 0.875rem;
	color: var(--color-text-secondary);
	margin: 0;
}

.category-count {
	font-size: 0.875rem;
	color: var(--color-text-secondary);
	font-weight: 500;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
	.daily-card {
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
	}
}
</style>
