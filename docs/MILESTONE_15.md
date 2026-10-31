# Milestone 15 — protected-route authorization

The access JWT verification now requires a configured authorized party/client and exact scope, in addition to signed issuer, audience and expiry validation. The protected evidence-count endpoint is disabled by default and requires:

- MYTAX_ENABLE_AUTH_PREVIEW=true
- MYTAX_OIDC_ISSUER=https://<keycloak>/realms/<realm>
- MYTAX_OIDC_AUDIENCE=mytax-api
- MYTAX_OIDC_ALLOWED_CLIENT=mytax-web
- MYTAX_OIDC_REQUIRED_SCOPE=mytax:evidence:read
- DATABASE_URL for a restricted, tenant-isolated PostgreSQL role.

Keycloak audience and scope mappers must be configured to supply these claims on **access tokens**. `azp` is preferred; `client_id` fallback is supported. Only use this client fallback if explicitly permitted by a verified deployment policy. `typ` must be Bearer or at+jwt when provided; deployment should enforce an access-token-specific header/type contract before production.

Unit tests exercise unauthorized client, absent scope, misleading scope substring, and ID-token type. Live signed JWT/JWKS and full HTTP integration with a test Keycloak realm are still outstanding. In particular, a passing JWT signature by itself does not prove an actor can view a tenant's evidence.

No real tax receipts or taxpayer PII should be sent yet.
