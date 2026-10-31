# Form BE YA 2025 — source-traceable preflight

Official explanatory notes: https://www.hasil.gov.my/media/pshpbomm/explanatorynotes_be2025_2.pdf

This milestone provides a **partial** Form BE field crosswalk (A6, B1, B2, B14–B19, B21) and a separate preflight endpoint. It does NOT claim a complete crosswalk or tax filing readiness.

Primary source, printed PDF page 9:
- B15 dividend apportionment; B16 separately taxed at 2% (requires full special conditions and statutory calculation).
- B17/B18 progressive Part I computation; B19 total income tax.
- B21 self rebate RM400 if chargeable income <= RM35,000. Further conditions apply.
- B21 additional spouse rebate RM400 subject to a legally allowed RM4,000 spouse deduction and other conditions.

No rebates are actually applied by preflight; amounts are explicitly INDICATIVE.
Dividends always trigger review, as does lack of resident / non-business confirmation.
Income summation and statutory classification, deductions, relief verification, tax rounding, credits and PCB mapping remain incomplete.

Endpoints:
- GET /v1/formbe/fields?ya=2025
- POST /v1/formbe/preflight (JSON fields named in PreflightRequest)

Do not upload sensitive identity or taxpayer documents into these unauthenticated developer preview APIs.
