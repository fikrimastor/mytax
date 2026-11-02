# Act 53 cross-check against the LHDN YA 2025 table (draft)

Status: **DRAFT. Read by the agent only. Not reviewed by a tax professional. No rule is approved.**

## Sources and limits
Record, hashes and excerpts: `docs/sources/act53/README.md`.
- **Text A:** Income Tax Act 1967 (Act 53), AGC updated text **as at 21 May 2024** (latest amendment Act A1706). Not an authentic text.
- **Text B:** the same Act, AGC updated text **as at 1 January 2026** (latest amendment Act 875). Not an authentic text. Provenance: the PDF metadata says "Print To PDF"; the owner's screenshot of the LHDN Akta page lists Act 53 as at 1 January 2026 with a 3.50 MB download, matching this file's size. Consistent with it coming from LHDN, but no hash was compared, so it is not proven.
- **LHDN table:** `docs/sources/ya2025/hasil-user-supplied-2026-11-02.txt`, pasted by the repository owner.
- Text B shows the law **as it stands on 1 January 2026**. It includes amendments that start in YA 2026 and later, so a value in Text B is **not automatically a YA 2025 value**. The in-force dates come from the "amending law" and "list of sections amended" tables inside Text B.
- Amending laws named in Text B: Act 862 "Finance Act 2024" (YA 2025 onward, with some items for YA 2025 to 2027), Act 874 "Finance Act 2025" (YA 2026 onward, some YA 2026 and 2027), Acts 863 and 875 (tax collection and enforcement). The Finance Act texts themselves were **not** supplied, so which paragraph changed in which Act is only partly visible.
- Text A and B both call Act 851 by different names ("Finance (No. 2) Act 2023" in a footnote, "Finance Act 2024" in a table). Not resolved.
- Registry amounts stay the LHDN YA 2025 values. Each registry entry note carries a finding kind (see the code in `policy/registry.go`).

## Finding kinds
| Kind | Meaning |
|---|---|
| `ACT53_CONFIRMED_YA2025` | Text B gives the same value and its own wording or the amending-law table places it in YA 2025 |
| `ACT53_YA_UNCERTAIN` | Text B gives the same value, but the text does not show if it applied in YA 2025 or only from YA 2026 |
| `ACT53_DIFFERS` | Text B gives a different value. Unresolved |
| `ACT53_SCOPE` | Amount agrees, scope or conditions differ or are not modelled |

## Reliefs: Text A and Text B against the 31 registry entries
| Registry ID | LHDN table | Text A (May 2024) | Text B (Jan 2026) | Kind |
|---|---|---|---|---|
| SELF | RM9,000 | s.46(1)(a) RM9,000 | same | unflagged, match |
| PARENTS | RM8,000 | RM8,000, "parents" | RM8,000, **parents or grandparents**; exam includes NPRA-registered vaccination | SCOPE |
| PARENTS_CHECKUP | RM1,000 | RM1,000 | same | match |
| DISABILITY_EQUIPMENT | RM6,000 | RM6,000 | same | match |
| DISABLED_SELF | RM7,000 | further RM6,000 | further **RM7,000** | YA_UNCERTAIN |
| SELF_EDUCATION / SELF_UPSKILL | RM7,000 / RM2,000 | same, (f)(iii) YA 2023 to 2026 | same | match |
| MEDICAL / VACCINATION / DENTAL | RM10,000 / 1,000 / 1,000 | same | same | match |
| MEDICAL_SCREENING | RM1,000 in the RM10,000 | exam, COVID-19 test, mental health | exam or disease detection test fee, self-testing device, mental health | SCOPE |
| MEDICAL_LEARNING_DISABILITY | RM6,000 | RM4,000 | **RM10,000**, part of the RM10,000 in (g) | **DIFFERS** |
| LIFESTYLE | RM2,500 | s.46(1)(p) RM2,500 | identical operative text | match |
| SPORTS | RM1,000 | s.46(1)(u) | same, now includes parents | match |
| EV_COMPOST | RM2,500 | EV charging only | EV YA 2023 to 2027; compost machine YA 2025 to 2027, **once only**; grinder and CCTV YA 2026 to 2027; total RM2,500 | CONFIRMED |
| CHILDCARE | RM3,000 | RM2,000 (+1,000 for YA 2020 to 2024) | RM3,000, no year limit; also care centres for child up to 12 | YA_UNCERTAIN |
| BREASTFEEDING | RM1,000 | s.46(1)(q) | identical | match |
| SSPN | RM8,000 net | RM8,000, footnote YA 2019 to 2024 | RM8,000 for **YA 2025, 2026 and 2027** | CONFIRMED |
| LIFE_EPF | RM7,000 (4,000 + 3,000) | s.49 | same | match |
| EDU_MED_INSURANCE | RM4,000 | s.49(1B) RM3,000 | s.49(1B) **RM4,000** | YA_UNCERTAIN |
| PRS | RM3,000 | s.49(1D) | same | match |
| SOCSO | RM350 | s.46(1)(n) | identical | match |
| SPOUSE | RM4,000 | s.47(1)(a), (3) | same | match |
| DISABLED_SPOUSE | RM6,000 | further RM5,000 | further **RM6,000**; s.47 last amended by Act 862, YA 2025 onward | CONFIRMED |
| CHILD_UNDER_18 / CHILD_STUDY_BASIC | RM2,000 | s.48(2)(a) | same | match |
| CHILD_STUDY_HIGHER | RM8,000 | s.48(3)(a)(i) | same | match |
| DISABLED_CHILD | RM8,000 | RM6,000 | **RM8,000**; s.48 last amended by Act 862, YA 2025 onward | CONFIRMED |
| DISABLED_CHILD_STUDY | RM8,000 additional | s.48(3)(a)(ii) | same | match |
| FIRST_HOME_INTEREST_LOW / MID | RM7,000 / RM5,000 | not in the Act | s.46B (Act 862, in force 1 Jan 2025): citizen and resident, first residential property, SPA 1 Jan 2025 to 31 Dec 2027, 3 consecutive years, no income from the property | CONFIRMED |

