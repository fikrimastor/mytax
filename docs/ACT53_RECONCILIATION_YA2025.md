# Act 53 cross-check against the LHDN YA 2025 table (draft)

Status: **DRAFT. Read by the agent only. Not reviewed by a tax professional. No rule is approved.**

## Sources and limits
- Act text: Income Tax Act 1967 (Act 53), AGC updated text **as at 21 May 2024**, **not an authentic text**. Record, hashes and excerpts: `docs/sources/act53/`.
- LHDN YA 2025 table: `docs/sources/ya2025/hasil-user-supplied-2026-11-02.txt` (pasted by the repository owner).
- The Act text ends at Act A1706 (in force 21 May 2024). Amendments after that date are **not in it**. The document is also inconsistent about Act 851: a footnote calls it "Finance (No. 2) Act 2023", the amendments table calls it "Finance Act 2024". Not resolved.
- Where the two sources disagree, **neither is treated as right**. The registry keeps the LHDN value and flags the difference in the entry note (`ACT53_DIFFERS`).

## (a) Reliefs, s.46 to s.49 versus the 31 registry entries
| Registry ID | Registry (LHDN table) | Act 53 text | Result |
|---|---|---|---|
| SELF | RM9,000 | s.46(1)(a) RM9,000 | Match |
| PARENTS / PARENTS_CHECKUP | RM8,000 / RM1,000 | s.46(1)(c) RM8,000; exam RM1,000 (proviso (e)) | Match. Act also needs parents resident in Malaysia, treatment in Malaysia, practitioner certification |
| DISABILITY_EQUIPMENT | RM6,000 | s.46(1)(d) RM6,000 | Match |
| DISABLED_SELF | RM7,000 | s.46(1)(e) "a further RM6,000" | **Differs** |
| SELF_EDUCATION / SELF_UPSKILL | RM7,000 / RM2,000 | s.46(1)(f) RM7,000; (f)(iii) RM2,000 for YA 2023 to 2026 | Match |
| MEDICAL / VACCINATION / DENTAL | RM10,000 / 1,000 / 1,000 | s.46(1)(g) total RM10,000 (was RM8,000 before Finance Act 2023); (iii), (iv) RM1,000 each | Match |
| MEDICAL_SCREENING | RM1,000 inside RM10,000 | s.46(1)(h) RM1,000, "part of" the RM10,000 | Match on cap and sharing. **Scope differs:** Act lists complete medical exam, COVID-19 test, mental health only; LHDN also lists self-test equipment and disease-detection test fees |
| MEDICAL_LEARNING_DISABILITY | RM6,000 | s.46(1)(ha) RM4,000, part of the RM10,000 | **Differs** |
| LIFESTYLE | RM2,500 | s.46(1)(p) RM2,500 | Match (see `RELIEF_LIFESTYLE_YA2025.md`) |
| SPORTS | RM1,000 | s.46(1)(u) RM1,000, separate paragraph | Match |
| EV_COMPOST | RM2,500 | s.46(1)(v) EV charging, RM2,500, YA 2023 to 2027 | Match for EV. **Composter not found** |
| CHILDCARE | RM3,000 | s.46(1)(r) RM2,000 plus a further RM1,000 for YA 2020 to 2024 | **Differs** |
| BREASTFEEDING | RM1,000 | s.46(1)(q) RM1,000, child 2 and below, not in the next YA | Match |
| SSPN | RM8,000 net | s.46(1)(k) RM8,000 deposited for his child, less withdrawals; footnote says the amendment applies to YA 2019 to 2024 | **YA 2025 basis not in this text** |
| LIFE_EPF | RM7,000 (4,000 + 3,000) | s.49(1)(a) RM3,000, (1)(b) RM4,000, (1)(c) RM4,000 widow/orphan pension, total RM7,000 (1A)(a) | Match on RM7,000 total and components. Widow/orphan pension component not in the registry |
| EDU_MED_INSURANCE | RM4,000 | s.49(1B) RM3,000 | **Differs** |
| PRS | RM3,000 | s.49(1D) RM3,000 | Match |
| SOCSO | RM350 | s.46(1)(n) RM350 | Match |
| SPOUSE | RM4,000 | s.47(1)(a), (3) RM4,000 including alimony | Match. Not allowed if wife is assessed separately, s.47(4) |
| DISABLED_SPOUSE | RM6,000 | s.47(1)(b) "a further RM5,000" | **Differs** |
| CHILD_UNDER_18 / CHILD_STUDY_BASIC | RM2,000 | s.48(2)(a) RM2,000 | Match |
| CHILD_STUDY_HIGHER | RM8,000 | s.48(3)(a)(i) four times RM2,000 | Match |
| DISABLED_CHILD | RM8,000 | s.48(2)(b) RM6,000 | **Differs** |
| DISABLED_CHILD_STUDY | RM8,000 additional | s.48(3)(a)(ii) further RM8,000 | Match |
| FIRST_HOME_INTEREST_LOW / MID | RM7,000 / RM5,000 | **No matching provision.** s.46B covers only SPA dated 10 Mar 2009 to 31 Dec 2010 (RM10,000, 3 years) | **Not found** |

