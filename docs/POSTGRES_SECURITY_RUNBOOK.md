# PostgreSQL security setup & tests

This runbook is for LOCAL/TEST databases only. No production database configured.

1. As database administrator create an isolated owner/migrator account and role `mytax_app` with `NOSUPERUSER NOBYPASSRLS`; ensure `mytax_app` is not the table owner.
2. Run `000001_security_foundation.up.sql`, then `000002_app_privileges.up.sql` as owner.
3. Open a transaction as the application role, then set the verified subject with `SELECT set_config('app.subject', $1, true)` through a parameterized query; all protected queries must use that **same transaction**.
4. Check that reads with no subject return no rows. After controlled membership provisioning, assert tenant A cannot access tenant B's evidence and audit events, including INSERT and SELECT. Verify transaction reset, cross-request pooling and concurrent access.
5. Verify app role cannot insert/update memberships, insert audit rows, or bypass tenant restrictions. Verify owner and migrations are never used by API code.
6. Current RLS will not create tenants or memberships through the app role. Provisioning must be designed and reviewed separately.
7. Store OIDC issuer and audience in protected server configuration. JWT signature / JWKS verification, expiry, azp, nonce as applicable, and key rotation are REQUIRED before real authentication.

These instructions are test guidance, not evidence that the tests have run. SQL static checks are included, but **actual PostgreSQL isolation tests have not yet been run**. No real taxpayer records or uploaded documents are permitted at this stage.
