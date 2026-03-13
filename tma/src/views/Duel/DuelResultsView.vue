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

const resultIcon = computed(() => {
	if (isDraw.value) return 'i-heroicons-minus-circle'
	return didWin.value ? 'i-heroicons-trophy' : 'i-heroicons-x-circle'
})

const resultColor = computed(() => {
	if (isDraw.value) return 'text-gray-500 dark:text-gray-400'
	return didWin.value ? 'text-yellow-500' : 'text-red-500'
})

const resultText = computed(() => {
	if (isDraw.value) return t('duel.draw')
	return didWin.value ? t('duel.victory') : t('duel.defeat')
})

const mmrChange = computed(() => gameData.value?.mmrChange ?? 0)
const mmrChangeColor = computed(() =>
	mmrChange.value >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400',
)

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
	<div class="min-h-screen bg-(--ui-bg) flex flex-col">
		<!-- Header -->
		<div class="px-4 py-3 flex items-center justify-between">
			<button class="p-2 -ml-2" @click="handleHome">
				<UIcon name="i-heroicons-x-mark" class="size-6" />
			</button>
			<h1 class="text-lg font-semibold">{{ t('duel.gameResult') }}</h1>
			<div class="w-10" />
		</div>

		<!-- Result Hero -->
		<div class="flex-1 flex flex-col items-center justify-center p-6">
			<!-- Result Icon -->
			<div class="mb-6">
				<UIcon :name="resultIcon" :class="resultColor" class="size-24" />
			</div>

			<!-- Result Text -->
			<h2 class="text-4xl font-bold mb-4">{{ resultText }}</h2>

			<!-- Score -->
			<div class="flex items-center gap-6 mb-6">
				<div class="text-center">
					<p class="text-sm text-gray-500 dark:text-gray-400">{{ t('duel.you') }}</p>
					<p class="text-5xl font-bold text-primary">
						{{ gameData?.playerScore ?? 0 }}
					</p>
				</div>
				<span class="text-2xl text-(--ui-text-dimmed)">-</span>
				<div class="text-center">
					<p class="text-sm text-gray-500 dark:text-gray-400">
						{{ gameData?.opponent?.username ?? t('duel.opponent') }}
					</p>
					<p class="text-5xl font-bold text-orange-500">
						{{ gameData?.opponentScore ?? 0 }}
					</p>
				</div>
			</div>

			<!-- MMR Change -->
			<div class="mb-8">
				<p :class="mmrChangeColor" class="text-2xl font-bold">
					{{
						t('duel.mmrChange', { sign: mmrChange >= 0 ? '+' : '', amount: mmrChange })
					}}
				</p>
			</div>
		</div>

		<!-- Actions -->
		<div class="p-4 space-y-3">
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

			<UButton
				v-else-if="rematchStatus === 'pending'"
				icon="i-heroicons-clock"
				color="gray"
				variant="soft"
				size="xl"
				block
				disabled
			>
				{{ t('duel.waitingOpponent') }}
			</UButton>

			<UAlert v-if="rematchError" color="red" variant="soft" class="mb-2">
				{{ rematchError }}
			</UAlert>

			<!-- Share Button -->
			<UButton
				icon="i-heroicons-share"
				color="gray"
				variant="soft"
				size="lg"
				block
				@click="handleShare"
			>
				{{ t('duel.shareResult') }}
			</UButton>

			<!-- Back to Lobby -->
			<UButton
				icon="i-heroicons-arrow-left"
				color="gray"
				variant="ghost"
				size="lg"
				block
				@click="handleBackToLobby"
			>
				{{ t('duel.backToLobby') }}
			</UButton>
		</div>
	</div>
</template>
