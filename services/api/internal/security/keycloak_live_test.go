package security

import (
 "context"
 "encoding/json"
 "net/http"
 "net/url"
 "os"
 "strings"
 "testing"
 "time"
 "github.com/coreos/go-oidc/v3/oidc"
)

// Live Keycloak tests. They need the disposable realm from
// docker-compose.keycloak.yml, and are skipped (NOT passed) without
// MYTAX_TEST_KEYCLOAK_ISSUER. CI must assert that variable is set.
const liveAudience = "mytax-api"

func liveIssuer(t *testing.T) string {
 t.Helper()
 issuer := os.Getenv("MYTAX_TEST_KEYCLOAK_ISSUER")
 if issuer == "" { t.Skip("MYTAX_TEST_KEYCLOAK_ISSUER not set: live Keycloak NOT verified") }
 return issuer
}

// liveVerifier is test-only: production NewAccessVerifier requires HTTPS.
func liveVerifier(t *testing.T, issuer, client, scope string) *AccessVerifier {
 t.Helper()
 ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
 defer cancel()
 provider, err := oidc.NewProvider(ctx, issuer)
 if err != nil { t.Fatal(err) }
 return &AccessVerifier{issuer: issuer, audience: liveAudience, client: client, scope: scope,
  verifier: provider.Verifier(&oidc.Config{ClientID: liveAudience})}
}

func liveToken(t *testing.T, issuer, client, user, pass string) (access, id string) {
 t.Helper()
 form := url.Values{"grant_type": {"password"}, "client_id": {client}, "username": {user}, "password": {pass}, "scope": {"openid"}}
 resp, err := http.Post(issuer+"/protocol/openid-connect/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
 if err != nil { t.Fatal(err) }
 defer resp.Body.Close()
 var body struct { AccessToken string `json:"access_token"`; IDToken string `json:"id_token"`; Error string `json:"error_description"` }
 if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatal(err) }
 if resp.StatusCode != 200 { t.Fatalf("token request failed %d: %s", resp.StatusCode, body.Error) }
 return body.AccessToken, body.IDToken
}

func TestLiveKeycloakValidAccessToken(t *testing.T) {
 issuer := liveIssuer(t)
 v := liveVerifier(t, issuer, "mytax-web", "mytax.api")
 access, _ := liveToken(t, issuer, "mytax-web", "alice", "ci-only-alice-pw")
 id, err := v.Verify(context.Background(), access)
 if err != nil { t.Fatalf("valid alice token rejected: %v", err) }
 if id.Subject == "" || !id.Verified || id.Issuer != issuer { t.Fatalf("unexpected identity %+v", id) }
}
