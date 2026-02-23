<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import DailyChallengeCard from '@/components/DailyChallenge/DailyChallengeCard.vue'
import MarathonCard from '@/components/Marathon/MarathonCard.vue'
import DuelCard from '@/components/Duel/DuelCard.vue'
import GameModeCard from '@/components/shared/GameModeCard.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const { currentUser, isAuthenticated } = useAuth()

// Player ID для composables (из auth)
const playerId = currentUser.value?.id || 'guest'
</script>

<template>
	<div class="mx-auto max-w-[800px]">
		<!-- User Info (optional) -->
		<div
			v-if="isAuthenticated && currentUser"
			class="mb-6 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg"
		>
			<UUser
				:name="currentUser.username"
				:description="currentUser.telegramUsername"
				:avatar="{ src: currentUser.avatarUrl, alt: currentUser.username }"
				size="lg"
			/>
		</div>

		<!-- ========================================
         ZONE 1: Daily Challenge
         ======================================== -->
		<section class="mb-8">
			<h2 class="text-xl font-bold mb-4 flex items-center gap-2">
				<UIcon name="i-heroicons-calendar-days" class="size-6 text-primary" />
				{{ t('home.todaysChallenge') }}
			</h2>
			<DailyChallengeCard :player-id="playerId" />
		</section>

		<!-- ========================================
         ZONE 2: Game Modes
         ======================================== -->
		<section class="mb-8">
			<h2 class="text-xl font-bold mb-4 flex items-center gap-2">
				<UIcon name="i-heroicons-puzzle-piece" class="size-6 text-primary" />
				{{ t('home.gameModes') }}
			</h2>

			<div class="space-y-3">
				<!-- Solo Marathon -->
				<MarathonCard :player-id="playerId" />

				<!-- PvP Duel -->
				<DuelCard :player-id="playerId" />

				<!-- Coming Soon: Party Mode -->
				<GameModeCard
					:title="t('home.partyMode')"
					icon="i-heroicons-user-group"
					:description="t('home.partyModeDesc')"
					:disabled="true"
					:badge="t('home.comingSoon')"
					badge-color="yellow"
				/>

				<!-- Coming Soon: Tournament -->
				<GameModeCard
					:title="t('home.tournament')"
					icon="i-heroicons-trophy"
					:description="t('home.tournamentDesc')"
					:disabled="true"
					:badge="t('home.comingSoon')"
					badge-color="yellow"
				/>
			</div>
		</section>
	</div>
</template>
