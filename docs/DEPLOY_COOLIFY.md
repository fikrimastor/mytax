# Deploy to Coolify (web + API)

Status: **developer preview**. Not tax advice. Do not enable uploads or real taxpayer data.

## Steps
1. In Coolify: New Resource → Public/Private Repository → `fikrimastor/mytax`, branch `main`.
2. Build Pack: **Docker Compose**. Compose file: `/docker-compose.yml`.
3. Set domains: `web` → e.g. `https://mytax.example.com`, `api` → e.g. `https://api.mytax.example.com`.
   Coolify fills `SERVICE_URL_WEB` and `SERVICE_URL_API`; compose uses them for `MYTAX_CORS_ORIGINS` and `NUXT_PUBLIC_API_BASE`.
4. Do NOT add `MYTAX_ENABLE_AUTH_PREVIEW` or `MYTAX_ENABLE_EVIDENCE_PREVIEW`.
5. Deploy. Wait for both healthchecks to be healthy.

## Verify
- `curl https://api.mytax.example.com/healthz` → `{"status":"ok"}`
- `curl "https://api.mytax.example.com/v1/policy/status?ya=2025"` → `NOT_APPROVED`
- Open the web URL; the preview notice is shown and the form calls the API without CORS errors.

## Verification status
- Committed and locally tested with `docker compose` only.
- Not yet verified on a live Coolify instance. Check the `SERVICE_URL_*` variable behaviour against the Coolify docs for your version on first deploy.
- Not independently reviewed.
