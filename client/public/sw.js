/* Network-only: never persist private pages, API replies or credentials. */
self.addEventListener('install', () => self.skipWaiting())
self.addEventListener('activate', event => event.waitUntil(self.clients.claim()))
self.addEventListener('fetch', event => {
  if (event.request.mode !== 'navigate' || event.request.method !== 'GET') { return }
  event.respondWith(fetch(event.request).catch(() => new Response(
    '<!doctype html><html lang="en"><meta name="viewport" content="width=device-width,initial-scale=1"><title>GMWE</title><style>body{font:18px system-ui;margin:3rem;line-height:1.6}a{color:#166345}</style><h1>GMWE</h1><p>You are offline.</p><a href="/">Try again</a></html>',
    { status: 503, headers: { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store' } }
  )))
})
