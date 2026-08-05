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

For commercial licensing, please contact hello@fasttoken.example.com
*/
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { getCookie, removeCookie, setCookie } from '@/lib/cookies'

/**
 * 皮肤（Skin）系统 · P1 运行时引擎 + P2 数据驱动
 * --------------------------------------------------------------------------
 * 皮肤 = 一整集设计语言变量，由单个 data 属性 <html data-skin="..."> 驱动，
 * 经 styles/skins.css（内置 4 套）作用到全站。明暗轴由既有 ThemeProvider 负责，
 * 二者正交、互不覆盖。
 *
 * P2 起皮肤清单变为「配置即数据」：
 *   - 内置 4 套（neo/aurora/classic/midnight）仍是 bundled 基线，保证零回归与无 FOUC；
 *   - 数据库 ui_skins 表为「覆盖 + 扩展」层：可禁用内置、改名/覆盖其 css、或新增第 5 套；
 *   - 前端挂载时 GET /api/ui/skins 拉取，合并后注入一个 <style id="db-skin-overrides">
 *     （置于 bundled skins.css 之后生效）。运营改皮肤免部署，重载配置即热更新。
 */

export interface SkinMeta {
  id: string
  name: string
  description: string
}

/** 服务端下发的皮肤定义（含 css）。 */
export interface ServerSkin {
  key: string
  name: string
  description: string
  css: string
  enabled: boolean
  isDefault: boolean
  priority: number
}

/** 内置 4 套皮肤（DB 为空时的基线；DB 可禁用/改名/覆盖 css）。 */
export const BUNDLED_SKINS: SkinMeta[] = [
  { id: 'neo', name: 'Neo', description: '现代 SaaS / 专业（默认）' },
  { id: 'aurora', name: 'Aurora', description: '科技紫 / 渐变玻璃 / 年轻' },
  { id: 'classic', name: 'Classic', description: '暖纸 / 企业 / 阅读优先' },
  { id: 'midnight', name: 'Midnight', description: '深海军蓝 / 开发者 / 青绿' },
  { id: 'sunset', name: 'Sunset', description: '暖橙 / 陶土 / 活力' },
]

const SKIN_COOKIE_NAME = 'vite-ui-skin'
const SKIN_COOKIE_MAX_AGE = 60 * 60 * 24 * 365 // 1 year
const DB_SKIN_STYLE_ID = 'db-skin-overrides'

function getStoredSkin(): string | null {
  const value = getCookie(SKIN_COOKIE_NAME)
  return value || null
}

/** 把当前皮肤写到 <html data-skin>，供 skins.css 的 [data-skin=...] 选择器作用。 */
function applySkin(skin: string) {
  if (typeof document === 'undefined') return
  document.documentElement.setAttribute('data-skin', skin)
}

/**
 * 把服务端下发的 css（含内置覆盖 + 新增皮肤）注入一个 style 元素，置于 bundled
 * skins.css 之后生效。仅注入 enabled 且 css 非空的项；内置皮肤 css 为空则回退 bundled。
 */
function injectDbSkins(skins: ServerSkin[]) {
  if (typeof document === 'undefined') return
  let el = document.getElementById(DB_SKIN_STYLE_ID) as HTMLStyleElement | null
  if (!el) {
    el = document.createElement('style')
    el.id = DB_SKIN_STYLE_ID
    document.head.appendChild(el)
  }
  const css = skins
    .filter((s) => s.enabled && s.css && s.css.trim().length > 0)
    .map((s) => s.css)
    .join('\n')
  el.textContent = css
}

type SkinContextValue = {
  skin: string
  defaultSkin: string
  skins: SkinMeta[]
  setSkin: (skin: string) => void
  resetSkin: () => void
}

const INITIAL: SkinContextValue = {
  skin: 'neo',
  defaultSkin: 'neo',
  skins: BUNDLED_SKINS,
  setSkin: () => {},
  resetSkin: () => {},
}

