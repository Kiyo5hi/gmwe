# GMWE

## Responsive UI

Desktop and tablet navigation share the masthead, without a persistent sidebar.
Phones retain bottom navigation with safe-area padding. Reading widths are
specific to each view: Hitokoto has an 800px reading column, optional expanded
filters and member/date hierarchy; Story is a dated timeline; Account is a
640px profile surface. Home pairs the current entry with a compact player.
The player loads audio only on request and pauses when its view is left.
Light/dark colors are shared across controls and pages. Account shows the
member's name, not the identity provider brand.

After generating the client, run `node scripts/check-mobile-layout.mjs` and
`node scripts/check-startup.mjs` from `client/`. Set `PLAYWRIGHT_MODULE` when
Playwright is installed outside this project. The layout fixture covers Chromium
and WebKit at 320-1920px, including simulated iPhone insets, all four views,
collapsed/expanded filters and Account navigation.
`LAYOUT_SCREENSHOT` optionally selects an output path prefix. These are isolated
fixtures, not evidence of an actual iPhone login.

## Storage

The API uses SQLite through GORM. Set `DB_URI=file:/app/data/gmwe.db?mode=rw`
in production: a missing database fails startup rather than creating an empty
replacement. Mount the entire database directory on local persistent storage,
including SQLite's WAL files. Do not put the live database on SMB/NFS.

Connections enable foreign keys, a five-second busy timeout, WAL and FULL
synchronous durability. One database connection serializes this small app's
operations. Startup migrates the schema but never creates default users or
passwords. Existing user IDs and hashes must be imported before production use.

Text uniqueness uses `gmwe_unicode_ci`, implemented with Go x/text's Unicode
loose comparison and trailing ASCII-space trimming. Case/accent variants compare
equal. This is an explicit application policy, not an exact implementation of
MySQL's older Unicode 5.2 collation. Tests cover case, combining accents, author
relationships, ID sequences, transactions, restart schema migration and backup.
Use the application storage tool for integrity checks: generic SQLite clients
do not register this custom collation.

The image contains `/app/api/storage`:

```text
storage import /path/new.db < snapshot.json
storage check /path/existing.db
storage backup /path/existing.db /path/new-backup.db
```

Import requires a new file and preserves IDs, times, soft-deletes, password
hashes and next-ID sequences. Its JSON input is secret material. Check prints
only counts, a dataset digest and the SQLite version. Backup uses SQLite's
online backup API, checks the copy and checkpoints its WAL; never copy just the
main file of a running database. Protect backup files and their parent directory.

Build with `docker build -f deploy/Dockerfile .`; the API build runs
`go test ./api/...` and includes the C compiler needed by the official GORM
SQLite driver. The static Nuxt 3 client uses Node 24 and npm 11.19.1;
run `npm ci` followed by `npm run generate` in `client/`. Commit
`package-lock.json` with dependency changes. Do not regenerate a separate pnpm
lockfile: production and local builds use the same npm lock.

Authentication is separate from this storage change. The API must not be exposed
for anonymous writes. The managed homelab deployment keeps its public/private
Caddy write guards until Pocket-ID integration has been verified end to end.

## Pocket ID avatars

Account and Hitokoto author avatars use the existing server-side writer mapping.
An authenticated `/api/auth/avatar/:userID` route resolves only configured members
to Pocket ID's public profile-picture endpoint on the configured issuer. No
administrator API key or new database is needed. Unknown members return 404,
anonymous callers 401. The browser sends no referrer and falls back to initials
if the image fails. A five-minute URL bucket bounds stale CDN images on subsequent
page loads; an already open page does not receive pushed updates. This adapter
uses Pocket ID's avatar API, not a provider-independent OIDC picture claim.
