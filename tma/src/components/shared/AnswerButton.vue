<script setup lang="ts">
import { computed } from 'vue'
import type { InternalInfrastructureHttpHandlersAnswerDTO } from '@/api/generated'

interface Props {
  answer: InternalInfrastructureHttpHandlersAnswerDTO
  selected?: boolean
  disabled?: boolean
  showFeedback?: boolean
  isCorrect?: boolean | null
  label?: string // A, B, C, D
}

const props = withDefaults(defineProps<Props>(), {
  selected: false,
  disabled: false,
  showFeedback: false,
  isCorrect: null
})

const emit = defineEmits<{
  click: [answerId: string]
}>()

/**
 * 4 feedback states per docs/game_modes/daily_challenge/02_gameplay.md:
 * 1. Correct answer → green bg + border, full opacity, checkmark
 * 2. Selected + wrong → red bg + border, full opacity, cross
 * 3. Not selected + not correct → muted (opacity-40)
 * 4. Selected + correct → green bg + border (same as #1)
 */
const buttonClasses = computed(() => {
  const base = 'w-full p-4 rounded-xl border-2 text-left transition-all duration-300'

  // Feedback mode
  if (props.showFeedback) {
    if (props.isCorrect === true) {
      // Correct answer → green
      return `${base} border-green-500 bg-green-500/20 dark:bg-green-500/15`
    }
    if (props.isCorrect === false) {
      // Selected wrong → red
      return `${base} border-red-500 bg-red-500/20 dark:bg-red-500/15`
    }
    // Not selected + not correct → muted
    return `${base} border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 opacity-40`
  }

  // Selected state (before submit) — более яркий синий для заметности
  if (props.selected) {
    return `${base} border-primary-500 bg-primary-500/20 dark:bg-primary-500/25 ring-2 ring-primary-500/30`
  }

  // Default — interactive
  if (props.disabled) {
    return `${base} border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 opacity-50 cursor-not-allowed`
  }

  return `${base} border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:border-primary-500 hover:-translate-y-0.5 hover:shadow-md active:translate-y-0 cursor-pointer`
})

const badgeColor = computed(() => {
  if (props.showFeedback && props.isCorrect === true) return 'green'
  if (props.showFeedback && props.isCorrect === false) return 'red'
  if (props.selected) return 'primary'
  return 'gray'
})

const feedbackIcon = computed(() => {
  if (!props.showFeedback || props.isCorrect === null) return null
  return props.isCorrect ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'
})

const feedbackIconColor = computed(() => {
  if (props.isCorrect === true) return 'text-green-500'
  if (props.isCorrect === false) return 'text-red-500'
  return ''
})

const handleClick = () => {
  if (!props.disabled) {
    emit('click', props.answer.id)
  }
}
</script>

<template>
  <button
    type="button"
    :class="buttonClasses"
    :disabled="disabled || showFeedback"
    @click="handleClick"
  >
    <div class="flex items-center gap-3">
      <!-- Label badge (A, B, C, D) -->
      <UBadge v-if="label" :color="badgeColor" size="lg" class="shrink-0">
        {{ label }}
      </UBadge>

      <!-- Answer text -->
      <span class="flex-1 text-base font-medium leading-snug text-gray-900 dark:text-gray-100">
        {{ answer.text }}
      </span>

      <!-- Feedback icon -->
      <UIcon
        v-if="feedbackIcon"
        :name="feedbackIcon"
        :class="['size-6 shrink-0', feedbackIconColor]"
      />
    </div>
  </button>
</template>
