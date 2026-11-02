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
- The 21 May 2024 text predates later amendments. Those were checked against the 1 January 2026 text and the Finance Act 2024 and 2025 texts (see below).
- Read so far: s.6, s.6A, s.46, s.47, s.48, s.49, s.50, s.51 and Schedule 1 Part I paras 1 and 1A. Other sections have not been reconciled.
- Second excerpt file: `act53-excerpts-2.txt`.
- Text was extracted with `pdftotext -layout`; formatting may differ from the PDF.
- Reading is by the agent only. Not reviewed by a tax professional.

## Second source: Act 53 as at 1 January 2026
- Document: Income Tax Act 1967 (Act 53), "Online version of updated text of reprint", **as at 1 January 2026**, latest amendment Act 875. Cover says it is an AGC updated text and **not an authentic text**.
- Obtained from: a PDF named `71923` (853 pages, 3,665,458 bytes), uploaded by the repository owner on 2026-11-02 in a zip.
- Provenance: the PDF metadata shows producer "Microsoft: Print To PDF" and an individual's name as author (not recorded here). The repository owner later showed a screenshot of the LHDN "Akta" page (not stored here) listing "Akta Cukai Pendapatan 1967 (Akta 53), sebagaimana dikemaskini pada 1 Januari 2026" with a 3.50 MB download. The uploaded file is 3,665,458 bytes, which is 3.50 MB if 1 MB = 1,048,576 bytes, and its cover also says 1 January 2026. This is consistent with the file coming from that page, but it is **not proof**: no hash of the LHDN download was compared. The text is still an AGC updated text, not an authentic text.
- Hashes: zip `8030840315d9f5c04318d5810571f2a6a52d24afbedaa9a65dc2ed7939cb6720`; PDF `1c26fd5d25fd4d1c263de59c5ad66862fbb7a0aedde076f6956f8b4f7dda8202`; `act53-2026-excerpts.txt` `b06eba0baca4864b5b356b4089e0f9bb2d8d216c07c61b9ac86e463e409050ea`.
- The full PDF is not committed. Excerpts only.
- It contains amendments up to Act 875, including Act 862 ("Finance Act 2024", YA 2025 onward) and Act 874 ("Finance Act 2025", YA 2026 onward). It shows the law **as at 1 January 2026**, so a value in it is not automatically a YA 2025 value; use its in-force tables.
- Read: s.6A, s.46, s.46B, s.47 to s.51, Schedule 1 Part I para 1 and 1A, the amending-law table and the list-of-sections-amended rows for those sections.
- Full cross-check: `docs/ACT53_RECONCILIATION_YA2025.md`.

## Third source: Finance Act 2024 (Act 862) and Akta Kewangan 2025 (Akta 874)
- Uploaded by the repository owner on 2026-11-02 in a zip named `Archive.zip` (zip sha256 `a4be6230d28c10c2f907df20ce50c6bb162b4f9e6afc5c07df451dc240d27e18`).
- `Finance-Act-2024.pdf`: Finance Act 2024, Act 862, 33 pages, Royal Assent 24 December 2024, gazetted 31 December 2024. sha256 `184f137160ab332d34aa91de2820acc39eb11c280a772b94cc497074aa1103e9`.
- `Akta-Kewangan-2025-Akta-874.pdf`: Akta Kewangan 2025, Akta 874, in Bahasa Melayu (quoted amending provisions are in English), 25 pages, perkenan 27 December 2025, warta 31 December 2025. sha256 `869948f1ae964090aedd7e5afa2f63e533d4f0d4d292cfd2a9ae4bdbb9a6e68b`.
- Both carry Percetakan Nasional Malaysia Berhad copyright notices. No hash was compared with the official publication.
- The full PDFs are not committed. Excerpts: `finance-acts-2024-2025-excerpts.txt` (sha256 `65f1c909f825c6b25f52e048b7e4aeb1311efdd3500132cfb0e0fe5a2065e8e2`).
- Read: FA 2024 s.3 to s.16; FA 2025 s.3 to s.8 and s.17. These settle the year in which each change to s.46, s.46B, s.47, s.48, s.49 and Schedule 1 applies; see `docs/ACT53_RECONCILIATION_YA2025.md`.

## Differences noticed between the 21 May 2024 Act text and the LHDN YA 2025 table (first source only; mostly settled by the 2026 text, see the reconciliation doc)
These are open questions, not conclusions.
1. **Childcare, s.46(1)(r):** Act text says RM2,000, with "a further one thousand ringgit" for YA 2020 to 2024. The LHDN YA 2025 table shows RM3,000. Settled: Finance Act 2024 s.9(f) extends the further RM1,000 to YA 2027.
2. **EV charging, s.46(1)(v):** RM2,500 for YA 2023 to 2027 matches. The LHDN table also names "mesin kompos sisa makanan domestik" (food-waste composter); that is not in the text of (v).
3. **Lifestyle course, s.46(1)(p)(v):** excludes courses under s.46(1)(f)(iii). LHDN table says only "kursus peningkatan kemahiran". Model follows the Act.
4. **Footnote on the page where (p) starts:** the asterisk and the note "effect for YA 2016 until YA 2020" belong to paragraph (o) (parent RM1,500), not (p). Resolved on re-reading. See `docs/ACT53_RECONCILIATION_YA2025.md` for the full cross-check.
