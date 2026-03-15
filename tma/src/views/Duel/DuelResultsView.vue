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
	<div class="fixed inset-0 flex flex-col bg-(--ui-bg) z-40">
		<!-- Header -->
		<div
			class="flex items-center justify-between px-4 pt-3 pb-2 shrink-0 border-b border-(--ui-border)"
		>
			<button class="p-1 text-(--ui-text-muted) hover:text-(--ui-text)" @click="handleHome">
				<UIcon name="i-heroicons-x-mark" class="size-5" />
			</button>
			<span class="text-sm font-semibold text-(--ui-text-highlighted)">{{
				t('duel.gameResult')
			}}</span>
			<div class="w-7" />
		</div>

		<!-- Content centered vertically -->
		<div class="flex-1 flex flex-col items-center justify-center px-4">
			<div class="text-6xl mb-3">{{ isDraw ? '🤝' : didWin ? '🏆' : '💔' }}</div>
			<h2 class="text-2xl font-black text-(--ui-text-highlighted)">{{ resultText }}</h2>
			<span
				class="mt-2 mb-6 px-3 py-1 rounded-full text-sm font-bold"
				:class="
					mmrChange >= 0 ? 'bg-green-500/15 text-green-500' : 'bg-red-500/15 text-red-500'
				"
			>
				{{ t('duel.mmrChange', { sign: mmrChange >= 0 ? '+' : '', amount: mmrChange }) }}
			</span>

			<!-- Score card -->
			<div
				class="w-full max-w-[280px] rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border) p-4 mb-8"
			>
				<div class="flex items-center">
					<div class="text-center flex-1">
						<p class="text-xs text-(--ui-text-muted) mb-1">{{ t('duel.you') }}</p>
						<p class="text-4xl font-black text-(--ui-text-highlighted) tabular-nums">
							{{ gameData?.playerScore ?? 0 }}
						</p>
					</div>
					<span class="text-lg text-(--ui-text-dimmed) font-bold mx-2">—</span>
					<div class="text-center flex-1">
						<p class="text-xs text-(--ui-text-muted) mb-1 truncate">
							{{ gameData?.opponent?.username ?? t('duel.opponent') }}
						</p>
						<p class="text-4xl font-black text-(--ui-text-muted) tabular-nums">
							{{ gameData?.opponentScore ?? 0 }}
						</p>
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="w-full flex flex-col gap-3">
				<UButton
					v-if="rematchStatus === 'idle'"
					icon="i-heroicons-arrow-path"
					color="primary"
					size="lg"
					block
					:loading="isRematchLoading"
					@click="handleRematch"
				>
					{{ t('duel.requestRematch') }}
				</UButton>
				<div
					v-else-if="rematchStatus === 'pending'"
					class="flex items-center justify-center gap-2 py-3 text-(--ui-text-muted)"
				>
					<UIcon name="i-heroicons-clock" class="size-4 animate-pulse" />
					<span class="text-sm">{{ t('duel.waitingOpponent') }}</span>
				</div>
				<p v-if="rematchError" class="text-center text-red-500 text-xs">
					{{ rematchError }}
				</p>
				<div class="flex items-center justify-center gap-4 pt-1">
					<button
						class="text-sm text-(--ui-text-muted) hover:text-(--ui-text) flex items-center gap-1.5 transition-colors"
						@click="handleShare"
					>
						<UIcon name="i-heroicons-share" class="size-4" />
						{{ t('duel.shareResult') }}
					</button>
					<span class="w-px h-4 bg-(--ui-border)" />
					<button
						class="text-sm text-(--ui-text-muted) hover:text-(--ui-text) flex items-center gap-1.5 transition-colors"
						@click="handleBackToLobby"
					>
						<UIcon name="i-heroicons-arrow-left" class="size-4" />
						{{ t('duel.backToLobby') }}
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
