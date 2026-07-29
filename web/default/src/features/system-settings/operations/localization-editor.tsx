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

For commercial licensing, please contact support@example.com
*/
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import zh from '@/l10n/locales/zh.json'
import { STATIC_I18N_KEYS } from '@/l10n/static-keys'

const TARGET_LOCALES = [
  { code: 'en', label: 'English' },
  { code: 'ja', label: '日本語' },
  { code: 'ru', label: 'Русский' },
  { code: 'fr', label: 'Français' },
  { code: 'vi', label: 'Tiếng Việt' },
  { code: 'ar', label: 'العربية' },
]

// 全部可翻译 key（源文 = 中文），用于搜索与回退展示。
const ALL_KEYS: string[] = Array.from(
  new Set([...Object.keys(zh as Record<string, string>), ...STATIC_I18N_KEYS]),
)

export function LocalizationEditor() {
  const { t } = useTranslation()
  const [locale, setLocale] = useState('en')
  const [overrides, setOverrides] = useState<Record<string, string>>({})
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [query, setQuery] = useState('')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    api
      .get('/api/i18n/messages', { params: { locale } })
      .then((res: any) => {
        if (cancelled) return
        const msgs: Record<string, string> = res?.data?.messages ?? {}
        setOverrides(msgs)
        setDrafts(msgs)
      })
      .catch(() => {
        if (cancelled) return
        setOverrides({})
        setDrafts({})
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [locale])

  // 空搜索时只显示已覆盖项（列表小）；有搜索词时从全量 key 中匹配（上限 200）。
  const visibleKeys = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return Object.keys(overrides)
    return ALL_KEYS.filter((k) => k.toLowerCase().includes(q)).slice(0, 200)
  }, [query, overrides])

  const dirty = useMemo(() => {
    const changed: Record<string, string> = {}
    for (const k of Object.keys(drafts)) {
      if (drafts[k] !== overrides[k]) changed[k] = drafts[k]
    }
    return changed
  }, [drafts, overrides])

  const save = async () => {
    if (Object.keys(dirty).length === 0) return
    setSaving(true)
    try {
      await api.put('/api/option/i18n', { locale, messages: dirty })
      setOverrides(drafts)
      toast.success(t('已保存本地化覆盖'))
    } catch {
      toast.error(t('保存失败'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className='space-y-4'>
      <div className='flex flex-wrap items-center gap-3'>
        <select
          value={locale}
          onChange={(e) => setLocale(e.target.value)}
          className='h-9 rounded-md border bg-background px-3 text-sm'
        >
          {TARGET_LOCALES.map((l) => (
            <option key={l.code} value={l.code}>
              {l.label}
            </option>
          ))}
        </select>
        <Input
          placeholder={t('搜索文案（中文源文）')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className='max-w-xs'
        />
        <Button onClick={save} disabled={saving || Object.keys(dirty).length === 0}>
          {saving ? t('保存中...') : t('保存')}
          {Object.keys(dirty).length > 0 ? ` (${Object.keys(dirty).length})` : ''}
        </Button>
        <span className='text-xs text-muted-foreground'>
          {t('共')} {Object.keys(overrides).length} {t('条已覆盖')}
        </span>
      </div>

      {loading ? (
        <p className='text-sm text-muted-foreground'>{t('加载中...')}</p>
      ) : (
        <div className='max-h-[60vh] space-y-2 overflow-auto rounded-md border p-2'>
          {visibleKeys.length === 0 ? (
            <p className='text-sm text-muted-foreground'>
              {query ? t('无匹配文案') : t('该语言暂无覆盖，搜索后添加')}
            </p>
          ) : (
            visibleKeys.map((key) => (
              <div
                key={key}
                className='grid grid-cols-1 gap-1 border-b py-2 last:border-0 md:grid-cols-[1fr_1fr] md:items-center md:gap-3'
              >
                <div className='break-all text-xs text-muted-foreground'>{key}</div>
                <Input
                  value={drafts[key] ?? ''}
                  placeholder={key}
                  onChange={(e) =>
                    setDrafts((d) => ({ ...d, [key]: e.target.value }))
                  }
                />
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
