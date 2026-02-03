<script setup lang="ts">
import { computed, onMounted, ref, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'

interface Props {
	playerId: string
}

const props = defineProps<Props>()
const router = useRouter()

const {
	state,
	isPlaying,
	isGameOver,
	isLoading,
	lives,
	progressToRecord,
	checkStatus,
	initialize,
} = useMarathon(props.playerId)

// ===========================
// Bonus Display Config
// ===========================

const bonusList = [
	{ key: 'shield' as const, label: 'Shield', icon: 'i-heroicons-shield-check', color: 'text-blue-500', description: 'Absorbs 1 wrong answer' },
	{ key: 'fiftyFifty' as const, label: '50/50', icon: 'i-heroicons-scissors', color: 'text-yellow-500', description: 'Removes 2 wrong answers' },
	{ key: 'skip' as const, label: 'Skip', icon: 'i-heroicons-forward', color: 'text-green-500', description: 'Skip without penalty' },
	{ key: 'freeze' as const, label: 'Freeze', icon: 'i-heroicons-clock', color: 'text-cyan-500', description: '+5 seconds to timer' },
]

// ===========================
// Life Restore Timer
// ===========================

const timerInterval = ref<number | null>(null)
const timeToLifeRestore = ref(0)

const timeToLifeRestoreFormatted = computed(() => {
	const seconds = timeToLifeRestore.value
	const hours = Math.floor(seconds / 3600)
	const minutes = Math.floor((seconds % 3600) / 60)
	const secs = seconds % 60

	if (hours > 0) {
		return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
	}
	return `${minutes}:${secs.toString().padStart(2, '0')}`
})

const startTimer = () => {
	if (timerInterval.value) return
	timeToLifeRestore.value = lives.value.timeToNextLife

	timerInterval.value = window.setInterval(() => {
		if (timeToLifeRestore.value > 0) {
			timeToLifeRestore.value--
		} else {
			checkStatus()
			if (timerInterval.value) {
				clearInterval(timerInterval.value)
				timerInterval.value = null
			}
		}
	}, 1000)
}

const stopTimer = () => {
	if (timerInterval.value) {
		clearInterval(timerInterval.value)
		timerInterval.value = null
	}
}

// ===========================
// Computed
// ===========================

const livesLabel = computed(() => {
	// Active or game over: show actual lives from server
	if (isPlaying.value || isGameOver.value) {
		return lives.value.label || '❤️'.repeat(lives.value.maxLives)
	}
	// Idle: show full lives (new game always starts with 3)
	return '❤️'.repeat(lives.value.maxLives)
})

const showLifeTimer = computed(
	() => lives.value.currentLives < lives.value.maxLives && timeToLifeRestore.value > 0,
)

// ===========================
// Actions
// ===========================

const handleStart = () => {
	if (isLoading.value) return
	router.push({ name: 'marathon-category' })
}

const handleResume = () => {
	if (isLoading.value) return
	router.push({ name: 'marathon-play' })
}

const handleViewGameOver = () => {
	if (isLoading.value) return
	router.push({ name: 'marathon-gameover' })
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	await initialize()
	startTimer()
})

onBeforeUnmount(() => {
	stopTimer()
})
</script>

