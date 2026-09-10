import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const { chromium, webkit } = require(process.env.PLAYWRIGHT_MODULE || 'playwright')
const root = new URL('../.output/public/', import.meta.url)
for (const [engine, launcher] of Object.entries({ chromium, webkit })) {
  const browser = await launcher.launch(engine === 'chromium' ? { channel: 'chrome' } : {})
  let release
  try {
    const page = await browser.newPage({ viewport: { width: 390, height: 844 }, isMobile: true })
    let accountCalls = 0
    let sessionCalls = 0
    let expired = false
    const pendingProfile = new Promise(resolve => { release = resolve })
    await page.route('http://gmwe.test/**', async route => {
      const path = new URL(route.request().url()).pathname
      if (path === '/login') return route.fulfill({ contentType: 'text/html', body: '<h1>Login fixture</h1>' })
      if (path === '/api/auth/session') {
        sessionCalls++
        return route.fulfill({ status: expired ? 401 : 204 })
      }
      if (path === '/api/auth/me') {
        accountCalls++
        await pendingProfile
        return route.fulfill({ json: { user: { name: 'Member', subject: 'fixture', user_id: 1 }, csrf: 'fixture', profile: { first_name: 'Member', last_name: '', state: 'ready', loaded: true } } })
      }
      if (path.startsWith('/api/')) return route.fulfill({ json: { Data: path.endsWith('/users') ? [{ ID: 1, Name: 'Member' }] : [{ ID: 1, Content: 'Fixture entry', CreatedAt: '2026-09-09', User: { Name: 'Member' } }], Total: 1 } })
      const asset = path.includes('.') ? path.slice(1) : 'index.html'
      const ext = asset.split('.').pop()
      return route.fulfill({ body: await readFile(new URL(asset, root)), contentType: ({ js: 'text/javascript', css: 'text/css', html: 'text/html', png: 'image/png' })[ext] || 'application/octet-stream' })
    })
    await page.goto('http://gmwe.test/hitokoto')
    await page.locator('.entry-list li').waitFor()
    assert.equal(accountCalls, 0, 'hidden composer must not request a profile')
    assert.equal(sessionCalls, 0, 'initial document is already authenticated by Caddy')
    const accountRequest = page.waitForRequest(request => request.url().endsWith('/api/auth/me'))
    await page.locator('.app-navigation a[href="/account"]').click()
    await accountRequest
    await page.locator('.account-page').waitFor()
    assert.equal(await page.locator('.app-shell').count(), 1, 'slow profile must not hide the app shell')
    assert.ok(sessionCalls >= 1)
    release()
    await page.locator('.app-navigation a[href="/hitokoto"]').click()
    await page.locator('.entry-list li').waitFor()
    await page.locator('button[aria-controls="entry-composer"]').click()
    await page.locator('#entry-member option').waitFor({ state: 'attached' })
    const loadedCalls = accountCalls
    await page.locator('button[aria-controls="entry-composer"]').click()
    await page.locator('button[aria-controls="entry-composer"]').click()
    assert.equal(accountCalls, loadedCalls, 'reopening composer should preserve loaded state')
    expired = true
    await page.locator('.app-navigation a[href="/our-story"]').click()
    await page.waitForURL('http://gmwe.test/login')
    console.log('PASS', engine, 'initial shell/list without profile, gated tabs, slow Account, lazy composer, expired redirect')
  } finally { release?.(); await browser.close() }
}
