import './assets/main.css'

import { createApp } from 'vue'
import { VueQueryPlugin, type VueQueryPluginOptions } from '@tanstack/vue-query'
import App from './App.vue'
import router from './router'
import ui from '@nuxt/ui/vue-plugin'

const app = createApp(App)

// Конфигурация Vue Query
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

app.use(router)
app.use(ui)
app.use(VueQueryPlugin, vueQueryOptions)

app.mount('#app')
