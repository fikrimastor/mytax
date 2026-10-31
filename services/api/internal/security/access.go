package security

import (
 "context"
 "errors"
 "net/http"
 "strings"
 "github.com/coreos/go-oidc/v3/oidc"
)
// AccessVerifier is distinct from the ID-token verifier. It validates
// signature, issuer, expiry and a REQUIRED audience claim on access JWTs.
type AccessVerifier struct{issuer,audience,client,scope string; verifier *oidc.IDTokenVerifier}
func NewAccessVerifier(ctx context.Context,issuer,audience,client,scope string)(*AccessVerifier,error){
 if !strings.HasPrefix(issuer,"https://")||audience==""||client==""||scope=="" {return nil,errors.New("HTTPS issuer and audience required")}
 provider,err:=oidc.NewProvider(ctx,issuer);if err!=nil{return nil,err}
 return &AccessVerifier{issuer:issuer,audience:audience,client:client,scope:scope,verifier:provider.Verifier(&oidc.Config{ClientID:audience})},nil
}
func (v *AccessVerifier) Verify(ctx context.Context, raw string)(Identity,error){
 if v==nil||v.verifier==nil||raw==""{return Identity{},errors.New("verifier or token absent")}
 token,err:=v.verifier.Verify(ctx,raw);if err!=nil{return Identity{},err}
 // go-oidc verifies JWT issuer, signature, expiry and audience. Configure
 // Keycloak audience mapper to include the actual API resource audience.
 var claims AccessClaims
 if err:=token.Claims(&claims);err!=nil{return Identity{},err}
 if err:=AuthorizeAccess(claims,v.client,v.scope);err!=nil{return Identity{},err}
 id:=Identity{Subject:token.Subject,Issuer:token.Issuer,Audience:v.audience,Verified:true}
 if err:=id.Valid(v.issuer,v.audience);err!=nil{return Identity{},err}
 return id,nil
}
func Bearer(r *http.Request)(string,error){
 values,ok:=r.Header["Authorization"]
 if !ok||len(values)!=1{return "",errors.New("exactly one authorization header required")}
 parts:=strings.Fields(values[0])
 if len(parts)!=2||!strings.EqualFold(parts[0],"Bearer")||parts[1]==""||strings.Contains(parts[1],","){return "",errors.New("invalid bearer token")}
 return parts[1],nil
}
