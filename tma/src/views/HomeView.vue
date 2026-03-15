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
	<div class="home-view">
		<!-- Header -->
		<header class="home-header">
			<div class="home-header__left">
				<div class="home-header__logo">
					<UIcon name="i-heroicons-bolt" class="size-5 text-white" />
				</div>
				<span class="home-header__title">Quiz Sprint</span>
			</div>
			<div class="home-header__actions">
				<button class="home-header__icon-btn" aria-label="Search">
					<UIcon name="i-heroicons-magnifying-glass" class="size-5" />
				</button>
			</div>
		</header>

		<!-- Greeting -->
		<div v-if="isAuthenticated && currentUser" class="home-greeting">
			<UAvatar :src="currentUser.avatarUrl" :alt="currentUser.username" size="md" />
			<div class="home-greeting__text">
				<p class="home-greeting__name">{{ greeting }}, {{ currentUser.username }}!</p>
				<p class="home-greeting__sub">{{ t('home.subtitle') }}</p>
			</div>
		</div>

		<!-- Banner card -->
		<div class="home-banner">
			<div class="home-banner__content">
				<p class="home-banner__title">
					{{ t('home.bannerTitle', 'Play quiz with your friends now!') }}
				</p>
				<UButton
					size="sm"
					color="white"
					variant="solid"
					:label="t('home.bannerCta', 'Find Friends')"
					class="home-banner__btn"
				/>
			</div>
			<!-- Decorative avatar stack -->
			<div class="home-banner__avatars" aria-hidden="true">
				<div class="avatar-bubble avatar-bubble--1">
					<UIcon name="i-heroicons-user" class="size-4 text-white" />
				</div>
				<div class="avatar-bubble avatar-bubble--2">
					<UIcon name="i-heroicons-user" class="size-4 text-white" />
				</div>
				<div class="avatar-bubble avatar-bubble--3">
					<UIcon name="i-heroicons-user" class="size-4 text-white" />
				</div>
				<div class="avatar-bubble avatar-bubble--4">
					<UIcon name="i-heroicons-user" class="size-4 text-white" />
				</div>
				<div class="avatar-bubble avatar-bubble--5">
					<UIcon name="i-heroicons-user" class="size-4 text-white" />
				</div>
			</div>
		</div>

		<!-- Game Mode Cards -->
		<section class="home-section">
			<div class="home-section__list">
				<DailyChallengeCard :player-id="playerId" />
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
		</section>
	</div>
</template>

<style scoped>
.home-view {
	max-width: 800px;
	margin: 0 auto;
	display: flex;
	flex-direction: column;
	gap: 1.25rem;
	padding-bottom: 1rem;
}

/* ─── Header ─── */
.home-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.25rem 0;
}

.home-header__left {
	display: flex;
	align-items: center;
	gap: 0.625rem;
}

.home-header__logo {
	width: 2rem;
	height: 2rem;
	border-radius: 0.5rem;
	background: var(--ui-primary);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.home-header__title {
	font-size: 1.125rem;
	font-weight: 700;
	color: var(--ui-text-highlighted);
	letter-spacing: -0.01em;
}

.home-header__actions {
	display: flex;
	align-items: center;
	gap: 0.25rem;
}

.home-header__icon-btn {
	width: 2.25rem;
	height: 2.25rem;
	border-radius: 0.625rem;
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--ui-text-muted);
	background: none;
	border: none;
	cursor: pointer;
	transition:
		background-color 150ms ease,
		color 150ms ease;
}

.home-header__icon-btn:hover {
	background-color: var(--ui-bg-accented);
	color: var(--ui-text);
}

/* ─── Greeting ─── */
.home-greeting {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.home-greeting__text {
	display: flex;
	flex-direction: column;
	gap: 0.125rem;
}

.home-greeting__name {
	font-size: 1rem;
	font-weight: 700;
	color: var(--ui-text-highlighted);
	line-height: 1.25;
}

.home-greeting__sub {
	font-size: 0.8125rem;
	color: var(--ui-text-dimmed);
}

/* ─── Banner ─── */
.home-banner {
	position: relative;
	border-radius: 1rem;
	background: linear-gradient(135deg, #6c5ce7 0%, #8675f3 50%, #a397fa 100%);
	padding: 1.25rem 1.25rem 1.25rem 1.25rem;
	overflow: hidden;
	min-height: 7rem;
	display: flex;
	align-items: center;
}

.home-banner__content {
	position: relative;
	z-index: 1;
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
	max-width: 60%;
}

.home-banner__title {
	font-size: 1rem;
	font-weight: 700;
	color: #ffffff;
	line-height: 1.35;
}

.home-banner__btn {
	align-self: flex-start;
}

/* Decorative floating avatar bubbles */
.home-banner__avatars {
	position: absolute;
	inset: 0;
	pointer-events: none;
}

.avatar-bubble {
	position: absolute;
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.25);
	backdrop-filter: blur(4px);
	display: flex;
	align-items: center;
	justify-content: center;
}

.avatar-bubble--1 {
	width: 2.25rem;
	height: 2.25rem;
	top: 0.625rem;
	right: 1.5rem;
	background: rgba(255, 255, 255, 0.3);
}

.avatar-bubble--2 {
	width: 1.875rem;
	height: 1.875rem;
	top: 1.75rem;
	right: 3.75rem;
	background: rgba(255, 255, 255, 0.2);
}

.avatar-bubble--3 {
	width: 2rem;
	height: 2rem;
	top: 2.75rem;
	right: 0.75rem;
	background: rgba(255, 255, 255, 0.25);
}

.avatar-bubble--4 {
	width: 1.75rem;
	height: 1.75rem;
	bottom: 0.875rem;
	right: 3rem;
	background: rgba(255, 255, 255, 0.2);
}

.avatar-bubble--5 {
	width: 1.5rem;
	height: 1.5rem;
	bottom: 0.5rem;
	right: 1rem;
	background: rgba(255, 255, 255, 0.15);
}

/* ─── Sections ─── */
.home-section {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.home-section__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.home-section__title {
	font-size: 1rem;
	font-weight: 700;
	color: var(--ui-text-highlighted);
}

.home-section__list {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}
</style>