Result of the cross-check against Text B:
- 6 entries confirmed for YA 2025 (SSPN, EV_COMPOST, DISABLED_SPOUSE, DISABLED_CHILD, both first-home tiers).
- 3 entries match the amount, but the text does not show whether YA 2025 or YA 2026 introduced it (DISABLED_SELF, CHILDCARE, EDU_MED_INSURANCE). They need the Finance Act 2024 text or an LHDN ruling.
- 1 real difference left: MEDICAL_LEARNING_DISABILITY (Act RM10,000 against LHDN RM6,000). The registry keeps RM6,000, which is lower and fails closed.
- 2 scope notes (MEDICAL_SCREENING, PARENTS).
- The rest match on amount. Unmatched details are in the list below.

### New or changed conditions found in Text B
- s.46(1)(sa): RM1,000 for entrance fee to a tourist attraction or a cultural and arts programme. Not in the LHDN YA 2025 table; year not shown in the text. **Not modelled.**
- s.46(1)(v)(ii): compost machine is claimable **once**, in YA 2025, 2026 or 2027 (s.46(1)(1A)(b)). Not modelled.
- s.46B: no deduction if the property earns income; shared deduction between co-claimants by A x B / C. First-home eligibility rules now require `citizen`, `occupied_as_residence` and `no_income_from_property` (they were missing).
- s.46(1)(o) (parent RM1,500, YA 2016 to 2020 by footnote) is still printed in Text B with its footnote; it is absent from the LHDN YA 2025 table.
- Conditions in both texts not yet modelled: s.46(3) and s.50(2) spouse deeming; s.47(4) wife assessed separately; s.48(4) two claimants get 50% each; s.48(5) child income above the deduction; s.48(9) meaning of child; s.49(2) obligatory contributions only; s.51 deduction must be claimed.
- s.49 footnote in Text A mentions YA 2025; Text B's amendment table says paragraph 10(c) of Act 742 applies to s.49(1D)(b) until YA 2030. Meaning for PRS in YA 2025 not analysed.

## Rates and rebates
- **Schedule 1 Part I, para 1:** identical in Text A and Text B and equal to the LHDN table and to `calculation/rates.go`: 0% first RM5,000, then 1%, 3%, 6% on RM15,000 each, 11% on RM20,000, 19% on RM30,000, 25% on RM300,000, 26% on RM200,000, 28% on RM1,400,000, 30% above RM2,000,000. Text B's amendment table shows Schedule 1 was amended by Acts 862 and 874, but the amended parts are not identifiable here; the Part I paragraph 1 table itself did not change between the two texts.
- **Para 1A:** non-resident (other than a company) 30% flat. Not implemented.
- **s.6A:** RM400 rebate (individual allowed s.46(1)(a); spouse deduction under s.47(1) or (2); wife allowed s.45A), each only if chargeable income is RM35,000 or less; zakat or fitrah rebate; umrah departure-levy rebate. Excess rebate is not refundable or carried forward (s.6A(4)). Text B's amendment table shows s.6A last amended by Act 823 (YA 2019), so this is unchanged for YA 2025. Not implemented.
- **Rounding:** a text search for "nearest ringgit", "nearest sen", "rounded" and "rounding" found nothing in Text A or in Text B. The code rounds each band to the nearest sen; the statutory rule is **not established**.
- **YA 2023 and YA 2024:** LHDN lists the same table for YA 2023 to 2025 and Part I paragraph 1 is unchanged across both Act texts. The code still **rejects** those years. Enabling them is a decision for the repository owner, not made here.

## What is still missing
- Text of Act 862 (Finance Act 2024) and Act 874 (Finance Act 2025) to settle the three YA_UNCERTAIN entries and the learning-disability difference, and any LHDN Public Ruling for YA 2025.
- Reading of the rest of the Act (anything outside s.6, s.6A, s.46 to s.51, Schedule 1 Part I).
- Review by a qualified Malaysian tax professional.

## Verification status
- Committed: see git log.
- Locally tested: `go vet ./...` and `go test ./...` passed on 2026-11-02. Two PostgreSQL integration tests skip (no database), so they prove nothing.
- CI not yet run. No tax-professional review. Neither Act text is authentic.
