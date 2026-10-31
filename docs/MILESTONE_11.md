# Milestone 11 — PostgreSQL tenant data boundaries (foundation only)

Added PostgreSQL DDL for tenants, membership, evidence metadata and audit events, with row-level security (RLS) and FORCE RLS. Access is scoped through the transaction-local `app.subject` setting. This setting is **not authentication**: the API must first cryptographically verify an OIDC JWT (issuer, audience, signature, expiry), then assign the subject with SET LOCAL inside the same transaction. Never expose raw PostgreSQL connections to users.

No connection pool, database execution, authenticated API endpoint, object-storage integration, tenant provisioning, encrypted file storage or durable signed audit chain is installed yet. No real documents should be uploaded. The metadata table assumes object keys are opaque; it does not upload bytes. Database application roles must not own tables, have BYPASSRLS, or permission to change the subject outside controlled server code. Audit event integrity needs a trusted append-only/signed writer and storage outside the mutable database trust boundary.

## Release gates
- Test migrations against real PostgreSQL with separate owner/app roles.
- Verify negative tenant isolation, including pooled concurrent transactions.
- Validate OIDC using trusted JWKS and required claims; never trust client-provided subject headers.
- Provide audited tenant membership provisioning and per-tenant permissions.
- Object storage encryption, private buckets, short-lived upload URLs, malware checks and retention/deletion policy.
- PDPA review and independent penetration testing before any personal information is accepted.

This milestone is deliberately not a production-ready implementation.
