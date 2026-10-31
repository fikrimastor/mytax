# Coolify deployment (web + API) — design

Date: 2026-10-31
Status: approved by user in chat; awaiting spec review

## Goal

Deploy mytax to Coolify as a public **developer preview** using one `docker-compose.yml` with two services: `web` (Nuxt) and `api` (Go/Echo). No Postgres or Keycloak in this iteration.

## Non-goals

- Postgres, migrations, Keycloak, real JWT auth on the server (later iteration).
- Taxpayer uploads, real personal data (blocked by `AGENTS.md`).
- Verifying a live Coolify deployment from this session (Coolify MCP returns 401).

## Constraints from the codebase

- API reads `PORT` (default 8080). Auth preview needs `MYTAX_ENABLE_AUTH_PREVIEW=true` plus OIDC and `DATABASE_URL` vars; evidence preview needs `MYTAX_ENABLE_EVIDENCE_PREVIEW=true`. Both are off by default and must stay off.
- CORS is hardcoded to `http://localhost:3000` in `services/api/cmd/server/main.go`; a real domain would be blocked by browsers.
- Web reads the API address from `NUXT_PUBLIC_API_BASE` (Nuxt `runtimeConfig.public.apiBase`), overridable at runtime.
- No Dockerfile, compose file, or lockfiles exist yet (`go.sum`, `package-lock.json`).
- Repository is public: no secrets in git.

## Chosen approach

One `docker-compose.yml` at the repo root, deployed with Coolify's "Docker Compose" build pack. Rejected: two separate Coolify apps (two resources to manage), Nixpacks (little control in a monorepo).

## Components

1. `services/api/Dockerfile` — multi-stage; `CGO_ENABLED=0` Go build; small runtime image; non-root user; `HEALTHCHECK` on `/healthz`.
2. `apps/web/Dockerfile` — multi-stage; `npm ci`, `nuxt build`; run `node .output/server/index.mjs`; non-root user; healthcheck on `/`.
3. `docker-compose.yml` — services `api` and `web`; Coolify domain variables `SERVICE_FQDN_API_8080` and `SERVICE_FQDN_WEB_3000`; `web` gets `NUXT_PUBLIC_API_BASE` pointing to the API domain; `api` gets `MYTAX_CORS_ORIGINS` set to the web domain; preview flags not set.
4. `.dockerignore` per service, root `.env.example` (names and safe defaults only, preview flags documented as `false` with a warning).
5. `docs/DEPLOY_COOLIFY.md` — manual Coolify steps (create resource, set domains, env vars, deploy, check health).
6. Code change: `main.go` reads `MYTAX_CORS_ORIGINS` (comma-separated); empty falls back to `http://localhost:3000`. One unit test for origin parsing.
7. Web shows a "developer preview, do not enter real data" notice if `app.vue` does not already have one (check before adding).

## Fail-closed defaults

- Auth and evidence previews stay disabled in compose and `.env.example`.
- API starts and serves `/healthz` and `/v1/policy/status` (returns `NOT_APPROVED`, `tax_calculation_enabled: false`) with no database.
- Coolify waits for healthchecks before routing traffic.

## Verification

- Local: `docker compose build && docker compose up`, then curl `/healthz` and `/v1/policy/status`, check web returns 200, and check a CORS preflight from the configured origin.
- `go test ./...` still passes.
- Report separately (per `AGENTS.md`): committed, tested locally, live Coolify deployment (not verified), independently reviewed (no).

## Commit plan

1. Lockfiles: `services/api/go.sum`, `apps/web/package-lock.json`, tidy `services/api/go.mod`.
2. CORS env var + test.
3. Dockerfiles, compose, `.dockerignore`, `.env.example`, `docs/DEPLOY_COOLIFY.md`, preview notice if needed.

## Open items

- Real domain names are chosen by the user in Coolify; the compose file uses Coolify-generated variables only.
