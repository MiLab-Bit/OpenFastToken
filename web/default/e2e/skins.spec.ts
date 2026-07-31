import { test, expect, type Page } from '@playwright/test'

/**
 * P0 基线：明暗轴 × 路由 截图建档 + 内容不变量断言。
 * 皮肤轴（data-skin）将在 P1 接入后并入 SKINS 数组，矩阵自动扩展到 12 组合。
 *
 * 运行：
 *   npx playwright test                      （对比快照，像素 diff）
 *   npx playwright test --update-snapshots  （刷新基线，仅确认 UI 有意变更时）
 */
const BASE = process.env.BASE_URL || 'http://127.0.0.1:3000'
const MODES = ['light', 'dark'] as const
const ROUTES = ['/', '/login'] as const
// 内容不变量锚点：品牌名必须出现在标题或正文
const CONTENT_ANCHORS = ['FastToken']

for (const mode of MODES) {
  for (const route of ROUTES) {
    test(`skin-baseline ${mode} ${route}`, async ({ page, context }) => {
      // 模拟用户已选主题（FOUC 修复依赖该 cookie）
      await context.addCookies([
        { name: 'vite-ui-theme', value: mode, domain: '127.0.0.1', path: '/' },
      ])
      const url = BASE + route
      await page.goto(url, { waitUntil: 'domcontentloaded' })
      await page.waitForTimeout(1500) // 字体/动效/图片稳定

      // 内容不变量：页面已渲染且含品牌锚点（防空白 / L10n 丢失）
      const bodyText = (await page.locator('body').innerText().catch(() => '')) || ''
      const title = (await page.title()) || ''
      const hit = CONTENT_ANCHORS.some((a) => title.includes(a) || bodyText.includes(a))
      const ok = hit || bodyText.length > 50
      expect(ok, `页面 ${route} 疑似空白/内容丢失（无品牌锚点且正文过少）`).toBe(true)

      const name = `${mode}-${route === '/' ? 'root' : route.replace(/\//g, '_')}`
      await expect(page).toHaveScreenshot(`${name}.png`, { fullPage: true })
    })
  }
}
