<script setup lang="ts">
import { computed } from 'vue'
import type { InternalInfrastructureHttpHandlersReviewAnswerDTO } from '@/api/generated'
import QuestionCard from '@/components/shared/QuestionCard.vue'
import AnswerButton from '@/components/shared/AnswerButton.vue'

interface Props {
  reviewAnswer: InternalInfrastructureHttpHandlersReviewAnswerDTO
  questionNumber: number
  totalQuestions: number
}

const props = defineProps<Props>()

// ===========================
// Computed
// ===========================

const answerLabels = ['A', 'B', 'C', 'D']

const isCorrect = computed(() => props.reviewAnswer.isCorrect)

const resultBadge = computed(() => {
  if (isCorrect.value) {
    return {
      label: 'Correct',
      color: 'green' as const,
      icon: 'i-heroicons-check-circle'
    }
  }
  return {
    label: 'Wrong',
    color: 'red' as const,
    icon: 'i-heroicons-x-circle'
  }
})

const findAnswerById = (answerId: string) => {
  return props.reviewAnswer.question.answers.find(a => a.id === answerId)
}
</script>

<template>
  <UCard class="review-answer-card">
    <!-- Header with result badge -->
    <template #header>
      <div class="review-header">
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

    <div class="review-content">
      <!-- Question -->
      <div class="question-section">
        <h3 class="question-text">{{ reviewAnswer.question.text }}</h3>
      </div>

      <!-- Answers -->
      <div class="answers-section">
        <AnswerButton
          v-for="(answer, index) in reviewAnswer.question.answers"
          :key="answer.id"
          :answer="answer"
          :selected="answer.id === reviewAnswer.playerAnswerId"
          :show-feedback="true"
          :is-correct="answer.id === reviewAnswer.correctAnswerId ? true : (answer.id === reviewAnswer.playerAnswerId && !isCorrect ? false : null)"
          :label="answerLabels[index]"
          :disabled="true"
        />
      </div>

      <!-- Explanation (if player was wrong) -->
      <div v-if="!isCorrect" class="explanation-section">
        <UAlert color="blue" variant="soft" icon="i-heroicons-light-bulb">
          <template #title>Correct Answer</template>
          <template #description>
            <p>
              The correct answer was:
              <strong class="ml-1">
                {{ findAnswerById(reviewAnswer.correctAnswerId)?.text }}
              </strong>
            </p>
          </template>
        </UAlert>
      </div>

      <!-- Points info -->
      <div class="points-section">
        <div class="points-earned" :class="{ 'no-points': !isCorrect }">
          <UIcon
            :name="isCorrect ? 'i-heroicons-star' : 'i-heroicons-x-mark'"
            class="size-5"
          />
          <span class="points-text">
            {{ isCorrect ? `+${reviewAnswer.pointsEarned} points` : 'No points earned' }}
          </span>
        </div>
      </div>
    </div>
  </UCard>
</template>

<style scoped>
.review-answer-card {
  width: 100%;
}

.review-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.review-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.question-section {
  padding: 0.5rem 0;
}

.question-text {
  font-size: 1.125rem;
  font-weight: 600;
  line-height: 1.6;
  color: rgb(var(--color-gray-900));
}

.answers-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.explanation-section {
  padding-top: 0.5rem;
}

.points-section {
  padding-top: 0.5rem;
  border-top: 1px solid rgb(var(--color-gray-200));
}

.points-earned {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border-radius: 0.5rem;
  background: rgb(var(--color-green-50));
  color: rgb(var(--color-green-700));
  font-weight: 600;
}

.points-earned.no-points {
  background: rgb(var(--color-gray-100));
  color: rgb(var(--color-gray-600));
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .question-text {
    color: rgb(var(--color-gray-100));
  }

  .points-section {
    border-top-color: rgb(var(--color-gray-700));
  }

  .points-earned {
    background: rgb(var(--color-green-900) / 0.3);
    color: rgb(var(--color-green-400));
  }

  .points-earned.no-points {
    background: rgb(var(--color-gray-800));
    color: rgb(var(--color-gray-400));
  }
}

/* Mobile optimizations */
@media (max-width: 640px) {
  .question-text {
    font-size: 1rem;
  }

  .answers-section {
    gap: 0.5rem;
  }
}
</style>
