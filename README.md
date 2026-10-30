# mytax

Open-source Malaysian personal income tax claim checker — early development, **not tax advice** and **not affiliated with LHDN**.

## Safety first
- OCR/AI never authorizes a claim.
- The API never approves reliefs or calculates final tax without a verified, year-specific LHDN rule package.
- No taxpayer login or secure document storage yet: **do not submit real personal data, EA forms, receipts or IC numbers**.

## Stack
- Nuxt 3 + Nuxt UI frontend in `apps/web`
- Go 1.23 + Echo API in `services/api`
- PostgreSQL planned (not implemented)

## Run
```sh
cd services/api && go mod tidy && go test ./... && go run ./cmd/server
# separately
cd apps/web && npm install && npm run dev
```
Local API: `http://localhost:8080`. Override frontend API with `NUXT_PUBLIC_API_BASE`.

## API
- `GET /healthz`
- `GET /v1/policy/status?ya=2025`
- `POST /v1/claims/evaluate` (generic evidence and duplicate flags only)

## Roadmap
1. Source and professionally review LHDN official YA-specific rules.
2. Build versioned policy packages with source proofs and unit/golden tests.
3. Implement detailed eligibility, shared caps and statutory computation.
4. Add secure user accounts, encrypted evidence and receipt ingestion.
5. Map verified amounts to Form BE, then introduce business/Form B support.

See `docs/COMPLIANCE.md` for policy approval requirements.
