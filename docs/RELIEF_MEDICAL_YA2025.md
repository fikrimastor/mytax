# Medical reliefs, YA 2025 — draft model

Status: **DRAFT. Not approved, not professionally reviewed, never reduces chargeable income.**

## Sources

| Source | Use | Retrieval |
|---|---|---|
| Act 53, May 2024 text (`docs/sources/act53/act53-excerpts-2.txt`) as amended by Finance Act 2024 s.9(a)–(d) (`docs/sources/act53/finance-acts-2024-2025-excerpts.txt`) | Statutory text for YA 2025 | local |
| Act 53, 1 Jan 2026 text (`docs/sources/act53/act53-2026-excerpts.txt`) | Only to show what changes in YA 2026 (Finance Act 2025 s.6). **Not** a YA 2025 source | local |
| LHDN Public Ruling No. 7/2025, 7th edition, 5 December 2025 | Definitions (§3.6 disabled person), interpretation (§6.2–§6.8) and worked examples | IRBM URL `https://www.hasil.gov.my/wp-content/uploads/pr-7-2025.pdf` returned HTTP 500 on 2026-11-07 and 2026-11-16. Mirror used: MAICSA `https://maicsa.org.my/media/10741/technical_announcement_251218_1_3.pdf` (the Moore mirror now returns 404). The PDF is not committed; the source record, the SHA-256 `47e07419363bc88b0fbffc404a92fc625bf96c84a4f3878ee8f0e5da30e5d10b` and the §3.6–§3.9 and §6.2–§6.8 excerpt are in `docs/sources/pr7-2025/`, labelled as a mirror | PDF 2026-11-16; excerpt read by the agent |
| Nota Penerangan Borang BE, Tahun Taksiran 2025 | Cross-check of caps and LHDN definitions | `https://ef.hasil.gov.my/eBE2026/Pdf/Nota_BE.pdf`, accessed 2026-11-07. Warning: `https://ef.hasil.gov.my/eBE2025/Pdf/Nota_BE_e.pdf` serves the **YA 2024** notes and must not be cited for YA 2025 |

Two limits apply to all of the above:

- The Act text in the repo is the AGC updated text, **not an authentic text**.
- The ruling is an administrative interpretation of the Director General under s.138A, **not law**.

The mirror has not been compared with the IRBM file, because the IRBM URL does not respond. Re-check the hash once it does.

## What is modelled

All rules live in `internal/policy` (`YA2025Draft()`). Amounts are integer sen; RM1 = 100 sen. Every limit is per taxpayer, one `Evaluate` call per taxpayer (PR §6.6.1, Examples 8, 13, 17, 20).

