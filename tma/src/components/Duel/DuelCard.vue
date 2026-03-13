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
	tickets,
	pendingChallenges,
	outgoingReadyChallenges,
	hasActiveDuel,
	mmr,
	leagueLabel,
	leagueIcon,
	seasonWins,
	seasonLosses,
	winRate,
	isLoading,
	canPlay,
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

const buttonColor = computed(() => {
	if (outgoingReadyChallenges.value.length > 0) return 'green'
	if (pendingChallenges.value.length > 0) return 'orange'
	return 'primary'
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
	<UCard>
		<!-- Header -->
		<template #header>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2.5">
					<UIcon name="i-heroicons-bolt" class="size-5 text-orange-500" />
					<h3 class="text-base font-semibold">{{ t('duel.title') }}</h3>
				</div>
				<UBadge v-if="totalAlerts > 0" color="orange" variant="soft" size="sm">
					{{ t('duel.pendingCount', { count: totalAlerts }) }}
				</UBadge>
			</div>
		</template>

		<!-- Body -->
		<div class="space-y-4">
			<!-- Player Rating -->
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<span class="text-2xl">{{ leagueIcon }}</span>
					<div>
						<p class="font-semibold">{{ leagueLabel }}</p>
						<p class="text-sm text-(--ui-text-dimmed)">
							{{ t('duel.mmrValue', { mmr }) }}
						</p>
					</div>
				</div>
				<div class="text-right">
					<p class="text-sm font-medium">
						<span class="text-green-600 dark:text-green-400"
							>{{ seasonWins }}{{ t('duel.wIndicator') }}</span
						>
						<span class="text-(--ui-text-dimmed) mx-1">/</span>
						<span class="text-red-600 dark:text-red-400"
							>{{ seasonLosses }}{{ t('duel.lIndicator') }}</span
						>
					</p>
					<p class="text-xs text-(--ui-text-dimmed)">
						{{ winRateFormatted }} {{ t('duel.winRate') }}
					</p>
				</div>
			</div>

			<!-- Meta Info -->
			<div class="grid grid-cols-2 gap-4 pt-3 border-t border-(--ui-border)">
				<div class="text-center">
					<p class="text-xs text-(--ui-text-dimmed) mb-1">
						{{ t('duel.tickets') }}
					</p>
					<p class="text-sm font-semibold">
						<UIcon name="i-heroicons-ticket" class="inline size-3.5 text-primary" />
						{{ tickets }}
					</p>
				</div>

				<div class="text-center">
					<p class="text-xs text-(--ui-text-dimmed) mb-1">
						{{ t('duel.thisSeason') }}
					</p>
					<p class="text-sm font-semibold">
						<UIcon name="i-heroicons-trophy" class="inline size-3.5 text-yellow-500" />
						{{ t('duel.totalGames', { count: totalGames }) }}
					</p>
				</div>
			</div>
		</div>

		<!-- Footer -->
		<template #footer>
			<UButton
				:icon="buttonIcon"
				:color="buttonColor"
				:loading="isLoading"
				:disabled="!canPlay && !hasActiveDuel && totalAlerts === 0"
				block
				size="lg"
				@click="handleClick"
			>
				{{ buttonText }}
			</UButton>
		</template>
	</UCard>
</template>
