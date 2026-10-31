# Milestone 14 — CI-backed PostgreSQL isolation

The API workflow now starts PostgreSQL 16, provisions synthetic two-tenant fixtures, runs Go tests and fails if tenant isolation, restricted-role checks or transaction-local subject reset fail.

CI-only database password is deliberately synthetic and hardcoded **only** inside ephemeral test configuration; never use it in a real environment.

Security scope: these checks cover database-level RLS and privilege boundaries. They are **not** proof that Keycloak issuer/audience, signing key rotation, JWT scopes/roles or HTTP authentication are correct. The Keycloak integration remains disabled by default.

Required additional checks before real personal information:
- Live Keycloak/JWKS tests using an isolated test realm and intentionally invalid tokens.
- Verify owner/migrator role separation, tenant provisioning, malicious input, concurrent pooled tenant transactions.
- Configure rate limits, request caps, secure logging, secrets management, object storage and audit signatures.
- External penetration testing and PDPA compliance evaluation.
