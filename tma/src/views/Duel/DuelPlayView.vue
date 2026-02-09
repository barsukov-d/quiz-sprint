<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useDuelWebSocket } from '@/composables/useDuelWebSocket'
import { useGameTimer } from '@/composables/useGameTimer'
import GameTimer from '@/components/shared/GameTimer.vue'
import AnswerButton from '@/components/shared/AnswerButton.vue'
import QuestionCard from '@/components/shared/QuestionCard.vue'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()

const duelId = computed(() => route.params.duelId as string)
const playerId = computed(() => currentUser.value?.id ?? '')

// ===========================
// WebSocket
// ===========================

const {
  isConnected,
  isReconnecting,
  error,
  connect,
  match,
  currentQuestion,
  currentRound,
  countdownSeconds,
  opponentAnswered,
  lastRoundResult,
  myPlayer,
  opponent,
  myScore,
  opponentScore,
  isWaiting,
  isCountdown,
  isPlaying,
  isFinished,
  didWin,
  myMmrChange,
  sendAnswer,
} = useDuelWebSocket(duelId.value, playerId.value)

// ===========================
// Timer
// ===========================

const timer = useGameTimer({
  initialTime: 10,
  autoStart: false,
  onTimeout: () => handleTimeout(),
})

// ===========================
// UI State
// ===========================

const selectedAnswerId = ref<string | null>(null)
const showFeedback = ref(false)
const hasAnswered = ref(false)
const answerStartTime = ref(0)

// ===========================
// Computed
// ===========================

const totalRounds = computed(() => match.value?.totalRounds ?? 10)

const answerLabels = ['A', 'B', 'C', 'D']

// Cast question to expected DTO format
const formattedQuestion = computed(() => {
  if (!currentQuestion.value) return null
  return {
    id: currentQuestion.value.id,
    text: currentQuestion.value.text,
    answers: [],
    points: 100,
    position: currentRound.value,
  }
})

// Format answers for AnswerButton component
const formattedAnswers = computed(() => {
  if (!currentQuestion.value) return []
  return currentQuestion.value.answers.map((a, idx) => ({
    id: a.id,
    text: a.text,
    position: idx,
  }))
})

// ===========================
// Watch
// ===========================

// Watch for new round
watch(currentQuestion, (newQuestion) => {
  if (newQuestion) {
    // Reset UI state for new round
    selectedAnswerId.value = null
    showFeedback.value = false
    hasAnswered.value = false
    answerStartTime.value = Date.now()

    // Start timer
    timer.reset()
    timer.start()
  }
})

// Watch for round result
watch(lastRoundResult, (result) => {
  if (result) {
    showFeedback.value = true
    timer.stop()
  }
})

// Watch for match finished
watch(isFinished, (finished) => {
  if (finished) {
    timer.stop()
    // Navigate to results after short delay
    setTimeout(() => {
      router.push({ name: 'duel-results', params: { duelId: duelId.value } })
    }, 2000)
  }
})

// ===========================
// Actions
// ===========================

const handleAnswerSelect = async (answerId: string) => {
  if (hasAnswered.value || showFeedback.value) return

  selectedAnswerId.value = answerId
  hasAnswered.value = true

  const timeTaken = Date.now() - answerStartTime.value

  // Send answer via WebSocket
  sendAnswer(answerId, timeTaken)
}

const handleTimeout = () => {
  if (!hasAnswered.value) {
    // Auto-submit empty answer on timeout
    hasAnswered.value = true
    sendAnswer('', 10000)
  }
}

const isCorrectAnswer = (answerId: string) => {
  if (!lastRoundResult.value) return false
  return answerId === lastRoundResult.value.correctAnswerId
}

// ===========================
// Lifecycle
// ===========================

