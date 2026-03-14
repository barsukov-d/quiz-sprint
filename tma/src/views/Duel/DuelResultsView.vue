<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { useGetDuelGameGameid } from '@/api/generated/hooks/duelController/useGetDuelGameGameid'
import { usePostDuelGameGameidRematch } from '@/api/generated'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const { currentUser } = useAuth()

const duelId = computed(() => route.params.duelId as string)
const playerId = computed(() => currentUser.value?.id ?? '')
const { t } = useI18n()

const rematchMutation = usePostDuelGameGameidRematch()
const isRematchLoading = computed(() => rematchMutation.isPending.value)

const { data: gameResult } = useGetDuelGameGameid(
	{ gameId: duelId },
	computed(() => ({ playerId: playerId.value })),
)

// ===========================
// State
// ===========================

const rematchStatus = ref<'idle' | 'pending' | 'accepted' | 'declined'>('idle')
const rematchError = ref<string | null>(null)

// ===========================
// Computed
// ===========================

const gameData = computed(() => gameResult.value?.data)

const didWin = computed(() => gameData.value?.result === 'win')
const isDraw = computed(() => gameData.value?.result === 'draw')

const resultText = computed(() => {
	if (isDraw.value) return t('duel.draw')
	return didWin.value ? t('duel.victory') : t('duel.defeat')
})

const mmrChange = computed(() => gameData.value?.mmrChange ?? 0)
// ===========================
// Actions
// ===========================

const handleRematch = async () => {
	try {
		rematchStatus.value = 'pending'
		rematchError.value = null

		const response = await rematchMutation.mutateAsync({
			gameId: duelId.value,
			data: { playerId: playerId.value },
		})

		if (response?.data?.status === 'accepted' && response?.data?.gameId) {
			router.push({ name: 'duel-play', params: { duelId: response.data.gameId } })
		} else {
			router.push({ name: 'duel-lobby' })
		}
	} catch {
		router.push({ name: 'duel-lobby' })
	}
}

const handleBackToLobby = () => {
	router.push({ name: 'duel-lobby' })
}

const handleHome = () => {
	router.push({ name: 'home' })
}

const handleShare = () => {
	// TODO: Share victory card
	const text = didWin.value
		? `I just won a PvP Duel! ${gameData.value?.playerScore} - ${gameData.value?.opponentScore}`
		: `Just finished a PvP Duel: ${gameData.value?.playerScore} - ${gameData.value?.opponentScore}`

	if (navigator.share) {
		navigator.share({
			title: 'Quiz Sprint Duel',
			text,
		})
	}
}
</script>

<template>
	<div
		class="min-h-screen flex flex-col bg-gradient-to-br from-purple-600 via-purple-700 to-indigo-800"
	>
		<!-- Header -->
		<div class="px-4 py-3 flex items-center justify-between">
			<button class="p-2 -ml-2 text-white/70 hover:text-white" @click="handleHome">
				<UIcon name="i-heroicons-x-mark" class="size-6" />
			</button>
			<h1 class="text-lg font-semibold text-white">{{ t('duel.gameResult') }}</h1>
			<div class="w-10" />
		</div>

		<!-- Result Hero -->
		<div class="flex-1 flex flex-col items-center justify-center p-6">
			<!-- Result Icon -->
			<div class="mb-4 text-8xl">
				{{ isDraw ? '🤝' : didWin ? '🏆' : '💔' }}
			</div>

			<!-- Result Text -->
			<h2 class="text-4xl font-black text-white mb-2">{{ resultText }}</h2>

			<!-- MMR Change -->
			<p
				class="text-xl font-bold mb-8"
				:class="mmrChange >= 0 ? 'text-green-300' : 'text-red-300'"
			>
				{{ t('duel.mmrChange', { sign: mmrChange >= 0 ? '+' : '', amount: mmrChange }) }}
			</p>

			<!-- Score comparison card -->
			<div class="w-full max-w-xs bg-white/10 backdrop-blur rounded-2xl p-5">
				<div class="flex items-center justify-between">
					<div class="text-center flex-1">
						<p class="text-sm text-white/60 mb-1">{{ t('duel.you') }}</p>
						<p class="text-5xl font-black text-white tabular-nums">
							{{ gameData?.playerScore ?? 0 }}
						</p>
					</div>
					<span class="text-2xl text-white/40 font-bold px-4">—</span>
					<div class="text-center flex-1">
						<p class="text-sm text-white/60 mb-1 truncate">
							{{ gameData?.opponent?.username ?? t('duel.opponent') }}
						</p>
						<p class="text-5xl font-black text-white/80 tabular-nums">
							{{ gameData?.opponentScore ?? 0 }}
						</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Actions -->
		<div class="p-6 space-y-4">
			<!-- Rematch Button -->
			<UButton
				v-if="rematchStatus === 'idle'"
				icon="i-heroicons-arrow-path"
				color="primary"
				size="xl"
				block
				:loading="isRematchLoading"
				@click="handleRematch"
			>
				{{ t('duel.requestRematch') }}
			</UButton>

			<div
				v-else-if="rematchStatus === 'pending'"
				class="flex items-center justify-center gap-2 py-3 text-white/70"
			>
				<UIcon name="i-heroicons-clock" class="size-5 animate-pulse" />
				<span class="text-sm">{{ t('duel.waitingOpponent') }}</span>
			</div>

			<p v-if="rematchError" class="text-center text-red-300 text-sm">{{ rematchError }}</p>

			<!-- Text link actions -->
			<div class="flex items-center justify-center gap-6 pt-1">
				<button
					class="text-sm text-white/60 hover:text-white flex items-center gap-1.5 transition-colors"
					@click="handleShare"
				>
					<UIcon name="i-heroicons-share" class="size-4" />
					{{ t('duel.shareResult') }}
				</button>
				<span class="text-white/20">|</span>
				<button
					class="text-sm text-white/60 hover:text-white flex items-center gap-1.5 transition-colors"
					@click="handleBackToLobby"
				>
					<UIcon name="i-heroicons-arrow-left" class="size-4" />
					{{ t('duel.backToLobby') }}
				</button>
			</div>
		</div>
	</div>
</template>
