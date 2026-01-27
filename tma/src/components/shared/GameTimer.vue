<script setup lang="ts">
import { computed, watch, onMounted } from 'vue'
import { useGameTimer, type GameTimerOptions } from '@/composables/useGameTimer'

interface Props {
  initialTime: number
  autoStart?: boolean
  onTimeout?: () => void
  onTick?: (remainingTime: number) => void
  warningThreshold?: number
  showProgress?: boolean
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  autoStart: true,
  warningThreshold: 5,
  showProgress: true,
  size: 'md'
})

// ===========================
// Timer Composable
// ===========================

const timer = useGameTimer({
  initialTime: props.initialTime,
  autoStart: props.autoStart,
  onTimeout: props.onTimeout,
  onTick: props.onTick,
  warningThreshold: props.warningThreshold
})

// ===========================
// Computed
// ===========================

const timerClass = computed(() => {
  if (timer.isExpired.value) return 'timer-expired'
  if (timer.isWarning.value) return 'timer-warning'
  return 'timer-normal'
})

const progressColor = computed(() => {
  if (timer.isExpired.value) return 'red'
  if (timer.isWarning.value) return 'orange'
  return 'green'
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return {
        text: 'text-sm',
        icon: 'size-4'
      }
    case 'lg':
      return {
        text: 'text-2xl',
        icon: 'size-8'
      }
    default: // md
      return {
        text: 'text-xl',
        icon: 'size-6'
      }
  }
})

// ===========================
// Watch for initialTime changes
// ===========================

watch(() => props.initialTime, (newTime) => {
  timer.reset(newTime)
})

// ===========================
// Expose timer methods
// ===========================

defineExpose({
  start: timer.start,
  stop: timer.stop,
  pause: timer.pause,
  resume: timer.resume,
  reset: timer.reset,
  addTime: timer.addTime,
  remainingTime: timer.remainingTime,
  isRunning: timer.isRunning
})
</script>

<template>
  <div class="game-timer" :class="timerClass">
    <!-- Timer Display -->
    <div class="timer-display" :class="sizeClasses.text">
      <UIcon
        name="i-heroicons-clock"
        :class="[sizeClasses.icon, timer.isWarning ? 'animate-pulse' : '']"
      />
      <span class="timer-value font-mono font-bold">
        {{ timer.formattedTime }}
      </span>
    </div>

    <!-- Progress Bar -->
    <UProgress
      v-if="showProgress"
      :value="timer.progress"
      :color="progressColor"
      :class="timer.isWarning ? 'animate-pulse' : ''"
    />

    <!-- Warning Message -->
    <div v-if="timer.isWarning && !timer.isExpired" class="timer-warning-text">
      <UIcon name="i-heroicons-exclamation-triangle" class="size-4" />
      <span class="text-xs">Time running out!</span>
    </div>

    <!-- Expired Message -->
    <div v-if="timer.isExpired" class="timer-expired-text">
      <UIcon name="i-heroicons-x-circle" class="size-4" />
      <span class="text-xs font-semibold">Time's up!</span>
    </div>
  </div>
</template>

<style scoped>
.game-timer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.timer-display {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.5rem;
  border-radius: 0.5rem;
  transition: all 0.2s;
}

.timer-normal .timer-display {
  background: rgb(var(--color-green-50));
  color: rgb(var(--color-green-700));
}

.timer-warning .timer-display {
  background: rgb(var(--color-orange-50));
  color: rgb(var(--color-orange-700));
  animation: pulse 1s ease-in-out infinite;
}

.timer-expired .timer-display {
  background: rgb(var(--color-red-50));
  color: rgb(var(--color-red-700));
}

.timer-warning-text,
.timer-expired-text {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.25rem;
}

.timer-warning-text {
  color: rgb(var(--color-orange-600));
}

.timer-expired-text {
  color: rgb(var(--color-red-600));
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .timer-normal .timer-display {
    background: rgb(var(--color-green-900) / 0.3);
    color: rgb(var(--color-green-400));
  }

  .timer-warning .timer-display {
    background: rgb(var(--color-orange-900) / 0.3);
    color: rgb(var(--color-orange-400));
  }

  .timer-expired .timer-display {
    background: rgb(var(--color-red-900) / 0.3);
    color: rgb(var(--color-red-400));
  }

  .timer-warning-text {
    color: rgb(var(--color-orange-400));
  }

  .timer-expired-text {
    color: rgb(var(--color-red-400));
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}
</style>
