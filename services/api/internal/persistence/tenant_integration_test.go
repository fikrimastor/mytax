package persistence

import (
 "context"
 "os"
 "testing"
 "github.com/jackc/pgx/v5/pgxpool"
 "github.com/fikrimastor/mytax/services/api/internal/security"
)
// Run against a PostgreSQL fixture that has migrations applied, two tenants,
// and membership of test-subject in only tenant A. Use a restricted app role.
// Never run this test on a database containing taxpayer data.
func TestTenantIsolationIntegration(t *testing.T){
 url:=os.Getenv("MYTAX_TEST_DATABASE_URL")
 a:=os.Getenv("MYTAX_TEST_TENANT_A")
 b:=os.Getenv("MYTAX_TEST_TENANT_B")
 if url==""||a==""||b==""{t.Skip("integration DB fixture not configured")}
 ctx:=context.Background()
 pool,err:=pgxpool.New(ctx,url);if err!=nil{t.Fatal(err)};defer pool.Close()
 db,err:=NewTenantDB(pool,"https://issuer.example","mytax-api");if err!=nil{t.Fatal(err)}
 id:=security.Identity{Subject:"test-subject",Issuer:"https://issuer.example",Audience:"mytax-api",Verified:true}
 if _,err:=db.CountEvidence(ctx,id,a);err!=nil{t.Fatalf("authorized tenant failed: %v",err)}
 if _,err:=db.CountEvidence(ctx,id,b);err==nil{t.Fatal("cross-tenant read allowed")}
 // Verify that no transaction-local subject remains after pool reuse.
 var leaked string
 if err:=pool.QueryRow(ctx,"select current_setting('app.subject',true)").Scan(&leaked);err!=nil{t.Fatal(err)}
 if leaked!=""{t.Fatalf("subject leaked across transactions: %q",leaked)}
}
