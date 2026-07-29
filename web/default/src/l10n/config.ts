/*
Copyright (C) 2023-2026 OpenFastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './locales/en.json'
import fr from './locales/fr.json'
import ja from './locales/ja.json'
import ru from './locales/ru.json'
import vi from './locales/vi.json'
import zh from './locales/zh.json'
import ar from './locales/ar.json'
import { applyLocaleSideEffects } from '@/lib/locale'

export const SUPPORTED_LANGUAGES = ['zh', 'en', 'fr', 'ru', 'ja', 'vi', 'ar'] as const
export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number]

const resources = {
  en,
  zh,
  fr,
  ru,
  ja,
  vi,
  ar,
} as const

// 唯一持久化 key。刻意不使用 i18next 旧标记(i18nextLng / i18nextLngUserSet /
// i18nextLangExplicit)，以免历史 bug 残留把中文用户锁成英文。
const LANG_KEY = 'ft_lang'

function isSupported(lang: string | null): lang is SupportedLanguage {
  return lang !== null && (SUPPORTED_LANGUAGES as readonly string[]).includes(lang)
}

// 极简规则：默认中文。仅当用户在界面主动切换语言、并把选择写入 LANG_KEY 后，
// 才在下次加载时尊重该选择。旧浏览器的任何 en 残留都不会被读取。
function getInitialLanguage(): SupportedLanguage {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem(LANG_KEY)
    if (isSupported(saved)) return saved
  }
  return 'zh'
}

i18n.use(initReactI18next).init({
  resources,
  lng: getInitialLanguage(),
  fallbackLng: ['en', 'zh'],
  supportedLngs: ['en', 'zh', 'fr', 'ru', 'ja', 'vi', 'ar'],
  load: 'languageOnly', // Convert zh-CN -> zh
  nsSeparator: false, // Allow literal colons in keys (e.g., URLs, labels)
  keySeparator: false, // Flat keys (e.g. 'about.featureXxx.title') resolved literally
  debug: import.meta.env.DEV,
  interpolation: {
    escapeValue: false, // not needed for react as it escapes by default
  },
  react: {
    // DB 翻译覆盖是异步 addResourceBundle 注入的，必须绑定 store 事件才能触发已挂载组件重渲染
    bindI18nStore: 'added removed',
  },
})

// Apply per-language side effects (dayjs locale + <html lang/dir>) for the initial language.
applyLocaleSideEffects(getInitialLanguage())


// 运行时从后端 i18n_messages 表拉取覆盖项（配置即数据：后台改文案免部署）。
const i18nOverridesLoaded = new Set<string>()
export async function loadI18nOverrides(locale: string) {
  if (i18nOverridesLoaded.has(locale)) return
  try {
    const res = await fetch(`/api/i18n/messages?locale=${encodeURIComponent(locale)}`)
    if (!res.ok) return
    const payload = await res.json()
    const msgs = payload?.data?.messages
    if (msgs && typeof msgs === 'object') {
      i18n.addResourceBundle(locale, 'translation', msgs as Record<string, string>, true, true)
      i18nOverridesLoaded.add(locale)
    }
  } catch {
    /* 网络失败时回退到打包语言包，不影响可用性 */
  }
}

// 唯一修改语言的入口：切换语言 + 持久化到 LANG_KEY + 清除所有旧标记。
export function setLanguage(lang: SupportedLanguage) {
  i18n.changeLanguage(lang)
  loadI18nOverrides(lang)
  applyLocaleSideEffects(lang)
  if (typeof window !== 'undefined') {
    localStorage.setItem(LANG_KEY, lang)
    localStorage.removeItem('i18nextLng')
    localStorage.removeItem('i18nextLngUserSet')
    localStorage.removeItem('i18nextLangExplicit')
  }
}

export default i18n

// Kick off runtime translation overrides for the initial language (after declarations above).
loadI18nOverrides(getInitialLanguage())
