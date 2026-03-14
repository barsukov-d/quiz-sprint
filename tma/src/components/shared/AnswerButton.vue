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
	colorIndex?: number // 0=blue, 1=red, 2=orange, 3=green — for QuizPlay 2-col grid
	colorBar?: boolean // full-width colored bar style (Figma checkbox variant)
}

const props = withDefaults(defineProps<Props>(), {
	selected: false,
	disabled: false,
	showFeedback: false,
	isCorrect: null,
	colorIndex: undefined,
	colorBar: false,
})

const emit = defineEmits<{
	click: [answerId: string]
}>()

// Colors for the 4 answer slots in QuizPlay
const COLOR_MAP: Record<number, string> = {
	0: '#4A90D9',
	1: '#E74C3C',
	2: '#F39C12',
	3: '#2ECC71',
}

const isColorMode = computed(() => props.colorIndex !== undefined)
const isBarMode = computed(() => props.colorBar && props.colorIndex !== undefined)

/**
 * 4 feedback states per docs/game_modes/daily_challenge/02_gameplay.md:
 * 1. Correct answer → green bg + border, full opacity, checkmark
 * 2. Selected + wrong → red bg + border, full opacity, cross
 * 3. Not selected + not correct → muted (opacity-40)
 * 4. Selected + correct → green bg + border (same as #1)
 */
const buttonClasses = computed(() => {
	if (isBarMode.value) {
		// Full-width bar with uniform color, high-contrast text
		const base =
			'w-full rounded-xl py-4 px-5 text-left transition-all duration-300 flex items-center gap-3'
		if (props.showFeedback) {
			if (props.isCorrect === true)
				return `${base} bg-green-500 dark:bg-green-600 text-white opacity-100`
			if (props.isCorrect === false)
				return `${base} bg-red-500 dark:bg-red-600 text-white opacity-100`
			return `${base} bg-(--ui-bg-accented) text-(--ui-text-highlighted) opacity-30`
		}
		if (props.disabled)
			return `${base} bg-(--ui-bg-accented) text-(--ui-text-highlighted) cursor-not-allowed opacity-60`
		if (props.selected)
			return `${base} bg-primary-500 dark:bg-primary-600 text-white cursor-pointer ring-2 ring-primary-300 dark:ring-primary-400 scale-[1.02]`
		return `${base} bg-(--ui-bg-accented) text-(--ui-text-highlighted) cursor-pointer hover:bg-(--ui-bg-elevated) active:scale-[0.98]`
	}

	if (isColorMode.value) {
		// Large solid-color button for QuizPlay 2-column grid
		const base =
			'w-full rounded-2xl text-center transition-all duration-300 min-h-[120px] flex items-center justify-center'
		if (props.disabled || props.showFeedback) {
			return `${base} cursor-not-allowed opacity-70`
		}
		return `${base} cursor-pointer active:scale-95`
	}

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
		return `${base} border-(--ui-border) bg-(--ui-bg-elevated) opacity-40`
	}

	// Selected state (before submit) — более яркий синий для заметности
	if (props.selected) {
		return `${base} border-primary-500 bg-primary-500/20 dark:bg-primary-500/25 ring-2 ring-primary-500/30`
	}

	// Default — interactive
	if (props.disabled) {
		return `${base} border-(--ui-border) bg-(--ui-bg-elevated) opacity-50 cursor-not-allowed`
	}

	return `${base} border-(--ui-border) bg-(--ui-bg-elevated) hover:border-primary-500 hover:-translate-y-0.5 hover:shadow-md active:translate-y-0 cursor-pointer`
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

const colorStyle = computed(() => {
	if (!isColorMode.value || props.colorIndex === undefined) return {}
	const color = COLOR_MAP[props.colorIndex % 4]
	return { backgroundColor: color }
})

const handleClick = () => {
	if (!props.disabled) {
		emit('click', props.answer.id)
	}
}
</script>

<template>
	<!-- Bar mode: full-width uniform bar for all game modes -->
	<button
		v-if="isBarMode"
		type="button"
		:class="buttonClasses"
		:disabled="disabled || showFeedback"
		@click="handleClick"
	>
		<span class="flex-1 text-base font-semibold leading-snug">
			{{ answer.text }}
		</span>
		<!-- Selected checkmark -->
		<UIcon v-if="selected && !showFeedback" name="i-heroicons-check" class="size-5 shrink-0" />
		<!-- Feedback icons -->
		<UIcon
			v-if="showFeedback && isCorrect === true"
			name="i-heroicons-check-circle-solid"
			class="size-5 text-white shrink-0"
		/>
		<UIcon
			v-if="showFeedback && isCorrect === false"
			name="i-heroicons-x-circle-solid"
			class="size-5 text-white shrink-0"
		/>
	</button>

	<!-- Color mode: large solid-color button for QuizPlay 2-col grid -->
	<button
		v-else-if="isColorMode"
		type="button"
		:class="buttonClasses"
		:style="colorStyle"
		:disabled="disabled || showFeedback"
		@click="handleClick"
	>
		<span class="text-white text-lg font-bold text-center leading-snug px-3">
			{{ answer.text }}
		</span>
	</button>

	<!-- Default mode: existing behavior -->
	<button
		v-else
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
			<span class="flex-1 text-base font-medium leading-snug text-(--ui-text-highlighted)">
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
