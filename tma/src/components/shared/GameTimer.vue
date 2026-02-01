<script setup lang="ts">
import { computed, watch } from 'vue'
import { useGameTimer } from '@/composables/useGameTimer'

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
	size: 'md',
})

const timer = useGameTimer({
	initialTime: props.initialTime,
	autoStart: props.autoStart,
	onTimeout: props.onTimeout,
	onTick: props.onTick,
	warningThreshold: props.warningThreshold,
})

const timerColorClass = computed(() => {
	if (timer.isExpired.value) return 'text-red-500'
	if (timer.isWarning.value) return 'text-orange-500'
	return 'text-green-500'
})

const progressColor = computed(() => {
	if (timer.isExpired.value) return 'error'
	if (timer.isWarning.value) return 'warning'
	return 'success'
})

const sizeClass = computed(() => {
	switch (props.size) {
		case 'sm':
			return 'text-sm'
		case 'lg':
			return 'text-2xl'
		default:
			return 'text-lg'
	}
})

const progressTimer = computed(() => timer.progress.value)

watch(
	() => props.initialTime,
	(newTime) => {
		timer.reset(newTime)
	},
)

defineExpose({
	start: timer.start,
	stop: timer.stop,
	pause: timer.pause,
	resume: timer.resume,
	reset: timer.reset,
	addTime: timer.addTime,
	remainingTime: timer.remainingTime,
	isRunning: timer.isRunning,
})
</script>

<template>
	<div class="flex flex-col gap-1.5">
		<!-- Timer digits -->
		<div class="flex items-center justify-center gap-1.5">
			<UIcon
				name="i-heroicons-clock"
				:class="[
					timerColorClass,
					timer.isWarning.value && !timer.isExpired.value ? 'animate-pulse' : '',
					size === 'sm' ? 'size-4' : size === 'lg' ? 'size-7' : 'size-5',
				]"
			/>
			<span
				:class="[
					sizeClass,
					timerColorClass,
					'font-mono font-bold tabular-nums',
					timer.isWarning.value && !timer.isExpired.value ? 'animate-pulse' : '',
				]"
			>
				{{ timer.formattedTime }}
			</span>
		</div>

		<!-- Progress bar -->
		<UProgress v-if="showProgress" v-model="progressTimer" :color="progressColor" size="xs" />

		<!-- Expired label -->
		<div
			v-if="timer.isExpired.value"
			class="flex items-center justify-center gap-1 text-red-500"
		>
			<UIcon name="i-heroicons-x-circle" class="size-4" />
			<span class="text-xs font-semibold">Time's up!</span>
		</div>
	</div>
</template>
