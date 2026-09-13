import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const { chromium, webkit } = require(process.env.PLAYWRIGHT_MODULE || 'playwright')
const root = new URL('../.output/public/', import.meta.url)
const entries = Array.from({ length: 20 }, (_, i) => ({ ID: i + 1, Content: 'Mobile layout fixture', CreatedAt: '2026-09-09T00:00:00Z', User: { Name: 'Member' } }))

// Substitute known insets only in fixture CSS: desktop engines have no iPhone Home indicator.
for (const [engine, launcher] of Object.entries({ chromium, webkit })) {
  const browser = await launcher.launch(engine === 'chromium' ? { channel: 'chrome' } : {})
  try {
    for (const [width, height, top, bottom, side] of [[320, 700, 0, 0, 0], [390, 844, 47, 34, 0], [430, 932, 59, 34, 0], [740, 390, 0, 21, 44], [1024, 768, 0, 0, 0], [1440, 900, 0, 0, 0], [1920, 1080, 0, 0, 0]]) {
      const desktop = width >= 1024
      const page = await browser.newPage({ viewport: { width, height }, isMobile: !desktop, hasTouch: !desktop })
      await page.route('http://gmwe.test/**', async route => {
        const pathname = new URL(route.request().url()).pathname
        if (pathname.startsWith('/api/')) {
          const body = pathname.endsWith('/me') ? { user: { name: 'Member', subject: 'fixture' }, profile: { first_name: 'Test', last_name: 'Member', loaded: true }, csrf: 'fixture' } : { Data: pathname.endsWith('/users') ? [{ ID: 1, Name: 'Member' }] : entries, Total: entries.length }
          return route.fulfill({ json: body })
        }
        const asset = pathname.includes('.') ? pathname.slice(1) : 'index.html'
        let body = await readFile(new URL(asset, root))
        const ext = asset.split('.').pop()
        if (ext === 'css') {
          body = Buffer.from(body.toString().replace(/env\(safe-area-inset-(top|bottom|left|right)(?:,[^)]*)?\)/g, (_, edge) => `${({ top, bottom, left: side, right: side })[edge]}px`))
        }
        await route.fulfill({ body, contentType: ({ html: 'text/html', js: 'text/javascript', css: 'text/css', png: 'image/png' })[ext] || 'application/octet-stream' })
      })
      await page.goto('http://gmwe.test/hitokoto')
      await page.locator('.entry-list li').first().waitFor()
      assert.equal(await page.locator('meta[name=viewport]').count(), 1)
      assert.match(await page.locator('meta[name=viewport]').getAttribute('content'), /viewport-fit=cover/)
      const result = await page.evaluate(() => {
        const nav = document.querySelector('.app-navigation')
        const main = document.querySelector('.app-content')
        const shell = document.querySelector('.app-shell')
        const rect = node => { const r = node.getBoundingClientRect(); return { top: r.top, bottom: r.bottom, left: r.left, right: r.right } }
        return { nav: rect(nav), shell: rect(shell), main: rect(main), links: [...nav.querySelectorAll('a')].map(rect), scrollWidth: main.scrollWidth, clientWidth: main.clientWidth, topPadding: parseFloat(getComputedStyle(shell).paddingTop) }
      })
      assert.equal(result.topPadding, top)
      assert.ok(result.nav.bottom <= height + 1)
      assert.ok(Math.abs(result.nav.bottom - height) <= 1)
      if (desktop) {
        assert.ok(result.main.left >= result.nav.right - 1)
        assert.ok(result.main.right - result.main.left >= 800)
        assert.equal(await page.locator('.filters').evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length), 4)
      } else { assert.ok(result.main.bottom <= result.nav.top + 1) }
      assert.ok(result.scrollWidth <= result.clientWidth)
      for (const link of result.links) {
        assert.ok(link.bottom <= height - bottom + 1)
        assert.ok(link.left >= side && link.right <= width - side)
        assert.ok(link.bottom - link.top >= (desktop ? 56 : 64))
      }
      console.log('PASS', engine, { width, height, top, bottom, side })
      if (process.env.LAYOUT_SCREENSHOT && [390, 1440].includes(width)) await page.screenshot({ path: `${process.env.LAYOUT_SCREENSHOT}-${engine}-${width}.png` })
      await page.getByRole('link', { name: 'Account', exact: true }).click()
      await page.getByRole('heading', { name: 'Test Member', exact: true }).waitFor()
      assert.equal(await page.locator('.account-page').getByText('Pocket-ID', { exact: true }).count(), 0)
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true)
      await page.close()
    }
  } finally { await browser.close() }
}
