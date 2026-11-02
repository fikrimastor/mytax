# Act 53 cross-check against the LHDN YA 2025 table (draft)

Status: **DRAFT. Read by the agent only. Not reviewed by a tax professional. No rule is approved.**

## Sources and limits
Record, hashes and excerpts: `docs/sources/act53/README.md`.
- **Text A:** Income Tax Act 1967 (Act 53), AGC updated text as at **21 May 2024**. Not an authentic text.
- **Text B:** the same Act, AGC updated text as at **1 January 2026** (latest amendment Act 875). Not an authentic text. It shows the law as it stands on 1 January 2026, so a value in it is not automatically a YA 2025 value. Provenance: the PDF metadata says "Print To PDF"; the owner's screenshot of the LHDN Akta page lists Act 53 as at 1 January 2026 with a 3.50 MB download that matches this file's size. Consistent with it coming from LHDN, but no hash was compared.
- **Finance Act 2024 (Act 862):** Royal Assent 24 December 2024. Section 3 sets the commencement: s.4, 8, 11, 12, 13, 14, 15, 16, 17 and paragraphs 9(a), (b), (c), (d), (g) for YA 2025 onward; s.5 and s.10 from 1 January 2025; paragraphs 9(e), (f), (h) for YA 2025 to YA 2027.
- **Akta Kewangan 2025 (Akta 874, in Malay):** perkenan 27 December 2025. Section 3 sets: s.7, 8, 17 and subparagraphs 6(a)(i) to (iv) for YA 2026 onward; 6(a)(v) and 6(c) for YA 2026; 6(a)(vi) and 6(b) for YA 2026 and 2027. **Nothing in it starts in YA 2025.**
- Both Finance Acts are Percetakan Nasional texts supplied by the repository owner. No hash was compared with the official publication.
- Registry amounts stay the LHDN YA 2025 values. Each registry entry note carries a finding kind (code in `policy/registry.go`).

## Result
Every one of the 31 registry amounts now agrees with the Act texts for YA 2025. The three entries that were "year uncertain" and the learning-disability difference are settled by the Finance Acts:

| Entry | LHDN YA 2025 | Settled by |
|---|---|---|
| DISABLED_SELF | RM7,000 | FA 2024 s.9(b): RM6,000 raised to RM7,000, YA 2025 onward |
| DISABLED_SPOUSE | RM6,000 | FA 2024 s.11: RM5,000 to RM6,000, YA 2025 onward. FA 2024 s.8 makes the same change in s.45A(1), a wife claiming a disabled husband (RM4,000 plus a further RM6,000) |
| DISABLED_CHILD | RM8,000 | FA 2024 s.12: RM6,000 to RM8,000, YA 2025 onward |
| EDU_MED_INSURANCE | RM4,000 | FA 2024 s.13: RM3,000 to RM4,000, YA 2025 onward |
| MEDICAL_LEARNING_DISABILITY | RM6,000 | FA 2024 s.9(d): RM4,000 to RM6,000, YA 2025 onward. FA 2025 s.6(a)(iii) raises it to RM10,000 from **YA 2026 only**, which explains the RM10,000 in Text B |
| CHILDCARE | RM3,000 | FA 2024 s.9(f) extends the "further RM1,000" from YA 2024 to YA 2027, so YA 2025 is RM2,000 + RM1,000. FA 2025 s.6(a)(iv) (YA 2026) makes it flat RM3,000 and adds care centres for a child up to 12 |
| SSPN | RM8,000 net | FA 2024 s.9(e): YA 2025, 2026, 2027, deposits for his child |
| EV_COMPOST | RM2,500 | FA 2024 s.9(h): EV YA 2023 to 2027; compost machine YA 2025 to 2027, once only; total RM2,500 |
| FIRST_HOME_INTEREST_LOW / MID | RM7,000 / RM5,000 | FA 2024 s.10 replaces s.46B, in force 1 Jan 2025 |
| MEDICAL_SCREENING | RM1,000 in the RM10,000 | FA 2024 s.9(c): disease detection test fee and self-testing medical device added to s.46(1)(h) |
| PARENTS | RM8,000 | FA 2024 s.9(a): parents or grandparents; complete medical examination includes any vaccination |
| SPORTS | RM1,000 | FA 2024 s.9(g): parents added |

Everything else (SELF, DISABILITY_EQUIPMENT, SELF_EDUCATION / SELF_UPSKILL, MEDICAL, VACCINATION, DENTAL, LIFESTYLE, BREASTFEEDING, LIFE_EPF, PRS, SOCSO, SPOUSE, CHILD_UNDER_18, CHILD_STUDY_BASIC / HIGHER, DISABLED_CHILD_STUDY) matches the Act text (A and B) and no Finance Act 2024 or 2025 provision changes it for YA 2025 (I read FA 2024 s.3 to s.16 and FA 2025 s.3 to s.8 and s.17; FA 2024 s.6 on s.34 and s.7 on s.44 concern business and donations and were not analysed further). Section 46(1)(p) lifestyle is not touched by either Finance Act.

