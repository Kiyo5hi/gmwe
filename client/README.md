# GMWE static client

The four views use Vue 3, Vue Router and native CSS, built by Vite. Components,
reactive helpers and routes have explicit imports; there is no SSR or Node server
in production. `routes.ts` is shared by the router and the build script that emits
Caddy's existing direct-entry HTML paths. Private assets remain behind Caddy's
session check. Do not add a public SPA fallback around that check.

Use Node 24 and npm 11.19.1 with the committed `package-lock.json`:

```sh
npm ci
npm run check
npm run build
npm run preview
```

`npm run dev` serves localhost:3000 and proxies `/api` and `/login` to the local
Go API on localhost:5500. Real OIDC authentication requires the registered origin;
use the isolated browser fixtures for local UI work. Preview is a static test
server and does not provide production authentication.

Vue's type checker currently requires the TypeScript 6 compiler API; TypeScript 7
is not compatible with this pinned checker. Dependency updates must pass `check`
before replacing that pin. PWA icons are generated from the existing logo at build
or dev startup. `dist/` contains only public build artifacts, not source, secrets or
Git metadata; production still authenticates private content before serving it.

```sh
npx playwright install --with-deps chromium webkit
npm run test:browser
```

CI runs both browser engines before publishing an image. The fixtures cover slow
profiles, session-guarded navigation, safe areas at 320–1920px, four views, dark
mode, filters, paging, CSRF writes, failed/duplicate submissions, drafts and logout.
They never write production data. The service worker remains network-only and
never caches private data or retries writes. Real iPhone/Passkey login remains a
separate acceptance step.

For an existing browser installation, set `PLAYWRIGHT_MODULE` to its module path;
`BROWSER=chromium` or `BROWSER=webkit` selects one engine. An optional
`CHROMIUM_EXECUTABLE_PATH` selects a local Chromium binary. `LAYOUT_SCREENSHOT`
selects a screenshot output prefix. The layout/startup fixtures accept
`CLIENT_OUTPUT` for comparisons with a retained static build.
