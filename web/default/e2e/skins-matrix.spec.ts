import { test, expect } from '@playwright/test'

/**
 * P1 皮肤矩阵：4 皮肤 × 2 明暗 × 2 路由 = 32 组合 截图建档 + 内容不变量断言。
 * 作为后续阶段（P3 组件打磨 / P5 全量验证）的回归基线。
 *
 * 运行：
 *   npx playwright test skins-matrix.spec.ts --update-snapshots   （首次建档）
 *   npx playwright test skins-matrix.spec.ts                     （回归比对）
 *
 * 注：neo 组合应与 P0 基线（skins.spec.ts）像素一致；aurora/classic/midnight
 * 为本次新增皮肤，首次建档即基线。
 */
const BASE = process.env.BASE_URL || 'http://127.0.0.1:3000'
const SKINS = ['neo', 'aurora', 'classic', 'midnight'] as const
const MODES = ['light', 'dark'] as const
const ROUTES = ['/', '/login'] as const
// 内容不变量锚点：品牌名必须出现在标题或正文
const CONTENT_ANCHORS = ['FastToken']

for (const skin of SKINS) {
  for (const mode of MODES) {
    for (const route of ROUTES) {
      test(`skin ${skin} ${mode} ${route}`, async ({ page, context }) => {
        // 模拟用户已选主题与皮肤（FOUC 脚本依赖这两个 cookie）
        await context.addCookies([
          { name: 'vite-ui-theme', value: mode, domain: '127.0.0.1', path: '/' },
          { name: 'vite-ui-skin', value: skin, domain: '127.0.0.1', path: '/' },
        ])
        const url = BASE + route
        await page.goto(url, { waitUntil: 'domcontentloaded' })
        await page.waitForTimeout(1500) // 字体/动效/图片稳定

        // 内容不变量：页面已渲染且含品牌锚点（防空白 / L10n 丢失）
        const bodyText =
          (await page.locator('body').innerText().catch(() => '')) || ''
        const title = (await page.title()) || ''
        const hit = CONTENT_ANCHORS.some(
          (a) => title.includes(a) || bodyText.includes(a)
        )
        const ok = hit || bodyText.length > 50
        expect(
          ok,
          `页面 ${skin}/${mode}/${route} 疑似空白/内容丢失（无品牌锚点且正文过少）`
        ).toBe(true)

        const name = `${skin}-${mode}-${
          route === '/' ? 'root' : route.replace(/\//g, '_')
        }`
        await expect(page).toHaveScreenshot(`${name}.png`, { fullPage: true })
      })
    }
  }
}