Count of the 31 entries: 22 match on amount (EV_COMPOST and MEDICAL_SCREENING have scope gaps), 6 amounts differ (DISABLED_SELF, MEDICAL_LEARNING_DISABILITY, CHILDCARE, EDU_MED_INSURANCE, DISABLED_SPOUSE, DISABLED_CHILD), 1 has no YA 2025 basis in the text (SSPN), 2 not found (both first-home interest tiers).

Registry entries beyond the LHDN table: parent RM1,500 s.46(1)(o) is marked in the Act as effective YA 2016 to 2020 (footnote) and is correctly absent from the YA 2025 table.

### Conditions in the Act that the draft rules do not model yet
- s.46(3), s.50(2): spouse expenses and premiums are deemed to be the other spouse's where s.45(2) applies or the spender has no total income.
- s.47(4): no spouse deduction if the wife is assessed separately. s.47(6): foreign income test.
- s.48(1): "maintenance", unmarried, full-time instruction, articles or indentures; s.48(4): two claimants each get 50%; s.48(5): no deduction if the child's total income exceeds the deduction; s.48(9): "child" means legitimate child, step-child or adopted child.
- s.49(2): only obligatory contributions count (except voluntary EPF); s.49(3), (4): insurance must be on the individual's or spouse's life; education/medical insurance for self, spouse or child.
- s.51: a deduction must be claimed.
- A footnote on s.49 says the Finance Act 2012 amendment applies "until the year of assessment 2021 and in respect of paragraph 10(c) until the year of assessment 2025". What this means for YA 2025 is **unresolved**.

## (b) Rates and rebates
- **Schedule 1 Part I, para 1** (rates "for every person"): 0% on first RM5,000, then 1%, 3%, 6% on RM15,000 each, 11% on RM20,000, 19% on RM30,000, 25% on RM300,000, 26% on RM200,000, 28% on RM1,400,000, 30% above RM2,000,000. **Matches** `calculation/rates.go` and its boundary tests. The Act text carries no year restriction on this table and no footnote about YA 2025.
- **Para 1A:** a non-resident (other than a company) is taxed at a flat 30%. Not implemented; the API only does resident gross tax.
- **s.6A:** RM400 rebate for an individual allowed s.46(1)(a), for a spouse deduction (s.47(1) or (2)), and for a wife allowed s.45A, each only if chargeable income is RM35,000 or less. Also a zakat/fitrah rebate for obligatory payments to a religious authority established by written law (s.6A(3)), and an umrah departure-levy rebate (s.6A(2A)). s.6A(4): any excess rebate is **not refundable and not carried forward**. Rebates apply before PCB set-off (s.110) and credits (s.132, s.133). None of this is implemented.
- **Rounding:** a text search for "nearest ringgit", "nearest sen" and "rounded" found nothing relevant. The code rounds each band to the nearest sen; the statutory rounding rule is **not established** by this text.
- **YA 2023 and YA 2024:** the LHDN page lists the same table for YA 2023, 2024 and 2025, and the Act text as at May 2024 shows the same Schedule. The code still **rejects** those years. Enabling them is a decision for the repository owner, not made here.

## (c) Amendments after 21 May 2024
Not done. The text needed is not available to the agent:
- `lom.agc.gov.my` and `hasil.gov.my` block direct access from this environment.
- The Act text ends at Act A1706.

Values that most likely depend on a later amendment (not verified; this is the reason for the differences above): DISABLED_SELF, MEDICAL_LEARNING_DISABILITY, CHILDCARE, EDU_MED_INSURANCE, DISABLED_SPOUSE, DISABLED_CHILD, SSPN for YA 2025, the EV/composter wording, and first-home interest relief.

**To finish (c) the repository owner needs to supply:** the Finance Act enacted after May 2024 that amends the Income Tax Act 1967 (the one carrying the Budget 2025 income tax measures), and any later amending Acts or LHDN Public Rulings for YA 2025. Name or number is not stated here because the agent could not confirm it.

## Verification status
- Committed: see git log.
- Locally tested: `go vet ./...` and `go test ./...` passed on 2026-11-02. Two PostgreSQL integration tests skip (no database), so they prove nothing.
- CI not yet run. No tax-professional review. The Act text is not authentic.
