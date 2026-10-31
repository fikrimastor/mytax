# Milestone 12 — OIDC verifier + tenant transaction boundary

The Go library now contains:
- OIDC ID-token verification using provider discovery and signed key verification, issuer, audience and expiry checks.
- A tenant-scoped read-only PostgreSQL transaction with SET LOCAL equivalent via set_config(...,true).
- Membership verification within that transaction and read-only evidence metadata count.
- Fail-closed unit tests.

**Not production-ready**: OIDC ID tokens are not interchangeable with Keycloak OAuth access tokens. A user-facing API should normally validate access tokens and their authorized audience/scopes using a separately reviewed implementation. The new verifier is NOT wired into any public HTTP endpoint. No untrusted header can supply the subject. No database URI, credentials or personal records are committed.

PostgreSQL migration/test requirements remain:
- real RLS isolation integration tests with PostgreSQL and separate owner/app roles;
- authenticated server-controlled subject context, non-bypass role and membership provisioning;
- OIDC live issuer/JWKS test, expiry, signing algorithms and intended token type;
- encrypted evidence storage, audit integrity and PDPA review.

Do not accept taxpayer PII or uploads until all gates pass.
