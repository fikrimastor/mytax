package persistence
import ("testing";"github.com/fikrimastor/mytax/services/api/internal/security")
func TestDependencyAndIdentityFailClosed(t *testing.T){
 if _,err:=NewTenantDB(nil,"issuer","aud");err==nil{t.Fatal("nil pool allowed")}
 i:=security.Identity{Subject:"x",Issuer:"issuer",Audience:"aud"}
 if err:=security.RequireTenant(i,"issuer","aud","tenant");err==nil{t.Fatal("unverified identity allowed")}
}
