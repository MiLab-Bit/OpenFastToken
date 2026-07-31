import { test, expect } from '@playwright/test'

test('probe sunset', async ({ page, context }) => {
  const errs: string[] = []
  page.on('pageerror', (e) =>
    errs.push(
      'pageerror:' +
        e.message +
        '\nfile:' +
        (e as Error & { filename?: string }).filename +
        '\nline:' +
        (e as Error & { lineno?: number }).lineno +
        '\nstack:' +
        (e.stack || 'no-stack'),
    ),
  )
  page.on('console', (m) => {
    if (m.type() === 'error' || m.type() === 'warning') {
      const loc = m.location()
      errs.push(
        m.type() + ':' + m.text() + ' @ ' + (loc.url || '?') + ':' + (loc.lineNumber || '?'),
      )
    }
  })
  await context.addCookies([
    { name: 'vite-ui-skin', value: 'sunset', url: 'http://127.0.0.1:3000' },
  ])
  await page.goto('http://127.0.0.1:3000/')
  await page.waitForTimeout(3000)
  const ds = await page.locator('html').evaluate((el) => el.getAttribute('data-skin'))
  const styleLen = await page.evaluate(
    () => (document.getElementById('db-skin-overrides')?.textContent?.length ?? 0),
  )
  const bodyBg = await page.evaluate(
    () => getComputedStyle(document.body).backgroundColor,
  )
  const hasSunset = await page.evaluate(
    () => !!document.getElementById('db-skin-overrides')?.textContent?.includes('sunset'),
  )
  console.log('DATA-SKIN=', ds)
  console.log('STYLE_LEN=', styleLen)
  console.log('HAS_SUNSET=', hasSunset)
  console.log('BODY_BG=', bodyBg)
  console.log('ERRORS=', JSON.stringify(errs))
})