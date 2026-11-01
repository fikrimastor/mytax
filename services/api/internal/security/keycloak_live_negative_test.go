package security

import (
 "context"
 "strings"
 "testing"
 "time"
 "github.com/coreos/go-oidc/v3/oidc"
)

// Negative matrix against a real Keycloak. Each case asserts the SPECIFIC
// rejection reason so a case cannot pass for the wrong reason.
func requireRejected(t *testing.T, err error, reason string) {
 t.Helper()
 if err == nil { t.Fatalf("token was ACCEPTED, expected rejection containing %q", reason) }
 if !strings.Contains(err.Error(), reason) { t.Fatalf("rejected for the wrong reason: got %q, want substring %q", err.Error(), reason) }
}

func TestLiveKeycloakNegativeMatrix(t *testing.T) {
 issuer := liveIssuer(t)
 ctx := context.Background()
 alice, aliceID := liveToken(t, issuer, "mytax-web", "alice", "ci-only-alice-pw")

 t.Run("client not allowed (azp)", func(t *testing.T) {
  other, _ := liveToken(t, issuer, "other-client", "alice", "ci-only-alice-pw")
  _, err := liveVerifier(t, issuer, "mytax-web", "mytax.api").Verify(ctx, other)
  requireRejected(t, err, "unauthorized client")
 })
 t.Run("required scope missing", func(t *testing.T) {
  _, err := liveVerifier(t, issuer, "mytax-web", "mytax.admin").Verify(ctx, alice)
  requireRejected(t, err, "required scope missing")
 })
 t.Run("token without API audience", func(t *testing.T) {
  noScope, _ := liveToken(t, issuer, "no-scope-client", "alice", "ci-only-alice-pw")
  _, err := liveVerifier(t, issuer, "no-scope-client", "mytax.api").Verify(ctx, noScope)
  requireRejected(t, err, "audience")
 })
 t.Run("ID token used as access token", func(t *testing.T) {
  _, err := liveVerifier(t, issuer, "mytax-web", "mytax.api").Verify(ctx, aliceID)
  requireRejected(t, err, "audience")
 })
 t.Run("wrong issuer (other realm)", func(t *testing.T) {
  masterIssuer := strings.Replace(issuer, "/realms/mytax-ci", "/realms/master", 1)
  _, err := liveVerifier(t, masterIssuer, "mytax-web", "mytax.api").Verify(ctx, alice)
  requireRejected(t, err, "different provider")
 })
 t.Run("expired", func(t *testing.T) {
  provider, err := oidc.NewProvider(ctx, issuer)
  if err != nil { t.Fatal(err) }
  late := func() time.Time { return time.Now().Add(10 * time.Minute) } // lifespan is 300s
  v := &AccessVerifier{issuer: issuer, audience: liveAudience, client: "mytax-web", scope: "mytax.api",
   verifier: provider.Verifier(&oidc.Config{ClientID: liveAudience, Now: late})}
  _, err = v.Verify(ctx, alice)
  requireRejected(t, err, "expired")
 })
 t.Run("tampered signature", func(t *testing.T) {
  parts := strings.Split(alice, ".")
  sig := []byte(parts[2])
  if sig[10] == 'A' { sig[10] = 'B' } else { sig[10] = 'A' }
  _, err := liveVerifier(t, issuer, "mytax-web", "mytax.api").Verify(ctx, parts[0]+"."+parts[1]+"."+string(sig))
  requireRejected(t, err, "signature")
 })
 t.Run("malformed and empty", func(t *testing.T) {
  v := liveVerifier(t, issuer, "mytax-web", "mytax.api")
  if _, err := v.Verify(ctx, "not-a-jwt"); err == nil { t.Fatal("malformed token accepted") }
  if _, err := v.Verify(ctx, ""); err == nil { t.Fatal("empty token accepted") }
 })
}
