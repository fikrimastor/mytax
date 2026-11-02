# Act 53 source record

## What this is
- Document: Income Tax Act 1967 (Act 53), "Online version of updated text of reprint", **as at 21 May 2024**. The file's own cover says it is an updated text by the Attorney General's Chambers and **is not an authentic text**. Latest amendment shown: Act A1706.
- Obtained from: PDF named `20240521-akta-cukai-pendapatan-1967-akta-53.pdf`, uploaded by the repository owner on 2026-11-02 in a zip. Stated origin: https://www.hasil.gov.my/wp-content/uploads/20240521-akta-cukai-pendapatan-1967-akta-53.pdf (the agent could not download it: the site returned "The URL you requested has been blocked").
- PDF size 4,581,879 bytes, 818 pages.

## Hashes
| Item | SHA256 |
|---|---|
| Uploaded zip | `53416295556e130b18d5ac5bceb2169d9da88f50ae485fa7591eb39530daf61c` |
| PDF inside the zip | `771ced020407dd6686fc493d6f9c1940bc29296fed0072422fb881c42737c49e` |
| `act53-s46-excerpt.txt` (this folder) | `36bb6dc60878bf9e4906aa4f3045e8f31a89a122864ecb48842e38d4d3ec7331` |
| `act53-excerpts-2.txt` (this folder) | `9734a346b3e566595f208ab79ff0284dc5c215522deab47878d60fa331432e43` |

The hashes prove those files only. They do not prove the PDF matches the live HASiL or AGC page.

## Not committed
The full PDF is not in this repository. Only the short excerpt used for citation is. Keep the original PDF where the hash can be re-checked.

## Limits
- Text predates amendments after 21 May 2024. Which later amendments affect YA 2025 has **not** been checked.
- Read so far: s.6, s.6A, s.46, s.47, s.48, s.49, s.50, s.51 and Schedule 1 Part I paras 1 and 1A. Other sections have not been reconciled.
- Second excerpt file: `act53-excerpts-2.txt`.
- Text was extracted with `pdftotext -layout`; formatting may differ from the PDF.
- Reading is by the agent only. Not reviewed by a tax professional.

## Differences noticed between this Act text and the LHDN YA 2025 table
These are open questions, not conclusions.
1. **Childcare, s.46(1)(r):** Act text says RM2,000, with "a further one thousand ringgit" for YA 2020 to 2024. The LHDN YA 2025 table shows RM3,000. This text gives no basis for RM3,000 in YA 2025.
2. **EV charging, s.46(1)(v):** RM2,500 for YA 2023 to 2027 matches. The LHDN table also names "mesin kompos sisa makanan domestik" (food-waste composter); that is not in the text of (v).
3. **Lifestyle course, s.46(1)(p)(v):** excludes courses under s.46(1)(f)(iii). LHDN table says only "kursus peningkatan kemahiran". Model follows the Act.
4. **Footnote on the page where (p) starts:** the asterisk and the note "effect for YA 2016 until YA 2020" belong to paragraph (o) (parent RM1,500), not (p). Resolved on re-reading. See `docs/ACT53_RECONCILIATION_YA2025.md` for the full cross-check.
