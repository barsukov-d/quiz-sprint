<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useI18n } from 'vue-i18n'

interface Props {
	playerId: string
}

const props = defineProps<Props>()
const router = useRouter()
const { t } = useI18n()

const { state, isPlaying, isGameOver, isLoading, progressToRecord, initialize } = useMarathon(
	props.playerId,
)

// ===========================
// Bonus Display Config
// ===========================

const bonusList = [
	{
		key: 'shield' as const,
		get label() { return t('daily.shieldName') },
		icon: 'i-heroicons-shield-check',
		color: 'text-blue-500',
		get description() { return t('daily.shieldDesc') },
	},
	{
		key: 'fiftyFifty' as const,
		get label() { return t('daily.fiftyfiftyName') },
		icon: 'i-heroicons-scissors',
		color: 'text-yellow-500',
		get description() { return t('daily.fiftyfiftyDesc') },
	},
	{
		key: 'skip' as const,
		get label() { return t('daily.skipName') },
		icon: 'i-heroicons-forward',
		color: 'text-green-500',
		get description() { return t('daily.skipDesc') },
	},
	{
		key: 'freeze' as const,
		get label() { return t('daily.freezeName') },
		icon: 'i-heroicons-clock',
		color: 'text-cyan-500',
		get description() { return t('daily.freezeDesc') },
	},
]

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
})
</script>

<template>
	<UCard>
		<!-- Header -->
		<template #header>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2.5">
					<span class="text-2xl">🏃</span>
					<h3 class="text-lg font-bold">{{ t('marathon.title') }}</h3>
				</div>
			</div>
		</template>

		<!-- ==================== -->
		<!-- IN PROGRESS STATE    -->
		<!-- ==================== -->
		<div v-if="isPlaying" class="space-y-4">
			<!-- Current game stats -->
			<div class="flex items-center justify-between">
				<div class="text-center flex-1">
					<p class="text-xs text-gray-500 dark:text-gray-400">{{ t('marathon.score') }}</p>
					<p class="text-2xl font-bold text-primary">{{ state.score }}</p>
				</div>
				<div class="w-px h-10 bg-gray-200 dark:bg-gray-700" />
				<div class="text-center flex-1">
					<p class="text-xs text-gray-500 dark:text-gray-400">{{ t('marathon.question') }}</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-gray-100">
						{{ state.totalQuestions }}
					</p>
				</div>
			</div>

			<!-- Progress to record (if has personal best) -->
			<div v-if="state.personalBest && state.personalBest > 0">
				<div class="flex justify-between text-xs mb-1">
					<span class="text-gray-500 dark:text-gray-400">
						{{ t('marathon.recordLabel', { score: state.score, best: state.personalBest }) }}
					</span>
					<span
						:class="
							progressToRecord >= 100
								? 'text-green-500 font-semibold'
								: 'text-gray-500 dark:text-gray-400'
						"
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
					:class="
						state.bonusInventory[b.key] > 0
							? 'bg-gray-50 dark:bg-gray-800'
							: 'bg-gray-50/50 dark:bg-gray-800/50 opacity-40'
					"
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
				<p class="text-sm text-gray-500 dark:text-gray-400">{{ t('marathon.gameOver') }}</p>
				<p class="text-3xl font-bold text-primary mt-1">{{ state.score }}</p>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
					{{ t('marathon.questionsAnswered', { count: state.totalQuestions }) }}
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
				<span class="text-sm text-gray-500 dark:text-gray-400">{{ t('marathon.personalBest') }}</span>
				<span class="font-bold text-yellow-500"> 🏆 {{ state.personalBest }} </span>
			</div>

			<!-- Bonuses available -->
			<div>
				<p class="text-xs text-gray-500 dark:text-gray-400 mb-2">{{ t('marathon.bonuses') }}</p>
				<div class="flex gap-2">
					<div
						v-for="b in bonusList"
						:key="b.key"
						class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs"
						:class="
							state.bonusInventory[b.key] > 0
								? 'bg-gray-50 dark:bg-gray-800'
								: 'bg-gray-50/50 dark:bg-gray-800/50 opacity-40'
						"
						:title="b.description"
					>
						<UIcon :name="b.icon" :class="b.color" class="w-4 h-4" />
						<span class="font-semibold text-gray-900 dark:text-gray-100">
							{{ state.bonusInventory[b.key] }}
						</span>
					</div>
				</div>
			</div>

			<!-- Rules hint (only when no personal best = likely new player) -->
			<div
				v-if="state.personalBest === null || state.personalBest === 0"
				class="text-xs text-gray-500 dark:text-gray-400 space-y-1"
			>
				<p>{{ t('marathon.energyRule1') }}</p>
				<p>{{ t('marathon.energyRule2') }}</p>
				<p>{{ t('marathon.difficultyRule') }}</p>
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
				{{ t('marathon.continueMarathon') }}
			</UButton>

			<!-- Game over: new run (primary) + view results (secondary) -->
			<div v-else-if="isGameOver" class="flex flex-col gap-2">
				<UButton
					icon="i-heroicons-arrow-path"
					color="primary"
					:loading="isLoading"
					block
					size="lg"
					@click="handleStart"
				>
					{{ t('marathon.newRun') }}
				</UButton>
				<UButton
					icon="i-heroicons-flag"
					color="neutral"
					variant="ghost"
					:loading="isLoading"
					block
					size="sm"
					@click="handleViewGameOver"
				>
					{{ t('marathon.viewResults') }}
				</UButton>
			</div>

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
				{{ t('marathon.startMarathon') }}
			</UButton>
		</template>
	</UCard>
</template>
