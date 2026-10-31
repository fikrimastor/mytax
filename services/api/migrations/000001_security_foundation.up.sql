-- M11 foundation. Run with a migration runner under a dedicated owner role.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE tenants (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE tenant_memberships (
 tenant_id uuid NOT NULL REFERENCES tenants(id),
 subject_id text NOT NULL CHECK (length(subject_id) BETWEEN 1 AND 255),
 role text NOT NULL CHECK (role IN ('owner','member')),
 PRIMARY KEY (tenant_id,subject_id)
);
CREATE TABLE evidence_metadata (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 tenant_id uuid NOT NULL REFERENCES tenants(id),
 claim_id uuid NOT NULL,
 sha256_hex char(64) NOT NULL CHECK (sha256_hex ~ '^[0-9a-f]{64}$'),
 object_key text NOT NULL,
 byte_size bigint NOT NULL CHECK (byte_size >= 0),
 content_type text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(tenant_id,object_key)
);
CREATE INDEX evidence_tenant_claim_idx ON evidence_metadata(tenant_id,claim_id);
CREATE TABLE audit_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 tenant_id uuid NOT NULL REFERENCES tenants(id),
 subject_id text NOT NULL,
 action text NOT NULL,
 resource_id text NOT NULL,
 event_digest char(64) NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_tenant_id_idx ON audit_events(tenant_id,id);
-- Application role must not own tables or have BYPASSRLS.
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence_metadata ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;
ALTER TABLE tenant_memberships FORCE ROW LEVEL SECURITY;
ALTER TABLE evidence_metadata FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_events FORCE ROW LEVEL SECURITY;
-- Current membership is determined from a transaction-local, verified OIDC subject.
-- The database alone DOES NOT authenticate subject_id. Untrusted clients must never set app.subject.
CREATE POLICY membership_read ON tenant_memberships FOR SELECT
 USING (subject_id = nullif(current_setting('app.subject', true),''));
CREATE POLICY tenants_read ON tenants FOR SELECT
 USING (EXISTS (SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=tenants.id AND m.subject_id=nullif(current_setting('app.subject',true),'')));
CREATE POLICY evidence_read ON evidence_metadata FOR SELECT
 USING (EXISTS (SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=evidence_metadata.tenant_id AND m.subject_id=nullif(current_setting('app.subject',true),'')));
CREATE POLICY evidence_write ON evidence_metadata FOR INSERT
 WITH CHECK (EXISTS (SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=evidence_metadata.tenant_id AND m.subject_id=nullif(current_setting('app.subject',true),'')));
CREATE POLICY audit_read ON audit_events FOR SELECT
 USING (EXISTS (SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=audit_events.tenant_id AND m.subject_id=nullif(current_setting('app.subject',true),'')));
-- Audit insertion through a privileged service only, not directly via app role.
-- No UPDATE or DELETE grants to application roles. Separate trusted provisioning for tenants/memberships.
