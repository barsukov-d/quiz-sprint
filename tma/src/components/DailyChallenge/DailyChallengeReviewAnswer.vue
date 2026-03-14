<script setup lang="ts">
import { computed } from 'vue'
import type { InternalInfrastructureHttpHandlersAnsweredQuestionDTO } from '@/api/generated'
import { useI18n } from 'vue-i18n'

interface Props {
	answeredQuestion: InternalInfrastructureHttpHandlersAnsweredQuestionDTO
	questionNumber: number
	totalQuestions: number
}

const props = defineProps<Props>()
const { t } = useI18n()

// ===========================
// Computed
// ===========================

const resultBadge = computed(() => {
	if (props.answeredQuestion.isCorrect) {
		return {
			label: t('daily.correctLabel'),
			color: 'green' as const,
			icon: 'i-heroicons-check-circle',
		}
	}
	return {
		label: t('daily.wrongLabel'),
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
	<div
		class="w-full rounded-(--ui-radius) border border-(--ui-border) overflow-hidden"
		:class="
			answeredQuestion.isCorrect
				? 'border-l-4 border-l-green-500'
				: 'border-l-4 border-l-red-500'
		"
	>
		<!-- Header -->
		<div
			class="flex justify-between items-center px-4 py-2.5 bg-(--ui-bg-elevated) border-b border-(--ui-border)"
		>
			<span class="text-xs text-(--ui-text-dimmed)">
				{{ t('shared.questionOf', { current: questionNumber, total: totalQuestions }) }}
			</span>
			<div class="flex items-center gap-1.5">
				<UIcon
					:name="resultBadge.icon"
					class="w-4 h-4"
					:class="answeredQuestion.isCorrect ? 'text-green-500' : 'text-red-500'"
				/>
				<span
					class="text-xs font-semibold"
					:class="
						answeredQuestion.isCorrect
							? 'text-green-600 dark:text-green-400'
							: 'text-red-600 dark:text-red-400'
					"
					>{{ resultBadge.label }}</span
				>
			</div>
		</div>

		<div class="px-4 py-3 flex flex-col gap-3 bg-(--ui-bg)">
			<!-- Question -->
			<p class="text-sm font-semibold leading-snug text-(--ui-text-highlighted)">
				{{ answeredQuestion.questionText }}
			</p>

			<!-- Answers -->
			<div class="flex flex-col gap-2">
				<!-- Player answer -->
				<div
					class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm"
					:class="
						answeredQuestion.isCorrect
							? 'bg-green-500/10 border border-green-500/30 text-green-700 dark:text-green-400'
							: 'bg-red-500/10 border border-red-500/30 text-red-700 dark:text-red-400'
					"
				>
					<UIcon
						:name="
							answeredQuestion.isCorrect
								? 'i-heroicons-check-circle'
								: 'i-heroicons-x-circle'
						"
						class="w-4 h-4 shrink-0"
					/>
					<span class="font-medium">{{ answeredQuestion.playerAnswerText }}</span>
				</div>

				<!-- Correct answer (if wrong) -->
				<div
					v-if="!answeredQuestion.isCorrect"
					class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm bg-green-500/10 border border-green-500/30 text-green-700 dark:text-green-400"
				>
					<UIcon name="i-heroicons-check-circle" class="w-4 h-4 shrink-0" />
					<span class="font-medium">{{ answeredQuestion.correctAnswerText }}</span>
				</div>
			</div>

			<!-- Stats row -->
			<div class="flex items-center gap-2 pt-1 border-t border-(--ui-border)">
				<div class="flex items-center gap-1.5 text-xs text-(--ui-text-muted)">
					<UIcon name="i-heroicons-clock" class="w-3.5 h-3.5" />
					<span>{{ formattedTime }}</span>
				</div>
				<div
					class="flex items-center gap-1.5 text-xs ml-auto"
					:class="
						answeredQuestion.isCorrect
							? 'text-green-600 dark:text-green-400'
							: 'text-(--ui-text-dimmed)'
					"
				>
					<UIcon
						:name="
							answeredQuestion.isCorrect ? 'i-heroicons-star' : 'i-heroicons-x-mark'
						"
						class="w-3.5 h-3.5"
					/>
					<span>
						{{
							answeredQuestion.isCorrect
								? t('daily.pointsEarned', { points: answeredQuestion.pointsEarned })
								: t('daily.noPoints')
						}}
					</span>
				</div>
			</div>
		</div>
	</div>
</template>
