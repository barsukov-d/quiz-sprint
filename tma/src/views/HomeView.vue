<script setup lang="ts">
import { computed } from 'vue'
import { useAuth } from '@/composables/useAuth'
import DailyChallengeCard from '@/components/DailyChallenge/DailyChallengeCard.vue'
import MarathonCard from '@/components/Marathon/MarathonCard.vue'
import DuelCard from '@/components/Duel/DuelCard.vue'
import GameModeCard from '@/components/shared/GameModeCard.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const { currentUser, isAuthenticated } = useAuth()

const playerId = currentUser.value?.id || 'guest'

const greeting = computed(() => {
	const hour = new Date().getHours()
	if (hour < 6) return t('home.greetingNight')
	if (hour < 12) return t('home.greetingMorning')
	if (hour < 18) return t('home.greetingDay')
	return t('home.greetingEvening')
})
</script>

<template>
	<div class="mx-auto max-w-[800px] space-y-5">
		<!-- Greeting -->
		<div v-if="isAuthenticated && currentUser" class="flex items-center gap-3">
			<UAvatar :src="currentUser.avatarUrl" :alt="currentUser.username" size="lg" />
			<div>
				<p class="text-lg font-bold text-(--ui-text-highlighted)">
					{{ greeting }}, {{ currentUser.username }}!
				</p>
				<p class="text-sm text-(--ui-text-dimmed)">{{ t('home.subtitle') }}</p>
			</div>
		</div>

		<!-- Daily Challenge — featured card -->
		<DailyChallengeCard :player-id="playerId" />

		<!-- Game Modes -->
		<div class="space-y-3">
			<MarathonCard :player-id="playerId" />
			<DuelCard :player-id="playerId" />

			<GameModeCard
				:title="t('home.partyMode')"
				icon="i-heroicons-user-group"
				:description="t('home.partyModeDesc')"
				:disabled="true"
				:badge="t('home.comingSoon')"
				badge-color="yellow"
			/>

			<GameModeCard
				:title="t('home.tournament')"
				icon="i-heroicons-trophy"
				:description="t('home.tournamentDesc')"
				:disabled="true"
				:badge="t('home.comingSoon')"
				badge-color="yellow"
			/>
		</div>
	</div>
</template>
