package security

import (
 "errors"
 "strings"
)
// AccessClaims are extracted only after cryptographic JWT verification.
type AccessClaims struct {
 AuthorizedParty string `json:"azp"`
 ClientID string `json:"client_id"`
 Scope string `json:"scope"`
 Typ string `json:"typ"`
}
func AuthorizeAccess(claims AccessClaims,requiredClient,requiredScope string)error{
 if requiredClient==""||requiredScope=="" {return errors.New("authorization policy is not configured")}
 if claims.Typ!="" && claims.Typ!="Bearer" && claims.Typ!="at+jwt" {return errors.New("unexpected token type")}
 client:=claims.AuthorizedParty
 if client=="" {client=claims.ClientID}
 if client!=requiredClient {return errors.New("unauthorized client")}
 for _,scope:=range strings.Fields(claims.Scope) {if scope==requiredScope{return nil}}
 return errors.New("required scope missing")
}
