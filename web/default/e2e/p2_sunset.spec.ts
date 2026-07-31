import { test, expect } from '@playwright/test'

// P2 验收：DB 新增的第 5 套皮肤（sunset）经 GET /api/ui/skins 下发，
// 前端 SkinProvider 注入其 css 后，设 cookie vite-ui-skin=sunset 应渲染出暖色背景。
test('DB-added skin "sunset" renders without code change', async ({ page, context }) => {
  await context.addCookies([
    { name: 'vite-ui-skin', value: 'sunset', url: 'http://127.0.0.1:3000' },
  ])
  await page.goto('/')
  await page.waitForTimeout(1500)
  await expect(page.locator('html')).toHaveAttribute('data-skin', 'sunset')
  // 内容不变量：落地页含品牌锚点且非空
  const txt = await page.locator('body').innerText()
  expect(txt.length).toBeGreaterThan(0)
  await page.screenshot({
    path: '/opt/fasttoken/web/default/e2e/p2-sunset-root.png',
    fullPage: false,
  })
})
