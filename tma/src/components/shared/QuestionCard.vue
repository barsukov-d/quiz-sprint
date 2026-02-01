<script setup lang="ts">
import type { InternalInfrastructureHttpHandlersQuestionDTO } from '@/api/generated'

interface Props {
  question: InternalInfrastructureHttpHandlersQuestionDTO
  questionNumber?: number
  totalQuestions?: number
  showBadge?: boolean
  points?: number
}

withDefaults(defineProps<Props>(), {
  showBadge: true
})
</script>

<template>
  <div class="py-6">
    <!-- Optional badge header -->
    <div v-if="showBadge && questionNumber && totalQuestions" class="flex items-center justify-between mb-3">
      <UBadge color="primary" variant="subtle">
        Question {{ questionNumber }} / {{ totalQuestions }}
      </UBadge>
      <UBadge v-if="points" color="yellow" variant="subtle">
        <UIcon name="i-heroicons-star" class="size-3" />
        {{ points }} pts
      </UBadge>
    </div>

    <!-- Question text — primary focus -->
    <p class="text-xl sm:text-2xl font-semibold leading-relaxed text-gray-900 dark:text-gray-100">
      {{ question.text }}
    </p>
  </div>
</template>
