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
		ui({
			ui: {
				colors: {
					primary: 'indigo',
					secondary: 'violet',
					success: 'emerald',
					warning: 'amber',
					error: 'rose',
					info: 'sky',
					neutral: 'slate',
				},
				// ─── Component overrides (Quizzo design patterns) ───
				button: {
					slots: {
						base: 'rounded-full font-semibold',
					},
				},
				card: {
					slots: {
						root: 'rounded-2xl',
					},
				},
				badge: {
					variants: {
						size: {
							xs: { base: 'rounded-full' },
							sm: { base: 'rounded-full' },
							md: { base: 'rounded-full' },
							lg: { base: 'rounded-full' },
							xl: { base: 'rounded-full' },
						},
					},
				},
				modal: {
					slots: {
						content: 'rounded-2xl',
					},
				},
				alert: {
					slots: {
						root: 'rounded-xl',
					},
				},
				input: {
					slots: {
						base: 'rounded-xl',
					},
				},
				select: {
					slots: {
						base: 'rounded-xl',
					},
				},
				progress: {
					slots: {
						root: 'rounded-full',
						indicator: 'rounded-full',
					},
				},
			},
		}),
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
			'localhost',
			'dev.quiz-sprint-tma.online',
			'quiz-sprint-tma.online',
		],
		// HMR configuration for Caddy reverse proxy
	},
})
