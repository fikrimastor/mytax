# YA 2025 source reconciliation (draft)

Status: **DRAFT. Nothing here is approved or professionally reviewed.**

## Source used
- Snapshot: `docs/sources/ya2025/hasil-user-supplied-2026-11-02.txt`
- SHA256 of the snapshot file: `eeb2b22c53c5f7c09884e48fbcd81af34e3dff6b1355cd8a7f84fb2ababec0d5`
- Stated origin: https://www.hasil.gov.my/individu/kadar-cukai/ and https://www.hasil.gov.my/individu/pelepasan-cukai/
- Supplied by: repository owner, pasted on 2026-11-02.

### Provenance limits (read before trusting)
- The text was pasted by a person. Automation could not fetch it: hasil.gov.my answered the agent with "The URL you requested has been blocked". The hash proves the file, not the live page.
- The pasted text carries no effective date, no section/paragraph citation and no Public Ruling reference.
- Statutory basis (Income Tax Act 1967 sections, Finance Act amendments) has not been checked yet. `lom.agc.gov.my` is reachable from the agent session and can be used for that.

## 1. Rate table (resident individual)
The pasted heading says the table applies to YA 2023, 2024 and 2025.

| Check | Result |
|---|---|
| Band limits and rates in `calculation/rates.go` (`ya2025`) | Match the pasted table (0, 1, 3, 6, 11, 19, 25, 26, 28, 30 %) |
| Cumulative tax at each upper bound (150, 600, 1,500, 3,700, 9,400, 84,400, 136,400, 528,400) | Recomputed by hand from the band rows: consistent |
| Boundary tests in `calculation/rates_test.go` | Cover every upper bound; `go test` on `internal/calculation` passed locally on 2026-11-02 |
| YA 2023 and YA 2024 | **Still rejected by code on purpose.** The source header lists them, but statute and effective-date checks are not done. Decide explicitly before enabling. |

Gross marginal tax only. Not covered by this source: rebates, non-resident rates, rounding rule, PCB, special cases.

## 2. Relief amounts (YA 2025)
All 31 entries in `policy.ReliefRegistry2025()` have a `MaxSen` that matches the pasted amount (item numbers refer to the pasted table). All remain `DRAFT_REVIEW_REQUIRED`.

Items to confirm, not decided here:
1. **Item 7 shares the RM10,000 medical cap?** The pasted row shows `10,000 (Terhad)` for both items 6 and 7, and the registry models item 7 as a RM1,000 child of `MEDICAL`. Needs confirmation from the ruling or notes that these share one RM10,000 cap.
2. **Item 17 (life insurance and EPF):** two components, RM4,000 (EPF/mandatory) and RM3,000 (life/takaful/extra EPF), total RM7,000. The registry has only the RM7,000 parent. The category wording ("penjawat awam skim berpencen, selain ... dan bekerja sendiri") is truncated in the paste and must be read from the source.
3. **Item 9 (lifestyle):** four sub-items (reading material, PC/phone/tablet, internet, skills course). The draft catalog has rules for books and internet only.
4. **Item 10 (sports):** RM1,000 additional and separate from the RM2,500 lifestyle cap, as registry models it. Confirm from the source.
5. **Item 11 (breastfeeding):** child 2 years and below; allowed once every 2 YA. Needs a stateful check across years, not yet modelled.
6. **Item 12 (childcare):** child 6 years and below, registered provider.
7. **Item 16b/16c:** age, marital status, institution recognition and disabled-child additional RM8,000 conditions are not modelled as facts.
8. **Item 22 (first home interest):** sale and purchase agreement dated 1 Jan 2025 to 31 Dec 2027; RM7,000 (price up to RM500,000) or RM5,000 (above RM500,000 up to RM750,000). Date window is not enforced in code.
9. Required facts in the draft catalog are yes/no flags. They do not yet model dates, ages, beneficiaries, spouse/child allocation or cross-category exclusions.

### Act 53 cross-check (added after reading s.46)
Source: `docs/sources/act53/README.md` (updated text as at 21 May 2024, not authentic, predates later amendments). Only s.46(1)(f), (p), (r), (t), (u), (v) were read.
- Lifestyle (p): RM2,500 shared cap and four items confirmed; see `docs/RELIEF_LIFESTYLE_YA2025.md`.
- Up-skilling (f)(iii): RM2,000 for YA 2023 to 2026 inside the RM7,000 education cap: consistent with the registry.
- Sports (u): RM1,000 in its own paragraph, separate from (p): consistent with the registry.
- EV charging (v): RM2,500 for YA 2023 to 2027: consistent. Composter is **not** in this text.
- **Childcare (r): registry says RM3,000 (LHDN table); Act text gives RM2,000 plus a further RM1,000 only for YA 2020 to 2024. Unresolved; treat the registry value as unconfirmed by the Act.**
- Breastfeeding (q): RM1,000, child aged 2 and below, not allowed for the next YA: consistent.

## 3. Not covered by the pasted source
Tax rebates, zakat, spouse/child allocation rules, joint or separate assessment, non-resident treatment, income classification, PCB, statutory rounding, YA 2026 values.

## 4. Next steps
- Statutory text for s.46 is now in hand (see above). Still to read: s.47, s.48, s.49, Schedule 1 rates, and amendments after 21 May 2024.
- Ask the repository owner to save the live pages and the YA 2025 Form BE explanatory notes (PDF) so a real file hash exists.
- Model one relief end to end (lifestyle is the smallest) with facts, caps and boundary tests.
- Send this document and the fixtures to a qualified Malaysian tax professional. Status stays `Draft` until they approve.
