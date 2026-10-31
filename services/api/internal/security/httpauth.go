package security

import (
 "context"
 "net/http"
 "github.com/labstack/echo/v4"
)
// TokenVerifier permits integration tests to inject a deterministic verifier.
// Production must supply the cryptographic AccessVerifier.
type TokenVerifier interface { Verify(context.Context,string)(Identity,error) }
func RequireAccess(verifier TokenVerifier,next echo.HandlerFunc) echo.HandlerFunc {
 return func(c echo.Context)error{
  if verifier==nil{return c.JSON(http.StatusServiceUnavailable,map[string]string{"error":"authentication unavailable"})}
  token,err:=Bearer(c.Request())
  if err!=nil{return c.JSON(http.StatusUnauthorized,map[string]string{"error":"unauthorized"})}
  id,err:=verifier.Verify(c.Request().Context(),token)
  if err!=nil||!id.Verified{return c.JSON(http.StatusUnauthorized,map[string]string{"error":"unauthorized"})}
  c.Set("verified_identity",id)
  return next(c)
 }
}
