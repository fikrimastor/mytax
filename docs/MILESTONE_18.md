# Milestone 18 — live Keycloak and Keycloak-to-RLS end-to-end tests

## What exists
- `docker-compose.keycloak.yml` (dev/test only, never deploy) and the synthetic realm `services/api/testdata/keycloak/mytax-ci-realm.json` (users `alice`, `bob`; clients `mytax-web`, `other-client`, `no-scope-client`; audience `mytax-api`; scope `mytax.api`). All passwords are disposable test values.
- `keycloak_live_test.go`: real access token accepted by `AccessVerifier`.
- `keycloak_live_negative_test.go`: rejection of disallowed `azp`, missing scope, missing API audience, ID token as access token, other-realm issuer, expired (verifier clock shifted), tampered signature, malformed/empty. Each case asserts its specific reason.
- `keycloak_live_rotation_test.go`: new higher-priority RSA key added via the admin API while a verifier is running; new-kid and old-kid tokens both verify without restart.
- `keycloak_e2e_test.go`: real token -> `RequireAccess` -> `TenantDB` -> PostgreSQL RLS with the restricted `mytax_app` role. Covers alice/bob cross-tenant denial, missing/garbage token, malformed tenant id, spoofed `X-User-ID`/`X-Tenant-ID` headers, wrong-client and no-scope tokens, 60 concurrent requests over a 3-connection pool, and `app.subject` leakage.
- `export_test.go` exposes an HTTP-capable verifier constructor to test files only. Production `NewAccessVerifier` still requires an HTTPS issuer.

## Lazy verifier (API can boot before Keycloak)
`oidc.NewProvider` does network discovery, so building the verifier at boot crashed the API whenever Keycloak was not ready yet (for example Coolify starting both together).
- `security.ValidateAccessConfig` (HTTPS issuer, audience, client, scope) now runs **at startup**, so a misconfiguration still crashes at boot.
- `security.LazyVerifier` performs only the discovery lazily. Until it succeeds every protected request is rejected with **503** `authentication unavailable` (`ErrVerifierUnavailable`); it is never treated as authenticated. Attempts are single-flight, rate-limited (10s attempt timeout, 5s minimum retry in `cmd/server/main.go`), and the verifier is cached after the first success. A missing token is still 401.
- Tested: fake-clock unit tests for backoff, caching, single-flight under 20 concurrent callers, timeout and nil verifier (also `-race` x20); a live test where a simulated outage returns 503 and then a real Keycloak token is accepted without restart; and mutation checks (removing the 503 mapping, the single-flight guard or the backoff each makes a test fail).
- Smoke-tested on the built binary: with `MYTAX_OIDC_ISSUER=https://127.0.0.1:1/...` it starts, `/healthz` is 200, the secure route is 503 (about 0.5 ms inside the backoff window) and 401 without a token; with an `http://` issuer it still exits at boot.
- Not covered: behaviour after a *successful* build when Keycloak later goes down (go-oidc caches keys; not tested), and the lazy path against a real HTTPS issuer (needs the staging Keycloak).

## Findings
- Declaring `clientScopes` in a realm import stops Keycloak creating built-ins, so tokens had **no `sub`** until a `basic` scope with `oidc-sub-mapper` was added. The old synthetic-JWKS tests could not catch this.
- Mutation check: removing the membership check or the `azp` check makes the e2e test fail with the expected messages. With the membership check removed, RLS alone still returned count 0 (no data leaked).

## CI result
GitHub Actions run 98 (commit `04f6a69`, workflow "API and database security tests") passed. Job logs confirm: `docker compose -f docker-compose.keycloak.yml up` pulled `quay.io/keycloak/keycloak:26.5.2`, the `mytax-ci` realm was ready after ~24s, and `TestLiveKeycloakValidAccessToken`, `TestLiveKeycloakNegativeMatrix`, `TestLiveKeycloakKeyRotation` and `TestKeycloakToRLSEndToEnd` all ran and passed. The step fails if any gate variable is empty or any test reports SKIP. (The PostgreSQL `invalid input syntax for type uuid: "not-a-uuid"` log line is the intended malformed-tenant-id case.) Runs 95-97 were green with the older workflow and did NOT run the live tests; they are not evidence.

## Not verified
- `docker compose up` has been confirmed on a GitHub-hosted runner only, not on a developer machine. The sandbox that wrote this had no Docker daemon and ran Keycloak 26.5.2 from the release zip.
- Expiry uses a shifted verifier clock, not real elapsed time. Removal of the *old* signing key, token revocation/user disable, and logout behaviour are untested (access JWTs stay valid until `exp`).
- The e2e route mirrors the one in `cmd/api/main.go` but is not the production binary, which needs an HTTPS issuer. No disposable-HTTPS or Coolify/staging Keycloak test yet. No independent review.

## Local run
```
docker compose -f docker-compose.keycloak.yml up -d
export MYTAX_TEST_KEYCLOAK_ISSUER=http://localhost:8081/realms/mytax-ci
export MYTAX_TEST_KEYCLOAK_ADMIN_USER=ci-admin MYTAX_TEST_KEYCLOAK_ADMIN_PASSWORD=ci-only-admin-pw
# Postgres fixture as in .github/workflows/api.yml, plus the superuser URL used only for seeding:
export MYTAX_TEST_ADMIN_DATABASE_URL='postgres://postgres:ci-postgres-only@localhost:5432/mytax_ci?sslmode=disable'
cd services/api && go test -count=1 -v ./internal/security/ ./internal/persistence/
```
