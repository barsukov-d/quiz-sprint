<script setup lang="ts">
import { computed } from 'vue'
import type { InternalInfrastructureHttpHandlersAnsweredQuestionDTO } from '@/api/generated'

interface Props {
	answeredQuestion: InternalInfrastructureHttpHandlersAnsweredQuestionDTO
	questionNumber: number
	totalQuestions: number
}

const props = defineProps<Props>()

// ===========================
// Computed
// ===========================

const resultBadge = computed(() => {
	if (props.answeredQuestion.isCorrect) {
		return {
			label: 'Correct',
			color: 'green' as const,
			icon: 'i-heroicons-check-circle',
		}
	}
	return {
		label: 'Wrong',
		color: 'red' as const,
		icon: 'i-heroicons-x-circle',
	}
})

const formattedTime = computed(() => {
	const ms = props.answeredQuestion.timeTaken
	const seconds = Math.floor(ms / 1000)
	return `${seconds}s`
})
</script>

<template>
	<UCard class="w-full">
		<!-- Header with result badge -->
		<template #header>
			<div class="flex justify-between items-center">
				<UBadge color="gray" variant="subtle">
					Question {{ questionNumber }} / {{ totalQuestions }}
				</UBadge>
				<UBadge
					:color="resultBadge.color"
					:icon="resultBadge.icon"
					variant="solid"
					size="lg"
				>
					{{ resultBadge.label }}
				</UBadge>
			</div>
		</template>

		<div class="flex flex-col gap-6">
			<!-- Question -->
			<div class="py-2">
				<h3 class="text-lg font-semibold leading-relaxed text-gray-900 dark:text-gray-100">
					{{ answeredQuestion.questionText }}
				</h3>
			</div>

			<!-- Your Answer -->
			<div class="flex flex-col gap-2">
				<div
					class="text-sm font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wider"
				>
					Your Answer:
				</div>
				<div
					class="flex items-center gap-3 p-4 rounded-lg border-2"
					:class="
						answeredQuestion.isCorrect
							? 'bg-green-50 dark:bg-green-900/30 border-green-500 dark:border-green-600 text-green-700 dark:text-green-400'
							: 'bg-red-50 dark:bg-red-900/30 border-red-500 dark:border-red-600 text-red-700 dark:text-red-400'
					"
				>
					<UIcon
						:name="
							answeredQuestion.isCorrect
								? 'i-heroicons-check-circle'
								: 'i-heroicons-x-circle'
						"
						class="text-xl flex-shrink-0"
					/>
					<span class="font-medium text-[15px]">{{
						answeredQuestion.playerAnswerText
					}}</span>
				</div>
			</div>

			<!-- Correct Answer (if wrong) -->
			<div v-if="!answeredQuestion.isCorrect" class="flex flex-col gap-2">
				<div
					class="text-sm font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wider"
				>
					Correct Answer:
				</div>
				<div
					class="flex items-center gap-3 p-4 rounded-lg border-2 bg-green-50 dark:bg-green-900/30 border-green-500 dark:border-green-600 text-green-700 dark:text-green-400"
				>
					<UIcon name="i-heroicons-check-circle" class="text-xl flex-shrink-0" />
					<span class="font-medium text-[15px]">{{
						answeredQuestion.correctAnswerText
					}}</span>
				</div>
			</div>

			<!-- Stats -->
			<div class="flex gap-4 pt-2 border-t border-gray-200 dark:border-gray-700">
				<div
					class="flex items-center gap-2 px-3 py-2 rounded-md bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 text-sm font-medium"
				>
					<UIcon name="i-heroicons-clock" class="w-4 h-4" />
					<span>{{ formattedTime }}</span>
				</div>
				<div
					class="flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium"
					:class="
						answeredQuestion.isCorrect
							? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400'
							: 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'
					"
				>
					<UIcon
						:name="
							answeredQuestion.isCorrect ? 'i-heroicons-star' : 'i-heroicons-x-mark'
						"
						class="w-4 h-4"
					/>
					<span>
						{{
							answeredQuestion.isCorrect
								? `+${answeredQuestion.pointsEarned} points`
								: 'No points'
						}}
					</span>
				</div>
			</div>
		</div>
	</UCard>
</template>