### Finding kinds used in the registry notes
| Kind | Meaning | Used now |
|---|---|---|
| `ACT53_CONFIRMED_YA2025` | Act and amending Finance Act text give the value and place it in YA 2025 | 12 entries |
| `ACT53_SCOPE` | Amount agrees; scope or conditions need care or are not modelled | MEDICAL_VACCINATION |
| `ACT53_YA_UNCERTAIN`, `ACT53_DIFFERS` | Kept for future use | none |

### Things that start in YA 2026 and must **not** be applied to YA 2025
- Learning disability RM10,000 (FA 2025 s.6(a)(iii)).
- Childcare for a child up to 12 at a care centre (FA 2025 s.6(a)(iv)).
- s.46(1)(sa), RM1,000 for entrance to a tourist attraction or a cultural and arts programme (YA 2026 only).
- Food waste grinder and CCTV under s.46(1)(v)(iii), (iv) (YA 2026 and 2027).
- Vaccines: from YA 2026 only vaccines registered with the National Pharmaceutical Regulatory Agency qualify under s.46(1)(g)(iii), replacing the eight-vaccine list. For YA 2025 the eight listed vaccines apply.
- s.49(3a) and (5), (6): life insurance on the life of a child and a "child" definition for s.49 (YA 2026).
- Schedule 1 Part XXIII: 2% on profits above RM100,000 that an LLP partner receives (YA 2026).

### New YA 2025 provision that affects tax payable (not relief)
- **Schedule 1 Part XXII (FA 2024 s.16, YA 2025 onward) with s.6(1)(r):** "notwithstanding Part I", an individual's dividend from a company that is deemed derived from Malaysia (s.14), **in excess of RM100,000, is charged at 2%**. The text says the chargeable income for this is as prescribed by the Minister where the individual has other income. Not computed anywhere in this repository. The gross-tax notice now says so.

## Conditions in the Act not yet modelled
- s.46(3), s.50(2): spouse expenses and premiums are deemed to be the other spouse's where s.45(2) applies or the spender has no total income.
- s.47(4): no spouse deduction if the wife is assessed separately. s.47(6): foreign income test.
- s.48(1): maintenance, unmarried, full-time instruction, articles or indentures; s.48(4): two claimants get 50% each; s.48(5): no deduction if the child's total income exceeds the deduction; s.48(9): "child" means legitimate child, step-child or adopted child.
- s.49(2): only obligatory contributions count (except voluntary EPF); s.49(3), (4): insurance on the individual's or spouse's life; education or medical insurance for self, spouse or child. s.51: a deduction must be claimed.
- s.46(1)(1A)(b): the compost machine deduction can be claimed once in YA 2025, 2026 or 2027.
- s.46B(3): co-claimants share the first-home deduction by A x B / C. First-home eligibility rules require `citizen`, `occupied_as_residence` and `no_income_from_property`.
- s.46(1)(o) (parent RM1,500, YA 2016 to 2020 by footnote) is printed in the Act but is absent from the LHDN YA 2025 table, which is consistent.
- A footnote in the Act text mentions paragraph 10(c) of Act 742 applying to s.49(1D)(b) until YA 2030. Meaning for PRS in YA 2025 not analysed.

## Rates and rebates
- **Schedule 1 Part I, para 1:** identical in Text A, Text B and the LHDN table, and equal to `calculation/rates.go`: 0% first RM5,000, then 1%, 3%, 6% on RM15,000 each, 11% on RM20,000, 19% on RM30,000, 25% on RM300,000, 26% on RM200,000, 28% on RM1,400,000, 30% above RM2,000,000. **Finance Act 2024 s.16 and Finance Act 2025 s.17 add new Parts (XXII and XXIII) and do not change Part I.**
- **Para 1A:** a non-resident (other than a company) is taxed at 30% flat. Not implemented.
- **s.6A:** RM400 rebate (individual allowed s.46(1)(a); spouse deduction under s.47(1) or (2); wife allowed s.45A), each only if chargeable income is RM35,000 or less; zakat or fitrah rebate; umrah departure-levy rebate. The excess is not refundable or carried forward (s.6A(4)). Last amended by Act 823 (YA 2019). Not implemented.
- **Rounding:** a text search for "nearest ringgit", "nearest sen", "rounded" and "rounding" found nothing in Text A or Text B. The code rounds each band to the nearest sen. The statutory rule is **not established**.
- **YA 2023 and YA 2024:** LHDN lists the same table for YA 2023 to 2025 and Part I paragraph 1 is unchanged across all the texts read. The code still **rejects** those years. Enabling them is a decision for the repository owner, not made here.

## What is still missing
- Review by a qualified Malaysian tax professional (nothing is approved).
- Reading of the rest of the Act beyond s.6, s.6A, s.46 to s.51 and Schedule 1 Part I; any LHDN Public Ruling for YA 2025.
- Modelling of the conditions listed above.
- A hash comparison of each supplied PDF with the official publication.

## Verification status
- Committed: see git log.
- Locally tested: `go vet ./...` and `go test ./...` passed on 2026-11-02. Two PostgreSQL integration tests skip (no database), so they prove nothing.
- CI not yet run. No tax-professional review. The two Act 53 texts are not authentic texts.
