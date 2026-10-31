package security

import "testing"
func TestAuthorizationMatrix(t *testing.T){
 cases:=[]struct{name string;claims AccessClaims;client,scope string;valid bool}{
  {"valid",AccessClaims{AuthorizedParty:"mytax-web",Scope:"openid mytax:evidence:read",Typ:"Bearer"},"mytax-web","mytax:evidence:read",true},
  {"wrong client",AccessClaims{AuthorizedParty:"another",Scope:"mytax:evidence:read"},"mytax-web","mytax:evidence:read",false},
  {"missing scope",AccessClaims{AuthorizedParty:"mytax-web",Scope:"openid"},"mytax-web","mytax:evidence:read",false},
  {"substring scope",AccessClaims{AuthorizedParty:"mytax-web",Scope:"mytax:evidence:read-more"},"mytax-web","mytax:evidence:read",false},
  {"id token",AccessClaims{AuthorizedParty:"mytax-web",Scope:"mytax:evidence:read",Typ:"ID"},"mytax-web","mytax:evidence:read",false},
  {"missing type",AccessClaims{AuthorizedParty:"mytax-web",Scope:"mytax:evidence:read"},"mytax-web","mytax:evidence:read",false},
  {"unconfigured",AccessClaims{AuthorizedParty:"mytax-web",Scope:"mytax:evidence:read"},"","mytax:evidence:read",false},
 }
 for _,tc:=range cases{err:=AuthorizeAccess(tc.claims,tc.client,tc.scope);if (err==nil)!=tc.valid{t.Errorf("%s: %v",tc.name,err)}}
}
