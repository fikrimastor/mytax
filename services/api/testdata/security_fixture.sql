-- TEST ONLY. Run as PostgreSQL superuser in isolated CI PostgreSQL.
CREATE ROLE mytax_app LOGIN PASSWORD 'ci-only-insecure-password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
\i migrations/000001_security_foundation.up.sql
\i migrations/000002_app_privileges.up.sql
INSERT INTO tenants(id) VALUES ('11111111-1111-4111-8111-111111111111'),('22222222-2222-4222-8222-222222222222');
INSERT INTO tenant_memberships(tenant_id,subject_id,role) VALUES ('11111111-1111-4111-8111-111111111111','test-subject','owner');
INSERT INTO evidence_metadata(tenant_id,claim_id,sha256_hex,object_key,byte_size,content_type)
VALUES ('11111111-1111-4111-8111-111111111111','aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',repeat('a',64),'ci-a',100,'application/pdf'),
('22222222-2222-4222-8222-222222222222','bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',repeat('b',64),'ci-b',200,'application/pdf');
