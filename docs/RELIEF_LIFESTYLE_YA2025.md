# Lifestyle relief, YA 2025 (item 9) — draft model

Status: **DRAFT. Not approved, not professionally reviewed, never reduces chargeable income.**

## Source
- Pasted HASiL text: `docs/sources/ya2025/hasil-user-supplied-2026-11-02.txt` (sha256 `eeb2b22c53c5f7c09884e48fbcd81af34e3dff6b1355cd8a7f84fb2ababec0d5`), item 9 "Gaya hidup".
- **Statutory section: not yet cited.** The agent could not read the Act: `lom.agc.gov.my` answers "Invalid request" to direct URLs and only serves links with server-issued tokens. A person should download the consolidated Income Tax Act 1967 (Act 53) and the relevant Finance Act amendments and save them under `docs/sources/`.

## What the source says (as pasted)
Expenses for self, spouse or child, shared cap **RM2,500**:
1. Books, journals, magazines, newspapers and similar publications (not prohibited reading material).
2. Personal computer, smartphone or tablet (not for business use).
3. Monthly internet subscription bill (in own name).
4. Skills improvement / self-development course fees.

Item 10 (additional sports relief, RM1,000) is separate and **not modelled here**.

## What is modelled
| Rule ID | Cap | Pool | Required facts (all must be true) |
|---|---|---|---|
| `LIFESTYLE_BOOKS` | RM2,500 | LIFESTYLE | existing facts, unchanged |
| `LIFESTYLE_DEVICE` (new) | RM2,500 | LIFESTYLE | resident, in the YA, qualifying beneficiary, qualifying device, not business use |
| `LIFESTYLE_INTERNET` | RM2,500 | LIFESTYLE | existing facts, unchanged |
| `LIFESTYLE_COURSE` (new) | RM2,500 | LIFESTYLE | resident, in the YA, qualifying beneficiary, recognised course |

Both `internal/policy` (`YA2025Draft`) and `internal/eligibility` (`Draft2025`) carry the new rules. Amounts are integer sen.

Behaviour:
- Shared cap of 250,000 sen across all four rules. Claims are allocated in claim-ID order, independent of input order.
- A claim with a missing or false fact, missing evidence, duplicate expense or unknown rule is `REVIEW_REQUIRED`, contributes 0 and does not use up the cap.
- Same `expense_key` cannot be claimed twice (eligibility engine). This also stops one course fee being claimed as lifestyle and as education up-skilling.
- Draft rules never produce an approved amount; `FilingReady` is always false.
- Only YA 2025 is accepted by the eligibility engine.

## Tests
`internal/eligibility/lifestyle_test.go` and `internal/policy/lifestyle_test.go`: cap boundaries (1 sen, cap-1, cap, cap+1, far above) for every rule, shared cap across the four rules, order independence, each required fact individually missing or false, evidence/expense-id/duplicate/unknown-rule fail-closed, separate pools, no approval leakage, bad input and unsupported year. Mutation checks (cap +1 sen, dropping a required fact) make these tests fail as intended.

## Known gaps (not decided, do not treat as rules)
1. Statutory citations missing (see above).
2. Meaning of "qualifying beneficiary" (self, spouse, child), spouse/child allocation and who may claim a child's expense are not modelled.
3. "Prohibited reading material", "business use", "recognised course" and device definitions are caller-supplied yes/no facts. Nothing verifies them.
4. Any frequency or per-item limit that may exist for devices or other items is **unknown**; not checked against any source.
5. Reimbursed or employer-paid expenses are not modelled.
6. Fact names are not uniform: older rules use `qualifying_publication`, `own_name`, `qualifying_internet`, `internet_in_own_name`; new ones use `qualifying_beneficiary`. The older books/internet rules do not yet require a beneficiary fact. Unifying them changes API behaviour and needs an explicit decision.
7. The `policy` engine relies on the caller's `Duplicate` flag; only the `eligibility` engine detects duplicates by `expense_key`.
8. Same-year overlap between lifestyle course fees and education up-skilling (RM2,000) is only prevented by identical `expense_key`.

## Verification status
- Committed: see git log for this branch.
- Locally tested: `go vet ./...` and `go test ./...` passed on 2026-11-02. Two PostgreSQL integration tests were **skipped** (database fixture not configured), so they prove nothing.
- CI: not yet run for this change.
- Independent / tax-professional review: not done.
