package security

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)
func TestMigrationSafetyMarkers(t *testing.T) {
 root:=filepath.Join("..","..","migrations")
 data,err:=os.ReadFile(filepath.Join(root,"000001_security_foundation.up.sql"))
 if err!=nil{t.Fatal(err)}
 sql:=string(data)
 for _,name:=range []string{"tenants","tenant_memberships","evidence_metadata","audit_events"} {
  if !strings.Contains(sql,"ALTER TABLE "+name+" FORCE ROW LEVEL SECURITY") {t.Errorf("missing FORCE RLS: %s",name)}
 }
 if strings.Contains(sql,"CREATE POLICY audit_write") {t.Fatal("unexpected unauthenticated audit insert policy")}
 privileges,err:=os.ReadFile(filepath.Join(root,"000002_app_privileges.up.sql"))
 if err!=nil{t.Fatal(err)}
 if strings.Contains(string(privileges),"GRANT INSERT ON audit_events TO mytax_app") {t.Fatal("audit append privilege is not yet authorized")}
}
