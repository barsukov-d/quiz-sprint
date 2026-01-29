<script setup lang="ts">
import type { InternalInfrastructureHttpHandlersQuestionDTO } from '@/api/generated'

interface Props {
  question: InternalInfrastructureHttpHandlersQuestionDTO
  questionNumber?: number
  totalQuestions?: number
  showBadge?: boolean
  points?: number
}

const props = withDefaults(defineProps<Props>(), {
  showBadge: true
})
</script>

<template>
  <UCard>
    <!-- Header with question number -->
    <template v-if="showBadge && questionNumber && totalQuestions" #header>
      <div class="flex items-center justify-between">
        <UBadge color="primary" variant="subtle">
          Question {{ questionNumber }} / {{ totalQuestions }}
        </UBadge>
        <UBadge v-if="points" color="yellow" variant="subtle">
          <UIcon name="i-heroicons-star" class="size-3" />
          {{ points }} pts
        </UBadge>
      </div>
    </template>

    <!-- Question Text -->
    <div class="question-content">
      <p class="question-text text-lg font-semibold leading-relaxed">
        {{ question.text }}
      </p>
    </div>
  </UCard>
</template>

<style scoped>
.question-content {
  padding: 1rem 0;
}

.question-text {
  color: rgb(var(--color-gray-900));
  min-height: 3rem;
}

@media (prefers-color-scheme: dark) {
  .question-text {
    color: rgb(var(--color-gray-100));
  }
}
</style>