<template>
	<UCard>
		<!-- Header -->
		<template #header>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2.5">
					<span class="text-2xl">🏃</span>
					<h3 class="text-lg font-bold">Marathon</h3>
				</div>
				<span class="text-lg" :title="`${lives.currentLives}/${lives.maxLives} lives`">
					{{ livesLabel }}
				</span>
			</div>
		</template>

		<!-- ==================== -->
		<!-- IN PROGRESS STATE    -->
		<!-- ==================== -->
		<div v-if="isPlaying" class="space-y-4">
			<!-- Current game stats -->
			<div class="flex items-center justify-between">
				<div class="text-center flex-1">
					<p class="text-xs text-gray-500 dark:text-gray-400">Score</p>
					<p class="text-2xl font-bold text-primary">{{ state.score }}</p>
				</div>
				<div class="w-px h-10 bg-gray-200 dark:bg-gray-700" />
				<div class="text-center flex-1">
					<p class="text-xs text-gray-500 dark:text-gray-400">Question</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-gray-100">
						{{ state.totalQuestions }}
					</p>
				</div>
			</div>

			<!-- Progress to record (if has personal best) -->
			<div v-if="state.personalBest && state.personalBest > 0">
				<div class="flex justify-between text-xs mb-1">
					<span class="text-gray-500 dark:text-gray-400">
						{{ state.score }}/{{ state.personalBest }} record
					</span>
					<span
						:class="progressToRecord >= 100
							? 'text-green-500 font-semibold'
							: 'text-gray-500 dark:text-gray-400'"
					>
						{{ progressToRecord }}%
					</span>
				</div>
				<UProgress
					v-model="progressToRecord"
					:color="progressToRecord >= 100 ? 'success' : 'primary'"
					size="xs"
				/>
			</div>

			<!-- Bonuses -->
			<div class="flex gap-2 justify-center">
				<div
					v-for="b in bonusList"
					:key="b.key"
					class="flex items-center gap-1 px-2 py-1 rounded-lg text-xs"
					:class="state.bonusInventory[b.key] > 0
						? 'bg-gray-50 dark:bg-gray-800'
						: 'bg-gray-50/50 dark:bg-gray-800/50 opacity-40'"
					:title="b.label"
				>
					<UIcon :name="b.icon" :class="b.color" class="w-4 h-4" />
					<span class="font-semibold text-gray-900 dark:text-gray-100">
						{{ state.bonusInventory[b.key] }}
					</span>
				</div>
			</div>
		</div>

		<!-- ==================== -->
		<!-- GAME OVER STATE      -->
		<!-- ==================== -->
		<div v-else-if="isGameOver" class="space-y-4">
			<div class="text-center">
				<p class="text-sm text-gray-500 dark:text-gray-400">Game Over</p>
				<p class="text-3xl font-bold text-primary mt-1">{{ state.score }}</p>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
					{{ state.totalQuestions }} questions answered
				</p>
			</div>
		</div>

		<!-- ==================== -->
		<!-- IDLE / READY STATE   -->
		<!-- ==================== -->
		<div v-else class="space-y-4">
			<!-- Personal Best -->
			<div
				v-if="state.personalBest !== null && state.personalBest > 0"
				class="flex items-center justify-between"
			>
				<span class="text-sm text-gray-500 dark:text-gray-400">Record</span>
				<span class="font-bold text-yellow-500">
					🏆 {{ state.personalBest }}
				</span>
			</div>

			<!-- Bonuses available -->
			<div>
				<p class="text-xs text-gray-500 dark:text-gray-400 mb-2">Bonuses</p>
				<div class="flex gap-2">
					<div
						v-for="b in bonusList"
						:key="b.key"
						class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs"
						:class="state.bonusInventory[b.key] > 0
							? 'bg-gray-50 dark:bg-gray-800'
							: 'bg-gray-50/50 dark:bg-gray-800/50 opacity-40'"
						:title="b.description"
					>
						<UIcon :name="b.icon" :class="b.color" class="w-4 h-4" />
						<span class="font-semibold text-gray-900 dark:text-gray-100">
							{{ state.bonusInventory[b.key] }}
						</span>
					</div>
				</div>
			</div>

			<!-- Life restore timer -->
			<div
				v-if="showLifeTimer"
				class="flex items-center gap-2 text-sm text-blue-500 dark:text-blue-400"
			>
				<UIcon name="i-heroicons-clock" class="size-4" />
				<span>Next life in <strong class="font-mono">{{ timeToLifeRestoreFormatted }}</strong></span>
			</div>

			<!-- Rules hint (only when no personal best = likely new player) -->
			<div
				v-if="state.personalBest === null || state.personalBest === 0"
				class="text-xs text-gray-500 dark:text-gray-400 space-y-1"
			>
				<p>3 lives, wrong answer = -1 life</p>
				<p>Difficulty increases over time</p>
				<p>Use bonuses strategically</p>
			</div>
		</div>

		<!-- Footer -->
		<template #footer>
			<!-- In progress: resume button -->
			<UButton
				v-if="isPlaying"
				icon="i-heroicons-play"
				color="primary"
				:loading="isLoading"
				block
				size="lg"
				@click="handleResume"
			>
				Continue Marathon
			</UButton>

			<!-- Game over: view results -->
			<UButton
				v-else-if="isGameOver"
				icon="i-heroicons-flag"
				color="red"
				:loading="isLoading"
				block
				size="lg"
				@click="handleViewGameOver"
			>
				View Results
			</UButton>

			<!-- Ready: start button (new game always starts with 3 lives) -->
			<UButton
				v-else
				icon="i-heroicons-bolt"
				color="primary"
				:loading="isLoading"
				block
				size="lg"
				@click="handleStart"
			>
				Start Marathon
			</UButton>
		</template>
	</UCard>
</template>
