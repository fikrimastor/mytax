# Medical relief batch for YA 2025 (sub-project 2)

Date: 2026-11-07. Status: design, awaiting review. Developer preview; not tax advice. Every rule in this batch stays **Draft**; nothing becomes Approved.

## Why

`internal/policy` is now the only rule catalog and the only `Evaluate` (PR #5). `YA2025Draft()` covers just two medical limbs (`MEDICAL_VACCINATION`, `MEDICAL_DENTAL`), although `registry.go` already lists the YA 2025 caps for parents' medical, disability equipment, the disabled individual, medical screening and learning disability. This batch moves those reliefs into the engine, with facts and sources, so claims are screened instead of only listed.

## Goal and non-goals

Goal: Draft rules for Income Tax Act 1967 s.46(1)(c), (d), (e), (g)(i), (g)(ii), (h) and (ha) for YA 2025, with the two engine features they need (a fixed-amount rule and nested groups), backed by golden fixtures taken from LHDN Public Ruling No. 7/2025.

Non-goals: the web frontend; disabled spouse and disabled child (s.45A, s.47(1)(b), s.48(2)(b), a later "status reliefs" batch); joint assessment and s.46(3) spouse deeming; YA 2026 changes; any Approved package.

## Sources

| Source | Use | Retrieval |
|---|---|---|
| Act 53, May 2024 text (`docs/sources/act53/act53-excerpts-2.txt`) as amended by Finance Act 2024 s.9(a)–(d) (`finance-acts-2024-2025-excerpts.txt`) | Statutory text for YA 2025 | local |
| Act 53, 1 Jan 2026 text (`act53-2026-excerpts.txt`) | Only to show what changes in YA 2026 (Finance Act 2025 s.6). **Not** a YA 2025 source | local |
| LHDN Public Ruling No. 7/2025, 7th edition, 5 Dec 2025 | Definitions (§3.6 disabled person), interpretation (§6.2–§6.8) and worked examples | IRBM URL `https://www.hasil.gov.my/wp-content/uploads/pr-7-2025.pdf` returned HTTP 500 on 2026-11-07. Content-identical mirror `https://moore.com.my/taxflash/Pr-No.7-2025.pdf` used; snapshot and SHA-256 committed under `docs/sources/` and labelled as a mirror |
| Nota Penerangan Borang BE, Tahun Taksiran 2025 | Cross-check of caps and LHDN definitions | `https://ef.hasil.gov.my/eBE2026/Pdf/Nota_BE.pdf`, accessed 2026-11-07. Note: `eBE2025/Pdf/Nota_BE_e.pdf` serves the **YA 2024** notes and must not be cited for YA 2025 |

The Act text in the repo is the AGC updated text, not an authentic text, and the ruling is an administrative interpretation (s.138A). Both limits are stated in `docs/RELIEF_MEDICAL_YA2025.md`.

## Design

### 1. Rules and groups (all Draft, YA 2025)

| Rule | Provision | Group | Rule limit |
|---|---|---|---|
| `PARENTS_MEDICAL` (medical, dental, special needs) | s.46(1)(c)(i) | `PARENTS` (RM8,000) | RM8,000 |
| `PARENTS_CARER` | s.46(1)(c)(ii) | `PARENTS` | RM8,000 |
| `PARENTS_CHECKUP` (complete examination, including any vaccination) | s.46(1)(c) proviso (e); FA 2024 s.9(a); PR §6.2.3(c) | `PARENTS` | RM1,000 |
| `DISABILITY_EQUIPMENT` | s.46(1)(d); PR §6.3 | `DISABILITY_EQUIPMENT` (RM6,000) | RM6,000 |
| `DISABLED_SELF` (fixed amount) | s.46(1)(e); FA 2024 s.9(b); PR §6.4 | none | RM7,000, fixed |
| `MEDICAL_SERIOUS_DISEASE` | s.46(1)(g)(i); PR §6.6.2 | `MEDICAL_SELF_FAMILY` (RM10,000) | RM10,000 |
| `MEDICAL_FERTILITY` | s.46(1)(g)(ii); PR §6.6.3 | `MEDICAL_SELF_FAMILY` | RM10,000 |
| `MEDICAL_VACCINATION` (existing) | s.46(1)(g)(iii) | `MEDICAL_SELF_FAMILY` | RM1,000 |
| `MEDICAL_DENTAL` (existing) | s.46(1)(g)(iv) | `MEDICAL_SELF_FAMILY` | RM1,000 |
| `MEDICAL_SCREENING_EXAM` | s.46(1)(h)(i); PR §6.7 | `MEDICAL_SCREENING` (RM1,000, parent `MEDICAL_SELF_FAMILY`) | RM1,000 |
| `MEDICAL_SCREENING_DETECTION_TEST` | s.46(1)(h)(i) as amended by FA 2024 s.9(c) | `MEDICAL_SCREENING` | RM1,000 |
| `MEDICAL_SCREENING_SELF_TEST_DEVICE` | s.46(1)(h)(ii) as amended by FA 2024 s.9(c) | `MEDICAL_SCREENING` | RM1,000 |
| `MEDICAL_SCREENING_MENTAL_HEALTH` | s.46(1)(h)(iii) | `MEDICAL_SCREENING` | RM1,000 |
| `MEDICAL_LEARNING_DISABILITY` | s.46(1)(ha); FA 2024 s.9(d); PR §6.8 | `MEDICAL_SELF_FAMILY` | RM6,000 |

All limits are per taxpayer (PR §6.6.1, Examples 8, 13, 17, 20). Each rule carries `Sources` entries for the Act section and the PR 7/2025 paragraph.

### 2. Interpretations where the law and the ruling are silent

Each is the stricter reading or one the engine already enforces. Each is listed in `docs/RELIEF_MEDICAL_YA2025.md` and in the rule's `Source.Section` note.

1. **RM8,000 under (c) is one cap for all parents and grandparents together.** The Act and PR §6.2.1 do not say "per parent"; PR Example 3 applies one combined cap; s.46(1)(ha) proviso (e) shows the Act says "notwithstanding … more than one" when it means otherwise.
2. **No sharing between siblings under (c).** Neither the five provisos of (c) nor PR §6.2 apportion the deduction; the equal-apportionment proviso exists only in the expired s.46(1)(o). Each claimant claims what they paid (`paid_by_claimant`), up to their own cap.
3. **One expense cannot be claimed under both (c) and (d).** PR 7/2025's Appendix equipment list for (c) overlaps the (d) list and gives no priority. The engine's existing duplicate detection (same `expense_key` → `DUPLICATE_EXPENSE`) blocks the second claim.
4. **Serious disease = the diseases named in PR §6.6.2** (AIDS, Parkinson's disease, cancer, kidney disease, leukemia). "Other similar diseases" points to an IRBM guideline not held in the repo; such claims do not set `qualifying_serious_disease` and end as `REVIEW_REQUIRED`.

### 3. Engine change A: fixed-amount rules

- `Rule` gains `FixedAmount bool` (`json:"fixed_amount,omitempty"`).
- A claim on a fixed-amount rule must have `amount_sen` equal to 0. The engine uses `LimitSen` as the claimed amount. Any other value makes `Evaluate` return an error (HTTP 400), the same treatment as a non-positive amount on an expense rule.
- `expense_key` is still required (for `DISABLED_SELF`: the DSW certificate number) and `evidence_reference` is still required (the certificate). A second `DISABLED_SELF` claim with a different key gets 0 because the rule limit is spent.
- `ValidateRelease` rejects a fixed-amount rule with a zero `LimitSen` or with a group.

### 4. Engine change B: nested groups

- `Group` gains `Parent string` (`json:"parent,omitempty"`).
- A claim's allowed amount is the minimum of the rule's remaining limit, its group's remaining limit and every ancestor group's remaining limit. On allocation, the amount is subtracted from the rule, its group and all ancestors.
- Package validation (`Evaluate` setup checks and `ValidateRelease`) rejects an unknown parent, a cycle, and a child group whose limit exceeds its parent's.
- Allocation order stays claim-ID order (documented today). LHDN Example 19 gives the same RM10,000 total in any order; the fixture orders its claim IDs as the ruling presents them.

### 5. Facts

New canonical facts, added to `facts.go` in sorted order. The existing `qualifying_beneficiary` states that the beneficiary is one the rule allows (parents/grandparents for (c); self, spouse, child or parent who is disabled for (d); self, spouse or child for (g)(i), (iii), (iv) and (h); self or spouse for (g)(ii); child for (ha)). Existing `resident`, `within_year` and `no_other_claim` keep their meaning.

| Fact | Rules |
|---|---|
| `paid_by_claimant` | every expense rule in this batch (PR Examples 8, 9, 20: receipt in the claimant's name) |
| `medical_practitioner_certified` | `PARENTS_*`, `MEDICAL_SERIOUS_DISEASE`, `MEDICAL_FERTILITY` |
| `parent_resident_in_malaysia` | `PARENTS_*` |
| `treatment_in_malaysia` | `PARENTS_*`, `MEDICAL_LEARNING_DISABILITY` |
| `qualifying_medical_treatment` | `PARENTS_MEDICAL` (medical, dental or special-needs expense as listed by PR §6.2 and its Appendix) |
| `qualifying_carer` | `PARENTS_CARER` (carer is not the claimant, spouse or child; certificate, receipt or work permit) |
| `qualifying_examination` | `PARENTS_CHECKUP`, `MEDICAL_SCREENING_EXAM` (a complete examination; laboratory tests alone do not qualify, PR §6.7.4) |
| `disability_certified_by_jkm` | `DISABILITY_EQUIPMENT`, `DISABLED_SELF` (PR §3.6) |
| `qualifying_supporting_equipment` | `DISABILITY_EQUIPMENT` (necessary basic supporting equipment; spectacles excluded, PR §6.3.2). Separate from the lifestyle `qualifying_device` |
| `qualifying_serious_disease` | `MEDICAL_SERIOUS_DISEASE` (section 2, item 4) |
| `qualifying_fertility_treatment`, `individual_married` | `MEDICAL_FERTILITY` |
| `qualifying_detection_test` | `MEDICAL_SCREENING_DETECTION_TEST` |
| `qualifying_self_test_device` | `MEDICAL_SCREENING_SELF_TEST_DEVICE` (registered under the Medical Device Act 2012 [Act 737]; not for business use) |
| `qualifying_mental_health_provider` | `MEDICAL_SCREENING_MENTAL_HEALTH` (provider types in s.46(1)(h)(iii), PR §6.7.5) |
| `child_age_18_or_below`, `qualifying_learning_disability`, `allied_health_practitioner_registered` | `MEDICAL_LEARNING_DISABILITY` (six categories in PR §6.8.1; practitioners in PR §6.8.3) |

The existing test that every rule fact is canonical and every canonical fact is used keeps passing because rules and facts land together.

### 6. API changes (developer preview)

- `GET /v1/policy/catalog`: new rules; rules may carry `fixed_amount`; groups may carry `parent`.
- `POST /v1/eligibility/preview` and `POST /v1/assessment/preview`: accept the new rule IDs. A fixed-amount claim sends `amount_sen: 0` or omits it.
- No other request or response field changes.

## Testing and fail-closed

1. **Golden fixtures from PR 7/2025** in `services/api/internal/policy/testdata/pr7_2025_cases.json`, checking each line amount and the total. One `Evaluate` call per taxpayer.

   | Example | Expected |
   |---|---|
   | 3 (parents: 4,500 + 1,000 + vaccination 2,800) | 4,500 + 1,000 + 1,000 = RM6,500 |
   | 4 (wheelchair, not DSW-registered) | 0, `REVIEW_REQUIRED` |
   | 8 (leukemia: Jasmine 10,000, Isaac 4,000) | 10,000 and 4,000 |
   | 9 (IVF 20,000, receipt in Fahim's name) | Fahim 10,000; Anita 0 (`paid_by_claimant` false) |
   | 10 (separate receipts 10,000 each) | 10,000 each |
   | 11 (fertility 12,000 + son's serious disease 6,000) | RM10,000 total |
   | 13 (vaccination 500 + dental 1,400 + serious disease 5,000) | 500 + 1,000 + 5,000 = RM6,500 |
   | 14 (psychological consultation 800 + examination 800) | RM1,000 total |
   | 16 (son's serious illness 12,500 + check-up 750 + PCR test 300) | 10,000 total; screening lines limited by what remains |
   | 17 (serious disease 9,000 + mammogram 200 + devices 700 + children's devices 400) | RM10,000 total |
   | 18 (child vaccination 300 + early intervention 7,000) | 300 + 6,000 = RM6,300 |
   | 19 (own serious disease 9,800 + child vaccination 300 + early intervention 7,000) | 9,800 + 300 + 0 = RM10,000 |
   | 20 (Peter 7,000, Jill 3,000 for one child) | 6,000 and 3,000 |

   Examples 12 and 15 depend on joint assessment (s.46(3) deeming), which is out of scope; they are listed in the doc as not covered. Where an example's split between limbs is not stated, the fixture records the encoding chosen.
2. **Boundaries**: exactly at and one sen over each cap: RM8,000 group, RM1,000 parents' examination, RM6,000 equipment, RM7,000 fixed, RM10,000 group, RM1,000 screening group, RM6,000 learning disability, RM1,000 vaccination and dental.
3. **Nested groups**: a full screening group gives the next screening claim 0 while a serious-disease claim still receives the rest of the RM10,000; package checks reject an unknown parent, a cycle and a child limit above its parent.
4. **Fixed amount**: amount 0 gives RM7,000 provisional; a non-zero amount is an error; a second claim gets 0; `ValidateRelease` rejects a grouped or zero-limit fixed rule.
5. **Fail-closed**: missing `disability_certified_by_jkm`, missing `qualifying_serious_disease` (unlisted disease), a laboratory-only test without `qualifying_examination`, a child over 18, and a fertility claim for a non-spouse all give `REVIEW_REQUIRED` and 0. The existing `TestDraftCatalogNeverProducesApprovedAmounts` covers the new rules.
6. **Verification before merge**: `gofmt -l`, `go vet ./...`, `go test -race -count=1 ./...`, the PostgreSQL integration tests on a throwaway PostgreSQL 16 with zero skips, and CI green.

## Docs updated in the same change

- New `docs/RELIEF_MEDICAL_YA2025.md`: rules, facts, the four interpretations, sources with access dates, the mirror and SHA-256 note, and the examples not covered.
- `docs/ROADMAP.md`: mark the medical batch as Draft-in-engine.
- `registry.go` notes: cite PR 7/2025 where it now settles a point; correct any reference to the eBE2025 URL as a YA 2025 source.
- `docs/ASSESSMENT_PREVIEW.md`: fixed-amount claim shape.

## Risks

- **Ruling read from a mirror.** Handled by labelling the snapshot as a mirror with its hash and re-checking against the IRBM URL when it responds.
- **Interpretations in section 2 may be wrong.** Each is the stricter reading or an existing engine guard, all stay Draft, and each is listed for tax-professional review.
- **Nested groups touch allocation for every rule.** Handled by the existing golden matrix (`relief_cases.json`) still passing unchanged, plus the new nested-group tests.
