package persistence

import (
 "context"
 "os"
 "testing"

 "github.com/jackc/pgx/v5/pgxpool"
)
func TestPostgresRoleAndRLSIntegration(t *testing.T) {
 uri:=os.Getenv("MYTAX_TEST_DATABASE_URL")
 if uri=="" {t.Skip("test database not configured")}
 ctx:=context.Background()
 pool,err:=pgxpool.New(ctx,uri);if err!=nil{t.Fatal(err)}
 defer pool.Close()
 var super,bypass bool
 var user string
 if err:=pool.QueryRow(ctx,"select current_user, rolsuper, rolbypassrls from pg_roles where rolname=current_user").Scan(&user,&super,&bypass);err!=nil{t.Fatal(err)}
 if user!="mytax_app" || super || bypass {t.Fatalf("unsafe application role %s super=%v bypass=%v",user,super,bypass)}
 var noSubject int
 if err:=pool.QueryRow(ctx,"select count(*) from evidence_metadata").Scan(&noSubject);err!=nil{t.Fatal(err)}
 if noSubject!=0{t.Fatalf("no-subject read exposed %d rows",noSubject)}
 var canInsertAudit,canEditMembers bool
 if err:=pool.QueryRow(ctx,"select has_table_privilege(current_user,'audit_events','INSERT'), has_table_privilege(current_user,'tenant_memberships','UPDATE')").Scan(&canInsertAudit,&canEditMembers);err!=nil{t.Fatal(err)}
 if canInsertAudit||canEditMembers{t.Fatal("application role has excessive privileges")}
}
