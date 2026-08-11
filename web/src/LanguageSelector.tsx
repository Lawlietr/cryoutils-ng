import { useI18n } from './i18n'
import { LOCALES } from './i18n/locales'

export default function LanguageSelector() {
  const { locale, setLocale } = useI18n()

  return (
    <select
      className="locale-select"
      value={locale}
      onChange={(e) => setLocale(e.target.value as typeof locale)}
      title="Language"
    >
      {Object.values(LOCALES).map((l) => (
        <option key={l.code} value={l.code}>
          {l.label}
        </option>
      ))}
    </select>
  )
}
