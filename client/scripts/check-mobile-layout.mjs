import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const { chromium, webkit } = require(process.env.PLAYWRIGHT_MODULE || 'playwright')
const root = new URL('../.output/public/', import.meta.url)
const examples = ['今天也想和你一起吃晚饭。', '把平凡的小事记下来，回头看的时候，原来已经一起走了这么远。', '周末的计划：睡个懒觉，再慢慢出门。', '下次还要一起去海边。']
const entries = Array.from({ length: 20 }, (_, i) => ({ ID: i + 1, Content: examples[i % examples.length], CreatedAt: '2026-09-09T00:00:00Z', User: { Name: i % 2 ? 'Kiyoshi Guo' : 'Acey Tong' } }))

// Substitute known insets only in fixture CSS: desktop engines have no iPhone Home indicator.
for (const [engine, launcher] of Object.entries({ chromium, webkit })) {
  const browser = await launcher.launch(engine === 'chromium' ? { channel: 'chrome' } : {})
  try {
    for (const [width, height, top, bottom, side] of [[320, 700, 0, 0, 0], [390, 844, 47, 34, 0], [430, 932, 59, 34, 0], [740, 390, 0, 21, 44], [768, 900, 0, 0, 0], [1024, 768, 0, 0, 0], [1440, 900, 0, 0, 0], [1920, 1080, 0, 0, 0]]) {
      const desktop = width >= 768
      const page = await browser.newPage({ viewport: { width, height }, isMobile: !desktop, hasTouch: !desktop })
      await page.route('http://gmwe.test/**', async route => {
        const pathname = new URL(route.request().url()).pathname
        if (pathname.startsWith('/api/')) {
          const body = pathname.endsWith('/me') ? { user: { name: 'Member', subject: 'fixture' }, profile: { first_name: 'Test', last_name: 'Member', loaded: true }, csrf: 'fixture' } : { Data: pathname.endsWith('/users') ? [{ ID: 1, Name: 'Member' }] : entries, Total: entries.length }
          return route.fulfill({ json: pathname === '/api/v1/hitokoto' ? { Data: entries[1] } : body })
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
      if (desktop) {
        assert.ok(result.nav.bottom <= result.main.top + 1)
        assert.ok(result.main.right - result.main.left >= 700)
        const brand = await page.locator('.wordmark').boundingBox()
        assert.ok(brand.x + brand.width < result.nav.left)
      } else { assert.ok(result.main.bottom <= result.nav.top + 1); assert.ok(Math.abs(result.nav.bottom - height) <= 1) }
      assert.equal(await page.locator('.filters').isVisible(), false)
      await page.getByRole('button', { name: '筛选条件', exact: true }).click()
      assert.equal(await page.locator('.filters').isVisible(), true)
      assert.equal(await page.locator('.filters').evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length), desktop ? 4 : 2)
      assert.equal(await page.locator('.app-content').evaluate(el => el.scrollWidth <= el.clientWidth), true)
      await page.getByRole('button', { name: '筛选条件', exact: true }).click()
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
      if (process.env.LAYOUT_SCREENSHOT && [390, 1440].includes(width)) {
        for (const theme of ['light', 'dark']) {
          await page.evaluate(theme => { document.documentElement.dataset.theme = theme }, theme)
          await page.screenshot({ path: `${process.env.LAYOUT_SCREENSHOT}-${engine}-${width}-account-${theme}.png` })
          await page.getByRole('link', { name: 'Hitokoto', exact: true }).click()
          await page.locator('.entry-list li').first().waitFor()
          await page.screenshot({ path: `${process.env.LAYOUT_SCREENSHOT}-${engine}-${width}-hitokoto-${theme}.png` })
          await page.getByRole('link', { name: 'Account', exact: true }).click()
          await page.getByRole('heading', { name: 'Test Member', exact: true }).waitFor()
        }
      }
      await page.getByRole('link', { name: 'Story', exact: true }).click()
      await page.locator('.story-timeline li').first().waitFor()
      assert.equal(await page.locator('.app-content').evaluate(el => el.scrollWidth <= el.clientWidth), true)
      await page.getByRole('link', { name: 'Home', exact: true }).click()
      await page.locator('blockquote').waitFor()
      assert.equal(await page.locator('.app-content').evaluate(el => el.scrollWidth <= el.clientWidth), true)
      if (process.env.LAYOUT_SCREENSHOT && [390, 1440].includes(width)) await page.screenshot({ path: `${process.env.LAYOUT_SCREENSHOT}-${engine}-${width}-home.png` })
      await page.close()
    }
  } finally { await browser.close() }
}
