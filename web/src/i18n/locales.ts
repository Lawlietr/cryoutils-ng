// Single source of truth for available locales.
// To add a new language:
//   1. Create public/locales/<code>.json
//   2. Add entry here (code must match filename)
// That's it. Locale type, selector dropdown, and default detection all update automatically.

export const LOCALES = {
  'en': { code: 'en', label: 'English' },
  'zh-TW': { code: 'zh-TW', label: '繁體中文' },
} as const

export type Locale = keyof typeof LOCALES

export type LocaleInfo = (typeof LOCALES)[Locale]

export function getLocaleByCode(code: string): LocaleInfo | undefined {
  return LOCALES[code as Locale]
}

export function resolveLocale(browserLang: string): Locale {
  const short = browserLang.slice(0, 2)
  // Exact match first, then prefix match (e.g. "zh-Hant" → "zh-TW")
  if (LOCALES[short as Locale]) return short as Locale
  for (const code of Object.keys(LOCALES)) {
    if (code.startsWith(short)) return code as Locale
  }
  return 'en' // fallback
}
