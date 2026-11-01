package security

import (
 "bytes"
 "context"
 "encoding/base64"
 "encoding/json"
 "net/http"
 "net/url"
 "os"
 "strings"
 "testing"
)

// Key rotation against a real Keycloak. Needs admin credentials of the
// disposable instance (MYTAX_TEST_KEYCLOAK_ADMIN_USER / _PASSWORD). Skipped
// (NOT passed) without them.
func tokenKid(t *testing.T, jwt string) string {
 t.Helper()
 raw, err := base64.RawURLEncoding.DecodeString(strings.Split(jwt, ".")[0])
 if err != nil { t.Fatal(err) }
 var h struct{ Kid string `json:"kid"` }
 if err := json.Unmarshal(raw, &h); err != nil { t.Fatal(err) }
 return h.Kid
}

func adminDo(t *testing.T, method, rawURL, bearer string, body any) *http.Response {
 t.Helper()
 var buf bytes.Buffer
 if body != nil { if err := json.NewEncoder(&buf).Encode(body); err != nil { t.Fatal(err) } }
 req, err := http.NewRequest(method, rawURL, &buf)
 if err != nil { t.Fatal(err) }
 req.Header.Set("Authorization", "Bearer "+bearer)
 req.Header.Set("Content-Type", "application/json")
 resp, err := http.DefaultClient.Do(req)
 if err != nil { t.Fatal(err) }
 return resp
}

func TestLiveKeycloakKeyRotation(t *testing.T) {
 issuer := liveIssuer(t)
 adminUser, adminPass := os.Getenv("MYTAX_TEST_KEYCLOAK_ADMIN_USER"), os.Getenv("MYTAX_TEST_KEYCLOAK_ADMIN_PASSWORD")
 if adminUser == "" || adminPass == "" { t.Skip("Keycloak admin credentials not set: key rotation NOT verified") }
 base := strings.Split(issuer, "/realms/")[0]

 // admin token from the master realm
 form := url.Values{"grant_type": {"password"}, "client_id": {"admin-cli"}, "username": {adminUser}, "password": {adminPass}}
 resp, err := http.Post(base+"/realms/master/protocol/openid-connect/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
 if err != nil { t.Fatal(err) }
 var at struct{ AccessToken string `json:"access_token"` }
 json.NewDecoder(resp.Body).Decode(&at); resp.Body.Close()
 if at.AccessToken == "" { t.Fatal("no admin token") }

 // verifier built BEFORE rotation and reused after it (no restart)
 v := liveVerifier(t, issuer, "mytax-web", "mytax.api")
 old, _ := liveToken(t, issuer, "mytax-web", "alice", "ci-only-alice-pw")
 if _, err := v.Verify(context.Background(), old); err != nil { t.Fatalf("pre-rotation token rejected: %v", err) }

 // add a new RSA key with higher priority than the default (100)
 realm := adminDo(t, "GET", base+"/admin/realms/mytax-ci", at.AccessToken, nil)
 var r struct{ ID string `json:"id"` }
 json.NewDecoder(realm.Body).Decode(&r); realm.Body.Close()
 comp := map[string]any{"name": "ci-rotated-rsa", "providerId": "rsa-generated", "providerType": "org.keycloak.keys.KeyProvider", "parentId": r.ID,
  "config": map[string][]string{"priority": {"200"}, "enabled": {"true"}, "active": {"true"}, "keySize": {"2048"}}}
 created := adminDo(t, "POST", base+"/admin/realms/mytax-ci/components", at.AccessToken, comp)
 created.Body.Close()
 if created.StatusCode != 201 { t.Fatalf("could not add key provider: HTTP %d", created.StatusCode) }
 loc := created.Header.Get("Location")
 t.Cleanup(func() { if loc != "" { adminDo(t, "DELETE", loc, at.AccessToken, nil).Body.Close() } })

 fresh, _ := liveToken(t, issuer, "mytax-web", "alice", "ci-only-alice-pw")
 if tokenKid(t, fresh) == tokenKid(t, old) { t.Fatal("rotation did not change the signing key (kid identical)") }
 if _, err := v.Verify(context.Background(), fresh); err != nil { t.Fatalf("token signed by NEW key rejected by long-lived verifier: %v", err) }
 if _, err := v.Verify(context.Background(), old); err != nil { t.Fatalf("token signed by OLD (still published) key rejected: %v", err) }
}
