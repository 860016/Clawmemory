import { createI18n } from 'vue-i18n'
import zh from './zh'
import en from './en'

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('clawmemory-locale') || 'zh',
  fallbackLocale: 'en',
  messages: { zh, en },
})

export default i18n

export function setLocale(locale: string) {
  i18n.global.locale.value = locale
  localStorage.setItem('clawmemory-locale', locale)
}

export function getLocale(): string {
  return i18n.global.locale.value
}
