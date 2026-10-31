# Milestone 17 — cryptographic JWT/JWKS integration test harness

The `security/jwks_integration_test.go` suite hosts a local OpenID discovery + JWKS server, signs synthetic RS256 JWTs, and runs the real go-oidc verification path. It checks valid tokens and rejection of:
- altered/wrong RSA signatures;
- expired tokens;
- wrong audience and issuer;
- unauthorized client (`azp`);
- absent required scope;
- ID-token type instead of access-token type;
- malformed compact JWT.

The test issuer uses HTTP on loopback **only** within a local test and constructs the verifier directly. Production `NewAccessVerifier` still requires HTTPS.

**Not yet included:** live Keycloak, key rotation, JWKS caching behavior, real HTTP+database integration, token revocation, and role/scope authorization beyond exact client and scope. No taxpayer data, secrets or private keys are committed; generated keys exist only in test memory.

The tests must execute in CI before this milestone can be considered verified. Do not treat committed tests alone as passing results.
