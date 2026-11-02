# mytax roadmap and open items

Status date: 2026-11-02. Developer preview. Not tax advice, not affiliated with LHDN. **No real taxpayer data, receipts or EA forms are accepted.**

Legend for verification (never merge these levels): **C** committed, **L** locally tested, **CI** CI passed, **Live** live integration tested, **Rev** independently reviewed.

## Where things stand
| Area | C | L | CI | Live | Rev |
|---|---|---|---|---|---|
| Go API unit tests (all packages) | yes | yes | yes (8 of 8 runs on branch `claude/quirky-wozniak-o9r7ih`, latest `d607860`) | no | no |
| PostgreSQL RLS tests (`TestTenantIsolationIntegration`, `TestPostgresRoleAndRLSIntegration`) | yes | skipped without a database | **passed on PostgreSQL 16 in CI** | n/a | no |
| Signed JWT / JWKS negative tests | yes | yes | yes | no (no Keycloak) | no |
| YA 2025 rate table and gross tax | yes | yes | yes | n/a | no |
| YA 2025 relief registry (31 entries) vs Act 53, Finance Acts 2024 and 2025 | yes | yes | yes | n/a | no |
| Coolify deployment | yes | Docker only | n/a | **no** | no |

The branch `claude/quirky-wozniak-o9r7ih` is 8 commits ahead of `main` and has **no pull request**. PR 2 (competitive landscape docs) is open and unmerged.

## Done so far (all Draft, nothing approved)
- Source records with hashes: pasted LHDN tables, Act 53 as at 21 May 2024 and 1 January 2026, Finance Act 2024 (Act 862), Akta Kewangan 2025 (Akta 874). See `docs/sources/`.
- All 31 YA 2025 registry amounts agree with the law texts; YA 2026-only items are listed so they are not applied to YA 2025 (`docs/ACT53_RECONCILIATION_YA2025.md`).
- Lifestyle relief modelled with Act citations and boundary tests (`docs/RELIEF_LIFESTYLE_YA2025.md`).
- First-home interest rule now has the s.46B conditions (citizen, occupied as residence, no income from the property).
- Candidate spouse and child deductions, provisional s.6A rebates, review-only dividend check (`docs/FAMILY_REBATE_DIVIDEND_YA2025.md`). Not wired into any endpoint.

## Open items

### 1. Verification and CI
- [ ] Open a pull request for the branch and merge to `main` once reviewed (not requested yet).
- [ ] Security tests listed in `AGENTS.md` that are **not** covered by the two PostgreSQL tests: concurrent pooled access, wrong-tenant **write** attempts (the app role has SELECT on four tables and INSERT on `evidence_metadata`, whose `WITH CHECK` policy no test exercises), a second subject in the second tenant. The existing tests cover the role flags, a no-subject read returning zero rows, a cross-tenant read being refused, and no subject leaking after pool reuse.
- [ ] Live Keycloak test (access token, client, scopes, tenant authorization). JWKS key rotation and caching tests. Token revocation. Roles beyond exact client and scope.
- [ ] Live Coolify deployment check (`SERVICE_URL_*` variables on the real instance).
- [ ] Compare each supplied PDF hash with the official publication. The PDFs are not in the repository.
- [ ] CI housekeeping: Node 20 deprecation warning for `actions/checkout@v4` and `actions/setup-go@v5`; a cache-key warning in the log. Many older Go files are not `gofmt`-formatted.

### 2. Tax rules
- [ ] **Professional review.** Nothing is approved. No release record, no `Approved` package, `ActivePackage` still returns an error by design.
- [ ] Official golden fixtures: LHDN worked examples to reconcile the whole calculation.
- [ ] LHDN Public Rulings for YA 2025 and the YA 2025 Form BE explanatory notes: not read. The Act text and Finance Acts were read for s.6, s.6A, s.46 to s.51 and Schedule 1 Part I only.
- [ ] **Rule coverage.** Rules exist only for lifestyle, education, vaccination and dental, first-home interest, and (in `internal/family`) spouse and child. No rule yet for: parents' medical, disability equipment, disabled individual, serious disease and fertility, medical screening, learning disability, sports, EV and compost, childcare, breastfeeding, SSPN, life insurance and EPF, education and medical insurance, PRS, SOCSO, self education fees.
- [ ] Conditions not yet modelled: s.46(3) and s.50(2) spouse deeming (needs the s.45(2) election context); s.49 insurance conditions and obligatory-contribution rule; s.51 deduction must be claimed; compost machine once across YA 2025 to 2027; breastfeeding equipment once per two years; spouse and child allocation across a whole household.
- [ ] Dividend above RM100,000 (2%): needs a decision on excess-only or whole dividend, per-company or total, and the Minister's prescription. Currently review only.
- [ ] Umrah departure-levy rebate (needs history across years). PCB and credits (s.132, s.133).
- [ ] Statutory rounding: no rule found in the texts read. The code rounds each band to the sen.
- [ ] Non-resident rate (30% flat, Schedule 1 para 1A) not implemented.
- [ ] YA 2023 and YA 2024: the code still rejects them; the owner decides whether to enable. A YA 2026 package would be built from Akta 874 and must not be copied from YA 2025.
- [ ] Fact vocabulary is inconsistent across rules (for example `qualifying_publication`, `qualifying_reading_material`, `internet_in_own_name`, `own_name`). Unifying changes API behaviour and needs a decision.
- [ ] Two engines (`policy` and `eligibility`) carry overlapping rules. Decide one source of truth.

### 3. Product wiring
- [ ] Connect `family`, rebates and the dividend review into `assessment` and the API, update API contracts and README.
- [ ] Frontend (`apps/web`) for statuses `verified`, `needs review`, `ineligible`, `unverified policy`, accessible and mobile friendly.

### 4. Gates before any real taxpayer data
(From `AGENTS.md`. None is operational.)
- [ ] Authenticated tenancy end to end, encrypted private storage, malware scanning, retention and deletion controls, audit logging, PDPA review, negative isolation tests.
- [ ] Independent threat review or penetration test.
- [ ] Preview endpoints stay disabled by default.

### 5. Claude connector (MCP server) — planned
See `docs/MCP_CONNECTOR_PLAN.md`. Plan only; phases P0 to P4 with gates. Depends on items 2, 3 and 4 above, in particular rule coverage and a canonical fact vocabulary.

## Suggested order
1. Pull request and review of the current branch; close the security test gaps in section 1.
2. Rule coverage for the most common receipt categories, with the canonical fact vocabulary.
3. Wire the engines into `assessment` and the API.
4. Live Keycloak and Coolify checks.
5. MCP phase P1 (synthetic data, local) once 2 and 3 are done.
6. Tax-professional review, then the gates in section 4, then any real user data.
