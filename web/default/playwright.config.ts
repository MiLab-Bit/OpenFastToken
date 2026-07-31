import { defineConfig, devices } from '@playwright/test'

/**
 * FastToken 视觉回归测试基座（P0 起）
 * - 对 明暗轴 × 皮肤轴 的组合截图建档（首次运行生成 baseline，之后做 diff）
 * - 内容不变量断言：确保页面可见文案与基线一致（L10n/内容不被破坏）
 * 运行：npx playwright test            （生成/对比快照）
 *       npx playwright test --update-snapshots （刷新基线，仅在确认 UI 有意变更时）
 */
export default defineConfig({
  testDir: './e2e',
  snapshotDir: './e2e/__snapshots__',
  timeout: 60_000,
  expect: { toHaveScreenshot: { maxDiffPixelRatio: 0.02 } },
  use: {
    baseURL: process.env.BASE_URL || 'http://127.0.0.1:3000',
    viewport: { width: 1280, height: 900 },
    // FOUC 修复依赖 cookie；测试前置 cookie 模拟用户已选主题
    actionTimeout: 15_000,
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
