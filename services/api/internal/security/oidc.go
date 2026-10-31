package security

import (
 "context"
 "errors"
 "strings"
 "github.com/coreos/go-oidc/v3/oidc"
)
// Verifier validates OIDC ID tokens only. Keycloak access tokens require
// a separately designed access-token verifier and authorization policy.
type Verifier struct{ issuer,audience string; verifier *oidc.IDTokenVerifier }
func NewOIDCVerifier(ctx context.Context,issuer,audience string)(*Verifier,error){
 if !strings.HasPrefix(issuer,"https://")||audience=="" {return nil,errors.New("HTTPS issuer and audience required")}
 p,err:=oidc.NewProvider(ctx,issuer);if err!=nil{return nil,err}
 return &Verifier{issuer:issuer,audience:audience,verifier:p.Verifier(&oidc.Config{ClientID:audience})},nil
}
func (v *Verifier) Verify(ctx context.Context,raw string)(Identity,error){
 if v==nil||v.verifier==nil||raw==""{return Identity{},errors.New("verifier/token missing")}
 token,err:=v.verifier.Verify(ctx,raw)
 if err!=nil{return Identity{},err}
 result:=Identity{Subject:token.Subject,Issuer:token.Issuer,Audience:v.audience,Verified:true}
 if err:=result.Valid(v.issuer,v.audience);err!=nil{return Identity{},err}
 return result,nil
}
