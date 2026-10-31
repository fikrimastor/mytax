# YA 2025 resident gross marginal-rate calculation

Official primary source: https://www.hasil.gov.my/individu/kadar-cukai/

The rate table is independently published for YA 2023, 2024 and 2025. This implementation explicitly supports only YA 2025 for now. Its cumulative threshold values are tested against the official table (RM0, RM150, RM600, RM1,500, RM3,700, RM9,400, RM84,400, RM136,400, RM528,400).

IMPORTANT: This calculates **gross marginal tax only** on a chargeable income number **provided by the caller**. The API cannot assert the user's actual chargeable income, because it has not implemented income classification, deduction or relief eligibility, rebates, credits, PCB, statutory rounding, joint assessments, nonresident treatment, special treatment, and filing reconciliation.

No endpoint is labeled refund or final tax payable. No YA 2026 rate inference is performed.

Testing gates before production: compare end-to-end statutory tax calculations with official BE instructions, including the LHDN tax table vs percentage rate application where relevant, rounding and rebates. Do not use this prototype to prepare returns.
