package security

import (
 "context"
 "github.com/coreos/go-oidc/v3/oidc"
)

// NewTestAccessVerifier is visible ONLY to this package's _test files
// (security_test). Production NewAccessVerifier still requires HTTPS; this
// exists so e2e tests can use a loopback Keycloak without weakening it.
func NewTestAccessVerifier(ctx context.Context, issuer, audience, client, scope string) (*AccessVerifier, error) {
 provider, err := oidc.NewProvider(ctx, issuer)
 if err != nil { return nil, err }
 return &AccessVerifier{issuer: issuer, audience: audience, client: client, scope: scope,
  verifier: provider.Verifier(&oidc.Config{ClientID: audience})}, nil
}
