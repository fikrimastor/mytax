package security_test

import (
 "context"
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/http/httptest"
 "net/url"
 "os"
 "strings"
 "sync"
 "sync/atomic"
 "testing"
 "time"

 "github.com/jackc/pgx/v5/pgxpool"
 "github.com/labstack/echo/v4"
 "github.com/fikrimastor/mytax/services/api/internal/persistence"
 "github.com/fikrimastor/mytax/services/api/internal/security"
)

// End-to-end: real Keycloak token -> real AccessVerifier -> HTTP route ->
// TenantDB -> real PostgreSQL RLS (restricted app role). Skipped (NOT passed)
// unless every variable is set:
//   MYTAX_TEST_KEYCLOAK_ISSUER, MYTAX_TEST_DATABASE_URL (restricted app role),
//   MYTAX_TEST_ADMIN_DATABASE_URL (superuser, used only to seed/clean),
//   MYTAX_TEST_TENANT_A, MYTAX_TEST_TENANT_B.
// The route below mirrors cmd/api/main.go's secure route; it is not the
// production binary itself (that requires an HTTPS issuer).
func e2eToken(t *testing.T, issuer, client, user, pass string) (token, sub string) {
 t.Helper()
 form := url.Values{"grant_type": {"password"}, "client_id": {client}, "username": {user}, "password": {pass}, "scope": {"openid"}}
 resp, err := http.Post(issuer+"/protocol/openid-connect/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
 if err != nil { t.Fatal(err) }
 defer resp.Body.Close()
 var b struct{ AccessToken string `json:"access_token"` }
 json.NewDecoder(resp.Body).Decode(&b)
 if resp.StatusCode != 200 || b.AccessToken == "" { t.Fatalf("token request failed: %d", resp.StatusCode) }
 raw, err := base64.RawURLEncoding.DecodeString(strings.Split(b.AccessToken, ".")[1])
 if err != nil { t.Fatal(err) }
 var c struct{ Sub string `json:"sub"` }
 json.Unmarshal(raw, &c)
 if c.Sub == "" { t.Fatal("token has no sub") }
 return b.AccessToken, c.Sub
}

func TestKeycloakToRLSEndToEnd(t *testing.T) {
 issuer := os.Getenv("MYTAX_TEST_KEYCLOAK_ISSUER")
 appURL, adminURL := os.Getenv("MYTAX_TEST_DATABASE_URL"), os.Getenv("MYTAX_TEST_ADMIN_DATABASE_URL")
 tenA, tenB := os.Getenv("MYTAX_TEST_TENANT_A"), os.Getenv("MYTAX_TEST_TENANT_B")
 if issuer == "" || appURL == "" || adminURL == "" || tenA == "" || tenB == "" {
  t.Skip("e2e env incomplete: Keycloak->RLS NOT verified")
 }
 ctx := context.Background()
 const audience = "mytax-api"

 aliceTok, aliceSub := e2eToken(t, issuer, "mytax-web", "alice", "ci-only-alice-pw")
 bobTok, bobSub := e2eToken(t, issuer, "mytax-web", "bob", "ci-only-bob-pw")

 // Seed with the superuser; the app role under test never writes these.
 admin, err := pgxpool.New(ctx, adminURL)
 if err != nil { t.Fatal(err) }
 t.Cleanup(admin.Close) // registered first => runs LAST (cleanups are LIFO)
 for _, s := range []struct{ tenant, sub string }{{tenA, aliceSub}, {tenB, bobSub}} {
  if _, err := admin.Exec(ctx, "insert into tenant_memberships(tenant_id,subject_id,role) values($1::uuid,$2,'owner') on conflict do nothing", s.tenant, s.sub); err != nil { t.Fatal(err) }
 }
 admin.Exec(ctx, "delete from evidence_metadata where object_key='e2e-b-extra'")
 // Extra evidence row so tenant B's count (2) differs from tenant A's (1).
 if _, err := admin.Exec(ctx, "insert into evidence_metadata(tenant_id,claim_id,sha256_hex,object_key,byte_size,content_type) values($1::uuid,gen_random_uuid(),repeat('c',64),'e2e-b-extra',1,'application/pdf')", tenB); err != nil { t.Fatal(err) }
 t.Cleanup(func() {
  if _, err := admin.Exec(ctx, "delete from evidence_metadata where object_key='e2e-b-extra'"); err != nil { t.Errorf("cleanup failed: %v", err) }
  admin.Exec(ctx, "delete from tenant_memberships where subject_id = any($1)", []string{aliceSub, bobSub})
 })

 pool, err := pgxpool.New(ctx, appURL+"&pool_max_conns=3")
 if err != nil { t.Fatal(err) }
 defer pool.Close()
 verifier, err := security.NewTestAccessVerifier(ctx, issuer, audience, "mytax-web", "mytax.api")
 if err != nil { t.Fatal(err) }
 tdb, err := persistence.NewTenantDB(pool, issuer, audience)
 if err != nil { t.Fatal(err) }

 e := echo.New()
 e.GET("/v1/secure/tenants/:tenant_id/evidence/count", security.RequireAccess(verifier, func(c echo.Context) error {
  id, ok := c.Get("verified_identity").(security.Identity)
  if !ok { return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"}) }
  count, err := tdb.CountEvidence(c.Request().Context(), id, c.Param("tenant_id"))
  if err != nil { return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"}) }
  return c.JSON(http.StatusOK, map[string]any{"count": count})
 }))
 srv := httptest.NewServer(e)
 defer srv.Close()

 call := func(token, tenant string, extra map[string]string) (int, int64) {
  req, _ := http.NewRequest("GET", srv.URL+"/v1/secure/tenants/"+tenant+"/evidence/count", nil)
  if token != "" { req.Header.Set("Authorization", "Bearer "+token) }
  for k, v := range extra { req.Header.Set(k, v) }
  resp, err := http.DefaultClient.Do(req)
  if err != nil { t.Fatal(err) }
  defer resp.Body.Close()
  var b struct{ Count int64 `json:"count"` }
  json.NewDecoder(resp.Body).Decode(&b)
  return resp.StatusCode, b.Count
 }
 expect := func(name string, gotStatus int, gotCount int64, wantStatus int, wantCount int64) {
  t.Helper()
  if gotStatus != wantStatus || gotCount != wantCount { t.Errorf("%s: got %d/count %d, want %d/count %d", name, gotStatus, gotCount, wantStatus, wantCount) }
 }

 s, c := call(aliceTok, tenA, nil);                                  expect("alice->A", s, c, 200, 1)
 s, c = call(aliceTok, tenB, nil);                                   expect("alice->B", s, c, 403, 0)
 s, c = call(bobTok, tenB, nil);                                     expect("bob->B", s, c, 200, 2)
 s, c = call(bobTok, tenA, nil);                                     expect("bob->A", s, c, 403, 0)
 s, c = call("", tenA, nil);                                         expect("no token", s, c, 401, 0)
 s, c = call("garbage", tenA, nil);                                  expect("garbage token", s, c, 401, 0)
 s, c = call(aliceTok, "not-a-uuid", nil);                           expect("bad tenant id", s, c, 403, 0)
 s, c = call(aliceTok, tenB, map[string]string{"X-User-ID": bobSub, "X-Tenant-ID": tenB}); expect("spoofed headers", s, c, 403, 0)
 other, _ := e2eToken(t, issuer, "other-client", "alice", "ci-only-alice-pw")
 s, c = call(other, tenA, nil);                                      expect("wrong client token", s, c, 401, 0)
 noScope, _ := e2eToken(t, issuer, "no-scope-client", "alice", "ci-only-alice-pw")
 s, c = call(noScope, tenA, nil);                                    expect("no-scope token", s, c, 401, 0)

 // Concurrent pooled access (3 connections, 60 interleaved requests): identity
 // must never leak between pooled connections.
 var wg sync.WaitGroup
 errs := make(chan string, 60)
 for i := 0; i < 30; i++ {
  wg.Add(2)
  go func() { defer wg.Done(); if s, c := call(aliceTok, tenA, nil); s != 200 || c != 1 { errs <- fmt.Sprintf("concurrent alice->A got %d/%d", s, c) } }()
  go func() { defer wg.Done(); if s, _ := call(bobTok, tenA, nil); s != 403 { errs <- fmt.Sprintf("concurrent bob->A got %d", s) } }()
 }
 wg.Wait(); close(errs)
 for m := range errs { t.Error(m) }

 // No transaction-local subject may survive on pooled connections.
 var leaked string
 if err := pool.QueryRow(ctx, "select coalesce(current_setting('app.subject',true),'')").Scan(&leaked); err != nil { t.Fatal(err) }
 if leaked != "" { t.Fatalf("subject leaked across transactions: %q", leaked) }
}

// The API may boot before Keycloak is ready. With a lazy verifier it must
// answer 503 (never 200/401-as-success) until the provider is reachable, then
// accept a real token without a restart.
func TestKeycloakLazyVerifierRecovers(t *testing.T) {
 issuer := os.Getenv("MYTAX_TEST_KEYCLOAK_ISSUER")
 if issuer == "" { t.Skip("MYTAX_TEST_KEYCLOAK_ISSUER not set: lazy verifier NOT verified live") }
 ctx := context.Background()
 tok, _ := e2eToken(t, issuer, "mytax-web", "alice", "ci-only-alice-pw")

 var reachable atomic.Bool // flips when "Keycloak becomes ready"
 lazy, err := security.NewLazyVerifier(func(c context.Context) (security.TokenVerifier, error) {
  if !reachable.Load() { return nil, errors.New("simulated: provider not ready") }
  return security.NewTestAccessVerifier(c, issuer, "mytax-api", "mytax-web", "mytax.api")
 }, 5*time.Second, 50*time.Millisecond)
 if err != nil { t.Fatal(err) }

 e := echo.New()
 e.GET("/p", security.RequireAccess(lazy, func(c echo.Context) error { return c.NoContent(http.StatusOK) }))
 status := func(token string) int {
  req := httptest.NewRequest(http.MethodGet, "/p", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  rec := httptest.NewRecorder()
  e.ServeHTTP(rec, req)
  return rec.Code
 }
 if got := status(tok); got != http.StatusServiceUnavailable { t.Fatalf("before provider is ready: got %d want 503", got) }
 reachable.Store(true)
 time.Sleep(80 * time.Millisecond) // past minRetry
 if got := status(tok); got != http.StatusOK { t.Fatalf("after provider is ready: got %d want 200", got) }
 if got := status(tok + "x"); got != http.StatusUnauthorized { t.Fatalf("forged token after recovery: got %d want 401", got) }
 _ = ctx
}
