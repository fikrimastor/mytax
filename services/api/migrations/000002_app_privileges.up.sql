-- Requires pre-created role mytax_app by database administrator.
-- SECURITY: mytax_app must be NOLOGIN or use tightly controlled credentials, NOBYPASSRLS,
-- not a superuser and not the table owner. Do not execute with untrusted inputs.
REVOKE ALL ON tenants, tenant_memberships, evidence_metadata, audit_events FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO mytax_app;
GRANT SELECT ON tenants,tenant_memberships,evidence_metadata,audit_events TO mytax_app;
GRANT INSERT ON evidence_metadata TO mytax_app;
-- No grant on audit_events INSERT until a trusted authenticated audit writer exists.
-- Do not grant INSERT/UPDATE/DELETE on memberships or tenants to application role.
