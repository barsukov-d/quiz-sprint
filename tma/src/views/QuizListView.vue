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
	<div class="mx-auto max-w-[800px]">
		<!-- Header -->
		<div class="flex items-center gap-3 px-4 py-4 border-b border-(--ui-border)">
			<UButton
				icon="i-heroicons-arrow-left"
				color="neutral"
				variant="ghost"
				size="md"
				@click="goBackToCategories"
			/>
			<h1 class="text-xl font-bold text-(--ui-text-highlighted)">
				{{ categoryName || t('quiz.title') }}
			</h1>
		</div>

		<!-- Loading state -->
		<div v-if="isLoading" class="flex justify-center items-center py-16">
			<UProgress animation="carousel" class="w-32" />
		</div>

		<!-- Error state -->
		<div v-else-if="isError" class="px-4 pt-6">
			<UAlert
				color="red"
				variant="soft"
				:title="t('quiz.loadError')"
				:description="error?.error.message || t('quiz.loadFailed')"
			/>
			<UButton
				color="red"
				class="mt-3"
				@click="
					() => {
						refetch()
					}
				"
			>
				{{ t('quiz.tryAgain') }}
			</UButton>
		</div>

		<!-- Success state — vertical list -->
		<div
			v-else-if="quizzes?.data && Array.isArray(quizzes.data) && quizzes.data.length > 0"
			class="flex flex-col gap-1 p-4"
		>
			<button
				v-for="(quiz, index) in quizzes.data"
				:key="quiz.id || index"
				class="flex items-center gap-4 p-3 rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border) text-left transition-transform active:scale-[0.98] hover:border-(--ui-border-accented)"
				@click="() => goToQuizDetails(quiz.id)"
			>
				<!-- Emoji placeholder / thumbnail -->
				<div
					class="relative shrink-0 w-20 h-16 rounded-lg bg-(--ui-bg-muted) flex items-center justify-center overflow-hidden"
				>
					<span class="text-3xl">📝</span>
					<div
						v-if="quiz.questionsCount"
						class="absolute bottom-1 right-1 bg-black/70 text-white text-[10px] font-bold px-1.5 py-0.5 rounded"
					>
						{{ quiz.questionsCount }} Qs
					</div>
				</div>

				<!-- Info -->
				<div class="flex-1 min-w-0">
					<h3 class="text-sm font-semibold text-(--ui-text-highlighted) truncate">
						{{ quiz.title || t('quiz.unnamed') }}
					</h3>
					<p
						v-if="quiz.description"
						class="text-xs text-(--ui-text-muted) mt-0.5 line-clamp-2"
					>
						{{ quiz.description }}
					</p>
					<div class="flex items-center gap-2 mt-1.5">
						<UBadge color="primary" variant="soft" size="xs">
							{{ t('quiz.questionsCount', { count: quiz.questionsCount || 0 }) }}
						</UBadge>
					</div>
				</div>

				<!-- Arrow -->
				<UIcon
					name="i-heroicons-chevron-right"
					class="size-4 text-(--ui-text-dimmed) shrink-0"
				/>
			</button>
		</div>

		<!-- Empty state -->
		<div v-else class="flex flex-col items-center py-16 text-(--ui-text-dimmed) px-4">
			<span class="text-5xl mb-4">📋</span>
			<p class="text-base font-semibold mb-1">{{ t('quiz.notFound') }}</p>
			<p class="text-sm text-center mb-6">{{ t('quiz.notFoundDesc') }}</p>
			<UButton @click="goBackToCategories">{{ t('quiz.backToCategories') }}</UButton>
		</div>
	</div>
</template>
