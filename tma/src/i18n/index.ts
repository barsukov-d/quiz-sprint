import { createI18n } from 'vue-i18n'
import { retrieveLaunchParams } from '@tma.js/sdk'
import en from './locales/en'
import ru from './locales/ru'

export type Locale = 'en' | 'ru'

export function detectLocale(): Locale {
	// 1. Manual override
	const stored = localStorage.getItem('locale')
	if (stored === 'en' || stored === 'ru') return stored

	// 2. Telegram user language
	try {
		const params = retrieveLaunchParams()
		const lang: string = params.tgWebAppData?.user?.language_code ?? ''
		if (lang.startsWith('ru')) return 'ru'
	} catch {
		// SDK not available outside Telegram
	}

	// 3. Default
	return 'en'
}

export const i18n = createI18n({
	legacy: false,
	locale: detectLocale(),
	availableLocales: ['en', 'ru'],
	messages: { en, ru },
})

export function setLocale(lang: Locale) {
	i18n.global.locale.value = lang
	localStorage.setItem('locale', lang)
}
