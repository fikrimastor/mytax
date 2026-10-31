package security

import (
 "context"
 "crypto/rand"
 "crypto/rsa"
 "crypto/sha256"
 "crypto"
 "encoding/base64"
 "encoding/json"
 "math/big"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
 "github.com/coreos/go-oidc/v3/oidc"
)

func enc(v any)string{b,_:=json.Marshal(v);return base64.RawURLEncoding.EncodeToString(b)}
func sign(t *testing.T,key *rsa.PrivateKey,kid string,claims map[string]any)string{
 t.Helper()
 head:=enc(map[string]any{"alg":"RS256","typ":"JWT","kid":kid})
 payload:=head+"."+enc(claims)
 digest:=sha256.Sum256([]byte(payload))
 sig,err:=rsa.SignPKCS1v15(rand.Reader,key,crypto.SHA256,digest[:]);if err!=nil{t.Fatal(err)}
 return payload+"."+base64.RawURLEncoding.EncodeToString(sig)
}
func TestSignedJWKSAccessTokens(t *testing.T){
 key,err:=rsa.GenerateKey(rand.Reader,2048);if err!=nil{t.Fatal(err)}
 other,err:=rsa.GenerateKey(rand.Reader,2048);if err!=nil{t.Fatal(err)}
 var issuer string
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Content-Type","application/json")
  switch r.URL.Path{
  case "/.well-known/openid-configuration":
   json.NewEncoder(w).Encode(map[string]any{"issuer":issuer,"jwks_uri":issuer+"/keys","authorization_endpoint":issuer+"/auth","token_endpoint":issuer+"/token"})
  case "/keys":
   json.NewEncoder(w).Encode(map[string]any{"keys":[]any{map[string]any{
    "kty":"RSA","kid":"k1","use":"sig","alg":"RS256",
    "n":base64.RawURLEncoding.EncodeToString(key.N.Bytes()),"e":base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
   }}})
  default:http.NotFound(w,r)
  }
 }))
 defer server.Close()
 issuer=server.URL
 // Test-only HTTP issuer; production constructor continues to require HTTPS.
 provider,err:=oidc.NewProvider(context.Background(),issuer);if err!=nil{t.Fatal(err)}
 verifier:=&AccessVerifier{issuer:issuer,audience:"mytax-api",client:"mytax-web",scope:"mytax:evidence:read",verifier:provider.Verifier(&oidc.Config{ClientID:"mytax-api"})}
 now:=time.Now().Unix()
 good:=func()map[string]any{return map[string]any{"iss":issuer,"sub":"test-user","aud":"mytax-api","iat":now,"exp":now+3600,"azp":"mytax-web","scope":"openid mytax:evidence:read","typ":"Bearer"}}
 cases:=[]struct{name string;claims map[string]any;key *rsa.PrivateKey;ok bool}{
  {"valid",good(),key,true},
  {"wrong signature",good(),other,false},
  {"expired",func()map[string]any{c:=good();c["exp"]=now-7200;return c}(),key,false},
  {"wrong audience",func()map[string]any{c:=good();c["aud"]="other-api";return c}(),key,false},
  {"wrong client",func()map[string]any{c:=good();c["azp"]="other-client";return c}(),key,false},
  {"missing scope",func()map[string]any{c:=good();c["scope"]="openid";return c}(),key,false},
  {"id token type",func()map[string]any{c:=good();c["typ"]="ID";return c}(),key,false},
  {"wrong issuer",func()map[string]any{c:=good();c["iss"]="https://unexpected.example";return c}(),key,false},
 }
 for _,tc:=range cases{t.Run(tc.name,func(t *testing.T){
  _,err:=verifier.Verify(context.Background(),sign(t,tc.key,"k1",tc.claims))
  if (err==nil)!=tc.ok{t.Fatalf("expected success=%v, got err=%v",tc.ok,err)}
 })}
 if _,err:=verifier.Verify(context.Background(),strings.Repeat("x",16));err==nil{t.Fatal("malformed token accepted")}
}
