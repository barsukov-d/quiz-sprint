import { describe, it, expect, beforeEach, vi } from 'vitest'

// Mock @tma.js/sdk before importing the module under test
vi.mock('@tma.js/sdk', () => ({
  retrieveLaunchParams: vi.fn(),
}))

import { retrieveLaunchParams } from '@tma.js/sdk'
import { detectLocale } from '@/i18n/index'

describe('detectLocale', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.mocked(retrieveLaunchParams).mockReturnValue({} as unknown as ReturnType<typeof retrieveLaunchParams>)
  })

  it('returns localStorage value when set to ru', () => {
    localStorage.setItem('locale', 'ru')
    expect(detectLocale()).toBe('ru')
  })

  it('returns localStorage value when set to en', () => {
    localStorage.setItem('locale', 'en')
    expect(detectLocale()).toBe('en')
  })

  it('falls back to Telegram languageCode ru', () => {
    vi.mocked(retrieveLaunchParams).mockReturnValue({
      initData: { user: { languageCode: 'ru' } },
    } as unknown as ReturnType<typeof retrieveLaunchParams>)
    expect(detectLocale()).toBe('ru')
  })

  it('normalizes ru-RU to ru', () => {
    vi.mocked(retrieveLaunchParams).mockReturnValue({
      initData: { user: { languageCode: 'ru-RU' } },
    } as unknown as ReturnType<typeof retrieveLaunchParams>)
    expect(detectLocale()).toBe('ru')
  })

  it('falls back to en for unknown language', () => {
    vi.mocked(retrieveLaunchParams).mockReturnValue({
      initData: { user: { languageCode: 'de' } },
    } as unknown as ReturnType<typeof retrieveLaunchParams>)
    expect(detectLocale()).toBe('en')
  })

  it('falls back to en when no Telegram data', () => {
    expect(detectLocale()).toBe('en')
  })

  it('localStorage takes priority over Telegram', () => {
    localStorage.setItem('locale', 'en')
    vi.mocked(retrieveLaunchParams).mockReturnValue({
      initData: { user: { languageCode: 'ru' } },
    } as unknown as ReturnType<typeof retrieveLaunchParams>)
    expect(detectLocale()).toBe('en')
  })
})
