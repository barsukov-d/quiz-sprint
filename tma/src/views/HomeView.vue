<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import DailyChallengeCard from '@/components/DailyChallenge/DailyChallengeCard.vue'
import MarathonCard from '@/components/Marathon/MarathonCard.vue'
import GameModeCard from '@/components/shared/GameModeCard.vue'

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
				Today's Challenge
			</h2>
			<DailyChallengeCard :player-id="playerId" />
		</section>

		<!-- ========================================
         ZONE 2: Game Modes
         ======================================== -->
		<section class="mb-8">
			<h2 class="text-xl font-bold mb-4 flex items-center gap-2">
				<UIcon name="i-heroicons-puzzle-piece" class="size-6 text-primary" />
				Game Modes
			</h2>

			<div class="space-y-3">
				<!-- Solo Marathon -->
				<MarathonCard :player-id="playerId" />

				<!-- Coming Soon: Quick Duel -->
				<GameModeCard
					title="Quick Duel"
					icon="i-heroicons-bolt"
					description="1v1 real-time battle against other players"
					:disabled="true"
					badge="Coming Soon"
					badge-color="yellow"
				/>

				<!-- Coming Soon: Party Mode -->
				<GameModeCard
					title="Party Mode"
					icon="i-heroicons-user-group"
					description="Multiplayer quiz party with friends"
					:disabled="true"
					badge="Coming Soon"
					badge-color="yellow"
				/>

				<!-- Coming Soon: Tournament -->
				<GameModeCard
					title="Tournament"
					icon="i-heroicons-trophy"
					description="Compete in weekly tournaments for prizes"
					:disabled="true"
					badge="Coming Soon"
					badge-color="yellow"
				/>
			</div>
		</section>
	</div>
</template>
