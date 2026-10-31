package security

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "testing"
 "github.com/labstack/echo/v4"
)
type fakeVerifier struct{valid bool}
func (v fakeVerifier)Verify(ctx context.Context,raw string)(Identity,error){
 if !v.valid||raw!="signed-test-token"{return Identity{},errors.New("invalid token")}
 return Identity{Subject:"fixture-user",Issuer:"https://example.test",Audience:"mytax-api",Verified:true},nil
}
func TestHTTPAccessGate(t *testing.T){
 tests:=[]struct{name,header string;enabled bool;want int}{
  {"missing","",true,401},
  {"invalid","Bearer forged",true,401},
  {"success","Bearer signed-test-token",true,200},
  {"disabled","Bearer signed-test-token",false,503},
 }
 for _,tt:=range tests{t.Run(tt.name,func(t *testing.T){
  e:=echo.New()
  handler:=func(c echo.Context)error{
   id,ok:=c.Get("verified_identity").(Identity)
   if !ok||id.Subject!="fixture-user"{return c.NoContent(500)}
   return c.NoContent(200)
  }
  var verifier TokenVerifier
  if tt.enabled{verifier=fakeVerifier{valid:true}}
  e.GET("/protected",RequireAccess(verifier,handler))
  req:=httptest.NewRequest(http.MethodGet,"/protected",nil)
  if tt.header!=""{req.Header.Set("Authorization",tt.header)}
  rec:=httptest.NewRecorder();e.ServeHTTP(rec,req)
  if rec.Code!=tt.want{t.Fatalf("got %d want %d",rec.Code,tt.want)}
 })}
}
