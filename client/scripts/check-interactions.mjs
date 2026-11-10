import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const { chromium, webkit } = require(process.env.PLAYWRIGHT_MODULE || 'playwright')
const root = new URL('../dist/', import.meta.url)
const members = [{ ID: 1, Name: 'Member' }, { ID: 2, Name: 'Second member' }]
const account = { user: { name: 'Member', subject: 'fixture', user_id: 1 }, csrf: 'fixture-csrf', profile: { first_name: 'Test', last_name: 'Member', loaded: true } }

for (const [engine, launcher] of Object.entries({ chromium, webkit }).filter(([name]) => !process.env.BROWSER || name === process.env.BROWSER)) {
  const browser = await launcher.launch(engine === 'chromium' && process.env.CHROMIUM_EXECUTABLE_PATH ? { executablePath: process.env.CHROMIUM_EXECUTABLE_PATH } : {})
  try {
    const context = await browser.newContext({ serviceWorkers: 'block', viewport: { width: 390, height: 844 } })
    const page = await context.newPage()
    const errors = []
    page.on('pageerror', error => errors.push(error.message))
    let writeStatus = 201, malformed = false, sessionStatus = 204, logoutStatus = 204
    const writes = [], lists = []
    await page.route('http://gmwe.test/**', async route => {
      const request = route.request(), url = new URL(request.url()), path = url.pathname
      if (path === '/login') return route.fulfill({ contentType: 'text/html', body: '<h1>Login fixture</h1>' })
      if (path === '/api/auth/session') {
        if (sessionStatus === 0) return route.abort('connectionfailed')
        return route.fulfill({ status: sessionStatus })
      }
      if (path === '/api/auth/me') return route.fulfill({ json: account })
      if (path.startsWith('/api/auth/avatar/')) return route.fulfill({ status: 404 })
      if (path === '/api/auth/logout') {
        assert.equal(request.method(), 'POST')
        assert.equal(request.headers()['x-csrf-token'], account.csrf)
        return route.fulfill({ status: logoutStatus })
      }
      if (path === '/api/v1/users') return route.fulfill({ json: { Data: members } })
      if (path === '/api/v1/hitokoto' && request.method() === 'POST') {
        assert.equal(request.headers()['x-csrf-token'], account.csrf)
        assert.equal(request.headers()['content-type'], 'application/json')
        const body = request.postDataJSON()
        assert.deepEqual(Object.keys(body).sort(), ['Content', 'UserID'])
        writes.push(body)
        return route.fulfill({ status: writeStatus, json: malformed ? { Data: {} } : { Data: { ID: 100, Content: body.Content, User: members.find(member => member.ID === body.UserID) } } })
      }
      if (path === '/api/v1/hitokotos') {
        lists.push(Object.fromEntries(url.searchParams))
        return route.fulfill({ json: { Data: [{ ID: 1, Content: 'Synthetic entry', CreatedAt: '2026-11-10T00:00:00Z', User: members[0] }], Total: 21 } })
      }
      if (path === '/api/v1/hitokoto') return route.fulfill({ json: { Data: { Content: 'Synthetic entry', User: members[0] } } })
      assert.equal(request.method(), 'GET', 'unexpected mutating request')
      const asset = path === '/' ? 'index.html' : path.includes('.') ? path.slice(1) : path.replace(/^\//, '').replace(/\/$/, '') + '/index.html'
      return route.fulfill({ body: await readFile(new URL(asset, root)), contentType: ({ js: 'text/javascript', css: 'text/css', html: 'text/html', png: 'image/png', webp: 'image/webp' })[asset.split('.').pop()] || 'application/octet-stream' })
    })
    await page.goto('http://gmwe.test/hitokoto/')
    await page.locator('.entry-list li').waitFor()
    assert.equal(await page.title(), 'Hitokoto | GMWE')
    await page.getByRole('button', { name: '筛选条件', exact: true }).click()
    await page.locator('#entry-search').fill('  filter text  ')
    await page.locator('.filters select').first().selectOption('2')
    await page.locator('.filters select').nth(1).selectOption('oldest')
    await page.locator('.date-filter input').first().fill('2026-10-01')
    await page.locator('.date-filter input').last().fill('2026-11-10')
    await page.getByRole('button', { name: '筛选', exact: true }).click()
    await page.waitForFunction(() => !document.querySelector('input[type=search]').disabled)
    assert.deepEqual(lists.at(-1), { page: '1', q: 'filter text', user_id: '2', from: '2026-10-01', to: '2026-11-10', sort: 'oldest' })
    await page.getByRole('button', { name: '下一页', exact: true }).click()
    await page.waitForFunction(() => document.querySelector('.pagination span').textContent.includes('2 / 2'))
    assert.equal(lists.at(-1).page, '2')
    await page.locator('button[aria-controls=entry-composer]').click()
    await page.locator('#entry-member option').first().waitFor({ state: 'attached' })
    await page.locator('#entry-member').selectOption('2')
    await page.locator('#entry').fill('  Saved fixture 🙂  ')
    await page.getByRole('button', { name: '保存', exact: true }).click()
    await page.getByRole('heading', { name: '已保存', exact: true }).waitFor()
    assert.deepEqual(writes.at(-1), { Content: 'Saved fixture 🙂', UserID: 2 })
    assert.equal(await page.locator('#entry').inputValue(), '')
    assert.equal(await page.evaluate(() => sessionStorage.getItem('gmwe:draft:fixture')), null)
    await page.waitForFunction(() => !document.querySelector('input[type=search]').disabled)
    assert.deepEqual(lists.at(-1), { page: '1', q: '', user_id: '', from: '', to: '', sort: 'newest' })

    for (const [status, message] of [[409, '这条内容已存在'], [400, '请检查所选成员'], [413, '请检查所选成员'], [500, '尚未确认保存成功'], [201, '尚未确认保存成功']]) {
      writeStatus = status; malformed = status === 201
      await page.locator('#entry').fill('Preserve draft ' + status)
      const count = writes.length
      await page.getByRole('button', { name: '保存', exact: true }).click()
      await page.locator('#entry-composer [role=alert]').filter({ hasText: message }).waitFor()
      assert.equal(await page.locator('#entry').inputValue(), 'Preserve draft ' + status)
      assert.equal(writes.length, count + 1, 'failed writes are never automatically retried')
    }
    await page.getByRole('link', { name: 'Story', exact: true }).click()
    await page.locator('.story-page').waitFor()
    await page.getByRole('link', { name: 'Hitokoto', exact: true }).click()
    await page.locator('.entry-list li').waitFor()
    await page.locator('button[aria-controls=entry-composer]').click()
    await page.waitForFunction(() => document.querySelector('#entry').value === 'Preserve draft 201')
    assert.equal(await page.locator('#entry-member').inputValue(), '2')

    for (const status of [401, 403]) {
      if (status === 403) {
        await page.reload()
        await page.locator('.entry-list li').waitFor()
        await page.locator('button[aria-controls=entry-composer]').click()
        await page.waitForFunction(() => !document.querySelector('#entry').disabled)
      }
      writeStatus = status; malformed = false
      await page.getByRole('button', { name: '保存', exact: true }).click()
      await page.locator('#entry-composer [role=alert]').filter({ hasText: '草稿已保留' }).waitFor()
      assert.equal(await page.locator('#entry').isDisabled(), true)
      assert.ok(await page.evaluate(() => sessionStorage.getItem('gmwe:draft:fixture')))
    }
    for (const status of [503, 0]) {
      sessionStatus = status
      await page.getByRole('link', { name: 'Account', exact: true }).click()
      await page.locator('.navigation-error').waitFor()
      assert.match(page.url(), /\/hitokoto\/?$/)
      assert.equal(await page.locator('.app-shell').count(), 1)
    }
    sessionStatus = 204
    await page.getByRole('link', { name: 'Account', exact: true }).click()
    await page.getByRole('heading', { name: 'Test Member', exact: true }).waitFor()
    logoutStatus = 500
    await page.getByRole('button', { name: '退出登录', exact: true }).click()
    await page.getByRole('alert').filter({ hasText: '退出失败' }).waitFor()
    assert.ok(await page.evaluate(() => sessionStorage.getItem('gmwe:draft:fixture')))
    await page.evaluate(() => sessionStorage.setItem('gmwe:draft:another-user', 'unrelated fixture'))
    logoutStatus = 204
    await page.getByRole('button', { name: '退出登录', exact: true }).click()
    await page.waitForURL('http://gmwe.test/login')
    assert.equal(await page.evaluate(() => sessionStorage.getItem('gmwe:draft:fixture')), null)
    assert.equal(await page.evaluate(() => sessionStorage.getItem('gmwe:draft:another-user')), 'unrelated fixture')
    assert.deepEqual(errors, [])
    console.log('PASS', engine, 'direct routes, filters/pagination, CSRF writes, error/draft preservation, session failures and logout')
    await context.close()
  } finally { await browser.close() }
}
