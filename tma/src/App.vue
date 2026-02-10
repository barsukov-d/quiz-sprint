<script setup lang="ts">
import eruda from 'eruda'
if (import.meta.env.DEV) {
	eruda.init()
}
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from './composables/useAuth'
import { usePostUserRegister } from './api/generated/hooks/userController/usePostUserRegister'
import BottomTabBar from './components/BottomTabBar.vue'

import { viewport } from '@tma.js/sdk'

viewport.safeAreaInsetTop()

const route = useRoute()
const router = useRouter()

// Show bottom navigation only on main screens, hide during quiz play and results
const showBottomNav = computed(() => {
	const hiddenRoutes = ['quiz-play', 'quiz-results', 'quiz-details']
	return !hiddenRoutes.includes(route.name as string)
})

const { isInitialized, getRawInitData, setCurrentUser, consumeStartParam } = useAuth()

const isLoading = ref(true)
const error = ref<string | null>(null)

// Мутация для регистрации пользователя
const { mutateAsync: registerUser } = usePostUserRegister()

// Handle deep link navigation
const handleDeepLink = (startParam: string) => {
	// Duel challenge link: duel_abc12345
	if (startParam.startsWith('duel_')) {
		console.log('🎮 Duel challenge detected, redirecting to lobby')
		router.push({
			name: 'duel-lobby',
			query: { challenge: startParam },
		})
		return
	}

	// Referral link: ref_user123
	if (startParam.startsWith('ref_')) {
		console.log('👥 Referral link detected:', startParam)
		// Store referral for later processing
		localStorage.setItem('referral_code', startParam)
		return
	}

	console.log('❓ Unknown deep link type:', startParam)
}

// Автоматическая регистрация/логин при загрузке
onMounted(async () => {
	try {
		// Ждем инициализации TMA
		if (!isInitialized.value) {
			console.warn('TMA not initialized yet')
			error.value = 'Приложение должно быть запущено в Telegram'
			isLoading.value = false
			return
		}

		// Проверяем наличие raw init data
		const rawInitData = getRawInitData()

		if (!rawInitData) {
			console.warn('No Telegram init data available')
			error.value = 'Нет данных авторизации из Telegram'
			isLoading.value = false
			return
		}

		console.log('Registering user with Telegram init data (signed by Telegram)')

		// Регистрируем/обновляем пользователя на backend
		// ⚠️ ВАЖНО: Данные НЕ в body, а в Authorization header!
		// Axios interceptor автоматически добавит: Authorization: tma <base64(init-data-raw)>
		// Backend должен:
		// 1. Декодировать base64
		// 2. Валидировать подпись init data
		// 3. Извлечь userId, username и т.д. из валидированных данных
		// 4. Создать/обновить пользователя
		const response = await registerUser()

		// Сохраняем данные пользователя в глобальном состоянии
		if (response?.data?.user) {
			setCurrentUser(response.data.user)
			console.log('User registered successfully:', response.data.user)

			if (response.data.isNewUser) {
				console.log('Welcome new user!')
			} else {
				console.log('Welcome back!')
			}

			// Handle deep link after successful registration
			const startParam = consumeStartParam()
			if (startParam) {
				console.log('🔗 Processing deep link:', startParam)
				handleDeepLink(startParam)
			}
		}
	} catch (err) {
		console.error('Failed to register user:', err)
		error.value = 'Не удалось зарегистрировать пользователя'
	} finally {
		isLoading.value = false
	}
})
</script>

<template>
	<UApp>
		<!-- Экран загрузки -->
		<div v-if="isLoading" class="loading-screen">
			<div class="loading-content">
				<div class="spinner"></div>
				<p>Загрузка...</p>
			</div>
		</div>

		<!-- Экран ошибки -->
		<div v-else-if="error" class="error-screen">
			<div class="error-content">
				<h2>Ошибка</h2>
				<p>{{ error }}</p>
				<p class="hint">Убедитесь, что приложение запущено в Telegram</p>
			</div>
		</div>

		<!-- Основное приложение -->
		<div v-else class="app-container p-4 pt-4 pb-20 sm:p-3 sm:pt-20">
			<RouterView />
			<BottomTabBar v-if="showBottomNav" />
		</div>
	</UApp>
</template>

<style scoped>
.loading-screen,
.error-screen {
	display: flex;
	align-items: center;
	justify-content: center;
	min-height: 100vh;
	padding: 20px;
}

.loading-content,
.error-content {
	text-align: center;
}

.spinner {
	width: 48px;
	height: 48px;
	margin: 0 auto 16px;
	border: 4px solid rgba(0, 0, 0, 0.1);
	border-left-color: var(--color-primary, #007aff);
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

@keyframes spin {
	to {
		transform: rotate(360deg);
	}
}

.error-content h2 {
	font-size: 24px;
	margin-bottom: 12px;
	color: #ff3b30;
}

.error-content p {
	margin-bottom: 8px;
	color: #333;
}

.error-content .hint {
	font-size: 14px;
	color: #999;
}

.app-container {
	min-height: 100vh;
	position: relative;
}
</style>
