// Live Keycloak tests. Skipped unless MYTAX_TEST_KEYCLOAK_ISSUER and
// MYTAX_TEST_KEYCLOAK_CA are set (CI starts Keycloak with a self-signed cert).
package livekeycloak

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/fikrimastor/mytax/services/api/internal/persistence"
	"github.com/fikrimastor/mytax/services/api/internal/security"
)

const (
	audience = "mytax-api"
	scope    = "mytax:evidence:read"
	tenantA  = "11111111-1111-4111-8111-111111111111"
	tenantB  = "22222222-2222-4222-8222-222222222222"
)

type env struct {
	issuer string
	ctx    context.Context // carries the CA-trusting HTTP client for go-oidc
	client *http.Client
}

func setup(t *testing.T) env {
	t.Helper()
	issuer, caPath := os.Getenv("MYTAX_TEST_KEYCLOAK_ISSUER"), os.Getenv("MYTAX_TEST_KEYCLOAK_CA")
	if issuer == "" || caPath == "" {
		t.Skip("live Keycloak not configured")
	}
	pem, err := os.ReadFile(caPath)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		t.Fatal("bad CA file")
	}
	c := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}
	return env{issuer, oidc.ClientContext(context.Background(), c), c}
}

func (e env) token(t *testing.T, clientID, scopes string) string {
	t.Helper()
	resp, err := e.client.PostForm(e.issuer+"/protocol/openid-connect/token", url.Values{
		"grant_type": {"password"}, "client_id": {clientID}, "username": {"ci-user"},
		"password": {"ci-only-password"}, "scope": {scopes},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.AccessToken == "" {
		t.Fatalf("token request for %s failed: status=%d err=%v %s", clientID, resp.StatusCode, err, body.Error)
	}
	return body.AccessToken
}

func (e env) verifier(t *testing.T, clientID string) *security.AccessVerifier {
	t.Helper()
	v, err := security.NewAccessVerifier(e.ctx, e.issuer, audience, clientID, scope)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// Cases 1+2: real Keycloak tokens, real claim mappers.
func TestLiveKeycloakTokens(t *testing.T) {
	e := setup(t)
	v := e.verifier(t, "mytax-web")

	id, err := v.Verify(e.ctx, e.token(t, "mytax-web", "openid "+scope))
	if err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if id.Subject != "33333333-3333-4333-8333-333333333333" || !id.Verified {
		t.Fatalf("unexpected identity %+v", id)
	}

	rejects := []struct{ name, token string }{
		{"missing scope", e.token(t, "mytax-web", "openid")},
		{"wrong client", e.token(t, "mytax-other", "openid "+scope)},
	}
	for _, tc := range rejects {
		if _, err := v.Verify(e.ctx, tc.token); err == nil {
			t.Errorf("%s: token accepted", tc.name)
		}
	}

	// Expiry: accepted while fresh, so the later rejection is due to exp only.
	short := e.verifier(t, "mytax-short")
	tok := e.token(t, "mytax-short", "openid "+scope)
	if _, err := short.Verify(e.ctx, tok); err != nil {
		t.Fatalf("fresh short-lived token rejected: %v", err)
	}
	time.Sleep(7 * time.Second)
	if _, err := short.Verify(e.ctx, tok); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expired token not rejected as expired: %v", err)
	}
}

// Case 3: Keycloak token -> Echo -> RequireAccess -> RLS (sub -> app.subject).
func TestLiveKeycloakHTTPAndRLS(t *testing.T) {
	e := setup(t)
	uri := os.Getenv("MYTAX_TEST_DATABASE_URL")
	if uri == "" {
		t.Skip("test database not configured")
	}
	pool, err := pgxpool.New(context.Background(), uri)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	db, err := persistence.NewTenantDB(pool, e.issuer, audience)
	if err != nil {
		t.Fatal(err)
	}

	srv := echo.New()
	srv.GET("/t/:tenant/count", security.RequireAccess(e.verifier(t, "mytax-web"), func(c echo.Context) error {
		n, err := db.CountEvidence(c.Request().Context(), c.Get("verified_identity").(security.Identity), c.Param("tenant"))
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		return c.JSON(http.StatusOK, map[string]int64{"count": n})
	}))
	get := func(path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		return rec
	}

	good := e.token(t, "mytax-web", "openid "+scope)
	if rec := get("/t/"+tenantA+"/count", good); rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != `{"count":1}` {
		t.Fatalf("member tenant: %d %s", rec.Code, rec.Body)
	}
	if rec := get("/t/"+tenantB+"/count", good); rec.Code != 403 {
		t.Fatalf("non-member tenant: %d %s", rec.Code, rec.Body)
	}
	if rec := get("/t/"+tenantA+"/count", ""); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	if rec := get("/t/"+tenantA+"/count", e.token(t, "mytax-other", "openid "+scope)); rec.Code != 401 {
		t.Fatalf("wrong client: %d", rec.Code)
	}
}
