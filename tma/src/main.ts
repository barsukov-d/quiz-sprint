import './assets/main.css'

import { createApp } from 'vue'
import { VueQueryPlugin, type VueQueryPluginOptions } from '@tanstack/vue-query'
import { init } from '@tma.js/sdk'
import App from './App.vue'
import router from './router'
import ui from '@nuxt/ui/vue-plugin'
import { useAuth } from './composables/useAuth'

// Инициализация Telegram Mini App
async function initializeApp() {
	try {
		// 1. Инициализируем TMA SDK (@tma.js/sdk v3)
		init()
		console.log('TMA SDK initialized')

		// 2. Инициализируем auth composable
		const { initializeTMA } = useAuth()
		await initializeTMA()

		// 3. Создаем Vue приложение
		const app = createApp(App)

		// 4. Конфигурация Vue Query
		const vueQueryOptions: VueQueryPluginOptions = {
			queryClientConfig: {
				defaultOptions: {
					queries: {
						staleTime: 1000 * 60 * 5, // 5 минут - данные свежие
						gcTime: 1000 * 60 * 10, // 10 минут - очистка кэша
						retry: 1, // 1 повтор при ошибке
						refetchOnWindowFocus: false, // Отключить refetch при фокусе (для TMA)
					},
					mutations: {
						retry: 0, // Не повторять мутации
					},
				},
			},
		}

		// 5. Подключаем плагины
		app.use(router)
		app.use(ui)
		app.use(VueQueryPlugin, vueQueryOptions)

		// 6. Монтируем приложение
		app.mount('#app')

		console.log('Vue app mounted')
	} catch (error) {
		console.error('Failed to initialize app:', error)
		// Показываем ошибку пользователю
		document.body.innerHTML = `
      <div style="padding: 20px; text-align: center;">
        <h1>Ошибка инициализации</h1>
        <p>Не удалось запустить приложение. Попробуйте перезагрузить страницу.</p>
        <pre style="text-align: left; background: #f5f5f5; padding: 10px;">${error}</pre>
      </div>
    `
	}
}

// Запускаем приложение
initializeApp()
