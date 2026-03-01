import { ref, computed } from 'vue'
import { retrieveLaunchParams } from '@tma.js/sdk'
import type { User as TelegramUser } from '@tma.js/sdk'
import type { InternalInfrastructureHttpHandlersUserDTO } from '@/api/generated/types/internalInfrastructureHttpHandlers/UserDTO'

// Тип для parsed init data
interface ParsedInitData {
	user: TelegramUser
	authDate: Date
	hash: string
	queryId?: string | null
}

// Глобальное состояние авторизации (shared между всеми компонентами)
const isInitialized = ref(false)
const telegramUser = ref<TelegramUser | null>(null)
const currentUser = ref<InternalInfrastructureHttpHandlersUserDTO | null>(null)
const authError = ref<string | null>(null)
const initDataRaw = ref<string | null>(null) // RAW init data для Authorization header
const startParam = ref<string | null>(null) // Deep link parameter from ?startapp=xxx

/**
 * Composable для управления авторизацией через Telegram Mini App
 *
 * Использование:
 * 1. В main.ts: await initializeTMA() перед mount
 * 2. В App.vue: await loginUser() для регистрации на backend
 * 3. В компонентах: const { isAuthenticated, user } = useAuth()
 */
export function useAuth() {
	/**
	 * Инициализация Telegram Mini App SDK
	 * Вызывать ОДИН РАЗ в main.ts перед mount приложения
	 */
	const initializeTMA = async () => {
		try {
			// Получаем launch params (включая RAW init data и parsed init data)
			const launchParams = retrieveLaunchParams()

			// 🔍 DEBUG: Смотрим что вернул retrieveLaunchParams
			console.log('🔍 Full launch params:', launchParams)
			// In @tma.js/sdk v3, startParam is named tgWebAppStartParam
			console.log('🔍 tgWebAppStartParam from SDK:', launchParams.tgWebAppStartParam)

			// Extract startParam for deep linking (from ?startapp=xxx)
			// In @tma.js/sdk v3, the property is tgWebAppStartParam (not startParam)
			if (launchParams.tgWebAppStartParam) {
				startParam.value = launchParams.tgWebAppStartParam
				console.log('✅ Deep link startParam:', startParam.value)
			}

			let rawData: string | undefined = undefined
			let parsedData: ParsedInitData | undefined = undefined

			// 🔧 WORKAROUND: Парсим из hash вручную
			// Telegram Desktop передает данные в hash параметрах
			if (typeof window !== 'undefined') {
				const hash = window.location.hash
				console.log('🔧 Trying to parse from hash:', hash)

				// Извлекаем tgWebAppData из hash
				const hashParams = new URLSearchParams(hash.substring(1)) // убираем #

				// Also extract startParam from hash as fallback
				if (!startParam.value) {
					const hashStartParam = hashParams.get('tgWebAppStartParam')
					if (hashStartParam) {
						startParam.value = hashStartParam
						console.log('✅ Deep link startParam from hash:', startParam.value)
					}
				}

				const tgWebAppData = hashParams.get('tgWebAppData')

				if (tgWebAppData) {
					// URL-декодируем данные - это и есть RAW init data!
					rawData = decodeURIComponent(tgWebAppData)
					console.log('✅ Extracted initDataRaw from hash:', rawData)

					// Парсим user data вручную из URL параметров
					const initParams = new URLSearchParams(rawData)
					const userJson = initParams.get('user')

					if (userJson) {
						const user = JSON.parse(userJson)
						parsedData = {
							user: {
								id: user.id,
								first_name: user.first_name,
								last_name: user.last_name,
								username: user.username,
								language_code: user.language_code,
								is_premium: user.is_premium || false,
								allows_write_to_pm: user.allows_write_to_pm || false,
								photo_url: user.photo_url,
							} as TelegramUser,
							authDate: new Date(parseInt(initParams.get('auth_date') || '0') * 1000),
							hash: initParams.get('hash') || '',
							queryId: initParams.get('query_id'),
						}
						console.log('✅ Parsed user data:', parsedData)
					}
				}
			}

			// Fallback: try localStorage (app resume after background reload)
			if (!rawData && typeof localStorage !== 'undefined') {
				const cached = localStorage.getItem('tma_init_data_raw')
				if (cached) {
					rawData = cached
					console.log('✅ Using cached init data from localStorage (app resume)')

					// Parse user from cached raw data
					const initParams = new URLSearchParams(rawData)
					const userJson = initParams.get('user')
					if (userJson) {
						const user = JSON.parse(userJson)
						parsedData = {
							user: {
								id: user.id,
								first_name: user.first_name,
								last_name: user.last_name,
								username: user.username,
								language_code: user.language_code,
								is_premium: user.is_premium || false,
								allows_write_to_pm: user.allows_write_to_pm || false,
								photo_url: user.photo_url,
							} as TelegramUser,
							authDate: new Date(parseInt(initParams.get('auth_date') || '0') * 1000),
							hash: initParams.get('hash') || '',
							queryId: initParams.get('query_id'),
						}
					}
				}
			}

			if (!rawData) {
				console.error('❌ TMA: No raw init data available!')
				console.error('❌ Could not find data in SDK or hash parameters')
				authError.value = 'No Telegram init data'
				isInitialized.value = true
				return false
			}

			// Сохраняем RAW init data для отправки на сервер
			initDataRaw.value = rawData

			// Сохраняем в localStorage для axios interceptor
			if (typeof localStorage !== 'undefined') {
				localStorage.setItem('tma_init_data_raw', rawData)
			}

			// Проверяем parsed user data
			if (!parsedData?.user) {
				console.warn('TMA: No user data available')
				authError.value = 'No Telegram user data'
				isInitialized.value = true
				return false
			}

			telegramUser.value = parsedData.user
			isInitialized.value = true

			console.log('TMA initialized:', {
				id: parsedData.user.id,
				username: parsedData.user.username,
				firstName: parsedData.user.first_name,
				hasRawInitData: !!initDataRaw.value,
			})

			return true
		} catch (error) {
			console.error('Failed to initialize TMA:', error)
			authError.value = 'TMA initialization failed'
			isInitialized.value = true
			return false
		}
	}

	/**
	 * Получить RAW init data для отправки на сервер
	 * Сервер сам распарсит и провалидирует подпись
	 */
	const getRawInitData = () => {
		return initDataRaw.value
	}

	/**
	 * Установить текущего пользователя после успешной регистрации/логина
	 */
	const setCurrentUser = (user: InternalInfrastructureHttpHandlersUserDTO) => {
		currentUser.value = user
		authError.value = null
	}

	/**
	 * Очистить состояние авторизации (logout)
	 */
	const clearAuth = () => {
		currentUser.value = null
		authError.value = null
	}

	/**
	 * Get and clear startParam (use once after handling deep link)
	 */
	const consumeStartParam = () => {
		const param = startParam.value
		startParam.value = null
		return param
	}

	// Computed свойства
	const isAuthenticated = computed(() => currentUser.value !== null)
	const userId = computed(() => currentUser.value?.id || null)

	return {
		// Состояние
		isInitialized,
		isAuthenticated,
		telegramUser,
		currentUser,
		userId,
		authError,
		initDataRaw,
		startParam,

		// Методы
		initializeTMA,
		getRawInitData,
		setCurrentUser,
		clearAuth,
		consumeStartParam,
	}
}
