/*
Copyright (C) 2023-2026 FastToken

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
import dayjs from '@/lib/dayjs'

// App language -> BCP-47 locale for Intl (number/date/currency formatting)
const INTL_LOCALE: Record<string, string> = {
  zh: 'zh-CN',
  en: 'en-US',
  ja: 'ja-JP',
  ru: 'ru-RU',
  fr: 'fr-FR',
  vi: 'vi-VN',
  ar: 'ar',
}

// App language -> dayjs locale name (must match an imported dayjs/locale pack)
const DAYJS_LOCALE: Record<string, string> = {
  zh: 'zh-cn',
  en: 'en',
  ja: 'ja',
  ru: 'ru',
  fr: 'fr',
  vi: 'vi',
  ar: 'ar',
}

const RTL_LANGS = new Set<string>(['ar'])

/** Current active app language (e.g. 'zh', 'en', 'ar'). */
export function getAppLanguage(): string {
  return i18n.resolvedLanguage || i18n.language || 'zh'
}

/** BCP-47 locale for Intl.NumberFormat / Intl.DateTimeFormat. */
export function getIntlLocale(): string {
  return INTL_LOCALE[getAppLanguage()] || 'zh-CN'
}

/** dayjs locale name for relative-time / calendar formatting. */
export function getDayjsLocale(): string {
  return DAYJS_LOCALE[getAppLanguage()] || 'zh-cn'
}

/** Whether the given (or current) language is right-to-left. */
export function isRtlLang(lang?: string): boolean {
  return RTL_LANGS.has(lang || getAppLanguage())
}

/**
 * Apply per-language side effects: switch dayjs locale and set <html lang/dir>.
 * Called once at startup (initial language) and on every setLanguage().
 */
export function applyLocaleSideEffects(lang: string): void {
  const dj = DAYJS_LOCALE[lang] || 'zh-cn'
  try {
    dayjs.locale(dj)
  } catch {
    /* locale pack may be missing; fall back to previous */
  }
  if (typeof document !== 'undefined') {
    document.documentElement.lang = lang
    document.documentElement.dir = isRtlLang(lang) ? 'rtl' : 'ltr'
  }
}
