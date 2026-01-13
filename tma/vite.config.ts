import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: true, // Allow external connections
    allowedHosts: [
      '.trycloudflare.com',
      'localhost',
      'dev.quiz-sprint-tma.online', // Development subdomain
      'quiz-sprint-tma.online', // Production domain
    ],
    hmr: {
      protocol: 'wss',
      host: 'dev.quiz-sprint-tma.online',
      clientPort: 443,
    },
  },
})
