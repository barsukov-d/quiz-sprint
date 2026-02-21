<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useMarathonSession } from '@/composables/useMarathonSession'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { currentUser } = useAuth()
const playerId = currentUser.value?.id || 'guest'

const {
  state,
  isLoading,
  canContinue,
  continueOffer,
  continueGame,
  reset,
  initialize,
} = useMarathon(playerId)

const session = useMarathonSession()
const resultRecorded = ref(false)

const gameOverResult = computed(() => state.value.gameOverResult)

const motivationalPrompt = computed(() =>
  session.getMotivationalPrompt(
    gameOverResult.value?.finalScore ?? state.value.score,
    state.value.personalBest,
  )
)

const handleContinueWithCoins = async () => {
  try {
    await continueGame('coins')
  } catch (error) {
    console.error('Failed to continue with coins:', error)
  }
}

const handleContinueWithAd = async () => {
  try {
    await continueGame('ad')
  } catch (error) {
    console.error('Failed to continue with ad:', error)
  }
}

const handleStartNewRun = () => {
  reset()
  router.push({ name: 'marathon-category' })
}

const handleBackToHome = () => {
  reset()
  session.resetSession()
  router.push({ name: 'home' })
}

onMounted(async () => {
  await initialize()

  if (!state.value.gameOverResult && state.value.status !== 'game-over') {
    router.push({ name: 'home' })
    return
  }

  if (!resultRecorded.value) {
    session.recordRunResult(
      gameOverResult.value?.finalScore ?? state.value.score,
      state.value.streakCount,
    )
    resultRecorded.value = true
  }
})
</script>

<template>
  <div class="min-h-screen mx-auto max-w-[800px] px-4 pt-14 pb-8 sm:px-3 sm:pt-12">
    <div class="flex flex-col items-center gap-6">
      <!-- Game Over Header -->
      <div class="text-center">
        <UIcon name="i-heroicons-trophy" class="size-16 text-yellow-500 mb-4" />
        <h1 class="text-2xl font-bold">Забег завершён</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">Марафон</p>
      </div>

      <!-- Score Card -->
      <UCard class="w-full">
        <div class="text-center space-y-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Правильных ответов</p>
            <p class="text-4xl font-bold text-primary">
              {{ gameOverResult?.finalScore ?? state.score }}
            </p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Вопросов</p>
              <p class="text-lg font-semibold">
                {{ gameOverResult?.totalQuestions ?? state.totalQuestions }}
              </p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Личный рекорд</p>
              <p class="text-lg font-semibold">
                <template v-if="gameOverResult?.isNewPersonalBest">
                  <span class="text-green-500">Новый рекорд! 🎉</span>
                </template>
                <template v-else>
                  {{ gameOverResult?.previousRecord ?? state.personalBest ?? '-' }}
                </template>
              </p>
            </div>
          </div>
        </div>
      </UCard>

      <!-- Continue Offer -->
      <UCard v-if="canContinue && continueOffer" class="w-full">
        <div class="text-center space-y-3">
          <h3 class="font-semibold">Продолжить текущий забег?</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Получи +1 ⚡ и продолжи с того же места
          </p>

          <div class="flex flex-col gap-2">
            <UButton
              color="primary"
              block
              size="lg"
              :loading="isLoading"
              icon="i-heroicons-currency-dollar"
              @click="handleContinueWithCoins"
            >
              Продолжить ({{ continueOffer.costCoins }} монет)
            </UButton>

            <UButton
              v-if="continueOffer.hasAd"
              color="gray"
              variant="soft"
              block
              size="lg"
              :loading="isLoading"
              icon="i-heroicons-play"
              @click="handleContinueWithAd"
            >
              Смотреть рекламу
            </UButton>
          </div>
        </div>
      </UCard>

      <!-- Session Stats (shown from 2nd run onward) -->
      <div
        v-if="session.runCount.value >= 2"
        class="w-full text-center text-sm text-gray-500 dark:text-gray-400"
      >
        {{ session.sessionLabel.value }}
      </div>

      <!-- Motivational Prompt -->
      <div class="w-full text-center text-sm font-medium text-primary">
        {{ motivationalPrompt }}
      </div>

      <!-- Actions -->
      <div class="w-full flex flex-col gap-2">
        <UButton
          color="primary"
          block
          size="lg"
          icon="i-heroicons-bolt"
          @click="handleStartNewRun"
        >
          ▶ Новый забег
        </UButton>

        <UButton
          color="gray"
          variant="ghost"
          block
          size="lg"
          icon="i-heroicons-home"
          @click="handleBackToHome"
        >
          На главную
        </UButton>
      </div>
    </div>
  </div>
</template>
