# Coolify Deployment (web + API) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deploy mytax web + API to Coolify as a fail-closed public developer preview using one `docker-compose.yml`.

**Architecture:** Two multi-stage Dockerfiles (Go API, Nuxt web) wired by a root compose file that uses Coolify `SERVICE_URL_*` variables for domains. API CORS origin becomes an env var. Auth and evidence previews stay off; no Postgres.

**Tech Stack:** Go 1.23 + Echo, Nuxt 3, Docker / Docker Compose, Coolify.

Spec: `docs/superpowers/specs/2026-10-31-coolify-deploy-design.md`

Working directory for all commands: repo root `/Users/fikrimastor/Sites/fikrimastor/mytax` unless stated. Append this trailer to every commit message:
`Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>`

## File structure

| File | Action | Responsibility |
|---|---|---|
| `services/api/go.mod`, `services/api/go.sum` | modify/create | tidy deps, lockfile |
| `apps/web/package-lock.json` | create | npm lockfile |
| `services/api/cmd/server/cors.go` | create | parse `MYTAX_CORS_ORIGINS` |
| `services/api/cmd/server/cors_test.go` | create | test for parsing |
| `services/api/cmd/server/main.go` | modify line 27 | use `corsOrigins()` |
| `services/api/Dockerfile`, `services/api/.dockerignore` | create | API image |
| `apps/web/Dockerfile`, `apps/web/.dockerignore` | create | web image |
| `apps/web/app.vue` | modify | global "no real data" notice |
| `docker-compose.yml`, `.env.example` | create | stack definition |
| `docs/DEPLOY_COOLIFY.md` | create | manual Coolify steps |

---

### Task 1: Commit lockfiles

**Files:** `services/api/go.mod`, `services/api/go.sum`, `apps/web/package-lock.json`

- [ ] **Step 1: Confirm state**

Run: `git status --short`
Expected: ` M services/api/go.mod`, `?? apps/web/package-lock.json`, `?? services/api/go.sum`

- [ ] **Step 2: Verify tests still pass with the tidy module**

Run: `cd services/api && go mod tidy && go test -count=1 ./... 2>&1 | tail -15`
Expected: every package `ok` (persistence RLS tests skip or pass depending on env; unset env is fine here).

- [ ] **Step 3: Commit**

```bash
git add services/api/go.mod services/api/go.sum apps/web/package-lock.json
git commit -m "chore(deps): add Go and npm lockfiles for reproducible builds

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Configurable CORS origins (TDD)

**Files:**
- Create: `services/api/cmd/server/cors.go`
- Create: `services/api/cmd/server/cors_test.go`
- Modify: `services/api/cmd/server/main.go:27`

- [ ] **Step 1: Write the failing test**

`services/api/cmd/server/cors_test.go`:

```go
package main

import (
	"reflect"
	"testing"
)

