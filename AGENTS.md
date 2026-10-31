# mytax — Agent Development Instructions

## Purpose and status
mytax is an **open-source Malaysian personal income tax calculation and claim-verification platform**. Its primary objective is to help taxpayers claim only reliefs they are legally entitled to, detect unsupported/duplicate claims, and explain calculations transparently.

**Current status: developer preview. Not affiliated with LHDN/HASiL. Not tax advice. Do not accept real taxpayer identifiers, EA forms, receipts or personal documents until security and tax compliance gates pass.**

## Source of truth and tax correctness
1. Use Malaysia's Income Tax Act 1967 and effective amendments and subsidiary legislation as primary legal authority, then authoritative LHDN Public Rulings, official guidelines, tax-year-specific Form BE/B explanatory notes, relief schedules and rate tables.
2. Cite the precise official document, URL, applicable assessment year, effective date, relevant section/paragraph and access date for each statutory rule. Preserve immutable source snapshots/hashes where permitted. Budget announcements and third-party calculators are research leads, not authorization.
3. Never copy YA 2025 rules to YA 2026 or another YA without explicit verification. Distinguish published rules, draft rules, assumptions, and uncertain interpretations.
4. Never let OCR, an LLM, user-supplied flags or receipts alone establish eligibility. Evidence of spending is not proof of legal entitlement.
5. Maintain rule versioning and an approval audit. A tax-rule release needs complete eligibility requirements, nested/shared caps, exclusions, spouse/child allocation, dates, source traces, professional review and official golden fixtures.
6. **Fail closed:** unapproved draft claims must not decrease chargeable income, create tax refunds, or appear as approved deductions. Unknown cases must be marked `REVIEW_REQUIRED`. Do not state that LHDN approved an internal decision.
7. Use integer sen or a correct decimal library, with statute-specific rounding. Separate income classification, deductions, individual reliefs, rebates, credits, PCB and final amounts; preserve each calculation's derivation.
8. All tax outputs must declare what is and is not validated. Do not label a gross-tax preview, provisional rebate, or Form BE worksheet as a filing-ready result.

Official starting references:
- https://www.hasil.gov.my/individu/pelepasan-cukai/
- https://www.hasil.gov.my/individu/kadar-cukai/
- https://mytax.hasil.gov.my/
- Tax-year-specific explanatory notes on hasil.gov.my (verify URLs and content before use).

## Technology and repository
- Frontend: Nuxt 3, Nuxt UI, TypeScript (`apps/web`).
- API: Go 1.23+, Echo (`services/api`).
- Persistence: PostgreSQL with strict tenant boundaries; pgx.
- Authentication: Keycloak/OIDC; validate signed access JWTs, issuer, audience, expiration, access-token type, authorized client and required scopes.
- Deployment target: Docker/Coolify; use environment variables and secret management.
- Keep policy logic deterministic, testable, separate from AI, and governed by immutable approved versions. AI is limited to extraction/classification assistance.

## Project structure
- `services/api/internal/policy`: draft relief rules, registry, approval contracts.
- `services/api/internal/eligibility`: expense candidate screening, shared caps, exclusions.
- `services/api/internal/calculation`: rate arithmetic, not full statutory liability.
- `services/api/internal/assessment`, `reconciliation`, `formbe`, `income`: partial workflows.
- `services/api/internal/evidence`: synthetic metadata-only prototype.
- `services/api/internal/security`, `persistence`: authentication and tenant transaction boundaries.
- `services/api/migrations`: PostgreSQL RLS and role grants.
- `docs/`: documented limitations, sources, test gates and milestone notes.
- `.github/workflows/api.yml`: Go + PostgreSQL security test workflow.

## Security and privacy
- The repository is **public**. Never commit tokens, secrets, passwords (other than explicitly synthetic disposable CI fixtures), private keys, personally identifying records or real tax documents.
- Never trust user-supplied subjects, tenant IDs, authorization claims, or an unsigned/unverified JWT.
- PostgreSQL queries involving tenant data must use a verified identity, restricted non-owner role, transaction-local subject and tested RLS membership boundary. Never expose database connections to clients.
- Do not enable document upload or persistent taxpayer records until authenticated tenancy, encrypted private storage, malware scanning, retention/deletion controls, audit logging, PDPA review and negative isolation tests are operational.
- Never claim hash-chained logs are tamper-proof without an independent trust anchor. Avoid sensitive data in logs.
- Preview endpoints should be disabled by default where they expose sensitive security workflows.

## Coding standards
- Make small, reviewable changes with clear commit messages. Preserve backward compatibility unless change is intentional and documented.
- Prefer concise Go domain packages and explicit error handling. Validate input at every boundary; reject negative/out-of-range money, duplicate IDs and unsupported YA.
- Use predictable ordering where claims compete for caps; document allocation behavior.
- Do not introduce arbitrary LLM-executed tax formulas or unverifiable assumptions.
- Keep accessible, mobile-responsive Nuxt interfaces with clear statuses: `verified`, `needs review`, `ineligible` or `unverified policy`. Do not imply certainty unsupported by tests.
- Update README/docs and API contracts when behavior changes.

## Testing and release gates
- Run `cd services/api && go mod tidy && go test ./...` and frontend build checks where applicable. Prefer a checked-in lockfile and deterministic CI.
- Add boundary tests for every statutory tax bracket, relief cap, nested cap, mutually exclusive category, duplicate claim, assessment year, eligibility fact, source and rounding rule.
- Require actual PostgreSQL two-tenant isolation tests (including no-subject, wrong tenant, concurrent pooled access and transaction context leakage); tests that skip are **not** proof of passing.
- Require signed JWT/JWKS negative tests: expiry, signature, issuer, audience, token type, client, scopes and key rotation; then test with disposable Keycloak realm.
- Never mark a task or security control as verified merely because files were committed. Report separately: committed, locally tested, CI passed, live integration tested, and independently reviewed.
- Tax release requires qualified Malaysian tax-professional review and reconciliation with official LHDN examples. Security release requires independent threat review/penetration testing.
- If CI or tools are unavailable, say exactly what was not verified and do not claim tests passed.

## Agent workflow
1. Read this file before making changes. `CLAUDE.md` references this file.
2. Inspect current code, migrations and docs; identify the smallest next safe milestone.
3. Confirm source and effective YA before coding any substantive tax rule.
4. Implement changes with unit/integration tests and source-linked documentation.
5. Run available checks; inspect results. Never represent missing CI results as success.
6. Commit to the appropriate branch only when authorized by the user's request; summarize commit SHA, files changed, verified test results, gaps and next release gates.

## Current development priority
Stabilize CI and signed-JWT/JWKS tests, test PostgreSQL RLS using disposable tenants, then live Keycloak access-token and tenant authorization. **Do not enable production taxpayer uploads yet.**
