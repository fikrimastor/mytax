# Milestone 13 — access-token verifier and database isolation test harness

This milestone adds a Keycloak-compatible JWT access-token verifier, strict Bearer parsing, and an opt-in RLS integration test.

**Important:** OIDC discovery-based access JWT verification assumes your Keycloak realm issues signed JWT access tokens with the API identifier in `aud`. Configure a Keycloak audience mapper for `mytax-api` or your chosen API audience. Issuer MUST be HTTPS and match the realm issuer exactly. The verifier uses go-oidc's JWT validation, but deployment must confirm token type, authorized client (azp if required), scopes, token revocation behavior and rotation policy. No API scopes or roles are enforced yet.

Integration test requires `MYTAX_TEST_DATABASE_URL`, `MYTAX_TEST_TENANT_A`, and `MYTAX_TEST_TENANT_B`. Provision test-subject as member of tenant A only; no real PII. Tests skip if not configured, and do not constitute passing PostgreSQL isolation validation in CI until a fixture is installed.

No real document upload or personal data intake allowed. Protected route enabled only if `MYTAX_ENABLE_AUTH_PREVIEW=true`, OIDC issuer/audience, and DATABASE_URL are configured. Use a restricted PostgreSQL role; avoid committing credentials. No explicit tenant provisioning endpoints exist.
