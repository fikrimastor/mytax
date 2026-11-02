# PostgreSQL security setup & tests

This runbook is for LOCAL/TEST databases only. No production database configured.

1. As database administrator create an isolated owner/migrator account and role `mytax_app` with `NOSUPERUSER NOBYPASSRLS`; ensure `mytax_app` is not the table owner.
2. Run `000001_security_foundation.up.sql`, then `000002_app_privileges.up.sql` as owner.
3. Open a transaction as the application role, then set the verified subject with `SELECT set_config('app.subject', $1, true)` through a parameterized query; all protected queries must use that **same transaction**.
4. Check that reads with no subject return no rows. After controlled membership provisioning, assert tenant A cannot access tenant B's evidence and audit events, including INSERT and SELECT. Verify transaction reset, cross-request pooling and concurrent access.
5. Verify app role cannot insert/update memberships, insert audit rows, or bypass tenant restrictions. Verify owner and migrations are never used by API code.
6. Current RLS will not create tenants or memberships through the app role. Provisioning must be designed and reviewed separately.
7. Store OIDC issuer and audience in protected server configuration. JWT signature / JWKS verification, expiry, azp, nonce as applicable, and key rotation are REQUIRED before real authentication.

These instructions are test guidance, not evidence that the tests have run. *(Updated 2026-11-02: the PostgreSQL isolation tests now run in CI against PostgreSQL 16 and passed on the branch `claude/quirky-wozniak-o9r7ih`; see the test matrix below for what they cover and what they do not.)* No real taxpayer records or uploaded documents are permitted at this stage.

## Two-tenant isolation test matrix (added 2026-11-02)
Fixture `services/api/testdata/security_fixture.sql`: tenant A with subject `test-subject`, tenant B with subject `test-subject-b`, one evidence row each, and the restricted role `mytax_app`. Tests live in `services/api/internal/persistence/tenant_matrix_integration_test.go` and always roll back, so the fixture never changes.

| Test | What it proves |
|---|---|
| `TestSymmetricTwoTenantIsolation` | Each subject reads only its own tenant, in both directions; the database alone hides the other tenant's rows in `tenants`, `tenant_memberships`, `evidence_metadata` |
| `TestSubjectWithoutMembershipOrWithoutSubjectSeesNothing` | Unknown, empty, unset, SQL-looking and case-changed subjects see zero rows in every table |
| `TestRLSRejectsCrossTenantWriteAndOwnWriteIsRolledBack` | The `WITH CHECK` policy refuses a write into the other tenant, with no subject, with an unknown subject; an own-tenant write is allowed and rolled back |
| `TestApplicationRoleCannotEscalateOrTamper` | The app role gets insufficient privilege (42501) for membership, tenant, audit and evidence changes, truncate, drop, disabling RLS, `BYPASSRLS` and `SET ROLE` |
| `TestSpoofedOrMalformedIdentityAndTenantNeverReachData` | Unverified identity, wrong issuer or audience, empty or padded subject, bad or injected tenant ID reach no data, and tables stay intact |
| `TestTransactionLocalSubjectDoesNotSurviveCommitRollbackOrError` | On a single reused connection, the subject never survives a commit, a denial, a SQL error, a failing callback or a panic |
| `TestConcurrentPooledAccessNeverLeaksOrCrossesTenants` | 24 goroutines over 4 pooled connections, mixing both tenants, denied cases and bare no-subject probes; no leakage, and every pooled connection ends clean |

Mutation checks run on a local PostgreSQL 16 (each one makes at least one test fail): session-level instead of transaction-local subject, removing the Go membership check, `WITH CHECK (true)`, `USING (true)` on evidence and on memberships, granting UPDATE on memberships, granting INSERT on audit events, and `BYPASSRLS` on the app role. The `WITH CHECK (true)` mutant was **not** caught before this matrix existed.

CI runs `go test -race` and fails the job on any skipped test.

Not covered: a live Keycloak, JWKS key rotation and caching, token revocation, tenant and membership provisioning (the app role cannot do it by design), and anything on a live deployment.
