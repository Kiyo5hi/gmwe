# GMWE

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