| Rule ID | Provision | Group | Cap | Required facts (all must be true) |
|---|---|---|---|---|
| `PARENTS_MEDICAL` (medical, dental, special needs) | s.46(1)(c)(i), FA 2024 s.9(a); PR §6.2 | `PARENTS` | RM8,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, medical_practitioner_certified, parent_resident_in_malaysia, treatment_in_malaysia, qualifying_medical_treatment, no_other_claim |
| `PARENTS_CARER` | s.46(1)(c)(ii), FA 2024 s.9(a); PR §6.2 | `PARENTS` | RM8,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, medical_practitioner_certified, parent_resident_in_malaysia, treatment_in_malaysia, qualifying_carer, no_other_claim |
| `PARENTS_CHECKUP` (complete examination, including any vaccination) | s.46(1)(c) proviso (e), FA 2024 s.9(a); PR §6.2.3(c) | `PARENTS` | RM1,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, medical_practitioner_certified, parent_resident_in_malaysia, treatment_in_malaysia, qualifying_examination, no_other_claim |
| `DISABILITY_EQUIPMENT` | s.46(1)(d); PR §6.3 | `DISABILITY_EQUIPMENT` | RM6,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, disability_certified_by_jkm, qualifying_supporting_equipment, no_other_claim |
| `DISABLED_SELF` (**fixed amount**) | s.46(1)(e), FA 2024 s.9(b); PR §6.4 | none | RM7,000 | resident, within_year, disability_certified_by_jkm |
| `MEDICAL_SERIOUS_DISEASE` | s.46(1)(g)(i); PR §6.6.2 | `MEDICAL_SELF_FAMILY` | RM10,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, medical_practitioner_certified, qualifying_serious_disease, no_other_claim |
| `MEDICAL_FERTILITY` | s.46(1)(g)(ii); PR §6.6.3 | `MEDICAL_SELF_FAMILY` | RM10,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, medical_practitioner_certified, qualifying_fertility_treatment, individual_married, no_other_claim |
| `MEDICAL_VACCINATION` (existing) | s.46(1)(g)(iii) | `MEDICAL_SELF_FAMILY` | RM1,000 | resident, within_year, qualifying_vaccine, qualifying_beneficiary, no_other_claim |
| `MEDICAL_DENTAL` (existing) | s.46(1)(g)(iv) | `MEDICAL_SELF_FAMILY` | RM1,000 | resident, within_year, qualifying_dental, qualifying_beneficiary, no_other_claim |
| `MEDICAL_SCREENING_EXAM` | s.46(1)(h)(i), FA 2024 s.9(c); PR §6.7 | `MEDICAL_SCREENING` | RM1,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, qualifying_examination, no_other_claim |
| `MEDICAL_SCREENING_DETECTION_TEST` | s.46(1)(h)(i) as amended by FA 2024 s.9(c); PR §6.7.6 | `MEDICAL_SCREENING` | RM1,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, qualifying_detection_test, no_other_claim |
| `MEDICAL_SCREENING_SELF_TEST_DEVICE` | s.46(1)(h)(ii) as amended by FA 2024 s.9(c); PR §6.7.7 | `MEDICAL_SCREENING` | RM1,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, qualifying_self_test_device, no_other_claim |
| `MEDICAL_SCREENING_MENTAL_HEALTH` | s.46(1)(h)(iii); PR §6.7.5 | `MEDICAL_SCREENING` | RM1,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, qualifying_mental_health_provider, no_other_claim |
| `MEDICAL_LEARNING_DISABILITY` | s.46(1)(ha), FA 2024 s.9(d); PR §6.8 | `MEDICAL_SELF_FAMILY` | RM6,000 | resident, within_year, qualifying_beneficiary, paid_by_claimant, treatment_in_malaysia, child_age_18_or_below, qualifying_learning_disability, allied_health_practitioner_registered, no_other_claim |

`MEDICAL_SCREENING` is **nested** in `MEDICAL_SELF_FAMILY`: `Group.Parent = "MEDICAL_SELF_FAMILY"`. A screening claim is therefore limited by three caps together — its rule limit, the RM1,000 screening group and the RM10,000 self/spouse/child group. The RM1,000 screening cap sits inside the RM10,000, as s.46(1)(h) says.

`PARENTS` is one RM8,000 cap for all parents and grandparents together (interpretation 1). `DISABILITY_EQUIPMENT` has a group of the same ID so (d) has its own pool.

Behaviour:
- Claims are allocated in claim-ID order, independent of input order.
- A claim with a missing or false fact, missing evidence, duplicate expense or unknown rule is `REVIEW_REQUIRED`, contributes 0 and does not use up a cap.
- The same `expense_key` cannot be claimed twice.
- Draft rules never produce an approved amount; `filing_ready` is always false.
- Only YA 2025 is accepted by the engine.

## Claim shape

Expense rules work as they do today: send the amount and a receipt reference.

```json
{"id":"c1","rule_id":"MEDICAL_SERIOUS_DISEASE","amount_sen":500000,"expense_key":"hospital-bill-1","evidence_reference":"receipt","facts":{"resident":true,"paid_by_claimant":true}}
```

