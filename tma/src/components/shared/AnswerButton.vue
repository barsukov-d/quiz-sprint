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

// ===========================
// Computed
// ===========================

const buttonColor = computed(() => {
  // Feedback mode (для Marathon)
  if (props.showFeedback && props.isCorrect !== null) {
    return props.isCorrect ? 'green' : 'red'
  }

  // Selected state
  if (props.selected) {
    return 'primary'
  }

  // Default
  return 'gray'
})

const buttonVariant = computed(() => {
  if (props.showFeedback || props.selected) {
    return 'solid' as const
  }
  return 'outline' as const
})

const buttonIcon = computed(() => {
  if (props.showFeedback && props.isCorrect !== null) {
    return props.isCorrect ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'
  }
  return undefined
})

const buttonClass = computed(() => {
  const classes = ['answer-button']

  if (props.selected) {
    classes.push('answer-selected')
  }

  if (props.showFeedback) {
    classes.push('answer-feedback')
  }

  return classes.join(' ')
})

// ===========================
// Methods
// ===========================

const handleClick = () => {
  if (!props.disabled) {
    emit('click', props.answer.id)
  }
}
</script>

<template>
  <button
    type="button"
    :class="buttonClass"
    :disabled="disabled"
    @click="handleClick"
  >
    <div class="answer-content">
      <!-- Label (A, B, C, D) -->
      <div v-if="label" class="answer-label">
        <UBadge :color="buttonColor" size="lg">
          {{ label }}
        </UBadge>
      </div>

      <!-- Answer Text -->
      <div class="answer-text">
        {{ answer.text }}
      </div>

      <!-- Feedback Icon -->
      <div v-if="showFeedback && buttonIcon" class="answer-icon">
        <UIcon :name="buttonIcon" class="size-6" />
      </div>
    </div>
  </button>
</template>

<style scoped>
.answer-button {
  width: 100%;
  padding: 1rem;
  border-radius: 0.75rem;
  border: 2px solid rgb(var(--color-gray-300));
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.answer-button:hover:not(:disabled) {
  border-color: rgb(var(--color-primary-500));
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.answer-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.answer-button.answer-selected {
  border-color: rgb(var(--color-primary-500));
  background: rgb(var(--color-primary-50));
}

.answer-button.answer-feedback {
  pointer-events: none;
}

.answer-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.answer-label {
  flex-shrink: 0;
}

.answer-text {
  flex: 1;
  font-size: 1rem;
  font-weight: 500;
  color: rgb(var(--color-gray-900));
  line-height: 1.5;
}

.answer-icon {
  flex-shrink: 0;
}

/* Feedback states */
.answer-button.answer-feedback[data-correct="true"] {
  border-color: rgb(var(--color-green-500));
  background: rgb(var(--color-green-50));
}

.answer-button.answer-feedback[data-correct="false"] {
  border-color: rgb(var(--color-red-500));
  background: rgb(var(--color-red-50));
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .answer-button {
    background: rgb(var(--color-gray-800));
    border-color: rgb(var(--color-gray-600));
  }

  .answer-button:hover:not(:disabled) {
    border-color: rgb(var(--color-primary-400));
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .answer-button.answer-selected {
    background: rgb(var(--color-primary-900) / 0.3);
    border-color: rgb(var(--color-primary-400));
  }

  .answer-text {
    color: rgb(var(--color-gray-100));
  }

  .answer-button.answer-feedback[data-correct="true"] {
    background: rgb(var(--color-green-900) / 0.3);
    border-color: rgb(var(--color-green-400));
  }

  .answer-button.answer-feedback[data-correct="false"] {
    background: rgb(var(--color-red-900) / 0.3);
    border-color: rgb(var(--color-red-400));
  }
}
</style>
