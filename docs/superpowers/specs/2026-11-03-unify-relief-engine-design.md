# Unify the relief engine (policy as the single engine and data source)

Date: 2026-11-03. Status: design, awaiting review. Developer preview; not tax advice.

## Why

Two packages evaluate relief claims and carry overlapping rules:

| | `internal/policy` | `internal/eligibility` |
|---|---|---|
| Serves | `GET /v1/policy/catalog`, `POST /v1/assessment/preview` | `POST /v1/eligibility/preview` |
| Rules | citations (`Sources`), Draft/Approved status | no citations |
| Duplicate expense | caller sends `duplicate` bool | server detects repeated `expense_key` |
| Evidence | caller sends `evidence_present` bool | server checks `evidence_reference` is set |
| Mutually exclusive reliefs | none | `ExclusiveFamily` |
| Overflow and claim-count guards | in `assessment` only | in `Evaluate` |

Every new relief rule would have to be written twice, and a parity test (`eligibility/catalog_parity_test.go`) is the only thing stopping drift. The rule coverage work (16 relief categories, `docs/ROADMAP.md` section 2) cannot start safely on top of that. This design removes the duplication first.

## Goal and non-goals

Goal: one rule catalog and one `Evaluate`, with no behaviour that is weaker than today's. Nothing becomes Approved.

Non-goals: new relief rules (next sub-project, medical batch), `family`/rebate/dividend wiring, the web frontend, authentication.

## Design

### 1. Structure
- `internal/policy` becomes the only package holding rule data (`Package`, `Rule`, `Group`, `Source`) and the only `Evaluate`.
- `internal/eligibility` is deleted. `cmd/server` calls `policy.Evaluate(policy.YA2025Draft(), claims)` for both `/v1/eligibility/preview` and (through `assessment`) `/v1/assessment/preview`.
- `policy.Rule` gains `ExclusiveFamily string`. The eligibility `Pool` concept maps to the existing policy `Group`; `MaxSen` maps to `LimitSen`.
- `Evaluate` carries over every eligibility safeguard: at most 1000 claims, an overflow guard on the running total, duplicate detection by `expense_key`, a required `evidence_reference`, mutually exclusive families, per-rule and per-group limits. The policy Draft/Approved handling and package validation stay.
- Year support stays at 2025 only: the endpoints keep rejecting other years exactly as now.

### 2. Claim and response shape (one shape for both endpoints)
Claim:

```json
{"id":"c1","rule_id":"MEDICAL_VACCINATION","amount_sen":50000,
 "expense_key":"inv-123","evidence_reference":"doc-9","facts":{"resident":true}}
```

- `evidence_present` and `duplicate` (caller-asserted booleans) are removed. A caller can no longer declare its own claim "not a duplicate".
- Decision: `claim_id`, `status`, `provisional_amount_sen`, `approved_amount_sen`, `reasons`.
- Result: policy's existing fields plus `provisional_total_sen` and `filing_ready` (always `false`).
- Reason codes:
  - kept from policy: `UNKNOWN_RELIEF_RULE`, `UNVERIFIED_FACT:<fact>`, `EVIDENCE_MISSING`, `DRAFT_POLICY_NOT_APPROVED`;
  - kept from eligibility: `EXPENSE_IDENTIFIER_MISSING`, `DUPLICATE_EXPENSE`, `MUTUALLY_EXCLUSIVE_RELIEF`, `LIMIT_REACHED`.

### 3. Rules and fact vocabulary
- Rules that exist only in eligibility move into `YA2025Draft()` under policy IDs: `EDUCATION_UPSKILL` becomes `EDUCATION_SELF_SKILLS`, `EDUCATION_OTHER` becomes `EDUCATION_SELF_OTHER`, and `MEDICAL_DENTAL` is added (limit RM1,000, group `MEDICAL_SELF_FAMILY`, facts `resident`, `within_year`, `qualifying_dental`, `qualifying_beneficiary`, `no_other_claim`).
- New file `internal/policy/facts.go` lists the canonical fact names as constants: `snake_case`; `qualifying_<thing>` for properties of the item or service; `<thing>_in_own_name` for ownership; shared facts `resident`, `within_year`, `no_other_claim`, `qualifying_beneficiary`.
- A test fails if any rule in any catalog requires a fact that is not in the canonical list.
- Where a merged rule had weaker facts in eligibility than in policy, the policy facts win (stricter). Where it had facts policy lacks, they are added to the policy rule.

### 4. API changes (developer preview; the only client sends `claims: []`)
- `POST /v1/eligibility/preview`: request unchanged. Response field names follow the policy shape; education rule IDs change; reason codes follow section 2.
- `POST /v1/assessment/preview`: claims use `expense_key` and `evidence_reference`.
- `GET /v1/policy/catalog`: rules gain `exclusive_family`; new and merged rules appear.
- `docs/ASSESSMENT_PREVIEW.md`, `docs/RELIEF_LIFESTYLE_YA2025.md`, `docs/ROADMAP.md` and the project-structure list in `AGENTS.md` (which still names `internal/eligibility`) are updated in the same change.

### 5. Testing and fail-closed
1. Before moving any code, add golden tests that record what `eligibility.Evaluate` returns today for a claim matrix: valid claim, over limit, group limit, duplicate `expense_key`, missing key, missing evidence, unknown rule, missing and false facts, both first-home tiers together (exclusive), total overflow, more than 1000 claims, duplicate claim ID, non-positive amount.
2. After the move, the unified `Evaluate` must give the same amounts and the same set of failing conditions for each case. The allowed differences are only the field and reason renames in section 2 and the merged-rule IDs and facts in section 3.
3. The duplicated rule tests in the two packages (`firsthome_test.go`, `lifestyle_test.go`) are merged into one set under `policy`. `catalog_parity_test.go` is deleted, because there is one catalog.
4. New test: no code path returns `Approved` or a non-zero `approved_amount_sen` for a Draft package.
5. Verification before merge: `go vet`, `go test -race ./...`, the PostgreSQL integration tests on a throwaway PostgreSQL 16 with zero skips, and CI passing.

## Risks and how they are handled
- The live preview endpoint changes shape: handled by the golden tests and by listing the exact differences above. Because the endpoint is a developer preview with no external client, a breaking change is accepted.
- A stricter merged rule changes which claims reach `REVIEW_REQUIRED`: accepted because the change is only toward failing closed.
- Hidden callers of `eligibility`: confirmed by search that only `cmd/server/main.go` imports it; the build proves no other caller remains.

## Out of scope and what follows
Sub-project 2 adds the medical batch (parents' medical, medical screening, serious disease and fertility, disability equipment, disabled individual, learning disability) on top of this engine, with its own spec and plan.