`DISABLED_SELF` is a fixed-amount rule. Send `amount_sen` 0 or omit it; the engine uses the RM7,000 rule limit. Any other amount returns HTTP 400.

```json
{"id":"c2","rule_id":"DISABLED_SELF","expense_key":"dsw-card-no","evidence_reference":"dsw-card","facts":{"resident":true,"within_year":true,"disability_certified_by_jkm":true}}
```

`expense_key` is still required: for `DISABLED_SELF` it is the DSW certificate number. `evidence_reference` is still required: the certificate itself. A second `DISABLED_SELF` claim with a different key gets 0, because the rule limit is spent.

`GET /v1/policy/catalog` marks such rules `"fixed_amount": true`; groups may carry `"parent"`.

## Interpretations for tax-professional review

Each is the stricter reading, or one the engine already enforces. Each is also repeated in the rule's `Source.Section` note. None is reviewed.

1. **RM8,000 under (c) is one cap for all parents and grandparents together.** The Act and PR §6.2.1 do not say "per parent". PR Example 3 applies one combined cap. s.46(1)(ha) proviso (e) shows the Act says "notwithstanding … more than one" when it means otherwise.
2. **No sharing between siblings under (c).** Neither the five provisos of (c) nor PR §6.2 apportion the deduction. The equal-apportionment proviso exists only in the expired s.46(1)(o). Each claimant claims what they paid (`paid_by_claimant`), up to their own cap.
3. **One expense cannot be claimed under both (c) and (d).** PR 7/2025's Appendix equipment list for (c) overlaps the (d) list and gives no priority. The engine's existing duplicate detection (same `expense_key` gives `DUPLICATE_EXPENSE`) blocks the second claim.
4. **Serious disease means the diseases named in PR §6.6.2**: AIDS, Parkinson's disease, cancer, kidney disease and leukemia. "Other similar diseases" points to an IRBM guideline not held in this repo. Such a claim does not set `qualifying_serious_disease` and ends as `REVIEW_REQUIRED`.

## Allocation order

Known limit: allocation is in claim-ID order. When a group cap binds, the split between lines depends on that order, while the group total does not. LHDN prescribes no order; PR 7/2025 Example 19 states only the RM10,000 total.

The fixtures order claim IDs as the ruling lists the expenses, and assert both the lines and the total. This limit is listed here for tax-professional review.

## Worked examples covered

`internal/policy/testdata/pr7_2025_cases.json` replays 17 cases (13 examples) from PR 7/2025. Each case is one taxpayer and one `Evaluate` call. Where the ruling does not state how an amount splits between limbs, the case records the encoding chosen.

