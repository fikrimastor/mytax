# Milestone 9 — expense-level relief candidate assessment

Primary official basis: https://www.hasil.gov.my/individu/pelepasan-cukai/ (YA 2025) and https://www.hasil.gov.my/media/d1rhwb05/pelepasan-cukai-tt-2025.pdf

POST /v1/eligibility/preview accepts assessment_year=2025 and claims with immutable ID, expense_key, rule_id, amount_sen, evidence_reference and verified facts. It deterministically orders by ID and checks shared lifestyle/education/medical caps, selected nested caps, duplicate expense identifiers and first-home tier exclusivity.

This is **not a complete or officially approved legal eligibility engine**. Facts are supplied by the user, not authenticated. Cross-spouse duplicate detection, receipt identity matching, overlapping medical sublimits, exceptions and all remaining relief rules require more work. Preview returns approved_total_sen=0, filing_ready=false and DRAFT_NOT_FOR_FILING irrespective of candidate values.

Never subtract provisional amounts from real statutory computations or describe them as allowable reliefs. Production activation requires full legal source review and professionally signed rule packages.
