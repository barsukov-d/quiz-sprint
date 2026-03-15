<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePvPDuel } from '@/composables/usePvPDuel'
import { useI18n } from 'vue-i18n'

interface Props {
	playerId: string
}

const props = defineProps<Props>()
const router = useRouter()
const { t } = useI18n()

// ===========================
// Composables
// ===========================

const {
	pendingChallenges,
	outgoingReadyChallenges,
	hasActiveDuel,
	mmr,
	leagueLabel,
	leagueIcon,
	seasonWins,
	seasonLosses,
	winRate,
	goToActiveDuel,
	initialize,
} = usePvPDuel(props.playerId)

// ===========================
// Computed
// ===========================

const buttonText = computed(() => {
	if (hasActiveDuel.value) return t('duel.continueDuel')
	if (outgoingReadyChallenges.value.length > 0) return t('duel.friendReady')
	if (pendingChallenges.value.length > 0)
		return t('duel.challengesCount', { count: pendingChallenges.value.length })
	return t('duel.findOpponent')
})

const buttonIcon = computed(() => {
	if (hasActiveDuel.value) return 'i-heroicons-play'
	if (outgoingReadyChallenges.value.length > 0) return 'i-heroicons-bolt'
	if (pendingChallenges.value.length > 0) return 'i-heroicons-bell-alert'
	return 'i-heroicons-magnifying-glass'
})

const totalAlerts = computed(
	() => pendingChallenges.value.length + outgoingReadyChallenges.value.length,
)

const totalGames = computed(() => seasonWins.value + seasonLosses.value)

const winRateFormatted = computed(() => {
	if (totalGames.value === 0) return '0%'
	return `${Math.round(winRate.value)}%`
})

// ===========================
// Actions
// ===========================

const handleClick = () => {
	if (hasActiveDuel.value) {
		goToActiveDuel()
	} else {
		router.push({ name: 'duel-lobby' })
	}
}

// ===========================
// Lifecycle
// ===========================

onMounted(async () => {
	try {
		await initialize()
	} catch (error) {
		console.error('Failed to initialize PvP Duel:', error)
	}
})
</script>

<template>
	<div
		class="rounded-(--ui-radius) overflow-hidden bg-(--ui-bg-elevated) border border-(--ui-border) cursor-pointer transition-all hover:shadow-lg hover:scale-[1.01] active:scale-[0.99]"
		@click="handleClick"
	>
		<!-- Header row -->
		<div class="flex items-center justify-between px-4 pt-4 pb-3">
			<div class="flex items-center gap-3">
				<div
					class="flex items-center justify-center size-10 rounded-xl bg-red-500/15 dark:bg-red-400/15"
				>
					<UIcon name="i-heroicons-bolt" class="size-5 text-red-500" />
				</div>
				<div>
					<h3 class="text-base font-bold text-(--ui-text-highlighted)">
						{{ t('duel.title') }}
					</h3>
					<p class="text-xs text-(--ui-text-dimmed)">
						{{ t('duel.mmrValue', { mmr }) }}
					</p>
				</div>
			</div>
			<UBadge v-if="totalAlerts > 0" color="orange" variant="soft" size="sm">
				{{ t('duel.pendingCount', { count: totalAlerts }) }}
			</UBadge>
			<UBadge v-else-if="hasActiveDuel" color="blue" variant="subtle" size="sm">
				<UIcon name="i-heroicons-play-circle" class="size-3.5 mr-0.5" />
				Live
			</UBadge>
		</div>

		<!-- Body -->
		<div class="px-4 pb-3">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<span class="text-lg">{{ leagueIcon }}</span>
					<span class="font-semibold text-sm">{{ leagueLabel }}</span>
				</div>
				<div class="text-right text-sm">
					<span class="text-green-500 font-medium">{{ seasonWins }}W</span>
					<span class="text-(--ui-text-dimmed) mx-0.5">/</span>
					<span class="text-red-500 font-medium">{{ seasonLosses }}L</span>
					<span class="text-(--ui-text-dimmed) ml-1 text-xs"
						>({{ winRateFormatted }})</span
					>
				</div>
			</div>
		</div>

		<!-- Footer action -->
		<div class="px-4 pb-4 pt-2 border-t border-(--ui-border-muted)">
			<div
				class="flex items-center justify-center gap-2 text-sm font-semibold"
				:class="totalAlerts > 0 ? 'text-orange-500' : 'text-primary'"
			>
				<UIcon :name="buttonIcon" class="size-4" />
				<span>{{ buttonText }}</span>
			</div>
		</div>
	</div>
</template>