| Example | Taxpayer | Expected lines | Total | Encoding chosen |
|---|---|---|---|---|
| 3 | Zubir | grandparents' medical 4,500; parents' dental 1,000; vaccination 1,000 (`LIMIT_REACHED`) | RM6,500 | Grandparents' medical and parents' dental are separate `PARENTS_MEDICAL` expenses; vaccination is `PARENTS_CHECKUP` (a complete examination includes vaccination) |
| 4 | Megat | wheelchair 0 (`UNVERIFIED_FACT:disability_certified_by_jkm`) | RM0 | The ruling gives no price; 300,000 sen is chosen |
| 8 | Jasmine | leukemia 10,000 | RM10,000 | — |
| 8 | Isaac | leukemia 4,000 | RM4,000 | — |
| 9 | Fahim | IVF 10,000 | RM10,000 | — |
| 9 | Anita | IVF 0 (`UNVERIFIED_FACT:paid_by_claimant`) | RM0 | The receipt is in Fahim's name, so `paid_by_claimant` is false for Anita |
| 10 | Fahim | IVF 10,000 | RM10,000 | — |
| 10 | Anita | IVF 10,000 | RM10,000 | — |
| 11 | Azli | son's serious disease 6,000; fertility 4,000 (`LIMIT_REACHED`) | RM10,000 | Claim IDs follow the ruling's table order; the ruling states only the RM10,000 total |
| 13 | Khadijah | vaccination 500; dental 1,000 (`LIMIT_REACHED`); serious disease 5,000 | RM6,500 | — |
| 14 | John | psychological consultation 800; examination 200 (`LIMIT_REACHED`) | RM1,000 | The ruling states only the RM1,000 total; claim IDs follow the text order |
| 16 | Diana | serious disease 10,000 (`LIMIT_REACHED`); vaccination 0; check-up 0; PCR test 0 | RM10,000 | The ruling lists 12,500 (serious disease 12,000 plus additional vaccination 500) as one line restricted to 10,000 and gives no amount for the screening lines; encoded as four claims in that order |
| 17 | Jasmin | serious disease 9,000; mammogram 200; devices 700; children's devices 100 (`LIMIT_REACHED`) | RM10,000 | The ruling states the RM10,000 total; per-line split follows claim-ID order |
| 18 | Anis | vaccination 300; early intervention 6,000 (`LIMIT_REACHED`) | RM6,300 | — |
| 19 | Anis | serious disease 9,800; influenza vaccination 200 (`LIMIT_REACHED`); early intervention 0 (`LIMIT_REACHED`) | RM10,000 | The ruling's line column (9,800, 300, 6,000) is before the RM10,000 group cap; the engine applies the cap per line in claim-ID order |
| 20 | Peter | early intervention 6,000 (`LIMIT_REACHED`) | RM6,000 | — |
| 20 | Jill | early intervention 3,000 | RM3,000 | — |

Not covered: Examples 12 and 15 (joint assessment, s.46(3) deeming); Examples 5–7 (disabled individual and spouse status, later batch).

## Not modelled

- Disabled spouse and disabled child: s.45A, s.47(1)(b), s.48(2)(b). A later "status reliefs" batch.
- Joint assessment and s.46(3) spouse deeming. This is why Examples 12 and 15 are absent.
- YA 2026 changes from Finance Act 2025 s.6: NPRA-registered vaccines replace the named list, and learning disability becomes RM10,000. YA 2025 keeps RM6,000.
- "Other similar diseases" under (g)(i): the IRBM guideline is not held in this repo, so such claims fail closed.
- Apportionment between siblings under (c) (interpretation 2), and any priority between (c) and (d) (interpretation 3).
- Whether a carer's work permit alone proves a carer relationship, and the PR §6.2 Appendix expense list in detail: the caller supplies `qualifying_carer` and `qualifying_medical_treatment` as yes/no facts. Nothing verifies them.
- Reimbursed or employer-paid expenses, and frequency limits beyond those stated.

## Tests

- `internal/policy/medical_test.go` with `testdata/pr7_2025_cases.json`: the 17 PR 7/2025 cases above, per-line amounts and totals.
- `TestMedicalCapBoundaries`: every rule at its cap and one sen over.
- `TestMedicalGroupBoundaries`: the RM8,000, RM10,000 and RM1,000 group caps at the boundary and one sen over.
- `TestFullScreeningGroupLeavesSeriousDiseaseTheRest`, `TestSameExpenseCannotBeClaimedUnderParentsAndEquipment`: nesting and duplicate detection.
- `TestDisabledSelfIsAFixedRM7000`, `TestMedicalFailsClosedOnUnverifiedFacts`: the fixed amount and fail-closed facts.
- `engine_nested_test.go` and `engine_fixed_test.go`: the two engine features against synthetic packages.
- `engine_unified_test.go`: no draft rule produces an approved amount.

## Verification status

- Draft only. No Approved package, no release record; `ActivePackage` still returns an error by design.
- Act 53 reading and PR 7/2025 reading are by the agent only. The ruling came from a mirror.
- Independent / tax-professional review: **not done**. Tax-professional review is required before any rule leaves Draft.
- Local and CI results for the batch are in the pull request, not in this file.
