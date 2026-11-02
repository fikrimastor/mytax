# Milestone 18 — live Keycloak in CI

CI starts Keycloak 26 with a throwaway self-signed certificate (`localhost`, HTTPS), imports `services/api/testdata/keycloak/realm-mytax.json`, and runs `internal/livekeycloak`. Production code is unchanged: `NewAccessVerifier` still requires an HTTPS issuer; tests give go-oidc an HTTP client that trusts the CI CA.

Covered with real Keycloak tokens:
- valid access token (`aud=mytax-api`, `azp=mytax-web`, scope `mytax:evidence:read`, `typ=Bearer`) is accepted;
- missing scope, wrong client and expired token are rejected (expiry proven by accepting the same token while fresh);
- Keycloak token → Echo → `RequireAccess` → `TenantDB` → PostgreSQL RLS: member tenant returns its evidence count (proves `sub` → `app.subject`), non-member tenant 403, no token / wrong client 401.

Realm note: the realm defines its own client scopes, so Keycloak's built-in `basic` scope is absent; `mytax-subject` supplies the `sub` claim. A production realm must do the same.

The CI job fails if either live test is skipped.

**Not yet included:** key rotation (follow-up), token revocation (needs offline-vs-introspection design), JWKS caching (go-oidc caches internally).