func TestCORSOrigins(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty falls back to local dev", "", []string{"http://localhost:3000"}},
		{"blank falls back to local dev", "  , ,", []string{"http://localhost:3000"}},
		{"single origin", "https://mytax.example.com", []string{"https://mytax.example.com"}},
		{"trims and splits", " https://a.example.com , https://b.example.com ", []string{"https://a.example.com", "https://b.example.com"}},
		{"wildcard is rejected", "*", []string{"http://localhost:3000"}},
		{"wildcard among others is dropped", "https://a.example.com,*", []string{"https://a.example.com"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := corsOrigins(tc.raw); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("corsOrigins(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd services/api && go test ./cmd/server/ -run TestCORSOrigins`
Expected: FAIL, `undefined: corsOrigins`

- [ ] **Step 3: Implement**

`services/api/cmd/server/cors.go`:

```go
package main

import "strings"

const defaultCORSOrigin = "http://localhost:3000"

// corsOrigins parses a comma-separated MYTAX_CORS_ORIGINS value.
// Wildcards are dropped (fail closed); with nothing valid left, only local dev is allowed.
func corsOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o == "" || o == "*" {
			continue
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		return []string{defaultCORSOrigin}
	}
	return out
}
```

- [ ] **Step 4: Use it in main.go**

In `services/api/cmd/server/main.go` line 27, replace `AllowOrigins:[]string{"http://localhost:3000"}` with `AllowOrigins:corsOrigins(os.Getenv("MYTAX_CORS_ORIGINS"))`. Resulting line:

```go
 e.Use(middleware.Recover(),middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins:corsOrigins(os.Getenv("MYTAX_CORS_ORIGINS"))}))
```

- [ ] **Step 5: Run tests**

Run: `cd services/api && go vet ./... && go test -count=1 ./cmd/server/ ./internal/...  2>&1 | tail -15`
Expected: `ok` for `cmd/server` and all internal packages.

- [ ] **Step 6: Commit**

```bash
git add services/api/cmd/server/cors.go services/api/cmd/server/cors_test.go services/api/cmd/server/main.go
git commit -m "feat(api): read CORS origins from MYTAX_CORS_ORIGINS

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 3: API Dockerfile

**Files:** Create `services/api/Dockerfile`, `services/api/.dockerignore`

- [ ] **Step 1: Write `.dockerignore`**

`services/api/.dockerignore`:

```
testdata
*.log
.env
.env.*
```

- [ ] **Step 2: Write Dockerfile**

`services/api/Dockerfile`:

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/server /server
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/server"]
```

- [ ] **Step 3: Build and smoke test**

Run:
```bash
docker build -t mytax-api:test services/api
docker run --rm -d --name mytax-api-test -p 18080:8080 mytax-api:test
sleep 3; curl -s localhost:18080/healthz; curl -s "localhost:18080/v1/policy/status?ya=2025"
docker rm -f mytax-api-test
```
Expected: `{"status":"ok"}` and a `NOT_APPROVED` policy status.

- [ ] **Step 4: Commit**

```bash
git add services/api/Dockerfile services/api/.dockerignore
git commit -m "build(api): add multi-stage non-root Dockerfile

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Web Dockerfile and notice

**Files:** Create `apps/web/Dockerfile`, `apps/web/.dockerignore`; modify `apps/web/app.vue`

- [ ] **Step 1: Write `.dockerignore`**

`apps/web/.dockerignore`:

```
node_modules
.nuxt
.output
*.log
.env
.env.*
```

- [ ] **Step 2: Write Dockerfile**

`apps/web/Dockerfile`:

```dockerfile
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=build /app/.output ./.output
USER node
ENV NITRO_HOST=0.0.0.0 NITRO_PORT=3000
EXPOSE 3000
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/ >/dev/null || exit 1
CMD ["node", ".output/server/index.mjs"]
```

- [ ] **Step 3: Add a global preview notice**

Replace the contents of `apps/web/app.vue` with:

```vue
<template>
  <div>
    <p class="subtle" role="note" style="margin:0;padding:8px 16px;background:#fef3c7;color:#92400e;font-size:13px">
      Developer preview. Not tax advice and not affiliated with LHDN. Do not enter real personal data, IC numbers, EA forms or receipts.
    </p>
    <NuxtPage />
  </div>
</template>
```

- [ ] **Step 4: Build and smoke test**

Run:
```bash
docker build -t mytax-web:test apps/web
docker run --rm -d --name mytax-web-test -p 13000:3000 mytax-web:test
sleep 5; curl -s -o /dev/null -w "%{http_code}\n" localhost:13000/; curl -s localhost:13000/ | grep -c "Do not enter real personal data"
docker rm -f mytax-web-test
```
Expected: `200`, and the grep count is at least `1`.

- [ ] **Step 5: Commit**

```bash
git add apps/web/Dockerfile apps/web/.dockerignore apps/web/app.vue
git commit -m "build(web): add multi-stage non-root Dockerfile and preview notice

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Compose file, env example, Coolify guide

**Files:** Create `docker-compose.yml`, `.env.example`, `docs/DEPLOY_COOLIFY.md`

- [ ] **Step 1: Write `docker-compose.yml`**

```yaml
services:
  api:
    build:
      context: ./services/api
    restart: unless-stopped
    environment:
      - SERVICE_URL_API_8080
      - PORT=8080
      - MYTAX_CORS_ORIGINS=${SERVICE_URL_WEB}
      # Previews stay OFF. Do not set MYTAX_ENABLE_AUTH_PREVIEW or MYTAX_ENABLE_EVIDENCE_PREVIEW.
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/healthz"]
      interval: 15s
      timeout: 3s
      retries: 3
      start_period: 5s

  web:
    build:
      context: ./apps/web
    restart: unless-stopped
    depends_on:
      api:
        condition: service_healthy
    environment:
      - SERVICE_URL_WEB_3000
      - NUXT_PUBLIC_API_BASE=${SERVICE_URL_API}
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:3000/"]
      interval: 15s
      timeout: 3s
      retries: 3
      start_period: 10s
```

- [ ] **Step 2: Write `.env.example`**

```
# Local docker compose only. On Coolify the domain variables are generated for you.
# SAFE DEFAULTS: leave both previews OFF. Turning them on needs Postgres + Keycloak
# and must pass the security gates in AGENTS.md first.
MYTAX_ENABLE_AUTH_PREVIEW=false
MYTAX_ENABLE_EVIDENCE_PREVIEW=false
# Local compose values (Coolify sets these itself):
SERVICE_URL_WEB=http://localhost:3000
SERVICE_URL_API=http://localhost:8080
```

- [ ] **Step 3: Validate and run the stack locally**

Run:
```bash
docker compose --env-file .env.example config >/dev/null && echo CONFIG_OK
docker compose --env-file .env.example up -d --build
sleep 15; docker compose ps
curl -s localhost:8080/healthz 2>/dev/null || echo "api port not published (expected: compose does not publish ports)"
docker compose exec -T web wget -qO- http://api:8080/healthz
docker compose down
```
Expected: `CONFIG_OK`, both services `healthy`, and `{"status":"ok"}` from web reaching `api` by service name. (Ports are not published on purpose; Coolify's proxy routes traffic.)

- [ ] **Step 4: Write `docs/DEPLOY_COOLIFY.md`**

```markdown
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
```

- [ ] **Step 5: Commit**

```bash
git add docker-compose.yml .env.example docs/DEPLOY_COOLIFY.md
git commit -m "build(deploy): add Coolify docker-compose, env example and guide

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Final check

- [ ] **Step 1:** `cd services/api && go test -count=1 ./... 2>&1 | tail -15` → all `ok`.
- [ ] **Step 2:** `git status --short` → clean; `git log --oneline -6` shows the five commits above plus the spec.
- [ ] **Step 3:** Report to the user, separately: committed, tested locally, CI (after push), live Coolify (not verified), independently reviewed (no). Do not push unless asked.
