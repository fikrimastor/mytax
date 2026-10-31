// Package security defines boundaries for verified OIDC identities.
// Parsing a token or accepting an X-User-ID header is NOT authentication.
package security

import (
 "errors"
 "strings"
)
type Identity struct {Subject string;Issuer string;Audience string;Verified bool}
func (i Identity) Valid(expectedIssuer,expectedAudience string)error{
 if !i.Verified{return errors.New("identity must be cryptographically verified")}
 if i.Subject==""||strings.TrimSpace(i.Subject)!=i.Subject{return errors.New("invalid subject")}
 if expectedIssuer==""||expectedAudience==""||i.Issuer!=expectedIssuer||i.Audience!=expectedAudience{return errors.New("issuer or audience mismatch")}
 return nil
}
func RequireTenant(identity Identity,issuer,audience,tenantID string)error{
 if err:=identity.Valid(issuer,audience);err!=nil{return err}
 if tenantID==""||strings.ContainsAny(tenantID," \t\r\n"){return errors.New("invalid tenant")}
 // Membership is checked separately in the database under RLS.
 return nil
}
