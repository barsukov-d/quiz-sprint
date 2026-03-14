<script setup lang="ts">
import { useGetQuizId } from '@/api'
import { useRoute, useRouter } from 'vue-router'
import { useLastQuiz } from '@/composables/useLastQuiz'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { saveLastQuizId } = useLastQuiz()
const { t } = useI18n()

// Получаем ID квиза из URL
const quizId = route.params.id as string

// Save this quiz as the last viewed
saveLastQuizId(quizId)

// Загружаем детали квиза
const { data: quizData, isLoading, isError, error } = useGetQuizId({ id: quizId })

// Топ-лидеры будут в quizData.data.topScores
// const { data: leaderboardData } = useGetQuizIdLeaderboard({ id: quizId })

// Начать квиз
const startQuiz = () => {
	router.push({
		name: 'quiz-play',
		params: { id: quizId },
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
	<div class="min-h-screen bg-(--ui-bg) flex flex-col">
		<!-- Top bar -->
		<div class="flex items-center justify-between px-4 pt-4 pb-2">
			<UButton
				icon="i-heroicons-x-mark"
				color="neutral"
				variant="ghost"
				size="md"
				@click="goBack"
			/>
			<div class="flex items-center gap-2">
				<UButton icon="i-heroicons-star" color="warning" variant="ghost" size="md" />
				<UButton
					icon="i-heroicons-ellipsis-horizontal"
					color="neutral"
					variant="ghost"
					size="md"
				/>
			</div>
		</div>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex-1 flex justify-center items-center py-12">
			<UProgress animation="carousel" class="w-32" />
			<span class="ml-4 text-(--ui-text-muted)">{{ t('quiz.loadingQuiz') }}</span>
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="px-4 pt-2">
			<UAlert
				color="red"
				variant="soft"
				:title="t('quiz.loadError')"
				:description="error?.error.message || t('quiz.loadFailed')"
			/>
		</div>

		<!-- Success state -->
		<template v-else-if="quizData?.data?.quiz">
			<!-- Cover area: emoji gradient placeholder -->
			<div
				class="mx-4 rounded-2xl overflow-hidden bg-gradient-to-br from-yellow-400 to-amber-500 h-48 flex items-center justify-center mb-4"
			>
				<span class="text-8xl">🧠</span>
			</div>

			<!-- Scrollable content -->
			<div class="flex-1 overflow-y-auto px-4 pb-32">
				<!-- Title -->
				<h1 class="text-2xl font-bold text-(--ui-text-highlighted) mb-4">
					{{ quizData.data.quiz.title }}
				</h1>

				<!-- Stats row: 4 columns -->
				<div class="grid grid-cols-4 gap-2 py-4 border-y border-(--ui-border) mb-4">
					<div class="flex flex-col items-center gap-1">
						<span class="text-xl font-bold text-(--ui-text-highlighted)">
							{{ quizData.data.quiz.questions?.length || 0 }}
						</span>
						<span class="text-xs text-(--ui-text-muted)">{{
							t('quiz.questions')
						}}</span>
					</div>
					<div class="flex flex-col items-center gap-1">
						<span class="text-xl font-bold text-(--ui-text-highlighted)">
							{{ formatTime(quizData.data.quiz.timeLimit || 0) }}
						</span>
						<span class="text-xs text-(--ui-text-muted)">{{
							t('quiz.timeLimitLabel')
						}}</span>
					</div>
					<div class="flex flex-col items-center gap-1">
						<span class="text-xl font-bold text-(--ui-text-highlighted)">
							{{ quizData.data.quiz.passingScore || 0 }}%
						</span>
						<span class="text-xs text-(--ui-text-muted)">{{
							t('quiz.passingScoreLabel')
						}}</span>
					</div>
					<div class="flex flex-col items-center gap-1">
						<span class="text-xl font-bold text-(--ui-text-highlighted)">—</span>
						<span class="text-xs text-(--ui-text-muted)">{{ t('quiz.category') }}</span>
					</div>
				</div>

				<!-- Author/category section -->
				<div class="flex items-center gap-3 mb-4">
					<div
						class="w-10 h-10 rounded-full bg-(--ui-bg-accented) flex items-center justify-center text-lg shrink-0"
					>
						🎓
					</div>
					<div class="flex-1 min-w-0">
						<div class="font-semibold text-(--ui-text-highlighted) truncate">
							{{ quizData.data.quiz.categoryId || t('quiz.defaultCategory') }}
						</div>
						<div class="text-sm text-(--ui-text-muted) truncate">
							{{ quizData.data.quiz.description }}
						</div>
					</div>
				</div>

				<!-- Description label + text -->
				<div v-if="quizData.data.quiz.description" class="mb-4">
					<h2 class="text-base font-bold text-(--ui-text-highlighted) mb-1">
						{{ t('quiz.description') }}
					</h2>
					<p class="text-sm text-(--ui-text-muted)">
						{{ quizData.data.quiz.description }}
					</p>
				</div>

				<!-- Top 3 Leaderboard -->
				<div
					v-if="quizData?.data?.topScores && quizData.data.topScores.length > 0"
					class="mb-4"
				>
					<h2 class="text-base font-bold text-(--ui-text-highlighted) mb-3">
						{{ t('quiz.topLeaders') }}
					</h2>
					<div class="space-y-2">
						<div
							v-for="(entry, index) in quizData.data.topScores.slice(0, 3)"
							:key="entry.userId"
							class="flex items-center justify-between p-3 rounded-xl bg-(--ui-bg-elevated) border border-(--ui-border)"
						>
							<div class="flex items-center gap-3">
								<span class="text-2xl">
									{{ index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉' }}
								</span>
								<div>
									<div class="font-semibold text-(--ui-text-highlighted) text-sm">
										{{ entry.username || t('quiz.anonymous') }}
									</div>
									<div class="text-xs text-(--ui-text-dimmed)">
										{{
											new Date(entry.completedAt * 1000).toLocaleDateString()
										}}
									</div>
								</div>
							</div>
							<div class="font-bold text-(--ui-text-highlighted)">
								{{ entry.score }}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Sticky bottom: Start Quiz button -->
			<div
				class="fixed bottom-0 left-0 right-0 px-4 pb-6 pt-3 bg-(--ui-bg)/90 backdrop-blur-sm border-t border-(--ui-border)"
			>
				<UButton
					size="xl"
					color="primary"
					block
					class="rounded-full font-bold"
					@click="startQuiz"
				>
					{{ t('quiz.startQuiz') }}
				</UButton>
			</div>
		</template>
	</div>
</template>

<style scoped></style>
