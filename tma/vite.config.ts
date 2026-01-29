import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import ui from '@nuxt/ui/vite'
// import vueDevTools from 'vite-plugin-vue-devtools'
import { visualizer } from 'rollup-plugin-visualizer'

// https://vite.dev/config/
export default defineConfig({
	plugins: [
		vue(),
		ui(),
		// vueDevTools(),
		visualizer({
			open: false,
			filename: 'dist/stats.html',
			gzipSize: true,
			brotliSize: true,
		}),
	],
	resolve: {
		alias: {
			'@': fileURLToPath(new URL('./src', import.meta.url)),
		},
	},
	server: {
		host: true, // Allow external connections
		port: 5173,
		strictPort: true,
		allowedHosts: [
			'.trycloudflare.com',
			'localhost',
			'dev.quiz-sprint-tma.online', // Development subdomain
			'quiz-sprint-tma.online', // Production domain
		],
		// HMR configuration for Caddy reverse proxy
	},
})
