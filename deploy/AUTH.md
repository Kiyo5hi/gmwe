# GMWE Browser Authentication

GMWE is private to the two configured Pocket-ID users. The client is public,
uses authorization code with S256 PKCE and an exact callback, and requests only
openid, write:hitokoto and offline_access for its existing GMWE API resource.
The refresh credential stays server-side; no client secret, pairing endpoint or
connection download is used. Native consent is required at interactive login.

Caddy serves /login, the non-sensitive logo and exact PWA manifest/icon/worker
paths without a session; see the explicit matcher in deploy/Caddyfile.
All other static files, including JavaScript, photographs and music, use its
native forward_auth check at /api/auth/page-access. Failed checks redirect to
/login; private responses are no-store. API data reads and writes independently
require the Secure, HttpOnly, host-only SameSite=Lax session cookie. Bare bearer
tokens, including old Shortcut credentials, are rejected even if cryptographically
valid. GET /api/v1/ping remains an unauthenticated liveness endpoint without data.
Unknown/retired downloads and pairing paths return 410 at Caddy.

The server verifies issuer, resource, client, signature, expiry, nonce, at_hash,
scope and the exact user allowlist. It derives authors from the configured subject
mapping. Writes and logout require matching Origin and session-bound CSRF.
Sessions use SCS's official SQLite store at required SESSION_DB_PATH, separate
from the content database. The directory/file are 0700/0600. Browser cookies are
persistent, Secure, HttpOnly and host-only. Absolute lifetime is 30 days, with
seven days of inactivity ending a session sooner. OAuth access expiry is still
enforced: the server renews via Pocket-ID and rechecks subject/client/scope/audience.
Per-session request locking covers load, rotation, save and logout in this
single-instance deployment. Revoked refresh grants end the session; transient
provider failures return 503, not anonymous access or silent write retries.
Logout destroys the local session immediately. It does not sign out Pocket-ID
globally. Server restarts retain sessions; copying/restoring the session database
can restore credentials, so exclude it from exported content and clear it during
disaster recovery unless deliberately retaining authenticated device sessions.
Revocation of provider membership is not instantaneous for an already-issued
session token: remove the local subject mapping and restart for emergency denial.

The submission page preserves unsuccessful input and never retries writes
automatically. Duplicate content is identified by the existing database constraint.
The last confirmed entry is shown only after a valid 201 response with its ID.
Drafts are per-account, tab-scoped sessionStorage (no tokens), restored for up to
24 hours and cleared after success or explicit sign-out. Storage restrictions may
prevent draft persistence across navigation; input remains in memory on failures.

Previously distributed Apple Shortcuts are retired. Their files can be removed
from the operator's devices; they cannot access this session-only API. Git history
retains the old credential-free distribution artifact, not runtime connection files.
No user database rows or identity-provider registrations are deleted during retirement.

Verification: go test ./api/... covers JWT, cookie/CSRF, callback replay, page checks,
old bearer rejection, author assignment and duplicate/input boundaries. Infra
candidate checks use an online database copy; production verification must assert
private page redirects, denied data reads/writes and retired downloads, not just
the public liveness response. Real Passkey login and intended user submission
remain separate acceptance steps. Never generate admin sessions to bypass them.

## Home Screen App

The manifest uses standalone mode and a root start URL. Icons are reproducibly
generated from the existing logo by sharp during generate/build/dev. Login also
includes the manifest and Apple touch icon, so installation needs no bypass.
The service worker is network-only, with a generic offline navigation response.
It never caches content, auth replies or writes, and has no background sync.
Cookie persistence is separate from PWA installation; clearing device website
data or provider revocation can still require login. On iPhone, add GMWE to the
Home Screen and complete login inside that installed app. Real iPhone relaunch
and renewal are acceptance steps, not covered by desktop browser emulation.
