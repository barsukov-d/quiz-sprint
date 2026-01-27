<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyChallenge } from '@/composables/useDailyChallenge'
import { useAuth } from '@/composables/useAuth'
import GameTimer from '@/components/shared/GameTimer.vue'
import QuestionCard from '@/components/shared/QuestionCard.vue'
import AnswerButton from '@/components/shared/AnswerButton.vue'

// ===========================
// Auth & Router
// ===========================

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

// ===========================
// Daily Challenge Composable
// ===========================

const {
  currentQuestion,
  questionIndex,
  totalQuestions,
  timeLimit,
  isPlaying,
  isCompleted,
  submitAnswer,
  initialize
} = useDailyChallenge(playerId)

// ===========================
// Local State
// ===========================

const selectedAnswerId = ref<string | null>(null)
const isSubmitting = ref(false)
const showSubmittedFeedback = ref(false)
const timerRef = ref<InstanceType<typeof GameTimer> | null>(null)

// ===========================
// Computed
// ===========================

const answerLabels = ['A', 'B', 'C', 'D']

const canSubmit = computed(() => {
  return selectedAnswerId.value !== null && !isSubmitting.value && !showSubmittedFeedback.value
})

// ===========================
// Methods
// ===========================

const handleAnswerSelect = async (answerId: string) => {
  if (isSubmitting.value || showSubmittedFeedback.value || timerRef.value?.remainingTime.value === 0) return

  console.log('[Daily Challenge] Answer selected:', answerId)
  selectedAnswerId.value = answerId

  // Auto-submit immediately after selection
  await handleSubmit()
}

const handleSubmit = async () => {
  if (!canSubmit.value || !selectedAnswerId.value) return

  isSubmitting.value = true

  try {
    // Calculate time taken
    const timeTaken = timeLimit.value - (timerRef.value?.remainingTime.value || 0)

    // Submit answer
    await submitAnswer(selectedAnswerId.value, timeTaken)

    // Show "Answer submitted" feedback (no correctness)
    showSubmittedFeedback.value = true

    // Pause timer
    timerRef.value?.pause()

    // Wait 1.5 seconds before moving to next question
    setTimeout(() => {
      handleNextStep()
    }, 1500)
  } catch (error) {
    console.error('Failed to submit answer:', error)
    isSubmitting.value = false
  }
}

const handleTimeout = async () => {
  console.log('[Daily Challenge] Timer expired. Selected answer:', selectedAnswerId.value)

  // Auto-submit when timer expires
  if (selectedAnswerId.value) {
    console.log('[Daily Challenge] Auto-submitting selected answer...')
    await handleSubmit()
  } else {
    // If no answer selected, skip question (counts as wrong)
    console.log('[Daily Challenge] No answer selected, skipping question')

    // Show brief feedback that time expired
    showSubmittedFeedback.value = true

    // Wait 1 second before moving to next question
    setTimeout(() => {
      handleNextStep()
    }, 1000)
  }
}

const handleNextStep = () => {
  showSubmittedFeedback.value = false
  selectedAnswerId.value = null
  isSubmitting.value = false

  // Check if game is completed
  if (isCompleted.value) {
    // Navigate to results page
    router.push({ name: 'daily-challenge-results' })
  } else {
    // Reset timer for next question
    timerRef.value?.reset(timeLimit.value)
    timerRef.value?.start()
  }
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
  await initialize()

  // Check if game is in progress
  if (!isPlaying.value) {
    router.push({ name: 'home' })
    return
  }

  // Start timer
  if (timerRef.value) {
    timerRef.value.start()
  }
})

onUnmounted(() => {
  // Cleanup timer
  if (timerRef.value) {
    timerRef.value.stop()
  }
})
</script>

<template>
  <div class="daily-challenge-play">
    <!-- Loading State -->
    <div v-if="!currentQuestion" class="loading-container">
      <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
      <p class="text-gray-500 dark:text-gray-400 mt-4">Loading question...</p>
    </div>

    <!-- Game View -->
    <div v-else class="game-container">
      <!-- Header: Progress & Timer -->
      <div class="game-header">
        <div class="progress-info">
          <UBadge color="primary" size="lg">
            Question {{ questionIndex + 1 }} / {{ totalQuestions }}
          </UBadge>
          <UProgress
            :value="((questionIndex + 1) / totalQuestions) * 100"
            color="primary"
            class="mt-2"
          />
        </div>

        <GameTimer
          ref="timerRef"
          :initial-time="timeLimit"
          :auto-start="false"
          :warning-threshold="5"
          :show-progress="true"
          size="md"
          :on-timeout="handleTimeout"
        />
      </div>

      <!-- Question Card -->
      <div class="question-section">
        <QuestionCard
          :question="currentQuestion"
          :question-number="questionIndex + 1"
          :total-questions="totalQuestions"
          :show-badge="false"
        />
      </div>

      <!-- Answer Buttons -->
      <div class="answers-section">
        <AnswerButton
          v-for="(answer, index) in currentQuestion.answers"
          :key="answer.id"
          :answer="answer"
          :selected="selectedAnswerId === answer.id"
          :disabled="isSubmitting || showSubmittedFeedback || (timerRef?.remainingTime.value === 0)"
          :label="answerLabels[index]"
          @click="handleAnswerSelect"
        />
      </div>

      <!-- Status Section -->
      <div class="submit-section">
        <!-- Loading State -->
        <UAlert
          v-if="isSubmitting"
          color="gray"
          variant="soft"
          title="Submitting answer..."
          icon="i-heroicons-arrow-path"
        >
          <template #icon>
            <UIcon name="i-heroicons-arrow-path" class="animate-spin" />
          </template>
        </UAlert>

        <!-- Submitted Feedback -->
        <UAlert
          v-else-if="showSubmittedFeedback"
          color="blue"
          variant="solid"
          title="Answer Submitted!"
          description="Moving to next question..."
          icon="i-heroicons-check-circle"
        />

        <!-- Waiting for answer selection -->
        <UAlert
          v-else
          color="gray"
          variant="soft"
          title="Select your answer"
          description="Click on one of the options above"
          icon="i-heroicons-cursor-arrow-rays"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.daily-challenge-play {
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

.game-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.game-header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  background: rgb(var(--color-gray-50));
  border-radius: 0.75rem;
}

.progress-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.question-section {
  margin: 0.5rem 0;
}

.answers-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.submit-section {
  margin-top: 1rem;
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .game-header {
    background: rgb(var(--color-gray-800));
  }
}

/* Mobile optimizations */
@media (max-width: 640px) {
  .daily-challenge-play {
    padding: 0.75rem;
    padding-top: 5rem;
  }

  .game-header {
    padding: 0.75rem;
  }

  .answers-section {
    gap: 0.5rem;
  }
}
</style>
