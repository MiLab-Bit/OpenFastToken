import { test, expect } from '@playwright/test'

/**
 * Functional end-to-end smoke test (complements the existing visual-regression
 * specs). Asserts the deployed SPA actually renders with the FastToken brand,
 * which catches blank-page / broken-bundle regressions that pure unit tests miss.
 */
test('页面渲染且包含品牌名 FastToken', async ({ page }) => {
  await page.goto('/')
  // 等待应用挂载
  await page.waitForLoadState('networkidle').catch(() => {})
  await expect(page.locator('html')).toContainText('FastToken')
})

test('健康检查端点可用', async ({ request }) => {
  const res = await request.get('/api/payment/status')
  expect(res.status()).toBe(200)
  const body = await res.json()
  expect(body).toHaveProperty('ready')
})
