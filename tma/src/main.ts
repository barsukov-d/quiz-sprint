import { createApp } from 'vue'
// @ts-expect-error Module declaration for .vue files
import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(router)

app.mount('#app')
