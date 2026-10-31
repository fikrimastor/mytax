package security
import "testing"
func TestIdentityFailClosed(t *testing.T){
 i:=Identity{Subject:"u",Issuer:"https://issuer.example",Audience:"mytax",Verified:false}
 if RequireTenant(i,i.Issuer,i.Audience,"tenant")==nil{t.Fatal("unverified accepted")}
 i.Verified=true
 if RequireTenant(i,"https://wrong",i.Audience,"tenant")==nil{t.Fatal("issuer mismatch accepted")}
 if RequireTenant(i,i.Issuer,i.Audience,"")==nil{t.Fatal("empty tenant accepted")}
 if err:=RequireTenant(i,i.Issuer,i.Audience,"tenant");err!=nil{t.Fatal(err)}
}
