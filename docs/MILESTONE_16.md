# Milestone 16 — HTTP authentication gate and negative cases

Added injectable HTTP authentication boundary and httptest coverage for missing bearer, invalid token, valid verified identity and missing verifier. These tests use a **fake verifier**; they do not exercise signatures, JWKS discovery, Keycloak keys or token expiry.

The real access verifier now requires `typ` to be explicitly `Bearer` or `at+jwt` in the signed JWT claims and requires an exact client and scope. This is a Keycloak-specific compatibility requirement and must be checked against the realm's actual token claims before enabling the protected endpoint.

**Outstanding**: end-to-end signed JWT/JWKS rotation tests, actual Keycloak realm, database-backed HTTP permission tests, successful CI check and threat review. Do not claim the full Milestone 16 goal is complete before these gates pass.

No sensitive taxpayer evidence or real authentication credentials should be committed or uploaded.