onMounted(() => {
  connect()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex flex-col">
    <!-- Connection Status -->
    <div
      v-if="!isConnected || isReconnecting"
      class="bg-yellow-500 text-white text-center py-2 text-sm"
    >
      <UIcon name="i-heroicons-wifi" class="inline size-4 mr-1" />
      {{ isReconnecting ? 'Reconnecting...' : 'Connecting...' }}
    </div>

    <!-- Error -->
    <div v-if="error" class="bg-red-500 text-white text-center py-2 text-sm">
      {{ error }}
    </div>

    <!-- Waiting for Opponent -->
    <div v-if="isWaiting" class="flex-1 flex items-center justify-center p-4">
      <div class="text-center">
        <div class="animate-pulse mb-4">
          <UIcon name="i-heroicons-users" class="size-16 text-primary" />
        </div>
        <h2 class="text-xl font-bold mb-2">Waiting for opponent...</h2>
        <p class="text-gray-500">The duel will start when both players are ready</p>
      </div>
    </div>

    <!-- Countdown -->
    <div v-else-if="isCountdown" class="flex-1 flex items-center justify-center p-4">
      <div class="text-center">
        <p class="text-gray-500 mb-2">Get ready!</p>
        <p class="text-8xl font-bold text-primary animate-pulse">
          {{ countdownSeconds }}
        </p>
      </div>
    </div>

    <!-- Playing -->
    <template v-else-if="isPlaying && currentQuestion">
      <!-- Header: Players & Scores -->
      <div class="bg-white dark:bg-gray-800 shadow-sm px-4 py-3">
        <div class="flex items-center justify-between">
          <!-- My Player -->
          <div class="flex items-center gap-2">
            <div class="w-10 h-10 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center">
              <span class="text-lg">{{ myPlayer?.leagueIcon }}</span>
            </div>
            <div>
              <p class="font-medium text-sm">You</p>
              <p class="text-2xl font-bold text-primary">{{ myScore }}</p>
            </div>
          </div>

          <!-- VS / Round -->
          <div class="text-center">
            <p class="text-xs text-gray-500">Round</p>
            <p class="font-bold">{{ currentRound }}/{{ totalRounds }}</p>
          </div>

          <!-- Opponent -->
          <div class="flex items-center gap-2">
            <div class="text-right">
              <p class="font-medium text-sm">{{ opponent?.username }}</p>
              <p class="text-2xl font-bold text-orange-500">{{ opponentScore }}</p>
            </div>
            <div class="w-10 h-10 rounded-full bg-orange-100 dark:bg-orange-900 flex items-center justify-center relative">
              <span class="text-lg">{{ opponent?.leagueIcon }}</span>
              <!-- Answered indicator -->
              <div
                v-if="opponentAnswered"
                class="absolute -bottom-1 -right-1 w-4 h-4 bg-green-500 rounded-full flex items-center justify-center"
              >
                <UIcon name="i-heroicons-check" class="size-3 text-white" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Timer -->
      <div class="px-4 py-3">
        <GameTimer
          ref="timerRef"
          :initial-time="10"
          :auto-start="false"
          :warning-threshold="3"
          show-progress
          size="lg"
        />
      </div>

      <!-- Question -->
      <div class="flex-1 px-4 py-2">
        <QuestionCard
          :question="formattedQuestion!"
          :question-number="currentRound"
          :total-questions="totalRounds"
        />

        <!-- Answers -->
        <div class="mt-6 space-y-3">
          <AnswerButton
            v-for="(answer, index) in formattedAnswers"
            :key="answer.id"
            :answer="answer"
            :label="answerLabels[index]"
            :selected="selectedAnswerId === answer.id"
            :disabled="hasAnswered"
            :show-feedback="showFeedback"
            :is-correct="isCorrectAnswer(answer.id)"
            @click="handleAnswerSelect(answer.id)"
          />
        </div>
      </div>
    </template>

    <!-- Match Finished (brief summary before redirect) -->
    <div v-else-if="isFinished" class="flex-1 flex items-center justify-center p-4">
      <div class="text-center">
        <div class="mb-4">
          <UIcon
            :name="didWin ? 'i-heroicons-trophy' : 'i-heroicons-x-circle'"
            :class="didWin ? 'text-yellow-500' : 'text-red-500'"
            class="size-20"
          />
        </div>
        <h2 class="text-3xl font-bold mb-2">
          {{ didWin === null ? 'Draw!' : didWin ? 'Victory!' : 'Defeat' }}
        </h2>
        <p class="text-xl">
          {{ myScore }} - {{ opponentScore }}
        </p>
        <p
          :class="myMmrChange >= 0 ? 'text-green-600' : 'text-red-600'"
          class="text-lg font-semibold mt-2"
        >
          {{ myMmrChange >= 0 ? '+' : '' }}{{ myMmrChange }} MMR
        </p>
      </div>
    </div>
  </div>
</template>