const SkinContext = createContext<SkinContextValue>(INITIAL)

export function SkinProvider({ children }: { children: React.ReactNode }) {
  const [serverSkins, setServerSkins] = useState<ServerSkin[]>([])
  // 首次拉取解析完成前不要执行"皮肤失效→回退默认"，避免 DB-only 皮肤在异步加载窗口被错误回退成 neo。
  const [serverLoaded, setServerLoaded] = useState(false)
  const [skin, _setSkin] = useState<string>(() => getStoredSkin() || 'neo')

  // 拉取服务端皮肤定义（配置即数据），与内置 4 套合并。
  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const res = await fetch('/api/ui/skins')
        if (!res.ok) return
        const payload = await res.json()
        const list: ServerSkin[] = payload?.data?.skins || []
        if (cancelled) return
        setServerSkins(list)
        setServerLoaded(true)
      } catch {
        /* 网络失败回退到内置 4 套，不影响可用性 */
        if (!cancelled) setServerLoaded(true)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  // 计算有效皮肤列表：内置 4 套为基线，DB 可禁用/改名/覆盖 css，DB-only 新皮肤追加。
  const { effectiveSkins, defaultSkin } = useMemo(() => {
    const byKey = new Map<string, ServerSkin>()
    for (const s of serverSkins) byKey.set(s.key, s)
    const merged: SkinMeta[] = []
    let def = 'neo'
    for (const b of BUNDLED_SKINS) {
      const s = byKey.get(b.id)
      if (s && s.enabled === false) continue // DB 禁用了该内置皮肤
      merged.push({
        id: b.id,
        name: s?.name || b.name,
        description: s?.description || b.description,
      })
      if (s?.isDefault) def = b.id
    }
    // 追加 DB-only 新皮肤（key 不在内置中的）
    const builtinIds = new Set(BUNDLED_SKINS.map((b) => b.id))
    for (const s of serverSkins) {
      if (!builtinIds.has(s.key) && s.enabled) {
        merged.push({ id: s.key, name: s.name, description: s.description })
        if (s.isDefault) def = s.key
      }
    }
    return { effectiveSkins: merged, defaultSkin: def }
  }, [serverSkins])

  // 注入 DB css（覆盖/新增）。
  useEffect(() => {
    injectDbSkins(serverSkins)
  }, [serverSkins])

  // 若当前皮肤被禁用/失效，回退到默认皮肤。
  // 注意：必须等 serverLoaded 之后再判断，否则首次渲染时 serverSkins 尚为空，
  // 会把 DB-only 皮肤（cookie 已选）错误地回退成 neo。
  useEffect(() => {
    if (!serverLoaded) return
    const valid = effectiveSkins.some((s) => s.id === skin)
    if (!valid) _setSkin(defaultSkin)
  }, [effectiveSkins, skin, defaultSkin, serverLoaded])

  // 应用当前皮肤到 <html>；FOUC 脚本已在 hydration 前设过初值，这里保持同步。
  useEffect(() => {
    applySkin(skin)
  }, [skin])

  const setSkin = useCallback((next: string) => {
    setCookie(SKIN_COOKIE_NAME, next, SKIN_COOKIE_MAX_AGE)
    _setSkin(next)
  }, [])

  const resetSkin = useCallback(() => {
    // 清掉 cookie，回退到默认皮肤（默认可能来自数据库）。
    removeCookie(SKIN_COOKIE_NAME)
    _setSkin(defaultSkin)
  }, [defaultSkin])

  const value = useMemo<SkinContextValue>(
    () => ({
      skin,
      defaultSkin,
      skins: effectiveSkins,
      setSkin,
      resetSkin,
    }),
    [skin, defaultSkin, effectiveSkins, setSkin, resetSkin]
  )

  return <SkinContext.Provider value={value}>{children}</SkinContext.Provider>
}

export function useSkin(): SkinContextValue {
  return useContext(SkinContext)
}
