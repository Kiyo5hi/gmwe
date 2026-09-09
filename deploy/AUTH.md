# GMWE Authentication

The API fails closed without `OIDC_CONFIG_FILE`. Reads remain public; writes
require a verified user. The JSON file is mounted read-only and contains:

```json
{
  "issuer": "https://identity.example.org",
  "client_id": "registered-public-client",
  "resource": "https://gmwe.example.org/api",
  "origin": "https://gmwe.example.org",
  "writers": [{"subject": "immutable-idp-user-id", "user_id": 1, "name": "Writer"}]
}
```

Use an exact `/api/auth/callback` redirect under `origin`, a public PKCE-required
client, and a restricted writer group. Grant only user-delegated
`write:hitokoto` on the configured API resource; do not enable M2M access.
Map immutable subjects to existing database authors, never names supplied by
the caller. Keep private identifiers/config in the infrastructure repository.

Web login uses authorization code + S256 PKCE, state and nonce. Access tokens
remain in an in-memory SCS session, with a Secure/HttpOnly/host-only cookie.
Session writes require exact Origin and CSRF headers. Sessions expire within
one hour; service restarts require web login again. No session database or
client secret is required. Do not enable request/query/header/body logging on
authentication endpoints or cache their responses.

API calls use Pocket-ID's signed access token, not its ID token. The API checks
RS256 signature, `at+jwt` type, issuer, resource audience, expiry, issue/not-before
times, exact client ID, scope, and allowed subject. An already-issued token can
remain usable until expiry after group membership or consent is revoked.
Immediate emergency denial requires removing that subject from the local
configuration and restarting GMWE, or restoring the reverse-proxy write gate.

## Apple Shortcuts

`/account` starts authorization-code + S256 PKCE pairing for a signed-in writer.
It explicitly requests consent for `openid offline_access write:hitokoto` with
the exact API resource. Authorize with the same Pocket-ID user; the verified
callback downloads `gmwe-shortcut.json` directly. It contains a native
Pocket-ID refresh token, not a GMWE API key. Treat the file as a password: keep it
in private device storage, not Git, messages, shared folders, or request logs.
Pair each device separately. Never share one rotating refresh token between
concurrent shortcuts.

The Shortcut must perform these operations:

1. Load its private connection JSON.
2. POST a form to `token_endpoint` with `grant_type=refresh_token`, `client_id`,
   `refresh_token`, and `resource` from that JSON. Use HTTPS only.
3. Require a successful OAuth response. If it includes a new `refresh_token`,
   replace the saved token before sending content. Do not print/show token values.
4. POST JSON `{"Content":"the new entry"}` to `api_endpoint`, using
   `Authorization: Bearer <access_token>` and `Content-Type: application/json`.
   Omit `UserID`; the verified account selects the author.
5. Treat HTTP 201 as success, 409 as duplicate. On OAuth failure, re-pair rather
   than falling back to anonymous writes. Do not blindly retry a failed write.

Pairing-start requires a web session and CSRF token, never only an API bearer.
State, nonce and the PKCE verifier expire after ten minutes and are consumed on
callback. The original writer subject must match both validated tokens. Refresh
tokens stay out of URLs and application logs and are not persisted by GMWE.
The legacy device polling endpoint returns 410 to authenticated callers.
Pocket-ID 2.14.0's Fosite device request sanitization drops `resource` before
storage, causing custom API scopes to fail on its device-info page. Pairing uses
the ordinary supported browser flow instead, without patching the provider or
weakening the resource/scope/user checks.
The connection file is configuration, not an installable Apple `.shortcut` file;
the actual Shortcut must still be updated and tested on the operator's device.

## Rollout

Test the new image against an isolated database copy, including authenticated
and denied writes. Preserve an online backup and the previous image/config.
Keep public and private write gates until verification. Permit auth routes for
login/pairing separately from content writes; do not open every POST route.
Never restore an old database over new accepted writes during rollback.
