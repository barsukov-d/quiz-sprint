<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { InternalInfrastructureHttpHandlersQuestionDTO } from '@/api/generated'

const { t } = useI18n()

interface Props {
	question: InternalInfrastructureHttpHandlersQuestionDTO
	questionNumber?: number
	totalQuestions?: number
	showBadge?: boolean
	points?: number
}

withDefaults(defineProps<Props>(), {
	showBadge: true,
})
</script>

<template>
	<div class="bg-(--ui-bg-elevated) rounded-(--ui-radius) p-6">
		<!-- Optional badge header -->
		<div
			v-if="showBadge && questionNumber && totalQuestions"
			class="flex items-center justify-between mb-4"
		>
			<UBadge color="primary" variant="subtle">
				{{ t('shared.questionOf', { current: questionNumber, total: totalQuestions }) }}
			</UBadge>
			<UBadge v-if="points" color="yellow" variant="subtle">
				<UIcon name="i-heroicons-star" class="size-3" />
				{{ t('shared.pts', { pts: points }) }}
			</UBadge>
		</div>

		<!-- Question text — primary focus -->
		<p class="text-xl font-bold leading-relaxed text-(--ui-text-highlighted) text-center">
			{{ question.text }}
		</p>
	</div>
</template>
