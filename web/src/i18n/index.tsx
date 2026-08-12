import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import { LOCALES_CODES, type Locale, type LocaleInfo } from './locales'

// ── Translation key structure (for type safety) ──────────────────────────────

export interface TranslationKeys {
  _label?: string  // Display label for language selector (optional, falls back to code)
  header: {
    locked: string
    unlocked: string
    sudoPassword: string
    unlock: string
  }
  status: {
    title: string
    swapFile: string
    swapSize: string
    zramSize: string
    zramActive: string
    totalSwap: string
    swappiness: string
    vram: string
    recommended: string
    default: string
    enabled: string
    disabled: string
    none: string
  }
  swap: {
    title: string
    swapSizeGB: string
    current: string
    resizing: string
    resizeSwap: string
    apply: string
    setting: string
  }
  memory: {
    title: string
    on: string
    off: string
  }
  vram: {
    title: string
    readOnly: string
  }
  presets: {
    title: string
    applying: string
    recommended: string
    stock: string
  }
  gamedata: {
    title: string
    select: string
    ssdLibrary: string
    externalLibrary: string
    syncing: string
    syncGameData: string
    cleaning: string
    cleanupOrphanedData: string
  }
  memoryParams: {
    hugepages: string
    shmem: string
    compaction_proactiveness: string
    defrag: string
    page_lock_unfairness: string
  }
  common: {
    loading: string
  }
}

export interface I18nContextType {
  locale: Locale
  t: TranslationKeys
  setLocale: (locale: Locale) => void
  tKey: (key: string) => string
}

const I18nContext = createContext<I18nContextType | null>(null)

export function useI18n(): I18nContextType {
  const ctx = useContext(I18nContext)
  if (!ctx) throw new Error('useI18n must be used within I18nProvider')
  return ctx
}

// Dot-notation key lookup: "status.recommended" → t.status.recommended
export function resolveKey(obj: unknown, key: string): string {
  const parts = key.split('.')
  let current = obj as Record<string, unknown>
  for (const part of parts) {
    if (current == null || typeof current !== 'object') return key
    current = current[part] as Record<string, unknown>
    if (current == null) return key
  }
  return typeof current === 'string' ? current : key
}

// ── Auto-discover locales at build time ──────────────────────────────────────

// Auto-discover locales at build time from src/locales/*.json
// (files are copied to public/locales/ by the build script for runtime serving)
const localeModules = import.meta.glob('/src/locales/*.json', { eager: true }) as Record<string, { default: TranslationKeys }>

const LOCALES: Record<Locale, LocaleInfo> = {} as Record<Locale, LocaleInfo>
for (const code of LOCALES_CODES) {
  const mod = localeModules[`/src/locales/${code}.json`]
  if (mod) {
    const data = mod.default as TranslationKeys
    LOCALES[code] = { code, label: data._label || code }
  }
}

export { LOCALES }

export function resolveLocale(browserLang: string): Locale {
  const short = browserLang.slice(0, 2)
  // Exact match first, then prefix match (e.g. "zh-Hant" → "zh-TW")
  if (LOCALES[short as Locale]) return short as Locale
  for (const code of LOCALES_CODES) {
    if (code.startsWith(short)) return code
  }
  return LOCALES_CODES[0] // fallback to first locale (en)
}

export function getLocaleByCode(code: string): LocaleInfo | undefined {
  return LOCALES[code as Locale]
}

// Re-export for convenience (LanguageSelector imports Locale from here)
export { LOCALES_CODES, type Locale } from './locales'

// ── Provider ─────────────────────────────────────────────────────────────────

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const stored = localStorage.getItem('cryoutils-ng-locale')
    if (stored && stored in LOCALES) return stored as Locale
    return resolveLocale(navigator.language)
  })

  const [translations, setTranslations] = useState<TranslationKeys | null>(null)

  useEffect(() => {
    let cancelled = false
    fetch(`/locales/${locale}.json`)
      .then((res) => res.json())
      .then((data) => {
        if (!cancelled) setTranslations(data as TranslationKeys)
      })
      .catch((err) => {
        console.error(`[i18n] Failed to load ${locale}:`, err)
      })
    return () => { cancelled = true }
  }, [locale])

  useEffect(() => {
    localStorage.setItem('cryoutils-ng-locale', locale)
  }, [locale])

  const setLocale = useCallback((l: Locale) => {
    setLocaleState(l)
  }, [])

  const tKey = useCallback(
    (key: string) => (translations ? resolveKey(translations, key) : ''),
    [translations]
  )

  const safeT = translations ?? {
    header: { locked: '', unlocked: '', sudoPassword: '', unlock: '' },
    status: { title: '', swapFile: '', swapSize: '', zramSize: '', zramActive: '', totalSwap: '', swappiness: '', vram: '', recommended: '', default: '', enabled: '', disabled: '', none: '' },
    swap: { title: '', swapSizeGB: '', current: '', resizing: '', resizeSwap: '', apply: '', setting: '' },
    memory: { title: '', on: '', off: '' },
    vram: { title: '', readOnly: '' },
    presets: { title: '', applying: '', recommended: '', stock: '' },
    gamedata: { title: '', select: '', ssdLibrary: '', externalLibrary: '', syncing: '', syncGameData: '', cleaning: '', cleanupOrphanedData: '' },
    memoryParams: { hugepages: '', shmem: '', compaction_proactiveness: '', defrag: '', page_lock_unfairness: '' },
    common: { loading: 'Loading...' },
  }

  const value: I18nContextType = {
    locale,
    t: safeT,
    setLocale,
    tKey,
  }

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}
