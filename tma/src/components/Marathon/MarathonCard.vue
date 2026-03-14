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
		get label() {
			return t('daily.shieldName')
		},
		icon: 'i-heroicons-shield-check',
		color: 'text-blue-500',
		get description() {
			return t('daily.shieldDesc')
		},
	},
	{
		key: 'fiftyFifty' as const,
		get label() {
			return t('daily.fiftyfiftyName')
		},
		icon: 'i-heroicons-scissors',
		color: 'text-yellow-500',
		get description() {
			return t('daily.fiftyfiftyDesc')
		},
	},
	{
		key: 'skip' as const,
		get label() {
			return t('daily.skipName')
		},
		icon: 'i-heroicons-forward',
		color: 'text-green-500',
		get description() {
			return t('daily.skipDesc')
		},
	},
	{
		key: 'freeze' as const,
		get label() {
			return t('daily.freezeName')
		},
		icon: 'i-heroicons-clock',
		color: 'text-cyan-500',
		get description() {
			return t('daily.freezeDesc')
		},
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

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	await initialize()
})
</script>

<template>
	<div
		class="rounded-(--ui-radius) overflow-hidden bg-(--ui-bg-elevated) border border-(--ui-border) transition-all hover:shadow-lg hover:scale-[1.01] active:scale-[0.99]"
	>
		<!-- Header row -->
		<div class="flex items-center justify-between px-4 pt-4 pb-3">
			<div class="flex items-center gap-3">
				<div
					class="flex items-center justify-center size-10 rounded-xl bg-orange-500/15 dark:bg-orange-400/15"
				>
					<span class="text-xl">🏃</span>
				</div>
				<div>
					<h3 class="text-base font-bold text-(--ui-text-highlighted)">
						{{ t('marathon.title') }}
					</h3>
					<p
						v-if="state.personalBest !== null && state.personalBest > 0"
						class="text-xs text-(--ui-text-dimmed)"
					>
						🏆 {{ t('marathon.personalBest') }}: {{ state.personalBest }}
					</p>
				</div>
			</div>
			<UBadge v-if="isPlaying" color="blue" variant="subtle" size="sm">
				<UIcon name="i-heroicons-play-circle" class="size-3.5 mr-0.5" />
				{{ state.score }} pts
			</UBadge>
			<UBadge v-else-if="isGameOver" color="red" variant="subtle" size="sm">
				{{ t('marathon.gameOver') }}
			</UBadge>
		</div>

		<!-- Body -->
		<div class="px-4 pb-3">
			<!-- In Progress: compact stats + bonuses -->
			<div v-if="isPlaying">
				<div class="flex items-center justify-between mb-2">
					<span class="text-sm text-(--ui-text-muted)">
						{{ t('marathon.question') }} #{{ state.totalQuestions }}
					</span>
					<div v-if="state.personalBest && state.personalBest > 0" class="text-right">
						<span
							class="text-xs"
							:class="
								progressToRecord >= 100
									? 'text-green-500 font-semibold'
									: 'text-(--ui-text-dimmed)'
							"
						>
							{{ progressToRecord }}% {{ t('marathon.personalBest').toLowerCase() }}
						</span>
					</div>
				</div>
				<!-- Bonuses row -->
				<div class="flex gap-2">
					<div
						v-for="b in bonusList"
						:key="b.key"
						class="flex items-center gap-1 px-2 py-1 rounded-lg text-xs"
						:class="
							state.bonusInventory[b.key] > 0
								? 'bg-(--ui-bg-muted)'
								: 'bg-(--ui-bg-muted) opacity-40'
						"
						:title="b.label"
					>
						<UIcon :name="b.icon" :class="b.color" class="w-4 h-4" />
						<span class="font-semibold text-(--ui-text-highlighted)">
							{{ state.bonusInventory[b.key] }}
						</span>
					</div>
				</div>
			</div>

			<!-- Game Over: score summary -->
			<div v-else-if="isGameOver" class="flex items-center justify-between">
				<div>
					<p class="text-2xl font-black text-primary">{{ state.score }}</p>
					<p class="text-xs text-(--ui-text-dimmed)">
						{{ t('marathon.questionsAnswered', { count: state.totalQuestions }) }}
					</p>
				</div>
			</div>

			<!-- Idle: bonuses preview -->
			<div v-else>
				<div class="flex gap-2">
					<div
						v-for="b in bonusList"
						:key="b.key"
						class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs bg-(--ui-bg-muted)"
						:title="b.description"
					>
						<UIcon :name="b.icon" :class="b.color" class="w-4 h-4" />
						<span class="font-semibold text-(--ui-text-highlighted)">
							{{ state.bonusInventory[b.key] }}
						</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Footer action -->
		<div
			class="px-4 pb-4 pt-2 border-t border-(--ui-border-muted) cursor-pointer"
			@click="isPlaying ? handleResume() : isGameOver ? handleStart() : handleStart()"
		>
			<div
				class="flex items-center justify-center gap-2 text-sm font-semibold"
				:class="isGameOver ? 'text-orange-500' : 'text-primary'"
			>
				<UIcon
					:name="
						isPlaying
							? 'i-heroicons-play'
							: isGameOver
								? 'i-heroicons-arrow-path'
								: 'i-heroicons-bolt'
					"
					class="size-4"
				/>
				<span>{{
					isPlaying
						? t('marathon.continueMarathon')
						: isGameOver
							? t('marathon.newRun')
							: t('marathon.startMarathon')
				}}</span>
			</div>
		</div>
	</div>
</template>
