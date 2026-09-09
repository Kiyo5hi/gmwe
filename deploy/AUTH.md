# GMWE Browser Authentication

GMWE is private to the two configured Pocket-ID users. The client is public,
uses authorization code with S256 PKCE and an exact callback, and requests only
openid and write:hitokoto for its existing GMWE API resource. No client secret,
offline_access, pairing endpoint or connection download is used.

Caddy serves only /login and the non-sensitive /gmwe.webp logo without a session.
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
Sessions are in memory and expire within an hour; a restart requires login again.
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
