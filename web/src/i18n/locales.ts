// ── Locale registry ──────────────────────────────────────────────────────────
// To add a language:
//   1. Create public/locales/<code>.json  (include "_label" for display name)
//   2. Add the code string to LOCALES_CODES below
// That's all. The app auto-discovers the new language.

export const LOCALES_CODES = ['en', 'zh-TW'] as const

export type Locale = (typeof LOCALES_CODES)[number]

export interface LocaleInfo {
  code: Locale
  label: string
}
