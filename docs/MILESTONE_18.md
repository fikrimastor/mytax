# Milestone 18 — live Keycloak and Keycloak-to-RLS end-to-end tests

## What exists
- `docker-compose.keycloak.yml` (dev/test only, never deploy) and the synthetic realm `services/api/testdata/keycloak/mytax-ci-realm.json` (users `alice`, `bob`; clients `mytax-web`, `other-client`, `no-scope-client`; audience `mytax-api`; scope `mytax.api`). All passwords are disposable test values.
- `keycloak_live_test.go`: real access token accepted by `AccessVerifier`.
- `keycloak_live_negative_test.go`: rejection of disallowed `azp`, missing scope, missing API audience, ID token as access token, other-realm issuer, expired (verifier clock shifted), tampered signature, malformed/empty. Each case asserts its specific reason.
- `keycloak_live_rotation_test.go`: new higher-priority RSA key added via the admin API while a verifier is running; new-kid and old-kid tokens both verify without restart.
- `keycloak_e2e_test.go`: real token -> `RequireAccess` -> `TenantDB` -> PostgreSQL RLS with the restricted `mytax_app` role. Covers alice/bob cross-tenant denial, missing/garbage token, malformed tenant id, spoofed `X-User-ID`/`X-Tenant-ID` headers, wrong-client and no-scope tokens, 60 concurrent requests over a 3-connection pool, and `app.subject` leakage.
- `export_test.go` exposes an HTTP-capable verifier constructor to test files only. Production `NewAccessVerifier` still requires an HTTPS issuer.

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
