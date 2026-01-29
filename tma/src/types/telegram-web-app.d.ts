// Type definitions for Telegram Web App API
// https://core.telegram.org/bots/webapps

interface TelegramWebAppUser {
	id: number
	is_bot?: boolean
	first_name: string
	last_name?: string
	username?: string
	language_code?: string
	is_premium?: boolean
	added_to_attachment_menu?: boolean
	allows_write_to_pm?: boolean
	photo_url?: string
}

interface TelegramWebAppInitData {
	query_id?: string
	user?: TelegramWebAppUser
	receiver?: TelegramWebAppUser
	chat?: any
	chat_type?: string
	chat_instance?: string
	start_param?: string
	can_send_after?: number
	auth_date: number
	hash: string
}

interface TelegramWebApp {
	initData: string
	initDataUnsafe: TelegramWebAppInitData
	version: string
	platform: string
	colorScheme: 'light' | 'dark'
	themeParams: any
	isExpanded: boolean
	viewportHeight: number
	viewportStableHeight: number
	headerColor: string
	backgroundColor: string
	isClosingConfirmationEnabled: boolean
	BackButton: any
	MainButton: any
	HapticFeedback: any

	ready(): void
	expand(): void
	close(): void
	enableClosingConfirmation(): void
	disableClosingConfirmation(): void
	onEvent(eventType: string, eventHandler: () => void): void
	offEvent(eventType: string, eventHandler: () => void): void
	sendData(data: string): void
	switchInlineQuery(query: string, choose_chat_types?: string[]): void
	openLink(url: string, options?: { try_instant_view?: boolean }): void
	openTelegramLink(url: string): void
	openInvoice(url: string, callback?: (status: string) => void): void
	showPopup(params: any, callback?: (button_id: string) => void): void
	showAlert(message: string, callback?: () => void): void
	showConfirm(message: string, callback?: (confirmed: boolean) => void): void
	showScanQrPopup(params: any, callback?: (data: string) => void): void
	closeScanQrPopup(): void
	readTextFromClipboard(callback?: (text: string) => void): void
	requestWriteAccess(callback?: (granted: boolean) => void): void
	requestContact(callback?: (granted: boolean) => void): void
}

interface Window {
	Telegram?: {
		WebApp: TelegramWebApp
	}
}
